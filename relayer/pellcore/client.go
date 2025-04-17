package pellcore

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"cosmossdk.io/simapp/params"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	enccodec "github.com/evmos/ethermint/encoding/codec"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/pell-chain/pellcore/app"
	"github.com/pell-chain/pellcore/pkg/authz"
	"github.com/pell-chain/pellcore/pkg/chains"
	pellcore_rpc "github.com/pell-chain/pellcore/pkg/rpc"
	"github.com/pell-chain/pellcore/relayer/chains/interfaces"
	"github.com/pell-chain/pellcore/relayer/config"
	pctx "github.com/pell-chain/pellcore/relayer/context"
	keyinterfaces "github.com/pell-chain/pellcore/relayer/keys/interfaces"
	lightclienttypes "github.com/pell-chain/pellcore/x/lightclient/types"
	relayertypes "github.com/pell-chain/pellcore/x/relayer/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

var _ interfaces.PellCoreBridger = &PellCoreBridge{}

// PellCoreBridge will be used to send tx to PellCore.
type PellCoreBridge struct {
	pellcore_rpc.Clients

	logger zerolog.Logger
	config config.ClientConfiguration

	cosmosClientContext cosmosclient.Context

	blockHeight   int64
	accountNumber map[authz.KeyType]uint64
	seqNumber     map[authz.KeyType]uint64

	encodingCfg          params.EncodingConfig
	keys                 keyinterfaces.ObserverKeys
	chainID              string
	chain                chains.Chain
	pellTxMsgLength      uint8 // max number of msg in a single pell transaction
	stop                 chan struct{}
	onBeforeStopCallback []func()

	mu sync.RWMutex

	// enableMockSDKClient is a flag that determines whether the mock cosmos sdk client should be used, primarily for
	// unit testing
	enableMockSDKClient bool
	mockSDKClient       rpcclient.Client
}

// grpc.WithInsecure()
var unsecureGRPC = grpc.WithTransportCredentials(insecure.NewCredentials())

type constructOpts struct {
	customAccountRetriever bool
	accountRetriever       cosmosclient.AccountRetriever
}

type Opt func(cfg *constructOpts)

// WithCustomAccountRetriever sets custom tendermint client
func WithCustomAccountRetriever(ac cosmosclient.AccountRetriever) Opt {
	return func(c *constructOpts) {
		c.customAccountRetriever = true
		c.accountRetriever = ac
	}
}

// NewClient create a new instance of Client
func NewClient(
	logger zerolog.Logger,
	keys keyinterfaces.ObserverKeys,
	chainIP string,
	signerName string,
	chainID string,
	hsmMode bool,
	pellTxMsgLength uint8,
	opts ...Opt,
) (*PellCoreBridge, error) {
	var constructOptions constructOpts
	for _, opt := range opts {
		opt(&constructOptions)
	}

	chain, err := chains.PellChainFromChainID(chainID)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid chain id %q", chainID)
	}

	log := logger.With().Str("module", "pellcoreClient").Logger()

	cfg := config.ClientConfiguration{
		ChainHost:    cosmosREST(chainIP),
		SignerName:   signerName,
		SignerPasswd: "password",
		ChainRPC:     tendermintRPC(chainIP),
		HsmMode:      hsmMode,
	}

	encodingCfg := app.MakeEncodingConfig()
	enccodec.RegisterLegacyAminoCodec(encodingCfg.Amino)
	enccodec.RegisterInterfaces(encodingCfg.InterfaceRegistry)

	xmsgtypes.RegisterInterfaces(encodingCfg.InterfaceRegistry)

	if encodingCfg.InterfaceRegistry != nil {
		authtypes.RegisterInterfaces(encodingCfg.InterfaceRegistry)
		authztypes.RegisterInterfaces(encodingCfg.InterfaceRegistry)
	}

	pellcoreClients, err := pellcore_rpc.NewGRPCClients(cosmosGRPC(chainIP), unsecureGRPC)
	if err != nil {
		return nil, errors.Wrap(err, "grpc dial fail")
	}

	accountsMap := make(map[authz.KeyType]uint64)
	seqMap := make(map[authz.KeyType]uint64)
	for _, keyType := range authz.GetAllKeyTypes() {
		accountsMap[keyType] = 0
		seqMap[keyType] = 0
	}

	cosmosContext, err := buildCosmosClientContext(chainID, keys, cfg, encodingCfg, constructOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build cosmos client context")
	}

	return &PellCoreBridge{
		Clients: pellcoreClients,
		logger:  log,
		config:  cfg,

		cosmosClientContext: cosmosContext,

		accountNumber: accountsMap,
		seqNumber:     seqMap,

		encodingCfg:     encodingCfg,
		keys:            keys,
		chainID:         chainID,
		chain:           chain,
		stop:            make(chan struct{}),
		pellTxMsgLength: pellTxMsgLength,

		enableMockSDKClient: false,
		mockSDKClient:       nil,
	}, nil
}

// buildCosmosClientContext constructs a valid context with all relevant values set
func buildCosmosClientContext(
	chainID string,
	keys keyinterfaces.ObserverKeys,
	config config.ClientConfiguration,
	encodingConfig params.EncodingConfig,
	opts constructOpts,
) (cosmosclient.Context, error) {
	if keys == nil {
		return cosmosclient.Context{}, errors.New("client key are not set")
	}

	addr, err := keys.GetSignerInfo().GetAddress()
	if err != nil {
		return cosmosclient.Context{}, errors.Wrap(err, "fail to get address from key")
	}

	// if password is needed, set it as input
	var (
		input   = strings.NewReader("")
		client  rpcclient.Client
		nodeURI string
	)
	password := keys.GetHotkeyPassword()
	if password != "" {
		input = strings.NewReader(fmt.Sprintf("%[1]s\n%[1]s\n", password))
	}

	{
		remote := config.ChainRPC
		if !strings.HasPrefix(config.ChainHost, "http") {
			remote = fmt.Sprintf("tcp://%s", remote)
		}
		nodeURI = remote

		wsClient, err := rpchttp.New(remote, "/websocket")
		if err != nil {
			return cosmosclient.Context{}, err
		}
		client = wsClient
	}

	var accountRetriever cosmosclient.AccountRetriever
	if opts.customAccountRetriever {
		accountRetriever = opts.accountRetriever
	} else {
		accountRetriever = authtypes.AccountRetriever{}
	}

	return cosmosclient.Context{
		Client:        client,
		NodeURI:       nodeURI,
		FromAddress:   addr,
		ChainID:       chainID,
		Keyring:       keys.GetKeybase(),
		BroadcastMode: "sync",
		HomeDir:       config.ChainHomeFolder,
		FromName:      config.SignerName,

		AccountRetriever: accountRetriever,

		Codec:             encodingConfig.Codec,
		InterfaceRegistry: encodingConfig.InterfaceRegistry,
		TxConfig:          encodingConfig.TxConfig,
		LegacyAmino:       encodingConfig.Amino,

		Input: input,
	}, nil
}

func (b *PellCoreBridge) UpdateChainID(chainID string) error {
	if b.chainID != chainID {
		b.chainID = chainID

		chain, err := chains.PellChainFromChainID(chainID)
		if err != nil {
			return fmt.Errorf("invalid chain id %s, %w", chainID, err)
		}
		b.chain = chain
	}

	return nil
}

// Chain returns the pellchain object
func (b *PellCoreBridge) Chain() chains.Chain {
	return b.chain
}

func (b *PellCoreBridge) GetLogger() *zerolog.Logger {
	return &b.logger
}

func (b *PellCoreBridge) WithKeys(keys keyinterfaces.ObserverKeys) {
	b.keys = keys
}

func (b *PellCoreBridge) GetKeys() keyinterfaces.ObserverKeys {
	return b.keys
}

// OnBeforeStop adds a callback to be called before the client stops.
func (c *PellCoreBridge) OnBeforeStop(callback func()) {
	c.onBeforeStopCallback = append(c.onBeforeStopCallback, callback)
}

// Stop stops the client and optionally calls the onBeforeStop callbacks.
func (c *PellCoreBridge) Stop() {
	c.logger.Info().Msgf("Stopping pellcore client")

	for i := len(c.onBeforeStopCallback) - 1; i >= 0; i-- {
		c.logger.Info().Int("callback.index", i).Msgf("calling onBeforeStopCallback")
		c.onBeforeStopCallback[i]()
	}

	close(c.stop)
}

// GetAccountNumberAndSequenceNumber We do not use multiple KeyType for now ,
// but this can be optionally used in the future to seprate TSS signer from Pellclient GRantee
func (b *PellCoreBridge) GetAccountNumberAndSequenceNumber(_ authz.KeyType) (uint64, uint64, error) {
	address, err := b.keys.GetAddress()
	if err != nil {
		return 0, 0, err
	}
	return b.cosmosClientContext.AccountRetriever.GetAccountNumberSequence(b.cosmosClientContext, address)
}

// SetAccountNumber sets the account number and sequence number for the given keyType
// todo remove method and make it part of the client constructor.
func (b *PellCoreBridge) SetAccountNumber(keyType authz.KeyType) error {
	address, err := b.keys.GetAddress()
	if err != nil {
		return err
	}

	accN, seq, err := b.cosmosClientContext.AccountRetriever.GetAccountNumberSequence(b.cosmosClientContext, address)
	if err != nil {
		return errors.Wrap(err, "fail to get account number and sequence number")
	}

	b.accountNumber[keyType] = accN
	b.seqNumber[keyType] = seq

	return nil
}

// WaitForPellcoreToCreateBlocks waits for pellcore to create blocks
func (b *PellCoreBridge) WaitForPellCoreToCreateBlocks(ctx context.Context) error {
	retryCount := 0
	for {
		block, err := b.GetLatestPellBlock(ctx)
		if err == nil && block.Header.Height > 1 {
			b.logger.Info().Msgf("Pellcore height: %d", block.Header.Height)
			break
		}
		retryCount++
		b.logger.Debug().Msgf("Failed to get latest Block , Retry : %d/%d", retryCount, DefaultRetryCount)
		if retryCount > ExtendedRetryCount {
			return fmt.Errorf("pellcore is not ready , Waited for %d seconds", DefaultRetryCount*DefaultRetryInterval)
		}
		time.Sleep(DefaultRetryInterval * time.Second)
	}
	return nil
}

// UpdatePellCoreContext updates core context
// pellcore stores core context for all clients
func (b *PellCoreBridge) UpdateAppContext(ctx context.Context, appContext *pctx.AppContext, init bool, logger zerolog.Logger) error {
	bn, err := b.GetBlockHeight(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get pellblock height")
	}

	plan, err := b.GetUpgradePlan(ctx)
	if err != nil {
		// if there is no active upgrade plan, plan will be nil, err will be nil as well.
		return errors.Wrap(err, "unable to get upgrade plan")
	}

	// Stop client and notify dependant services to stop (Orchestrator, Observers, and Signers)
	if plan != nil && bn == plan.Height-1 { // stop pellclients; notify operator to upgrade and restart
		b.logger.Warn().Msgf(
			"Active upgrade plan detected and upgrade height reached: %s at height %d; Stopping PellClient;"+
				" please kill this process, replace pellclientd binary with upgraded version, and restart pellclientd",
			plan.Name,
			plan.Height,
		)
		b.Stop()
		return nil
	}

	supportedChains, err := b.GetSupportedChains(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to fetch supported chains")
	}

	chainParams, err := b.GetChainParams(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to fetch chain params")
	}

	keyGen, err := b.GetKeyGen(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to fetch keygen from pellcore")
	}

	crosschainFlags, err := b.GetCrosschainFlags(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to fetch crosschain flags from pellcore")
	}

	tss, err := b.GetCurrentTss(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to fetch current TSS")
	}
	tssPubKey := tss.GetTssPubkey()

	newEVMParams := make(map[int64]*relayertypes.ChainParams)
	// check and update chain params for each chain
	for _, chainParam := range chainParams {
		if !chainParam.GetIsSupported() {
			logger.Warn().Int64("chain.id", chainParam.ChainId).Msg("Skipping unsupported chain")
			continue
		}

		// if chains.IsPellChain(chainParam.ChainId) {
		// 	continue
		// }

		err := relayertypes.ValidateChainParams(chainParam)
		if err != nil {
			logger.Warn().Err(err).Int64("chain.id", chainParam.ChainId).Msg("Skipping invalid chain params")
			continue
		}

		if chains.IsEVMChain(chainParam.ChainId) {
			newEVMParams[chainParam.ChainId] = chainParam
		}
	}

	newSupportedChains := make([]chains.Chain, len(supportedChains))
	for i, chain := range supportedChains {
		newSupportedChains[i] = *chain
	}

	verificationFlags, err := b.GetVerificationFlags(ctx)
	if err != nil {
		b.logger.Info().Msg("Unable to fetch verification flags from pellcore")

		// The block header functionality is currently disabled on the PellCore side
		// The verification flags might not exist and we should not return an error here to prevent the PellClient from starting
		// TODO: Uncomment this line when the block header functionality is enabled and we need to get the verification flags
		// return fmt.Errorf("failed to get verification flags: %w", err)

		verificationFlags = lightclienttypes.VerificationFlags{}
	}

	return appContext.Update(
		keyGen,
		newSupportedChains,
		newEVMParams,
		tssPubKey,
		crosschainFlags,
		verificationFlags,
		init,
		b.logger,
	)
}

func cosmosREST(host string) string {
	return fmt.Sprintf("%s:1317", host)
}

func cosmosGRPC(host string) string {
	return fmt.Sprintf("%s:9090", host)
}

func tendermintRPC(host string) string {
	return fmt.Sprintf("%s:26657", host)
}

func (b *PellCoreBridge) EnableMockSDKClient(client rpcclient.Client) {
	b.mockSDKClient = client
	b.enableMockSDKClient = true
}
