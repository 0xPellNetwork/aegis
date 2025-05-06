package keeper

import (
	"context"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	observertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// RateLimiterInput collects the input data for the rate limiter
func (k Keeper) RateLimiterInput(
	c context.Context,
	req *types.QueryRateLimiterInputRequest,
) (res *types.QueryRateLimiterInputResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Window <= 0 {
		return nil, status.Error(codes.InvalidArgument, "window must be positive")
	}

	// use default MaxPendingXmsgs if not specified or too high
	limit := req.Limit
	if limit == 0 || limit > MaxPendingXmsgs {
		limit = MaxPendingXmsgs
	}
	ctx := sdk.UnwrapSDKContext(c)

	// get current height and tss
	height := ctx.BlockHeight()
	if height <= 0 {
		return nil, status.Error(codes.OutOfRange, "height out of range")
	}
	tss, found := k.relayerKeeper.GetTSS(ctx)
	if !found {
		return nil, observertypes.ErrTssNotFound
	}

	// calculate the rate limiter sliding window left boundary (inclusive)
	leftWindowBoundary := height - req.Window + 1
	if leftWindowBoundary < 1 {
		leftWindowBoundary = 1
	}

	// the `limit` of pending result is reached or not
	maxCCTXsReached := func(cctxs []*types.Xmsg) bool {
		// #nosec G115 len always positive
		return uint32(len(cctxs)) > limit
	}

	// if a cctx falls within the rate limiter window
	isCCTXInWindow := func(cctx *types.Xmsg) bool {
		// #nosec G115 checked positive
		return cctx.InboundTxParams.InboundTxBlockHeight >= uint64(leftWindowBoundary)
	}

	// if a cctx is an outgoing cctx that orginates from PellChain
	// reverted incoming cctx has an external `SenderChainId` and should not be counted
	isCCTXOutgoing := func(cctx *types.Xmsg) bool {
		return chains.IsPellChain(cctx.InboundTxParams.SenderChainId)
	}

	// it is a past cctx if its nonce < `nonceLow`,
	isPastXmsg := func(cctx *types.Xmsg, nonceLow int64) bool {
		// #nosec G115 always positive
		return cctx.GetCurrentOutTxParam().OutboundTxTssNonce < uint64(nonceLow)
	}

	// get foreign chains and conversion rates of foreign coins
	externalSupportedChains := chains.FilterChains(
		k.GetRelayerKeeper().GetSupportedChains(ctx),
		chains.FilterExternalChains,
	)

	// query pending nonces of each foreign chain and get the lowest height of the pending cctxs
	lowestPendingXmsgHeight := int64(0)
	pendingNoncesMap := make(map[int64]observertypes.PendingNonces)
	for _, chain := range externalSupportedChains {
		pendingNonces, found := k.GetRelayerKeeper().GetPendingNonces(ctx, tss.TssPubkey, chain.Id)
		if !found {
			return nil, status.Error(codes.Internal, "pending nonces not found")
		}
		pendingNoncesMap[chain.Id] = pendingNonces

		// update lowest pending cctx height
		if pendingNonces.NonceLow < pendingNonces.NonceHigh {
			cctx, err := getXmsgByChainIDAndNonce(k, ctx, tss.TssPubkey, chain.Id, pendingNonces.NonceLow)
			if err != nil {
				return nil, err
			}
			// #nosec G115 len always in range
			cctxHeight := int64(cctx.InboundTxParams.InboundTxBlockHeight)
			if lowestPendingXmsgHeight == 0 || cctxHeight < lowestPendingXmsgHeight {
				lowestPendingXmsgHeight = cctxHeight
			}
		}
	}

	// define a few variables to be used in the query loops
	totalPending := uint64(0)
	cctxsMissed := make([]*types.Xmsg, 0)
	cctxsPending := make([]*types.Xmsg, 0)

	// query backwards for pending cctxs of each foreign chain
	for _, chain := range externalSupportedChains {
		// we should at least query 1000 prior to find any pending cctx that we might have missed
		// this logic is needed because a confirmation of higher nonce will automatically update the p.NonceLow
		// therefore might mask some lower nonce cctx that is still pending.
		pendingNonces := pendingNoncesMap[chain.Id]
		startNonce := pendingNonces.NonceHigh - 1
		endNonce := pendingNonces.NonceLow - MaxLookbackNonce
		if endNonce < 0 {
			endNonce = 0
		}

		// go all the way back to the left window boundary or `NonceLow - 1000`, depending on which on arrives first
		for nonce := startNonce; nonce >= 0; nonce-- {
			cctx, err := getXmsgByChainIDAndNonce(k, ctx, tss.TssPubkey, chain.Id, nonce)
			if err != nil {
				return nil, err
			}
			inWindow := isCCTXInWindow(cctx)
			isOutgoing := isCCTXOutgoing(cctx)
			isPast := isPastXmsg(cctx, pendingNonces.NonceLow)

			// we should at least go backwards by 1000 nonces to pick up missed pending cctxs
			// we might go even further back if the endNonce hasn't hit the left window boundary yet
			if nonce < endNonce && !inWindow {
				break
			}

			// sum up the cctxs' value if the cctx is outgoing, within the window and in the past
			if inWindow && isOutgoing && isPast {
			}

			// add cctx to corresponding list
			if IsPending(cctx) {
				totalPending++
				if isPast {
					cctxsMissed = append(cctxsMissed, cctx)
				} else {
					cctxsPending = append(cctxsPending, cctx)
				}
			}
		}
	}

	// sort the missed cctxs order by height (can sort by other criteria, for unit testability)
	SortXmsgsByHeightAndChainID(cctxsMissed)

	// sort the pending cctxs order by height (first come first serve)
	SortXmsgsByHeightAndChainID(cctxsPending)

	// we take all the missed cctxs (won't be a lot) for simplicity of the query, but we only take a `limit` number of pending cctxs
	if maxCCTXsReached(cctxsPending) {
		cctxsPending = cctxsPending[:limit]
	}

	return &types.QueryRateLimiterInputResponse{
		Height:                  height,
		XmsgsMissed:             cctxsMissed,
		XmsgsPending:            cctxsPending,
		TotalPending:            totalPending,
		LowestPendingXmsgHeight: lowestPendingXmsgHeight,
	}, nil
}

// ListPendingXmsgWithinRateLimit returns a list of pending cctxs that do not exceed the outbound rate limit
// a limit for the number of cctxs to return can be specified or the default is MaxPendingXmsgs
func (k Keeper) ListPendingXmsgWithinRateLimit(c context.Context, req *types.QueryListPendingXmsgWithinRateLimitRequest) (res *types.QueryListPendingXmsgWithinRateLimitResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// use default MaxPendingXmsgs if not specified or too high
	limit := req.Limit
	if limit == 0 || limit > MaxPendingXmsgs {
		limit = MaxPendingXmsgs
	}
	ctx := sdk.UnwrapSDKContext(c)

	// define a few variables to be used in the query loops
	limitExceeded := false
	totalPending := uint64(0)
	totalWithdrawInApell := sdkmath.NewInt(0)
	xmsgs := make([]*types.Xmsg, 0)
	foreignChains := k.relayerKeeper.GetSupportedForeignChains(ctx)

	// check rate limit flags to decide if we should apply rate limit
	applyLimit := true
	rateLimitFlags, found := k.GetRateLimiterFlags(ctx)
	if !found || !rateLimitFlags.Enabled {
		applyLimit = false
	}
	if rateLimitFlags.Rate.IsNil() || rateLimitFlags.Rate.IsZero() {
		applyLimit = false
	}

	// fallback to non-rate-limited query if rate limiter is disabled
	if !applyLimit {
		for _, chain := range foreignChains {
			resp, err := k.ListPendingXmsg(ctx, &types.QueryListPendingXmsgRequest{ChainId: chain.Id, Limit: limit})
			if err == nil {
				xmsgs = append(xmsgs, resp.Xmsg...)
				totalPending += resp.TotalPending
			}
		}
		return &types.QueryListPendingXmsgWithinRateLimitResponse{
			Xmsgs:             xmsgs,
			TotalPending:      totalPending,
			RateLimitExceeded: false,
		}, nil
	}

	// get current height and tss
	height := ctx.BlockHeight()
	if height <= 0 {
		return nil, status.Error(codes.OutOfRange, "height out of range")
	}
	tss, found := k.relayerKeeper.GetTSS(ctx)
	if !found {
		return nil, observertypes.ErrTssNotFound
	}

	// calculate the rate limiter sliding window left boundary (inclusive)
	leftWindowBoundary := height - rateLimitFlags.Window + 1
	if leftWindowBoundary < 0 {
		leftWindowBoundary = 0
	}

	// get the conversion rates for all foreign coins
	var blockLimitInApell sdkmath.Int
	var windowLimitInApell sdkmath.Int
	if applyLimit {
		// initiate block limit and window limit in apell
		blockLimitInApell = sdkmath.NewIntFromBigInt(rateLimitFlags.Rate.BigInt())
		windowLimitInApell = blockLimitInApell.Mul(sdkmath.NewInt(rateLimitFlags.Window))
	}

	// the criteria to stop adding xmsgs to the rpc response
	maxXmsgsReached := func(xmsgs []*types.Xmsg) bool {
		// #nosec G701 len always positive
		return uint32(len(xmsgs)) >= limit
	}

	// if a xmsg falls within the rate limiter window
	isXmsgInWindow := func(xmsg *types.Xmsg) bool {
		// #nosec G701 checked positive
		return xmsg.InboundTxParams.InboundTxBlockHeight >= uint64(leftWindowBoundary)
	}

	// if a xmsg is an outgoing xmsg that orginates from PellChain
	// reverted incoming xmsg has an external `SenderChainId` and should not be counted
	isXmsgOutgoing := func(xmsg *types.Xmsg) bool {
		return chains.IsPellChain(xmsg.InboundTxParams.SenderChainId)
	}

	// query pending nonces for each foreign chain and get the lowest height of the pending xmsgs
	lowestPendingXmsgHeight := int64(0)
	pendingNoncesMap := make(map[int64]observertypes.PendingNonces)
	for _, chain := range foreignChains {
		pendingNonces, found := k.GetRelayerKeeper().GetPendingNonces(ctx, tss.TssPubkey, chain.Id)
		if !found {
			return nil, status.Errorf(codes.Internal, "pending nonces not found: chain %d", chain.Id)
		}
		pendingNoncesMap[chain.Id] = pendingNonces

		// insert pending nonces and update lowest height
		if pendingNonces.NonceLow < pendingNonces.NonceHigh {
			xmsg, err := getXmsgByChainIDAndNonce(k, ctx, tss.TssPubkey, chain.Id, pendingNonces.NonceLow)
			if err != nil {
				return nil, err
			}
			// #nosec G701 len always in range
			xmsgHeight := int64(xmsg.InboundTxParams.InboundTxBlockHeight)
			if lowestPendingXmsgHeight == 0 || xmsgHeight < lowestPendingXmsgHeight {
				lowestPendingXmsgHeight = xmsgHeight
			}
		}
	}
	if len(pendingNoncesMap) == 0 {
		return nil, status.Error(codes.Internal, "pending nonces not found")
	}

	// invariant: for period of time >= `rateLimitFlags.Window`, the pellclient-side average withdraw rate should be <= `blockLimitInPell`
	// otherwise, this query should return empty result and wait for the average rate to drop below `blockLimitInPell`
	withdrawWindow := rateLimitFlags.Window
	withdrawLimitInApell := windowLimitInApell
	if lowestPendingXmsgHeight != 0 {
		// `pendingXmsgWindow` is the width of [lowestPendingXmsgHeight, height] window
		// if the window can be wider than `rateLimitFlags.Window`, we should adjust the total withdraw limit proportionally
		pendingXmsgWindow := height - lowestPendingXmsgHeight + 1
		if pendingXmsgWindow > rateLimitFlags.Window {
			withdrawWindow = pendingXmsgWindow
			withdrawLimitInApell = blockLimitInApell.Mul(sdkmath.NewInt(pendingXmsgWindow))
		}
	}

	// query backwards for potential missed pending xmsgs for each foreign chain
	for _, chain := range foreignChains {
		// we should at least query 1000 prior to find any pending xmsg that we might have missed
		// this logic is needed because a confirmation of higher nonce will automatically update the p.NonceLow
		// therefore might mask some lower nonce xmsg that is still pending.
		pendingNonces := pendingNoncesMap[chain.Id]
		startNonce := pendingNonces.NonceLow - 1
		endNonce := pendingNonces.NonceLow - MaxLookbackNonce
		if endNonce < 0 {
			endNonce = 0
		}

		// query xmsg by nonce backwards to the left boundary of the rate limit sliding window
		for nonce := startNonce; nonce >= 0; nonce-- {
			xmsg, err := getXmsgByChainIDAndNonce(k, ctx, tss.TssPubkey, chain.Id, nonce)
			if err != nil {
				return nil, err
			}
			inWindow := isXmsgInWindow(xmsg)
			isOutgoing := isXmsgOutgoing(xmsg)

			// we should at least go backwards by 1000 nonces to pick up missed pending xmsgs
			// we might go even further back if rate limiter is enabled and the endNonce hasn't hit the left window boundary yet
			// stop at the left window boundary if the `endNonce` hasn't hit it yet
			if nonce < endNonce && !inWindow {
				break
			}
			// sum up the xmsgs' value if the xmsg is outgoing and within the window
			if inWindow && isOutgoing &&
				rateLimitExceeded(
					chain.Id,
					xmsg,
					&totalWithdrawInApell,
					withdrawLimitInApell,
				) {
				limitExceeded = true
				continue
			}

			// only take a `limit` number of pending xmsgs as result but still count the total pending xmsgs
			if IsPending(xmsg) {
				totalPending++
				if !maxXmsgsReached(xmsgs) {
					xmsgs = append(xmsgs, xmsg)
				}
			}
		}
	}

	// remember the number of missed pending xmsgs
	missedPending := len(xmsgs)

	// query forwards for pending xmsgs for each foreign chain
	for _, chain := range foreignChains {
		pendingNonces := pendingNoncesMap[chain.Id]

		// #nosec G701 always in range
		totalPending += uint64(pendingNonces.NonceHigh - pendingNonces.NonceLow)

		// query the pending xmsgs in range [NonceLow, NonceHigh)
		for nonce := pendingNonces.NonceLow; nonce < pendingNonces.NonceHigh; nonce++ {
			xmsg, err := getXmsgByChainIDAndNonce(k, ctx, tss.TssPubkey, chain.Id, nonce)
			if err != nil {
				return nil, err
			}
			isOutgoing := isXmsgOutgoing(xmsg)

			// skip the xmsg if rate limit is exceeded but still accumulate the total withdraw value
			if isOutgoing && rateLimitExceeded(
				chain.Id,
				xmsg,
				&totalWithdrawInApell,
				withdrawLimitInApell,
			) {
				limitExceeded = true
				continue
			}
			// only take a `limit` number of pending xmsgs as result
			if maxXmsgsReached(xmsgs) {
				continue
			}
			xmsgs = append(xmsgs, xmsg)
		}
	}

	// if the rate limit is exceeded, only return the missed pending xmsgs
	if limitExceeded {
		xmsgs = xmsgs[:missedPending]
	}

	// sort the xmsgs by chain ID and nonce (lower nonce holds higher priority for scheduling)
	sort.SliceStable(xmsgs, func(i, j int) bool {
		if xmsgs[i].GetCurrentOutTxParam().ReceiverChainId == xmsgs[j].GetCurrentOutTxParam().ReceiverChainId {
			return xmsgs[i].GetCurrentOutTxParam().OutboundTxTssNonce < xmsgs[j].GetCurrentOutTxParam().OutboundTxTssNonce
		}
		return xmsgs[i].GetCurrentOutTxParam().ReceiverChainId < xmsgs[j].GetCurrentOutTxParam().ReceiverChainId
	})

	return &types.QueryListPendingXmsgWithinRateLimitResponse{
		Xmsgs:                 xmsgs,
		TotalPending:          totalPending,
		CurrentWithdrawWindow: withdrawWindow,
		CurrentWithdrawRate:   totalWithdrawInApell.Quo(sdkmath.NewInt(withdrawWindow)).String(),
		RateLimitExceeded:     limitExceeded,
	}, nil
}

// rateLimitExceeded accumulates the xmsg value and then checks if the rate limit is exceeded
// returns true if the rate limit is exceeded
func rateLimitExceeded(
	_ int64, // chainID
	_ *types.Xmsg, // xmsg
	currentXmsgValue *sdkmath.Int,
	withdrawLimitInApell sdkmath.Int,
) bool {
	*currentXmsgValue = currentXmsgValue.Add(sdkmath.NewInt(1))
	return currentXmsgValue.GT(withdrawLimitInApell)
}
