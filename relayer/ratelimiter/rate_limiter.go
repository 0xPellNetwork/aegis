// Package ratelimiter provides functionalities for rate limiting the cross-chain transactions
package ratelimiter

import (
	sdkmath "cosmossdk.io/math"

	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

// Input is the input data for the rate limiter
type Input struct {
	// pell chain height
	Height int64

	// the missed cctxs in range [?, NonceLow) across all chains
	XmsgsMissed []*xmsgtypes.Xmsg

	// the pending cctxs in range [NonceLow, NonceHigh) across all chains
	XmsgsPending []*xmsgtypes.Xmsg

	// the total value of the past cctxs within window across all chains
	PastXmsgsValue sdkmath.Int

	// the total value of the pending cctxs across all chains
	PendingXmsgsValue sdkmath.Int

	// the lowest height of the pending (not missed) cctxs across all chains
	LowestPendingXmsgHeight int64
}

// Output is the output data for the rate limiter
type Output struct {
	// the cctxs to be scheduled after rate limit check
	XmsgsMap map[int64][]*xmsgtypes.Xmsg

	// the current sliding window within which the withdrawals are considered by the rate limiter
	CurrentWithdrawWindow int64

	// the current withdraw rate (apell/block) within the current sliding window
	CurrentWithdrawRate sdkmath.Int

	// wehther the current withdraw rate exceeds the given rate limit or not
	RateLimitExceeded bool
}

// NewInput creates a rate limiter input from gRPC response
func NewInput(resp xmsgtypes.QueryRateLimiterInputResponse) (*Input, bool) {
	return &Input{
		Height:                  resp.Height,
		XmsgsMissed:             resp.XmsgsMissed,
		XmsgsPending:            resp.XmsgsPending,
		PastXmsgsValue:          sdkmath.ZeroInt(),
		PendingXmsgsValue:       sdkmath.ZeroInt(),
		LowestPendingXmsgHeight: resp.LowestPendingXmsgHeight,
	}, true
}

// IsRateLimiterUsable checks if the rate limiter is usable or not
func IsRateLimiterUsable(rateLimiterFlags xmsgtypes.RateLimiterFlags) bool {
	if !rateLimiterFlags.Enabled {
		return false
	}
	if rateLimiterFlags.Window <= 0 {
		return false
	}
	if rateLimiterFlags.Rate.IsNil() {
		return false
	}
	if rateLimiterFlags.Rate.IsZero() {
		return false
	}
	return true
}

// ApplyRateLimiter applies the rate limiter to the input and produces output
func ApplyRateLimiter(input *Input, window int64, rate sdkmath.Uint) *Output {
	// block limit and the window limit in apell
	blockLimitInApell := sdkmath.NewIntFromBigInt(rate.BigInt())
	windowLimitInApell := blockLimitInApell.Mul(sdkmath.NewInt(window))

	// invariant: for period of time >= `window`, the pellclient-side average withdraw rate should be <= `blockLimitInPell`
	// otherwise, pellclient should wait for the average rate to drop below `blockLimitInPell`
	withdrawWindow := window
	withdrawLimitInApell := windowLimitInApell
	if input.LowestPendingXmsgHeight != 0 {
		// If [input.LowestPendingXmsgHeight, input.Height] is wider than the given `window`, we should:
		// 1. use the wider window to calculate the average withdraw rate
		// 2. adjust the limit proportionally to fit the wider window
		pendingXmsgWindow := input.Height - input.LowestPendingXmsgHeight + 1
		if pendingXmsgWindow > window {
			withdrawWindow = pendingXmsgWindow
			withdrawLimitInApell = blockLimitInApell.Mul(sdkmath.NewInt(pendingXmsgWindow))
		}
	}

	// limit exceeded or not
	totalWithdrawInApell := input.PastXmsgsValue.Add(input.PendingXmsgsValue)
	limitExceeded := totalWithdrawInApell.GT(withdrawLimitInApell)

	// define the result cctx map to be scheduled
	cctxMap := make(map[int64][]*xmsgtypes.Xmsg)

	// addXmsgsToMap adds the given cctxs to the cctx map
	addXmsgsToMap := func(cctxs []*xmsgtypes.Xmsg) {
		for _, cctx := range cctxs {
			chainID := cctx.GetCurrentOutTxParam().ReceiverChainId
			if _, found := cctxMap[chainID]; !found {
				cctxMap[chainID] = make([]*xmsgtypes.Xmsg, 0)
			}
			cctxMap[chainID] = append(cctxMap[chainID], cctx)
		}
	}

	// schedule missed cctxs regardless of the `limitExceeded` flag
	addXmsgsToMap(input.XmsgsMissed)

	// schedule pending cctxs only if `limitExceeded == false`
	if !limitExceeded {
		addXmsgsToMap(input.XmsgsPending)
	}

	return &Output{
		XmsgsMap:              cctxMap,
		CurrentWithdrawWindow: withdrawWindow,
		CurrentWithdrawRate:   totalWithdrawInApell.Quo(sdkmath.NewInt(withdrawWindow)),
		RateLimitExceeded:     limitExceeded,
	}
}
