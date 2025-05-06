package keeper

import (
	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	pevmtypes "github.com/0xPellNetwork/aegis/x/pevm/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// PayGasAndUpdateXmsg updates the outbound tx with the new amount after paying the gas fee
// **Caller should feed temporary ctx into this function**
// chainID is the outbound chain chain id , this can be receiver chain for regular transactions and sender-chain to reverted transactions
func (k Keeper) PayGasAndUpdateXmsg(
	ctx sdk.Context,
	chainID int64,
	xmsg *types.Xmsg,
) error {
	return k.PayNativeGasAndUpdateXmsg(ctx, chainID, xmsg)
}

// ChainGasParams returns the params to calculates the fees for gas for a chain
// tha gas address, the gas limit, gas price and protocol flat fee are returned
func (k Keeper) getChainGasParams(
	ctx sdk.Context,
	chainID int64,
) (gasLimit, gasPrice math.Uint, err error) {

	// get the gas price
	gasPrice, isFound := k.GetMedianGasPriceInUint(ctx, chainID)
	if !isFound {
		return gasLimit, gasPrice, types.ErrUnableToGetGasPrice
	}
	gasLimit = math.NewUintFromBigInt(pevmtypes.PEVMGasLimit)

	return
}

// PayGasNativeAndUpdateXmsg updates the outbound tx with the new amount subtracting the gas fee
// **Caller should feed temporary ctx into this function**
func (k Keeper) PayNativeGasAndUpdateXmsg(
	ctx sdk.Context,
	chainID int64,
	xmsg *types.Xmsg,
) error {
	if chain := k.relayerKeeper.GetSupportedChainFromChainID(ctx, chainID); chain == nil {
		return relayertypes.ErrSupportedChains
	}

	// get gas params
	_, gasPrice, err := k.getChainGasParams(ctx, chainID)
	if err != nil {
		return cosmoserrors.Wrap(types.ErrCannotFindGasParams, err.Error())
	}

	// xmsg.GetCurrentOutTxParam().OutboundTxGasLimit = gasLimit.Uint64()
	xmsg.GetCurrentOutTxParam().OutboundTxGasPrice = gasPrice.String()

	return nil
}
