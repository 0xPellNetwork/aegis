package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p/core"
	maddr "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gitlab.com/thorchain/tss/go-tss/p2p"

	"github.com/0xPellNetwork/aegis/pkg/authz"
	"github.com/0xPellNetwork/aegis/pkg/constant"
	"github.com/0xPellNetwork/aegis/relayer/config"
	appcontext "github.com/0xPellNetwork/aegis/relayer/context"
	clientlogs "github.com/0xPellNetwork/aegis/relayer/logs"
	"github.com/0xPellNetwork/aegis/relayer/metrics"
	orchestrator "github.com/0xPellNetwork/aegis/relayer/orchestrator"
	"github.com/0xPellNetwork/aegis/relayer/pellcore"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

type Multiaddr = core.Multiaddr

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start PellClient Relayer",
	RunE:  start,
}

func init() {
	RootCmd.AddCommand(StartCmd)
}

func start(_ *cobra.Command, _ []string) error {
	if err := setHomeDir(); err != nil {
		return err
	}

	SetupConfigForTest()

	//Prompt for Hotkey and TSS key-share passwords
	hotkeyPass, tssKeyPass, err := promptPasswords()
	if err != nil {
		return err
	}

	//Load Config file given path
	cfg, err := config.Load(rootArgs.pellCoreHome)
	if err != nil {
		return err
	}

	logger, err := clientlogs.InitLogger(cfg)
	if err != nil {
		return errors.Wrap(err, "initLogger failed")
	}

	//Wait until pellcore has started
	if len(cfg.Peer) != 0 {
		if err := validatePeer(cfg.Peer); err != nil {
			return errors.Wrap(err, "unable to validate peer")
		}
	}

	masterLogger := logger.Std
	startLogger := masterLogger.With().Str("module", "startup").Logger()

	// Initialize core parameters from pellcore
	appContext := appcontext.NewAppContext(cfg, masterLogger)
	ctx := appcontext.WithAppContext(context.Background(), appContext)

	// Wait until pellcore is up
	waitForPellCore(cfg, startLogger)
	startLogger.Info().Msgf("PellCore is ready , Trying to connect to %s", cfg.Peer)

	telemetryServer := metrics.NewTelemetryServer()
	go func() {
		if err := telemetryServer.Start(); err != nil {
			startLogger.Error().Err(err).Msg("telemetryServer error")
			panic("telemetryServer error")
		}
	}()

	// CreatePellcoreClient:  pellcore client is used for all communication to pellcore , which this client connects to.
	// Pellcore accumulates votes , and provides a centralized source of truth for all clients
	pellcoreClient, err := CreatePellcoreClient(cfg, hotkeyPass, masterLogger)
	if err != nil {
		startLogger.Error().Err(err).Msg("CreatePellcoreClient error")
		return err
	}

	// Wait until pellcore is ready to create blocks
	if err = pellcoreClient.WaitForPellCoreToCreateBlocks(ctx); err != nil {
		startLogger.Error().Err(err).Msg("WaitForPellcoreToCreateBlocks error")
		return err
	}
	startLogger.Info().Msgf("Pellcore client is ready")

	// Set grantee account number and sequence number
	if err = pellcoreClient.SetAccountNumber(authz.PellClientGranteeKey); err != nil {
		startLogger.Error().Err(err).Msg("SetAccountNumber error")
		return err
	}

	// cross-check chainid
	res, err := pellcoreClient.GetNodeInfo(ctx)
	if err != nil {
		startLogger.Error().Err(err).Msg("GetNodeInfo error")
		return err
	}

	if strings.Compare(res.GetDefaultNodeInfo().Network, cfg.ChainID) != 0 {
		startLogger.Warn().Msgf("chain id mismatch, pellcore chain id %s, pellcore client chain id %s; reset pellcore client chain id",
			res.GetDefaultNodeInfo().Network, cfg.ChainID)

		cfg.ChainID = res.GetDefaultNodeInfo().Network
		if err := pellcoreClient.UpdateChainID(cfg.ChainID); err != nil {
			return err
		}
	}

	// CreateAuthzSigner : which is used to sign all authz messages . All votes broadcast to pellcore are wrapped in authz exec .
	// This is to ensure that the user does not need to keep their operator key online , and can use a cold key to sign votes
	signerAddress, err := pellcoreClient.GetKeys().GetAddress()
	if err != nil {
		startLogger.Error().Err(err).Msg("error getting signer address")
		return err
	}
	CreateAuthzSigner(pellcoreClient.GetKeys().GetOperatorAddress().String(), signerAddress)
	startLogger.Debug().Msgf("CreateAuthzSigner is ready")

	// Initialize core parameters from pellcore
	err = pellcoreClient.UpdateAppContext(ctx, appContext, true, startLogger)
	if err != nil {
		startLogger.Error().Err(err).Msg("Error getting core parameters")
		return err
	}
	startLogger.Info().Msgf("Config is updated from PellCore %s", maskCfg(cfg))

	go pellcoreClient.UpdateAppContextWorker(ctx)

	// Generate TSS address . The Tss address is generated through Keygen ceremony. The TSS key is used to sign all outbound transactions .
	// The hotkeyPk is private key for the Hotkey. The Hotkey is used to sign all inbound transactions
	// Each node processes a portion of the key stored in ~/.pellcored/.tss by default . Custom location can be specified in config file during init.
	// After generating the key , the address is set on the pellcore
	hotkeyPk, err := pellcoreClient.GetKeys().GetPrivateKey(hotkeyPass)
	if err != nil {
		startLogger.Error().Err(err).Msg("pellcore client GetPrivateKey error")
	}
	startLogger.Debug().Msgf("hotkeyPk %s", hotkeyPk.String())
	if len(hotkeyPk.Bytes()) != 32 {
		errMsg := fmt.Sprintf("key bytes len %d != 32", len(hotkeyPk.Bytes()))
		log.Error().Msgf(errMsg)
		return errors.New(errMsg)
	}
	priKey := secp256k1.PrivKey(hotkeyPk.Bytes()[:32])

	// Generate pre Params if not present already
	peers, err := initPeers(cfg.Peer)
	if err != nil {
		log.Error().Err(err).Msg("peer address error")
	}
	initPreParams(cfg.PreParamsPath)
	if cfg.P2PDiagnostic {
		err := RunDiagnostics(startLogger, peers, hotkeyPk, cfg)
		if err != nil {
			startLogger.Error().Err(err).Msg("RunDiagnostics error")
			return err
		}
	}

	m, err := metrics.NewMetrics()
	if err != nil {
		log.Error().Err(err).Msg("NewMetrics")
		return err
	}
	m.Start()

	metrics.Info.WithLabelValues(constant.Version).Set(1)
	metrics.LastStartTime.SetToCurrentTime()

	tssHistoricalList, err := pellcoreClient.GetTssHistory(ctx)
	if err != nil {
		startLogger.Error().Err(err).Msg("GetTssHistory error")
	}

	telemetryServer.SetIPAddress(cfg.PublicIP)

	// Creating a channel to listen for os signals (or other signals)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	// Create TSS server
	// Generate a new TSS if keygen is set and add it into the tss server
	// If TSS has already been generated, and keygen was successful ; we use the existing TSS
	tssServer, err := GenerateTss(
		ctx,
		masterLogger,
		pellcoreClient,
		peers,
		priKey,
		telemetryServer,
		tssHistoricalList,
		tssKeyPass,
		hotkeyPass,
	)
	if err != nil {
		return fmt.Errorf("SetupTSSServer error: %w", err)
	}

	if cfg.TestTssKeysign {
		err = TestTSS(tssServer, masterLogger)
		if err != nil {
			startLogger.Error().Err(err).Msgf("TestTSS error : %s", tssServer.CurrentPubkey)
		}
	}

	// Wait for TSS keygen to be successful before proceeding, This is a blocking thread only for a new keygen.
	// For existing keygen, this should directly proceed to the next step
	ticker := time.NewTicker(time.Second * 1)
	for range ticker.C {
		keyGen := appContext.GetKeygen()
		if keyGen.Status != relayertypes.KeygenStatus_SUCCESS {
			startLogger.Info().Msgf("Waiting for TSS Keygen to be a success, current status %s", keyGen.Status)
			continue
		}
		break
	}

	// Update Current TSS value from pellcore, if TSS keygen is successful, the TSS address is set on pell-core
	// Returns err if the RPC call fails as pell client needs the current TSS address to be set
	// This is only needed in case of a new Keygen , as the TSS address is set on pellcore only after the keygen is successful i.e enough votes have been broadcast
	currentTss, err := pellcoreClient.GetCurrentTss(ctx)
	if err != nil {
		startLogger.Error().Err(err).Msg("GetCurrentTSS error")
		return err
	}

	// Defensive check: Make sure the tssServer address is set to the current TSS address and not the newly generated one
	tssServer.CurrentPubkey = currentTss.TssPubkey
	if tssServer.EVMAddress() == (ethcommon.Address{}) {
		startLogger.Error().Msg("TSS address is not set in pellcore")
	} else {
		startLogger.Info().Msgf("Current TSS address \n ETH : %s \n PubKey : %s ",
			tssServer.EVMAddress(),
			tssServer.CurrentPubkey)
	}
	if len(appContext.GetEnabledChains()) == 0 {
		startLogger.Error().Interface("config", cfg).Msgf("No chains in updated config")
	}

	isObserver, err := isObserverNode(ctx, pellcoreClient)
	switch {
	case err != nil:
		startLogger.Error().Msgf("Unable to determine if node is an observer")
		return err
	case !isObserver:
		addr := pellcoreClient.GetKeys().GetOperatorAddress().String()
		startLogger.Info().Str("operator_address", addr).Msg("This node is not an observer. Exit 0")
		return nil
	}

	userDir, err := os.UserHomeDir()
	if err != nil {
		log.Error().Err(err).Msg("os.UserHomeDir")
		return err
	}
	dbpath := filepath.Join(userDir, ".pellclient/chainobserver")

	// Orchestrator wraps the pellcore client and adds the observers and signer maps to it.
	// This is the high level object used for CCTX interactions
	maestro, err := orchestrator.New(
		ctx,
		pellcoreClient,
		tssServer,
		dbpath,
		logger,
		telemetryServer,
	)
	if err != nil {
		startLogger.Error().Err(err).Msg("Unable to create orchestrator")
		return err
	}

	// Start orchestrator with all observers and signers
	if err := maestro.Start(ctx); err != nil {
		startLogger.Error().Err(err).Msg("Unable to start orchestrator")
		return err
	}

	startLogger.Info().Msgf("Pellclientd is running...")

	// ========================================================
	sig := <-signalChannel
	startLogger.Info().Msgf("stop signal received: %s", sig)

	pellcoreClient.Stop()

	return nil
}

func initPeers(peer string) (p2p.AddrList, error) {
	var peers p2p.AddrList

	if peer != "" {
		address, err := maddr.NewMultiaddr(peer)
		if err != nil {
			log.Error().Err(err).Msg("NewMultiaddr error")
			return p2p.AddrList{}, err
		}
		peers = append(peers, address)
	}
	return peers, nil
}

func initPreParams(path string) {
	if path != "" {
		path = filepath.Clean(path)
		log.Info().Msgf("pre-params file path %s", path)
		preParamsFile, err := os.Open(path)
		if err != nil {
			log.Error().Err(err).Msg("open pre-params file failed; skip")
		} else {
			bz, err := io.ReadAll(preParamsFile)
			if err != nil {
				log.Error().Err(err).Msg("read pre-params file failed; skip")
			} else {
				err = json.Unmarshal(bz, &preParams)
				if err != nil {
					log.Error().Err(err).Msg("unmarshal pre-params file failed; skip and generate new one")
					preParams = nil // skip reading pre-params; generate new one instead
				}
			}
		}
	}
}

// promptPasswords() This function will prompt for passwords which will be used to decrypt two key files:
// 1. HotKey
// 2. TSS key-share
func promptPasswords() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("HotKey Password: ")
	hotKeyPass, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	fmt.Print("TSS Password: ")
	TSSKeyPass, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	//trim delimiters
	hotKeyPass = strings.TrimSuffix(hotKeyPass, "\n")
	TSSKeyPass = strings.TrimSuffix(TSSKeyPass, "\n")

	return hotKeyPass, TSSKeyPass, err
}

// isObserverNode checks whether THIS node is an observer node.
func isObserverNode(ctx context.Context, client *pellcore.PellCoreBridge) (bool, error) {
	observers, err := client.GetObserverList(ctx)
	if err != nil {
		return false, errors.Wrap(err, "unable to get observers list")
	}

	operatorAddress := client.GetKeys().GetOperatorAddress().String()

	for _, observer := range observers {
		if observer == operatorAddress {
			return true, nil
		}
	}

	return false, nil
}
