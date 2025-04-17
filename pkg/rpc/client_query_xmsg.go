package rpc

import (
	"context"
	"sort"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// 32MB
var maxSizeOption = grpc.MaxCallRecvMsgSize(32 * 1024 * 1024)

// GetLastBlockHeight returns the pellchain block height
func (c *Clients) GetLastBlockHeight(ctx context.Context) ([]*types.LastBlockHeight, error) {
	resp, err := c.Xmsg.LastBlockHeightAll(ctx, &types.QueryAllLastBlockHeightRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get block height")
	}
	return resp.LastBlockHeight, nil
}

// GetBlockHeight returns the pellchain block height
func (c *Clients) GetBlockHeight(ctx context.Context) (int64, error) {
	height, err := c.Xmsg.LastPellHeight(
		ctx,
		&types.QueryLastPellHeightRequest{},
	)
	if err != nil {
		return 0, err
	}

	return height.Height, nil
}

// GetRateLimiterFlags returns the rate limiter flags
func (c *Clients) GetRateLimiterFlags(ctx context.Context) (types.RateLimiterFlags, error) {
	resp, err := c.Xmsg.RateLimiterFlags(ctx, &types.QueryRateLimiterFlagsRequest{})
	if err != nil {
		return types.RateLimiterFlags{}, errors.Wrap(err, "failed to get rate limiter flags")
	}

	return resp.RateLimiterFlags, nil
}

// GetRateLimiterInput returns input data for the rate limit checker
func (c *Clients) GetRateLimiterInput(ctx context.Context, window int64) (*types.QueryRateLimiterInputResponse, error) {
	in := &types.QueryRateLimiterInputRequest{Window: window}

	resp, err := c.Xmsg.RateLimiterInput(ctx, in, maxSizeOption)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rate limiter input")
	}

	return resp, nil
}

// GetAllCctx returns all cross chain transactions
func (c *Clients) GetAllXmsg(ctx context.Context) ([]*types.Xmsg, error) {
	resp, err := c.Xmsg.XmsgAll(ctx, &types.QueryAllXmsgRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all cross chain transactions")
	}
	return resp.Xmsgs, nil
}

func (c *Clients) GetXmsgByHash(ctx context.Context, sendHash string) (*types.Xmsg, error) {
	resp, err := c.Xmsg.Xmsg(ctx, &types.QueryGetXmsgRequest{Index: sendHash})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cctx by hash")
	}
	return resp.Xmsg, nil
}

func (c *Clients) GetXmsgByNonce(ctx context.Context, chainID int64, nonce uint64) (*types.Xmsg, error) {
	resp, err := c.Xmsg.XmsgByNonce(ctx, &types.QueryGetXmsgByNonceRequest{
		ChainId: chainID,
		Nonce:   nonce,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cctx by nonce")
	}
	return resp.Xmsg, nil
}

// ListPendingCCTXWithinRateLimit returns a list of pending cctxs that do not exceed the outbound rate limit
//   - The max size of the list is crosschainkeeper.MaxPendingXmsgs
//   - The returned `rateLimitExceeded` flag indicates if the rate limit is exceeded or not
func (c *Clients) ListPendingXmsgWithinRatelimit(ctx context.Context) (*types.QueryListPendingXmsgWithinRateLimitResponse, error) {

	resp, err := c.Xmsg.ListPendingXmsgWithinRateLimit(
		ctx,
		&types.QueryListPendingXmsgWithinRateLimitRequest{},
		maxSizeOption,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pending cctxs within rate limit")
	}
	return resp, nil
}

// ListPendingXmsg returns a list of pending cctxs for a given chainID
//   - The max size of the list is xmsgkeeper.MaxPendingXmsgs
func (c *Clients) ListPendingXmsg(ctx context.Context, chainID int64) ([]*types.Xmsg, uint64, error) {

	resp, err := c.Xmsg.ListPendingXmsg(
		ctx,
		&types.QueryListPendingXmsgRequest{ChainId: chainID},
		maxSizeOption,
	)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to get pending cctxs")
	}
	return resp.Xmsg, resp.TotalPending, nil
}

// GetOutboundTracker returns the outbound tracker for a chain and nonce
func (c *Clients) GetOutTxTracker(ctx context.Context, chain chains.Chain, nonce uint64) (*types.OutTxTracker, error) {
	resp, err := c.Xmsg.OutTxTracker(ctx, &types.QueryGetOutTxTrackerRequest{
		ChainId: chain.Id,
		Nonce:   nonce,
	})
	if err != nil {
		return nil, err
	}
	return &resp.OutTxTracker, nil
}

// GetInboundTrackersForChain returns the inbound trackers for a chain
func (c *Clients) GetInboundTrackersForChain(ctx context.Context, chainID int64) ([]types.InTxTracker, error) {
	resp, err := c.Xmsg.InTxTrackerAllByChain(ctx, &types.QueryAllInTxTrackerByChainRequest{ChainId: chainID})
	if err != nil {
		return nil, err
	}
	return resp.InTxTrackers, nil
}

// GetAllOutboundTrackerByChain returns all outbound trackers for a chain
func (c *Clients) GetAllOutTxTrackerByChain(ctx context.Context, chainID int64, order interfaces.Order) ([]types.OutTxTracker, error) {
	resp, err := c.Xmsg.OutTxTrackerAllByChain(ctx, &types.QueryAllOutTxTrackerByChainRequest{
		Chain: chainID,
		Pagination: &query.PageRequest{
			Key:        nil,
			Offset:     0,
			Limit:      2000,
			CountTotal: false,
			Reverse:    false,
		},
	})
	if err != nil {
		return nil, err
	}
	if order == interfaces.Ascending {
		sort.SliceStable(resp.OutTxTrackers, func(i, j int) bool {
			return resp.OutTxTrackers[i].Nonce < resp.OutTxTrackers[j].Nonce
		})
	}
	if order == interfaces.Descending {
		sort.SliceStable(resp.OutTxTrackers, func(i, j int) bool {
			return resp.OutTxTrackers[i].Nonce > resp.OutTxTrackers[j].Nonce
		})
	}
	return resp.OutTxTrackers, nil
}

func (c *Clients) GetLastBlockHeightByChain(ctx context.Context, chain chains.Chain) (*types.LastBlockHeight, error) {
	resp, err := c.Xmsg.LastBlockHeight(ctx, &types.QueryGetLastBlockHeightRequest{Index: chain.ChainName()})
	if err != nil {
		return nil, err
	}
	return resp.LastBlockHeight, nil
}

func (c *Clients) GetChainIndex(ctx context.Context, chainId int64) (*types.ChainIndex, error) {
	resp, err := c.Xmsg.ChainIndex(ctx, &types.QueryChainIndexRequest{ChainId: chainId})
	if err != nil {
		return nil, err
	}

	return &types.ChainIndex{
		ChainId:    uint64(chainId),
		CurrHeight: resp.CurrHeight,
	}, nil
}

// GetPellRechargeOperationIndex returns the pell token increment index
func (c *Clients) GetPellRechargeOperationIndex(ctx context.Context, chainId int64) (*types.PellRechargeOperationIndex, error) {
	resp, err := c.Xmsg.PellRechargeOperationIndex(ctx, &types.QueryPellRechargeOperationIndexRequest{ChainId: chainId})
	if err != nil {
		return nil, err
	}

	return &types.PellRechargeOperationIndex{
		ChainId:   uint64(chainId),
		CurrIndex: resp.CurrIndex,
	}, nil
}

// GetGasRechargeOperationIndex returns the gas token increment index
func (c *Clients) GetGasRechargeOperationIndex(ctx context.Context, chainId int64) (*types.GasRechargeOperationIndex, error) {
	resp, err := c.Xmsg.GasRechargeOperationIndex(ctx, &types.QueryGasRechargeOperationIndexRequest{ChainId: chainId})
	if err != nil {
		return nil, err
	}

	return &types.GasRechargeOperationIndex{
		ChainId:   uint64(chainId),
		CurrIndex: resp.CurrIndex,
	}, nil
}
