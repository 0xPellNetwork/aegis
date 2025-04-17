package keeper

import (
	"fmt"
	"time"

	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	pellchains "github.com/pell-chain/pellcore/pkg/chains"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

const (
	// RemainingFeesToStabilityPoolPercent is the percentage of remaining fees used to fund the gas stability pool
	RemainingFeesToStabilityPoolPercent = 95
)

// CheckAndUpdateXmsgGasPriceFunc is a function type for checking and updating the gas price of a xmsg
type CheckAndUpdateXmsgGasPriceFunc func(
	ctx sdk.Context,
	k Keeper,
	xmsg types.Xmsg,
	flags observertypes.GasPriceIncreaseFlags,
) (math.Uint, math.Uint, error)

// IterateAndUpdateXmsgGasPrice iterates through all xmsg and updates the gas price if pending for too long
// The function returns the number of xmsgs updated and the gas price increase flags used
func (k Keeper) IterateAndUpdateXmsgGasPrice(
	ctx sdk.Context,
	chains []*pellchains.Chain,
	updateFunc CheckAndUpdateXmsgGasPriceFunc,
) (int, observertypes.GasPriceIncreaseFlags) {
	// fetch the gas price increase flags or use default
	gasPriceIncreaseFlags := observertypes.DefaultGasPriceIncreaseFlags
	crosschainFlags, found := k.relayerKeeper.GetCrosschainFlags(ctx)
	if found && crosschainFlags.GasPriceIncreaseFlags != nil {
		gasPriceIncreaseFlags = *crosschainFlags.GasPriceIncreaseFlags
	}

	// skip if haven't reached epoch end
	if ctx.BlockHeight()%gasPriceIncreaseFlags.EpochLength != 0 {
		return 0, gasPriceIncreaseFlags
	}

	xmsgCount := 0

IterateChains:
	for _, chain := range chains {
		// support only external evm chains
		if pellchains.IsEVMChain(chain.Id) && !pellchains.IsPellChain(chain.Id) {
			res, err := k.ListPendingXmsg(sdk.UnwrapSDKContext(ctx), &types.QueryListPendingXmsgRequest{
				ChainId: chain.Id,
				Limit:   gasPriceIncreaseFlags.MaxPendingXmsgs,
			})
			if err != nil {
				ctx.Logger().Info("GasStabilityPool: fetching pending xmsg failed",
					"chainID", chain.Id,
					"err", err.Error(),
				)
				continue IterateChains
			}

			// iterate through all pending xmsg
			for _, pendingXmsg := range res.Xmsg {
				if pendingXmsg != nil {
					gasPriceIncrease, additionalFees, err := updateFunc(ctx, k, *pendingXmsg, gasPriceIncreaseFlags)
					if err != nil {
						ctx.Logger().Info("GasStabilityPool: updating gas price for pending xmsg failed",
							"xmsgIndex", pendingXmsg.Index,
							"err", err.Error(),
						)
						continue IterateChains
					}
					if !gasPriceIncrease.IsNil() && !gasPriceIncrease.IsZero() {
						// Emit typed event for gas price increase
						if err := ctx.EventManager().EmitTypedEvent(
							&types.EventXmsgGasPriceIncreased{
								XmsgIndex:        pendingXmsg.Index,
								GasPriceIncrease: gasPriceIncrease.String(),
								AdditionalFees:   additionalFees.String(),
							}); err != nil {
							ctx.Logger().Error(
								"GasStabilityPool: failed to emit EventXmsgGasPriceIncreased",
								"err", err.Error(),
							)
						}
						xmsgCount++
					}
				}
			}
		}
	}

	return xmsgCount, gasPriceIncreaseFlags
}

// CheckAndUpdateXmsgGasPrice checks if the retry interval is reached and updates the gas price if so
// The function returns the gas price increase and the additional fees paid from the gas stability pool
func CheckAndUpdateXmsgGasPrice(
	ctx sdk.Context,
	k Keeper,
	xmsg types.Xmsg,
	flags observertypes.GasPriceIncreaseFlags,
) (math.Uint, math.Uint, error) {
	// skip if gas price or gas limit is not set
	if xmsg.GetCurrentOutTxParam().OutboundTxGasPrice == "" || xmsg.GetCurrentOutTxParam().OutboundTxGasLimit == 0 {
		return math.ZeroUint(), math.ZeroUint(), nil
	}

	// skip if retry interval is not reached
	lastUpdated := time.Unix(xmsg.XmsgStatus.LastUpdateTimestamp, 0)
	if ctx.BlockTime().Before(lastUpdated.Add(flags.RetryInterval)) {
		return math.ZeroUint(), math.ZeroUint(), nil
	}

	// compute gas price increase
	chainID := xmsg.GetCurrentOutTxParam().ReceiverChainId
	medianGasPrice, isFound := k.GetMedianGasPriceInUint(ctx, chainID)
	if !isFound {
		return math.ZeroUint(), math.ZeroUint(), cosmoserrors.Wrap(
			types.ErrUnableToGetGasPrice,
			fmt.Sprintf("cannot get gas price for chain %d", chainID),
		)
	}
	gasPriceIncrease := medianGasPrice.MulUint64(uint64(flags.GasPriceIncreasePercent)).QuoUint64(100)

	// compute new gas price
	currentGasPrice, err := xmsg.GetCurrentOutTxParam().GetGasPrice()
	if err != nil {
		return math.ZeroUint(), math.ZeroUint(), err
	}
	newGasPrice := math.NewUint(currentGasPrice).Add(gasPriceIncrease)

	// check limit -- use default limit if not set
	gasPriceIncreaseMax := flags.GasPriceIncreaseMax
	if gasPriceIncreaseMax == 0 {
		gasPriceIncreaseMax = observertypes.DefaultGasPriceIncreaseFlags.GasPriceIncreaseMax
	}
	limit := medianGasPrice.MulUint64(uint64(gasPriceIncreaseMax)).QuoUint64(100)
	if newGasPrice.GT(limit) {
		return math.ZeroUint(), math.ZeroUint(), nil
	}

	// withdraw additional fees from the gas stability pool
	gasLimit := math.NewUint(xmsg.GetCurrentOutTxParam().OutboundTxGasLimit)
	additionalFees := gasLimit.Mul(gasPriceIncrease)

	// set new gas price and last update timestamp
	xmsg.GetCurrentOutTxParam().OutboundTxGasPrice = newGasPrice.String()
	xmsg.XmsgStatus.LastUpdateTimestamp = ctx.BlockHeader().Time.Unix()
	k.SetXmsg(ctx, xmsg)

	return gasPriceIncrease, additionalFees, nil
}
