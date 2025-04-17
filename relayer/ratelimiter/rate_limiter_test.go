package ratelimiter_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/relayer/ratelimiter"
	"github.com/pell-chain/pellcore/testutil/sample"
	xmsgkeeper "github.com/pell-chain/pellcore/x/xmsg/keeper"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

func Test_NewInput(t *testing.T) {
	// sample response
	response := xmsgtypes.QueryRateLimiterInputResponse{
		Height:                  10,
		XmsgsMissed:             []*xmsgtypes.Xmsg{sample.Xmsg_pell(t, "1-1")},
		XmsgsPending:            []*xmsgtypes.Xmsg{sample.Xmsg_pell(t, "1-2")},
		TotalPending:            7,
		LowestPendingXmsgHeight: 2,
	}

	t.Run("should create a input from gRPC response", func(t *testing.T) {
		filterInput, ok := ratelimiter.NewInput(response)
		require.True(t, ok)
		require.Equal(t, response.Height, filterInput.Height)
		require.Equal(t, response.XmsgsMissed, filterInput.XmsgsMissed)
		require.Equal(t, response.XmsgsPending, filterInput.XmsgsPending)
		require.Equal(t, response.LowestPendingXmsgHeight, filterInput.LowestPendingXmsgHeight)
	})
}

func Test_IsRateLimiterUsable(t *testing.T) {
	tests := []struct {
		name     string
		flags    xmsgtypes.RateLimiterFlags
		expected bool
	}{
		{
			name: "rate limiter is enabled",
			flags: xmsgtypes.RateLimiterFlags{
				Enabled: true,
				Window:  100,
				Rate:    sdkmath.NewUint(1e18), // 1 PELL/block
			},
			expected: true,
		},
		{
			name: "rate limiter is disabled",
			flags: xmsgtypes.RateLimiterFlags{
				Enabled: false,
			},
			expected: false,
		},
		{
			name: "rate limiter is enabled with 0 window",
			flags: xmsgtypes.RateLimiterFlags{
				Enabled: true,
				Window:  0,
			},
			expected: false,
		},
		{
			name: "rate limiter is enabled with nil rate",
			flags: xmsgtypes.RateLimiterFlags{
				Enabled: true,
				Window:  100,
				Rate:    sdkmath.Uint{},
			},
			expected: false,
		},
		{
			name: "rate limiter is enabled with zero rate",
			flags: xmsgtypes.RateLimiterFlags{
				Enabled: true,
				Window:  100,
				Rate:    sdkmath.NewUint(0),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usable := ratelimiter.IsRateLimiterUsable(tt.flags)
			require.Equal(t, tt.expected, usable)
		})
	}
}

func Test_ApplyRateLimiter(t *testing.T) {
	// define test chain ids
	ethChainID := chains.EthChain().Id
	pellChainID := chains.PellChainMainnet().Id

	// create 10 missed and 90 pending cctxs for eth chain, the coinType/amount does not matter for this test
	// but we still use a proper cctx value (0.5 PELL) to make the test more realistic
	ethXmsgsMissed := sample.CustomXmsgsInBlockRange(
		t,
		1,
		10,
		pellChainID,
		ethChainID,
		xmsgtypes.XmsgStatus_PENDING_OUTBOUND,
	)
	ethXmsgsPending := sample.CustomXmsgsInBlockRange(
		t,
		11,
		100,
		pellChainID,
		ethChainID,
		xmsgtypes.XmsgStatus_PENDING_OUTBOUND,
	)
	ethXmsgsAll := append(append([]*xmsgtypes.Xmsg{}, ethXmsgsMissed...), ethXmsgsPending...)

	// all missed cctxs and all pending cctxs across all chains
	allXmsgsMissed := xmsgkeeper.SortXmsgsByHeightAndChainID(
		append(append([]*xmsgtypes.Xmsg{}, ethXmsgsMissed...)))
	allXmsgsPending := xmsgkeeper.SortXmsgsByHeightAndChainID(
		append(append([]*xmsgtypes.Xmsg{}, ethXmsgsPending...)))

	// define test cases
	tests := []struct {
		name   string
		window int64
		rate   sdkmath.Uint
		input  ratelimiter.Input
		output ratelimiter.Output
	}{
		{
			name:   "should return all missed and pending cctxs",
			window: 100,
			rate:   sdkmath.NewUint(1e18), // 1 PELL/block
			input: ratelimiter.Input{
				Height:                  100,
				XmsgsMissed:             allXmsgsMissed,
				XmsgsPending:            allXmsgsPending,
				PastXmsgsValue:          sdkmath.NewInt(10).Mul(sdkmath.NewInt(1e18)), // 10 * 1 PELL
				PendingXmsgsValue:       sdkmath.NewInt(90).Mul(sdkmath.NewInt(1e18)), // 90 * 1 PELL
				LowestPendingXmsgHeight: 11,
			},
			output: ratelimiter.Output{
				XmsgsMap: map[int64][]*xmsgtypes.Xmsg{
					ethChainID: ethXmsgsAll,
				},
				CurrentWithdrawWindow: 100,                  // height [1, 100]
				CurrentWithdrawRate:   sdkmath.NewInt(1e18), // (10 + 90) / 100
				RateLimitExceeded:     false,
			},
		},
		{
			name:   "should monitor a wider window and adjust the total limit",
			window: 50,
			rate:   sdkmath.NewUint(1e18), // 1 PELL/block
			input: ratelimiter.Input{
				Height:                  100,
				XmsgsMissed:             allXmsgsMissed,
				XmsgsPending:            allXmsgsPending,
				PastXmsgsValue:          sdkmath.NewInt(0),                            // no past cctx in height range [51, 100]
				PendingXmsgsValue:       sdkmath.NewInt(90).Mul(sdkmath.NewInt(1e18)), // 90 * 1 PELL
				LowestPendingXmsgHeight: 11,
			},
			output: ratelimiter.Output{
				XmsgsMap: map[int64][]*xmsgtypes.Xmsg{
					ethChainID: ethXmsgsAll,
				},
				CurrentWithdrawWindow: 90,                   // [LowestPendingXmsgHeight, Height] = [11, 100]
				CurrentWithdrawRate:   sdkmath.NewInt(1e18), // 90 / 90 = 1 PELL/block
				RateLimitExceeded:     false,
			},
		},
		{
			name:   "rate limit is exceeded in given sliding window 100",
			window: 100,
			rate:   sdkmath.NewUint(1e18), // 1 PELL/block
			input: ratelimiter.Input{
				Height:                  100,
				XmsgsMissed:             allXmsgsMissed,
				XmsgsPending:            allXmsgsPending,
				PastXmsgsValue:          sdkmath.NewInt(11).Mul(sdkmath.NewInt(1e18)), // 11 PELL, increased value by 1 PELL
				PendingXmsgsValue:       sdkmath.NewInt(90).Mul(sdkmath.NewInt(1e18)), // 90 * 1 PELL
				LowestPendingXmsgHeight: 11,
			},
			output: ratelimiter.Output{ // should return missed cctxs only
				XmsgsMap: map[int64][]*xmsgtypes.Xmsg{
					ethChainID: ethXmsgsMissed,
				},
				CurrentWithdrawWindow: 100, // height [1, 100]
				CurrentWithdrawRate: sdkmath.NewInt(
					101e16,
				), // (11 + 90) / 100 = 1.01 PELL/block (exceeds 0.99 PELL/block)
				RateLimitExceeded: true,
			},
		},
		{
			name:   "rate limit is exceeded in wider window then the given sliding window 50",
			window: 50,
			rate:   sdkmath.NewUint(1e18), // 1 PELL/block
			input: ratelimiter.Input{
				Height:                  100,
				XmsgsMissed:             allXmsgsMissed,
				XmsgsPending:            allXmsgsPending,
				PastXmsgsValue:          sdkmath.NewInt(0),                            // no past cctx in height range [51, 100]
				PendingXmsgsValue:       sdkmath.NewInt(91).Mul(sdkmath.NewInt(1e18)), // 91 PELL, increased value by 1 PELL
				LowestPendingXmsgHeight: 11,
			},
			output: ratelimiter.Output{
				XmsgsMap: map[int64][]*xmsgtypes.Xmsg{
					ethChainID: ethXmsgsMissed,
				},
				CurrentWithdrawWindow: 90, // [LowestPendingXmsgHeight, Height] = [11, 100]
				CurrentWithdrawRate: sdkmath.NewInt(91).
					Mul(sdkmath.NewInt(1e18)).
					Quo(sdkmath.NewInt(90)),
				// 91 / 90 = 1.011111111111111111 PELL/block
				RateLimitExceeded: true,
			},
		},
		{
			name:   "should not exceed rate limit if we wait for 1 more block",
			window: 50,
			rate:   sdkmath.NewUint(1e18), // 1 PELL/block
			input: ratelimiter.Input{
				Height:                  101,
				XmsgsMissed:             allXmsgsMissed,
				XmsgsPending:            allXmsgsPending,
				PastXmsgsValue:          sdkmath.NewInt(0),                            // no past cctx in height range [52, 101]
				PendingXmsgsValue:       sdkmath.NewInt(91).Mul(sdkmath.NewInt(1e18)), // 91 PELL, increased value by 1 PELL
				LowestPendingXmsgHeight: 11,
			},
			output: ratelimiter.Output{
				XmsgsMap: map[int64][]*xmsgtypes.Xmsg{
					ethChainID: ethXmsgsAll,
				},
				CurrentWithdrawWindow: 91, // [LowestPendingXmsgHeight, Height] = [11, 101]
				CurrentWithdrawRate: sdkmath.NewInt(91).
					Mul(sdkmath.NewInt(1e18)).
					Quo(sdkmath.NewInt(91)),
				// 91 / 91 = 1.011 PELL/block
				RateLimitExceeded: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := ratelimiter.ApplyRateLimiter(&tt.input, tt.window, tt.rate)
			require.Equal(t, tt.output.XmsgsMap, output.XmsgsMap)
			require.Equal(t, tt.output.CurrentWithdrawWindow, output.CurrentWithdrawWindow)
			require.Equal(t, tt.output.CurrentWithdrawRate, output.CurrentWithdrawRate)
			require.Equal(t, tt.output.RateLimitExceeded, output.RateLimitExceeded)
		})
	}
}
