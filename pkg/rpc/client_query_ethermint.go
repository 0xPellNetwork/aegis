package rpc

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	"github.com/evmos/ethermint/x/feemarket/types"
)

func (c *Clients) GetFeemarketParams(ctx context.Context) (*types.Params, error) {
	resp, err := c.EthermintFeeMarket.Params(ctx, &types.QueryParamsRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get base gas price")
	}
	if resp.Params.BaseFee.IsNil() {
		return nil, fmt.Errorf("base fee is nil")
	}

	return &resp.Params, nil
}
