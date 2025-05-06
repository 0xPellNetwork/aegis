package keeper_test

import (
	"errors"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	testkeeper "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	observertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestKeeper_IterateAndUpdateXmsgGasPrice(t *testing.T) {
	k, ctx, _, zk := testkeeper.XmsgKeeper(t)

	// updateFuncMap tracks the calls done with xmsg index
	updateFuncMap := make(map[string]struct{})

	// failMap gives the xmsg index that should fail
	failMap := make(map[string]struct{})

	// updateFunc mocks the update function and keep track of the calls done with xmsg index
	updateFunc := func(
		ctx sdk.Context,
		k keeper.Keeper,
		xmsg types.Xmsg,
		flags observertypes.GasPriceIncreaseFlags,
	) (math.Uint, math.Uint, error) {
		if _, ok := failMap[xmsg.Index]; ok {
			return math.NewUint(0), math.NewUint(0), errors.New("failed")
		}

		updateFuncMap[xmsg.Index] = struct{}{}
		return math.NewUint(10), math.NewUint(10), nil
	}

	ethChain := chains.EthChain()
	bscChain := chains.BscMainnetChain()
	pellChain := chains.PellChainMainnet()
	// add some evm and non-evm chains
	supportedChains := []*chains.Chain{
		&ethChain,
		&bscChain,
		&pellChain,
	}

	// set pending xmsg
	tss := sample.Tss_pell()
	zk.ObserverKeeper.SetTSS(ctx, tss)
	createXmsgWithNonceRange(t, ctx, *k, 10, 15, chains.EthChain().Id, tss, zk)
	createXmsgWithNonceRange(t, ctx, *k, 30, 35, chains.BscMainnetChain().Id, tss, zk)
	createXmsgWithNonceRange(t, ctx, *k, 40, 45, chains.PellChainMainnet().Id, tss, zk)

	// set a xmsg where the update function should fail to test that the next xmsg are not updated but the next chains are
	failMap[sample.GetXmsgIndicesFromString_pell("1-12")] = struct{}{}

	// test that the default crosschain flags are used when not set and the epoch length is not reached
	ctx = ctx.WithBlockHeight(observertypes.DefaultCrosschainFlags().GasPriceIncreaseFlags.EpochLength + 1)

	xmsgCount, flags := k.IterateAndUpdateXmsgGasPrice(ctx, supportedChains, updateFunc)
	require.Equal(t, 0, xmsgCount)
	require.Equal(t, *observertypes.DefaultCrosschainFlags().GasPriceIncreaseFlags, flags)

	// test that custom crosschain flags are used when set and the epoch length is reached
	customFlags := observertypes.GasPriceIncreaseFlags{
		EpochLength:             100,
		RetryInterval:           time.Minute * 10,
		GasPriceIncreasePercent: 100,
		GasPriceIncreaseMax:     200,
		MaxPendingXmsgs:         10,
	}
	crosschainFlags := sample.CrosschainFlags_pell()
	crosschainFlags.GasPriceIncreaseFlags = &customFlags
	zk.ObserverKeeper.SetCrosschainFlags(ctx, *crosschainFlags)

	xmsgCount, flags = k.IterateAndUpdateXmsgGasPrice(ctx, supportedChains, updateFunc)
	require.Equal(t, 0, xmsgCount)
	require.Equal(t, customFlags, flags)

	// test that xmsg are iterated and updated when the epoch length is reached
	ctx = ctx.WithBlockHeight(observertypes.DefaultCrosschainFlags().GasPriceIncreaseFlags.EpochLength * 2)
	xmsgCount, flags = k.IterateAndUpdateXmsgGasPrice(ctx, supportedChains, updateFunc)

	// 2 eth + 5 bsc = 7
	require.Equal(t, 7, xmsgCount)
	require.Equal(t, customFlags, flags)

	// check that the update function was called with the xmsg index
	require.Equal(t, 7, len(updateFuncMap))
	require.Contains(t, updateFuncMap, sample.GetXmsgIndicesFromString_pell("1-10"))
	require.Contains(t, updateFuncMap, sample.GetXmsgIndicesFromString_pell("1-11"))

	require.Contains(t, updateFuncMap, sample.GetXmsgIndicesFromString_pell("56-30"))
	require.Contains(t, updateFuncMap, sample.GetXmsgIndicesFromString_pell("56-31"))
	require.Contains(t, updateFuncMap, sample.GetXmsgIndicesFromString_pell("56-32"))
	require.Contains(t, updateFuncMap, sample.GetXmsgIndicesFromString_pell("56-33"))
	require.Contains(t, updateFuncMap, sample.GetXmsgIndicesFromString_pell("56-34"))
}

func TestCheckAndUpdateXmsgGasPrice(t *testing.T) {
	sampleTimestamp := time.Now()
	retryIntervalReached := sampleTimestamp.Add(observertypes.DefaultGasPriceIncreaseFlags.RetryInterval + time.Second)
	retryIntervalNotReached := sampleTimestamp.Add(observertypes.DefaultGasPriceIncreaseFlags.RetryInterval - time.Second)

	tt := []struct {
		name                     string
		xmsg                     types.Xmsg
		flags                    observertypes.GasPriceIncreaseFlags
		blockTimestamp           time.Time
		medianGasPrice           uint64
		expectedGasPriceIncrease math.Uint
		expectedAdditionalFees   math.Uint
		isError                  bool
	}{
		{
			name: "can update gas price when retry interval is reached",
			xmsg: types.Xmsg{
				Index: "a1",
				XmsgStatus: &types.Status{
					LastUpdateTimestamp: sampleTimestamp.Unix(),
				},
				OutboundTxParams: []*types.OutboundTxParams{
					{
						ReceiverChainId:    42,
						OutboundTxGasLimit: 1000,
						OutboundTxGasPrice: "100",
					},
				},
			},
			flags:                    observertypes.DefaultGasPriceIncreaseFlags,
			blockTimestamp:           retryIntervalReached,
			medianGasPrice:           50,
			expectedGasPriceIncrease: math.NewUint(50),    // 100% medianGasPrice
			expectedAdditionalFees:   math.NewUint(50000), // gasLimit * increase
		},
		{
			name: "can update gas price at max limit",
			xmsg: types.Xmsg{
				Index: "a2",
				XmsgStatus: &types.Status{
					LastUpdateTimestamp: sampleTimestamp.Unix(),
				},
				OutboundTxParams: []*types.OutboundTxParams{
					{
						ReceiverChainId:    42,
						OutboundTxGasLimit: 1000,
						OutboundTxGasPrice: "100",
					},
				},
			},
			flags: observertypes.GasPriceIncreaseFlags{
				EpochLength:             100,
				RetryInterval:           time.Minute * 10,
				GasPriceIncreasePercent: 200, // Increase gas price to 100+50*2 = 200
				GasPriceIncreaseMax:     400, // Max gas price is 50*4 = 200
			},
			blockTimestamp:           retryIntervalReached,
			medianGasPrice:           50,
			expectedGasPriceIncrease: math.NewUint(100),    // 200% medianGasPrice
			expectedAdditionalFees:   math.NewUint(100000), // gasLimit * increase
		},
		{
			name: "default gas price increase limit used if not defined",
			xmsg: types.Xmsg{
				Index: "a3",
				XmsgStatus: &types.Status{
					LastUpdateTimestamp: sampleTimestamp.Unix(),
				},
				OutboundTxParams: []*types.OutboundTxParams{
					{
						ReceiverChainId:    42,
						OutboundTxGasLimit: 1000,
						OutboundTxGasPrice: "100",
					},
				},
			},
			flags: observertypes.GasPriceIncreaseFlags{
				EpochLength:             100,
				RetryInterval:           time.Minute * 10,
				GasPriceIncreasePercent: 100,
				GasPriceIncreaseMax:     0, // Limit should not be reached
			},
			blockTimestamp:           retryIntervalReached,
			medianGasPrice:           50,
			expectedGasPriceIncrease: math.NewUint(50),    // 100% medianGasPrice
			expectedAdditionalFees:   math.NewUint(50000), // gasLimit * increase
		},
		{
			name: "skip if max limit reached",
			xmsg: types.Xmsg{
				Index: "b0",
				XmsgStatus: &types.Status{
					LastUpdateTimestamp: sampleTimestamp.Unix(),
				},
				OutboundTxParams: []*types.OutboundTxParams{
					{
						ReceiverChainId:    42,
						OutboundTxGasLimit: 1000,
						OutboundTxGasPrice: "100",
					},
				},
			},
			flags: observertypes.GasPriceIncreaseFlags{
				EpochLength:             100,
				RetryInterval:           time.Minute * 10,
				GasPriceIncreasePercent: 200, // Increase gas price to 100+50*2 = 200
				GasPriceIncreaseMax:     300, // Max gas price is 50*3 = 150
			},
			blockTimestamp:           retryIntervalReached,
			medianGasPrice:           50,
			expectedGasPriceIncrease: math.NewUint(0),
			expectedAdditionalFees:   math.NewUint(0),
		},
		{
			name: "skip if gas price is not set",
			xmsg: types.Xmsg{
				Index: "b1",
				XmsgStatus: &types.Status{
					LastUpdateTimestamp: sampleTimestamp.Unix(),
				},
				OutboundTxParams: []*types.OutboundTxParams{
					{
						ReceiverChainId:    42,
						OutboundTxGasLimit: 100,
						OutboundTxGasPrice: "",
					},
				},
			},
			flags:                    observertypes.DefaultGasPriceIncreaseFlags,
			blockTimestamp:           retryIntervalReached,
			medianGasPrice:           100,
			expectedGasPriceIncrease: math.NewUint(0),
			expectedAdditionalFees:   math.NewUint(0),
		},
		{
			name: "skip if gas limit is not set",
			xmsg: types.Xmsg{
				Index: "b2",
				XmsgStatus: &types.Status{
					LastUpdateTimestamp: sampleTimestamp.Unix(),
				},
				OutboundTxParams: []*types.OutboundTxParams{
					{
						ReceiverChainId:    42,
						OutboundTxGasLimit: 0,
						OutboundTxGasPrice: "100",
					},
				},
			},
			flags:                    observertypes.DefaultGasPriceIncreaseFlags,
			blockTimestamp:           retryIntervalReached,
			medianGasPrice:           100,
			expectedGasPriceIncrease: math.NewUint(0),
			expectedAdditionalFees:   math.NewUint(0),
		},
		{
			name: "skip if retry interval is not reached",
			xmsg: types.Xmsg{
				Index: "b3",
				XmsgStatus: &types.Status{
					LastUpdateTimestamp: sampleTimestamp.Unix(),
				},
				OutboundTxParams: []*types.OutboundTxParams{
					{
						ReceiverChainId:    42,
						OutboundTxGasLimit: 0,
						OutboundTxGasPrice: "100",
					},
				},
			},
			flags:                    observertypes.DefaultGasPriceIncreaseFlags,
			blockTimestamp:           retryIntervalNotReached,
			medianGasPrice:           100,
			expectedGasPriceIncrease: math.NewUint(0),
			expectedAdditionalFees:   math.NewUint(0),
		},
		{
			name: "returns error if can't find median gas price",
			xmsg: types.Xmsg{
				Index: "c1",
				XmsgStatus: &types.Status{
					LastUpdateTimestamp: sampleTimestamp.Unix(),
				},
				OutboundTxParams: []*types.OutboundTxParams{
					{
						ReceiverChainId:    42,
						OutboundTxGasLimit: 1000,
						OutboundTxGasPrice: "100",
					},
				},
			},
			flags:          observertypes.DefaultGasPriceIncreaseFlags,
			blockTimestamp: retryIntervalReached,
			medianGasPrice: 0,
			isError:        true,
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			k, ctx := testkeeper.XmsgKeeperAllMocks(t)
			chainID := tc.xmsg.GetCurrentOutTxParam().ReceiverChainId
			previousGasPrice, err := tc.xmsg.GetCurrentOutTxParam().GetGasPrice()
			if err != nil {
				previousGasPrice = 0
			}

			// set median gas price if not zero
			if tc.medianGasPrice != 0 {
				k.SetGasPrice(ctx, types.GasPrice{
					ChainId:     chainID,
					Prices:      []uint64{tc.medianGasPrice},
					MedianIndex: 0,
				})

				// ensure median gas price is set
				medianGasPrice, isFound := k.GetMedianGasPriceInUint(ctx, chainID)
				require.True(t, isFound)
				require.True(t, medianGasPrice.Equal(math.NewUint(tc.medianGasPrice)))
			}

			// set block timestamp
			ctx = ctx.WithBlockTime(tc.blockTimestamp)

			// check and update gas price
			gasPriceIncrease, feesPaid, err := keeper.CheckAndUpdateXmsgGasPrice(ctx, *k, tc.xmsg, tc.flags)

			if tc.isError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// check values
			require.True(t, gasPriceIncrease.Equal(tc.expectedGasPriceIncrease), "expected %s, got %s", tc.expectedGasPriceIncrease.String(), gasPriceIncrease.String())
			require.True(t, feesPaid.Equal(tc.expectedAdditionalFees), "expected %s, got %s", tc.expectedAdditionalFees.String(), feesPaid.String())

			// check xmsg
			if !tc.expectedGasPriceIncrease.IsZero() {
				xmsg, found := k.GetXmsg(ctx, tc.xmsg.Index)
				require.True(t, found)
				newGasPrice, err := xmsg.GetCurrentOutTxParam().GetGasPrice()
				require.NoError(t, err)
				require.EqualValues(t, tc.expectedGasPriceIncrease.AddUint64(previousGasPrice).Uint64(), newGasPrice, "%d - %d", tc.expectedGasPriceIncrease.Uint64(), previousGasPrice)
				require.EqualValues(t, tc.blockTimestamp.Unix(), xmsg.XmsgStatus.LastUpdateTimestamp)
			}
		})
	}
}
