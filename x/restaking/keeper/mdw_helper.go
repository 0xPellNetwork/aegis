package keeper

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/pell-chain/pellcore/pkg/chains"
	pevmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	"github.com/pell-chain/pellcore/x/restaking/types"
	"github.com/pell-chain/pellcore/x/xmsg/keeper"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

// createAndSendXMsgFromEvent creates and sends an xmsg from an event
func (k *Keeper) processInboundEvent(
	ctx sdk.Context,
	txOrigin string,
	receiverChainID *big.Int,
	senderAddr ethcommon.Address,
	receiverAddr ethcommon.Address,
	message []byte,
	eventLog *ethtypes.Log,
) (*xmsgtypes.Xmsg, error) {
	if isValid := k.isWhitelistedRegistryRouter(ctx, senderAddr); !isValid {
		return nil, errors.New("sender contract address not whitelisted")
	}

	receiverChain := k.relayerKeeper.GetSupportedChainFromChainID(ctx, receiverChainID.Int64())
	if receiverChain == nil {
		return nil, errors.New("receiver chain not found")
	}

	chainParams, found := k.relayerKeeper.GetChainParamsByChainID(ctx, receiverChain.Id)
	if !found {
		return nil, errors.New("chain params not found")
	}

	if receiverChain.IsExternalChain() && (chainParams.ConnectorContractAddress == "") {
		return nil, errors.New("connector contract address not found")
	}

	senderChain, err := chains.PellChainFromChainID(ctx.ChainID())
	if err != nil {
		return nil, errors.New("sender chain not found")
	}

	pellSent := xmsgtypes.InboundPellEvent{
		PellData: &xmsgtypes.InboundPellEvent_PellSent{
			PellSent: &xmsgtypes.PellSent{
				TxOrigin:        txOrigin,
				Sender:          senderAddr.Hex(),
				ReceiverChainId: receiverChain.Id,
				Receiver:        receiverAddr.Hex(),
				Message:         base64.StdEncoding.EncodeToString(message),
				PellParams:      pevmtypes.ReceiveCall.String(),
			},
		},
	}

	msg := xmsgtypes.NewMsgVoteOnObservedInboundTx(
		"",
		senderAddr.Hex(),
		senderChain.Id,
		txOrigin,
		receiverAddr.Hex(),
		receiverChain.Id,
		eventLog.TxHash.String(),
		eventLog.BlockNumber,
		chainParams.GasLimit,
		eventLog.Index,
		pellSent,
	)

	return k.processInboundMsg(ctx, msg, receiverChain)
}

// processInboundMsg creates and sends an xmsg from a msg
func (k *Keeper) processInboundMsg(ctx sdk.Context, msg *xmsgtypes.MsgVoteOnObservedInboundTx, receiverChain *chains.Chain) (*xmsgtypes.Xmsg, error) {
	tss, found := k.relayerKeeper.GetTSS(ctx)
	if !found {
		return nil, errors.New("tss not found")
	}

	// Create a new xmsg with status as pending Inbound, this is created directly from the event without waiting for any observer votes
	xmsg, err := xmsgtypes.NewXmsg(ctx, *msg, tss.TssPubkey)
	if err != nil {
		return nil, err
	}

	xmsg.SetPendingOutbound("Cross-chain message ready for outbound processing from registry router")

	// Get gas price and amount
	gasPrice, found := k.xmsgKeeper.GetGasPrice(ctx, receiverChain.Id)
	if !found {
		return nil, errors.New("gas price not found")
	}

	xmsg.GetCurrentOutTxParam().OutboundTxGasPrice = fmt.Sprint(gasPrice.Prices[gasPrice.MedianIndex])
	keeper.EmitEventPellSent(ctx, xmsg)

	if err := k.xmsgKeeper.ProcessXmsg(ctx, xmsg, receiverChain); err != nil {
		return nil, err
	}

	return &xmsg, nil
}

// isWhitelistedRegistryRouter checks if a registry router address is whitelisted
func (k *Keeper) isWhitelistedRegistryRouter(ctx sdk.Context, addr ethcommon.Address) bool {
	// Get all whitelisted contract addresses
	whitelistedAddrs, err := k.GetAllRegistryRouterAddresses(ctx)
	if err != nil {
		return false
	}

	for _, whitelistedAddr := range whitelistedAddrs {
		if whitelistedAddr == addr {
			return true
		}
	}

	return false
}

// ConvertPubkeyRegistrationParamsFromEventToStore converts a pubkey registration params from event to store
func ConvertPubkeyRegistrationParamsFromEventToStore(params registryrouter.IRegistryRouterSyncPubkeyRegistrationParams) *types.PubkeyRegistrationParamsV2 {
	return &types.PubkeyRegistrationParamsV2{
		PubkeyG1: &types.G1PointV2{
			X: math.NewIntFromBigInt(params.PubkeyG1.X),
			Y: math.NewIntFromBigInt(params.PubkeyG1.Y),
		},
		PubkeyG2: &types.G2PointV2{
			X: []math.Int{
				math.NewIntFromBigInt(params.PubkeyG2.X[0]),
				math.NewIntFromBigInt(params.PubkeyG2.X[1]),
			},
			Y: []math.Int{
				math.NewIntFromBigInt(params.PubkeyG2.Y[0]),
				math.NewIntFromBigInt(params.PubkeyG2.Y[1]),
			},
		},
	}
}

// ConvertPubkeyRegistrationParamsFromStoreToEvent converts a pubkey registration params from store to event
func ConvertPubkeyRegistrationParamsFromStoreToEvent(params *types.PubkeyRegistrationParamsV2) registryrouter.IRegistryRouterSyncPubkeyRegistrationParams {
	return registryrouter.IRegistryRouterSyncPubkeyRegistrationParams{
		PubkeyG1: registryrouter.BN254G1Point{
			X: params.PubkeyG1.X.BigInt(),
			Y: params.PubkeyG1.Y.BigInt(),
		},
		PubkeyG2: registryrouter.BN254G2Point{
			X: [2]*big.Int{
				params.PubkeyG2.X[0].BigInt(),
				params.PubkeyG2.X[1].BigInt(),
			},
			Y: [2]*big.Int{
				params.PubkeyG2.Y[0].BigInt(),
				params.PubkeyG2.Y[1].BigInt(),
			},
		},
	}
}

// ConvertPoolParamsFromEventToStore converts a router strategy params from event to store
func ConvertPoolParamsFromEventToStore(params []registryrouter.IStakeRegistryRouterPoolParams) []*types.PoolParams {
	result := make([]*types.PoolParams, len(params))
	for i, param := range params {
		result[i] = &types.PoolParams{
			ChainId:    param.ChainId.Uint64(),
			Pool:       param.Pool.Hex(),
			Multiplier: param.Multiplier.Uint64(),
		}
	}
	return result
}

func ConvertPoolParamsFromStakeEventToStore(params []stakeregistryrouter.IStakeRegistryRouterPoolParams) []*types.PoolParams {
	result := make([]*types.PoolParams, len(params))
	for i, param := range params {
		result[i] = &types.PoolParams{
			ChainId:    param.ChainId.Uint64(),
			Pool:       param.Pool.Hex(),
			Multiplier: param.Multiplier.Uint64(),
		}
	}
	return result
}

// ConvertPoolParamsFromStoreToEvent converts a router strategy params from store to event
func ConvertPoolParamsFromStoreToEvent(params []*types.PoolParams) ([]registryrouter.IStakeRegistryRouterPoolParams, error) {
	result := make([]registryrouter.IStakeRegistryRouterPoolParams, len(params))
	for i, param := range params {
		result[i] = registryrouter.IStakeRegistryRouterPoolParams{
			ChainId:    new(big.Int).SetUint64(param.ChainId),
			Pool:       ethcommon.HexToAddress(param.Pool),
			Multiplier: new(big.Int).SetUint64(param.Multiplier),
		}
	}
	return result, nil
}

// ConvertPoolParamsFromStoreToStakeEvent converts a router strategy params from store to event
func ConvertPoolParamsFromStoreToStakeEvent(params []*types.PoolParams) ([]stakeregistryrouter.IStakeRegistryRouterPoolParams, error) {
	result := make([]stakeregistryrouter.IStakeRegistryRouterPoolParams, len(params))
	for i, param := range params {
		result[i] = stakeregistryrouter.IStakeRegistryRouterPoolParams{
			ChainId:    new(big.Int).SetUint64(param.ChainId),
			Pool:       ethcommon.HexToAddress(param.Pool),
			Multiplier: new(big.Int).SetUint64(param.Multiplier),
		}
	}
	return result, nil
}

// ConvertOperatorSetParamFromEventToStore converts an operator set param from event to store
func ConvertOperatorSetParamFromStoreToEvent(param *types.OperatorSetParam) registryrouter.IRegistryRouterOperatorSetParam {
	return registryrouter.IRegistryRouterOperatorSetParam{
		MaxOperatorCount:        param.MaxOperatorCount,
		KickBIPsOfOperatorStake: uint16(param.KickBipsOfOperatorStake),
		KickBIPsOfTotalStake:    uint16(param.KickBipsOfTotalStake),
	}
}

// IntersectBytes returns the intersection of two byte slices as a new slice.
func IntersectBytes(a, b []byte) []byte {
	// Use a map to mark existence of elements in 'a'.
	seen := make(map[byte]struct{})
	for _, val := range a {
		seen[val] = struct{}{}
	}

	// Check elements in 'b', if they exist in 'seen', append to result.
	var intersection []byte
	for _, val := range b {
		if _, ok := seen[val]; ok {
			intersection = append(intersection, val)
		}
	}
	return intersection
}
