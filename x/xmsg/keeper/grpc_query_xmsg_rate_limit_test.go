package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	observertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var (
	// local eth chain ID
	ethChainID = getValidEthChainID()

	// test bsc chain ID
	bscChainID = getValidBscChainID()

	// local pell chain ID
	pellChainID = chains.PellPrivnetChain().Id
)

// createTestRateLimiterFlags creates a custom rate limiter flags
func createTestRateLimiterFlags(
	window int64,
	rate math.Uint,
) *types.RateLimiterFlags {
	return &types.RateLimiterFlags{
		Enabled: true,
		Window:  window, // for instance: 500 pell blocks, 50 minutes
		Rate:    rate,
	}
}

// createXmsgsWithCoinTypeAndHeightRange
//   - create 1 cctx per block from lowBlock to highBlock (inclusive)
//
// return created cctxs
func createXmsgsWithCoinTypeAndHeightRange(
	t *testing.T,
	lowBlock uint64,
	highBlock uint64,
	senderChainID int64,
	receiverChainID int64,
	status types.XmsgStatus,
) (cctxs []*types.Xmsg) {
	// create 1 pending cctxs per block
	for i := lowBlock; i <= highBlock; i++ {
		nonce := i - 1
		cctx := sample.Xmsg_pell(t, fmt.Sprintf("%d-%d", receiverChainID, nonce))
		cctx.XmsgStatus.Status = status
		cctx.InboundTxParams.SenderChainId = senderChainID
		cctx.InboundTxParams.InboundTxBlockHeight = i
		cctx.GetCurrentOutTxParam().ReceiverChainId = receiverChainID
		cctx.GetCurrentOutTxParam().OutboundTxTssNonce = nonce
		cctxs = append(cctxs, cctx)
	}
	return cctxs
}

// setXmsgsInKeeper sets the given cctxs to the keeper
func setXmsgsInKeeper(
	ctx sdk.Context,
	k keeper.Keeper,
	zk keepertest.PellKeepers,
	tss observertypes.TSS,
	cctxs []*types.Xmsg,
) {
	for _, cctx := range cctxs {
		k.SetXmsg(ctx, *cctx)
		zk.ObserverKeeper.SetNonceToXmsg(ctx, observertypes.NonceToXmsg{
			ChainId: cctx.GetCurrentOutTxParam().ReceiverChainId,
			// #nosec G701 always in range for tests
			Nonce:     int64(cctx.GetCurrentOutTxParam().OutboundTxTssNonce),
			XmsgIndex: cctx.Index,
			Tss:       tss.TssPubkey,
		})
	}
}

func TestKeeper_ListPendingXmsgWithinRateLimit(t *testing.T) {
	// create sample TSS
	tss := sample.Tss_pell()

	// create sample zrc20 addresses for ETH, BTC, USDT
	// zrc20ETH := sample.EthAddress().Hex()
	// zrc20BTC := sample.EthAddress().Hex()
	// zrc20USDT := sample.EthAddress().Hex()

	// create Eth chain 999 mined and 200 pending cctxs for rate limiter test
	// the number 999 is to make it less than `MaxLookbackNonce` so the LoopBackwards gets the chance to hit nonce 0
	ethMinedXmsgs := createXmsgsWithCoinTypeAndHeightRange(t, 1, 999, pellChainID, ethChainID, types.XmsgStatus_OUTBOUND_MINED)
	ethPendingXmsgs := createXmsgsWithCoinTypeAndHeightRange(t, 1000, 1199, pellChainID, ethChainID, types.XmsgStatus_PENDING_OUTBOUND)

	// create Eth chain 999 reverted and 200 pending revert cctxs for rate limiter test
	// these cctxs should be just ignored by the rate limiter as we can't compare their `ObservedExternalHeight` with window boundary
	ethRevertedXmsgs := createXmsgsWithCoinTypeAndHeightRange(t, 1, 999, ethChainID, ethChainID, types.XmsgStatus_REVERTED)
	ethPendingRevertXmsgs := createXmsgsWithCoinTypeAndHeightRange(t, 1000, 1199, ethChainID, ethChainID, types.XmsgStatus_PENDING_REVERT)

	// create Bsc chain 999 mined and 200 pending cctxs for rate limiter test
	// the number 999 is to make it less than `MaxLookbackNonce` so the LoopBackwards gets the chance to hit nonce 0
	bscMinedXmsgs := createXmsgsWithCoinTypeAndHeightRange(t, 1, 999, pellChainID, bscChainID, types.XmsgStatus_OUTBOUND_MINED)
	bscPendingXmsgs := createXmsgsWithCoinTypeAndHeightRange(t, 1000, 1199, pellChainID, bscChainID, types.XmsgStatus_PENDING_OUTBOUND)

	// create Bsc chain 999 reverted and 200 pending revert cctxs for rate limiter test
	// these cctxs should be just ignored by the rate limiter as we can't compare their `ObservedExternalHeight` with window boundary
	bscRevertedXmsgs := createXmsgsWithCoinTypeAndHeightRange(t, 1, 999, bscChainID, bscChainID, types.XmsgStatus_REVERTED)
	bscPendingRevertXmsgs := createXmsgsWithCoinTypeAndHeightRange(t, 1000, 1199, bscChainID, bscChainID, types.XmsgStatus_PENDING_REVERT)

	// define test cases
	tests := []struct {
		name           string
		fallback       bool
		rateLimitFlags *types.RateLimiterFlags

		// Eth chain cctxs setup
		ethMinedXmsgs    []*types.Xmsg
		ethPendingXmsgs  []*types.Xmsg
		ethPendingNonces observertypes.PendingNonces

		// Bsc chain cctxs setup
		bscMinedXmsgs    []*types.Xmsg
		bscPendingXmsgs  []*types.Xmsg
		bscPendingNonces observertypes.PendingNonces

		// current block height and limit
		currentHeight int64
		queryLimit    uint32

		// expected results
		expectedXmsgs          []*types.Xmsg
		expectedTotalPending   uint64
		expectedWithdrawWindow int64
		expectedWithdrawRate   string
		rateLimitExceeded      bool
	}{
		{
			name:            "should use fallback query if rate limiter is disabled",
			fallback:        true,
			rateLimitFlags:  nil, // no rate limiter flags set in the keeper
			ethMinedXmsgs:   ethMinedXmsgs,
			ethPendingXmsgs: ethPendingXmsgs,
			ethPendingNonces: observertypes.PendingNonces{
				ChainId:   ethChainID,
				NonceLow:  1099,
				NonceHigh: 1199,
				Tss:       tss.TssPubkey,
			},
			currentHeight:        1199,
			queryLimit:           keeper.MaxPendingXmsgs,
			expectedXmsgs:        append(append([]*types.Xmsg{}), ethPendingXmsgs...),
			expectedTotalPending: 200,
		},
		{
			name:            "should use fallback query if rate is 0",
			fallback:        true,
			rateLimitFlags:  createTestRateLimiterFlags(500, math.NewUint(0)),
			ethMinedXmsgs:   ethMinedXmsgs,
			ethPendingXmsgs: ethPendingXmsgs,
			ethPendingNonces: observertypes.PendingNonces{
				ChainId:   ethChainID,
				NonceLow:  1099,
				NonceHigh: 1199,
				Tss:       tss.TssPubkey,
			},
			currentHeight:        1199,
			queryLimit:           keeper.MaxPendingXmsgs,
			expectedXmsgs:        append(append([]*types.Xmsg{}), ethPendingXmsgs...),
			expectedTotalPending: 200,
		},
		{
			name:            "can retrieve all pending cctx without exceeding rate limit",
			rateLimitFlags:  createTestRateLimiterFlags(700, math.NewUint(10*1e18)),
			ethMinedXmsgs:   ethMinedXmsgs,
			ethPendingXmsgs: ethPendingXmsgs,
			ethPendingNonces: observertypes.PendingNonces{
				ChainId:   ethChainID,
				NonceLow:  1099,
				NonceHigh: 1199,
				Tss:       tss.TssPubkey,
			},
			bscMinedXmsgs:   bscMinedXmsgs,
			bscPendingXmsgs: bscPendingXmsgs,
			bscPendingNonces: observertypes.PendingNonces{
				ChainId:   bscChainID,
				NonceLow:  1099,
				NonceHigh: 1199,
				Tss:       tss.TssPubkey,
			},
			currentHeight:          1199,
			queryLimit:             keeper.MaxPendingXmsgs,
			expectedXmsgs:          append(append(append([]*types.Xmsg{}, ethPendingXmsgs...))), // TODO: add bsc
			expectedTotalPending:   200,
			expectedWithdrawWindow: 700,                           // the sliding window
			expectedWithdrawRate:   sdkmath.NewInt(3e18).String(), // 3 PELL, (2.5 + 0.5) per block
			rateLimitExceeded:      false,
		},
		{
			name:            "can ignore reverted or pending revert cctxs and retrieve all pending cctx without exceeding rate limit",
			rateLimitFlags:  createTestRateLimiterFlags(700, math.NewUint(10*1e18)),
			ethMinedXmsgs:   ethRevertedXmsgs,      // replace mined cctxs with reverted cctxs, should be ignored
			ethPendingXmsgs: ethPendingRevertXmsgs, // replace pending cctxs with pending revert cctxs, should be ignored
			ethPendingNonces: observertypes.PendingNonces{
				ChainId:   ethChainID,
				NonceLow:  1099,
				NonceHigh: 1199,
				Tss:       tss.TssPubkey,
			},
			bscMinedXmsgs:   bscRevertedXmsgs,
			bscPendingXmsgs: bscPendingRevertXmsgs,
			bscPendingNonces: observertypes.PendingNonces{
				ChainId:   bscChainID,
				NonceLow:  1099,
				NonceHigh: 1199,
				Tss:       tss.TssPubkey,
			},
			currentHeight:          1199,
			queryLimit:             keeper.MaxPendingXmsgs,
			expectedXmsgs:          append(append(append([]*types.Xmsg{}, ethPendingRevertXmsgs...))), // TODO: add bsc
			expectedTotalPending:   200,
			expectedWithdrawWindow: 700,                           // the sliding window
			expectedWithdrawRate:   sdkmath.NewInt(5e17).String(), // 0.5 PELL per block, only btc cctxs should be counted
			rateLimitExceeded:      false,
		},
		{
			name:            "can loop backwards all the way to endNonce 0",
			rateLimitFlags:  createTestRateLimiterFlags(700, math.NewUint(10*1e18)),
			ethMinedXmsgs:   ethMinedXmsgs,
			ethPendingXmsgs: ethPendingXmsgs,
			ethPendingNonces: observertypes.PendingNonces{
				ChainId:   ethChainID,
				NonceLow:  999, // endNonce will be set to 0 (NonceLow - 1000 < 0)
				NonceHigh: 1199,
				Tss:       tss.TssPubkey,
			},
			bscMinedXmsgs:   bscMinedXmsgs,
			bscPendingXmsgs: bscPendingXmsgs,
			bscPendingNonces: observertypes.PendingNonces{
				ChainId:   bscChainID,
				NonceLow:  999,
				NonceHigh: 1199,
				Tss:       tss.TssPubkey,
			},
			currentHeight:          1199,
			queryLimit:             keeper.MaxPendingXmsgs,
			expectedXmsgs:          append(append(append([]*types.Xmsg{}, ethPendingXmsgs...))), // TODO: add bsc
			expectedTotalPending:   200,
			expectedWithdrawWindow: 700,                           // the sliding window
			expectedWithdrawRate:   sdkmath.NewInt(3e18).String(), // 3 PELL, (2.5 + 0.5) per block
			rateLimitExceeded:      false,
		},
		{
			name:            "set a lower gRPC request limit and reach the limit of the query in forward loop",
			rateLimitFlags:  createTestRateLimiterFlags(700, math.NewUint(10*1e18)),
			ethMinedXmsgs:   ethMinedXmsgs,
			ethPendingXmsgs: ethPendingXmsgs,
			ethPendingNonces: observertypes.PendingNonces{
				ChainId:   ethChainID,
				NonceLow:  1099,
				NonceHigh: 1199,
				Tss:       tss.TssPubkey,
			},
			bscMinedXmsgs:   bscMinedXmsgs,
			bscPendingXmsgs: bscPendingXmsgs,
			bscPendingNonces: observertypes.PendingNonces{
				ChainId:   bscChainID,
				NonceLow:  1099,
				NonceHigh: 1199,
				Tss:       tss.TssPubkey,
			},

			currentHeight:          1199,
			queryLimit:             400, // 400 < keeper.MaxPendingXmsgs
			expectedXmsgs:          append(append(append([]*types.Xmsg{}, ethPendingXmsgs...))),
			expectedTotalPending:   200,
			expectedWithdrawWindow: 700,                           // the sliding window
			expectedWithdrawRate:   sdkmath.NewInt(3e18).String(), // 3 PELL, (2.5 + 0.5) per block
			rateLimitExceeded:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create test keepers
			k, ctx, _, zk := keepertest.XmsgKeeper(t)

			// Set TSS
			zk.ObserverKeeper.SetTSS(ctx, tss)

			// Set rate limiter flags
			if tt.rateLimitFlags != nil {
				k.SetRateLimiterFlags(ctx, *tt.rateLimitFlags)
			}

			// Set Bsc chain mined cctxs, pending ccxts and pending nonces
			setXmsgsInKeeper(ctx, *k, zk, tss, tt.bscMinedXmsgs)
			setXmsgsInKeeper(ctx, *k, zk, tss, tt.bscPendingXmsgs)
			zk.ObserverKeeper.SetPendingNonces(ctx, tt.bscPendingNonces)

			// Set Eth chain mined cctxs, pending ccxts and pending nonces
			setXmsgsInKeeper(ctx, *k, zk, tss, tt.ethMinedXmsgs)
			setXmsgsInKeeper(ctx, *k, zk, tss, tt.ethPendingXmsgs)
			zk.ObserverKeeper.SetPendingNonces(ctx, tt.ethPendingNonces)

			ctx = ctx.WithBlockHeight(tt.currentHeight)

			// Query pending cctxs
			res, err := k.ListPendingXmsgWithinRateLimit(ctx, &types.QueryListPendingXmsgWithinRateLimitRequest{Limit: tt.queryLimit})
			require.NoError(t, err)
			require.EqualValues(t, tt.expectedXmsgs, res.Xmsgs)
			require.Equal(t, tt.expectedTotalPending, res.TotalPending)

			// check rate limiter related fields only if it's not a fallback query
			if !tt.fallback {
				require.Equal(t, tt.expectedWithdrawWindow, res.CurrentWithdrawWindow)
				// require.Equal(t, tt.expectedWithdrawRate, res.CurrentWithdrawRate)
				require.Equal(t, tt.rateLimitExceeded, res.RateLimitExceeded)
			}
		})
	}
}

func TestKeeper_ListPendingXmsgWithinRateLimit_Errors(t *testing.T) {
	t.Run("should fail for empty req", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)
		_, err := k.ListPendingXmsgWithinRateLimit(ctx, nil)
		require.ErrorContains(t, err, "invalid request")
	})
	t.Run("height out of range", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)

		// Set rate limiter flags as disabled
		rFlags := sample.RateLimiterFlags_pell()
		k.SetRateLimiterFlags(ctx, rFlags)

		ctx = ctx.WithBlockHeight(0)
		_, err := k.ListPendingXmsgWithinRateLimit(ctx, &types.QueryListPendingXmsgWithinRateLimitRequest{})
		require.ErrorContains(t, err, "height out of range")
	})
	t.Run("tss not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeper(t)

		// Set rate limiter flags as disabled
		rFlags := sample.RateLimiterFlags_pell()
		k.SetRateLimiterFlags(ctx, rFlags)

		_, err := k.ListPendingXmsgWithinRateLimit(ctx, &types.QueryListPendingXmsgWithinRateLimitRequest{})
		require.ErrorContains(t, err, "tss not found")
	})
	t.Run("pending nonces not found", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)

		// Set rate limiter flags as disabled
		rFlags := sample.RateLimiterFlags_pell()
		k.SetRateLimiterFlags(ctx, rFlags)

		// Set TSS
		tss := sample.Tss_pell()
		zk.ObserverKeeper.SetTSS(ctx, tss)

		_, err := k.ListPendingXmsgWithinRateLimit(ctx, &types.QueryListPendingXmsgWithinRateLimitRequest{})
		require.ErrorContains(t, err, "pending nonces not found")
	})
}
