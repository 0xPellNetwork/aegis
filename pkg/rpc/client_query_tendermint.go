package rpc

import (
	"context"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"

	"github.com/pell-chain/pellcore/pkg/retry"
)

// GetLatestPellBlock returns the latest pell block
func (c *Clients) GetLatestPellBlock(ctx context.Context) (*cmtservice.Block, error) {
	res, err := c.Tendermint.GetLatestBlock(ctx, &cmtservice.GetLatestBlockRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get latest pell block")
	}

	return res.SdkBlock, nil
}

// GetNodeInfo returns the node info
func (c *Clients) GetNodeInfo(ctx context.Context) (*cmtservice.GetNodeInfoResponse, error) {
	res, err := retry.DoTypedWithRetry(func() (*cmtservice.GetNodeInfoResponse, error) {
		return c.Tendermint.GetNodeInfo(ctx, &cmtservice.GetNodeInfoRequest{})
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to get node info")
	}

	return res, nil
}
