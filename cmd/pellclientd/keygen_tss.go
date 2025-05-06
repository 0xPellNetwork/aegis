package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/rs/zerolog"
	"github.com/tendermint/crypto/sha3"
	tsscommon "gitlab.com/thorchain/tss/go-tss/common"
	"gitlab.com/thorchain/tss/go-tss/keygen"
	"gitlab.com/thorchain/tss/go-tss/p2p"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	pctx "github.com/0xPellNetwork/aegis/relayer/context"
	"github.com/0xPellNetwork/aegis/relayer/metrics"
	"github.com/0xPellNetwork/aegis/relayer/pellcore"
	mc "github.com/0xPellNetwork/aegis/relayer/tss"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

func GenerateTss(
	ctx context.Context,
	logger zerolog.Logger,
	pellBridge *pellcore.PellCoreBridge,
	peers p2p.AddrList,
	priKey secp256k1.PrivKey,
	ts *metrics.TelemetryServer,
	tssHistoricalList []relayertypes.TSS,
	tssPassword string,
	hotkeyPassword string,
) (*mc.TSS, error) {
	keygenLogger := logger.With().Str("module", "keygen").Logger()
	app, err := pctx.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	tss, err := mc.NewTSS(
		ctx,
		peers,
		priKey,
		preParams,
		pellBridge,
		tssHistoricalList,
		tssPassword,
		hotkeyPassword,
	)
	if err != nil {
		keygenLogger.Error().Err(err).Msg("NewTSS error")
		return nil, err
	}
	ts.SetP2PID(tss.Server.GetLocalPeerID())

	// If Keygen block is set it will try to generate new TSS at the block
	// This is a blocking thread and will wait until the ceremony is complete successfully
	// If the TSS generation is unsuccessful , it will loop indefinitely until a new TSS is generated
	// Set TSS block to 0 using genesis file to disable this feature
	// Note : The TSS generation is done through the "hotkey" or "Pell-clientGrantee" This key needs to be present on the machine for the TSS signing to happen .
	// "PellClientGrantee" key is different from the "operator" key .The "Operator" key gives all pellclient related permissions such as TSS generation ,reporting and signing, INBOUND and OUTBOUND vote signing, to the "PellClientGrantee" key.
	// The votes to signify a successful TSS generation (Or unsuccessful) is signed by the operator key and broadcast to pellcore by the zetcalientGrantee key on behalf of the operator .
	ticker := time.NewTicker(time.Second * 1)
	triedKeygenAtBlock := false
	lastBlock := int64(0)
	for range ticker.C {
		// Break out of loop only when TSS is generated successfully, either at the keygenBlock or if it has been generated already , Block set as zero in genesis file
		// This loop will try keygen at the keygen block and then wait for keygen to be successfully reported by all nodes before breaking out of the loop.
		// If keygen is unsuccessful, it will reset the triedKeygenAtBlock flag and try again at a new keygen block.

		keyGen := app.GetKeygen()
		if keyGen.Status == relayertypes.KeygenStatus_SUCCESS {
			return tss, nil
		}
		// Arrive at this stage only if keygen is unsuccessfully reported by every node . This will reset the flag and to try again at a new keygen block
		if keyGen.Status == relayertypes.KeygenStatus_FAILED {
			triedKeygenAtBlock = false
			continue
		}
		// Try generating TSS at keygen block , only when status is pending keygen and generation has not been tried at the block
		if keyGen.Status == relayertypes.KeygenStatus_PENDING {
			// Return error if RPC is not working
			currentBlock, err := pellBridge.GetBlockHeight(ctx)
			if err != nil {
				keygenLogger.Error().Err(err).Msg("GetPellBlockHeight RPC  error")
				continue
			}
			// Reset the flag if the keygen block has passed and a new keygen block has been set . This condition is only reached if the older keygen is stuck at PendingKeygen for some reason
			if keyGen.BlockNumber > currentBlock {
				triedKeygenAtBlock = false
			}
			if !triedKeygenAtBlock {
				// If not at keygen block do not try to generate TSS
				if currentBlock != keyGen.BlockNumber {
					if currentBlock > lastBlock {
						lastBlock = currentBlock
						keygenLogger.Info().
							Msgf("Waiting For Keygen Block to arrive or new keygen block to be set. Keygen Block : %d Current Block : %d ChainID %s ",
								keyGen.BlockNumber, currentBlock, app.Config().ChainID)
					}
					continue
				}
				// Try keygen only once at a particular block, irrespective of whether it is successful or failure
				triedKeygenAtBlock = true
				err = keygenTss(ctx, keyGen, tss, keygenLogger)
				if err != nil {
					keygenLogger.Error().Err(err).Msg("keygenTss error")
					tssFailedVoteHash, err := pellBridge.PostVoteTSS(ctx, "", keyGen.BlockNumber, chains.ReceiveStatus_FAILED)
					if err != nil {
						keygenLogger.Error().Err(err).Msg("Failed to broadcast Failed TSS Vote to pellcore")
						return nil, err
					}
					keygenLogger.Info().Msgf("TSS Failed Vote: %s", tssFailedVoteHash)
					continue
				}

				newTss := mc.TSS{
					Server:        tss.Server,
					Keys:          tss.Keys,
					CurrentPubkey: tss.CurrentPubkey,
					Signers:       tss.Signers,
					CoreBridge:    nil,
				}

				// If TSS is successful , broadcast the vote to pellcore and set Pubkey
				tssSuccessVoteHash, err := pellBridge.PostVoteTSS(ctx, newTss.CurrentPubkey, keyGen.BlockNumber, chains.ReceiveStatus_SUCCESS)
				if err != nil {
					keygenLogger.Error().Err(err).Msg("TSS successful but unable to broadcast vote to pell-core")
					return nil, err
				}
				keygenLogger.Info().Msgf("TSS successful Vote: %s", tssSuccessVoteHash)
				err = SetTSSPubKey(tss, keygenLogger)
				if err != nil {
					keygenLogger.Error().Err(err).Msg("SetTSSPubKey error")
				}
				err = TestTSS(&newTss, keygenLogger)
				if err != nil {
					keygenLogger.Error().Err(err).Msgf("TestTSS error: %s", newTss.CurrentPubkey)
				}
				continue
			}
		}
		keygenLogger.Debug().Msgf("Waiting for TSS to be generated or Current Keygen to be be finalized. Keygen Block : %d ", keyGen.BlockNumber)
	}
	return nil, errors.New("unexpected state for TSS generation")
}

func keygenTss(ctx context.Context, keyGen relayertypes.Keygen, tss *mc.TSS, keygenLogger zerolog.Logger) error {
	keygenLogger.Info().Msgf("Keygen at blocknum %d , TSS signers %s ", keyGen.BlockNumber, keyGen.GranteePubkeys)
	var req keygen.Request
	req = keygen.NewRequest(keyGen.GranteePubkeys, keyGen.BlockNumber, "0.14.0")
	res, err := tss.Server.Keygen(req)
	if res.Status != tsscommon.Success || res.PubKey == "" {
		keygenLogger.Error().Msgf("keygen fail: reason %s blame nodes %s", res.Blame.FailReason, res.Blame.BlameNodes)
		// Need to broadcast keygen blame result here
		digest, err := digestReq(req)
		if err != nil {
			return err
		}
		index := fmt.Sprintf("keygen-%s-%d", digest, keyGen.BlockNumber)
		pellHash, err := tss.CoreBridge.PostBlameData(ctx, &res.Blame, tss.CoreBridge.Chain().Id, index)
		if err != nil {
			keygenLogger.Error().Err(err).Msg("error sending blame data to core")
			return err
		}

		// Increment Blame counter
		for _, node := range res.Blame.BlameNodes {
			metrics.TssNodeBlamePerPubKey.WithLabelValues(node.Pubkey).Inc()
		}

		keygenLogger.Info().Msgf("keygen posted blame data tx hash: %s", pellHash)
		return fmt.Errorf("keygen fail: reason %s blame nodes %s", res.Blame.FailReason, res.Blame.BlameNodes)
	}
	if err != nil {
		keygenLogger.Error().Msgf("keygen fail: reason %s ", err.Error())
		return err
	}
	// Keeping this line here for now, but this is redundant as CurrentPubkey is updated from pell-core
	tss.CurrentPubkey = res.PubKey
	tss.Signers = keyGen.GranteePubkeys

	// Keygen succeed! Report TSS address
	keygenLogger.Debug().Msgf("Keygen success! keygen response: %v", res)
	return nil
}

func SetTSSPubKey(tss *mc.TSS, logger zerolog.Logger) error {
	err := tss.InsertPubKey(tss.CurrentPubkey)
	if err != nil {
		logger.Error().Msgf("SetPubKey fail")
		return err
	}
	logger.Info().Msgf("TSS address in hex: %s", tss.EVMAddress().Hex())
	return nil

}

func TestTSS(tss *mc.TSS, logger zerolog.Logger) error {
	keygenLogger := logger.With().Str("module", "test-keygen").Logger()
	keygenLogger.Info().Msgf("KeyGen success ! Doing a Key-sign test")
	// KeySign can fail even if TSS keygen is successful, just logging the error here to break out of outer loop and report TSS
	err := mc.TestKeysign(tss.CurrentPubkey, tss.Server)
	if err != nil {
		return err
	}
	return nil
}

func digestReq(request keygen.Request) (string, error) {
	bytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(bytes)
	digest := hex.EncodeToString(hasher.Sum(nil))

	return digest, nil
}
