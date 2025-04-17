package observer

import (
	"context"
	"fmt"
	"time"

	"github.com/pell-chain/pellcore/relayer/chains/evm/rpc"
	"github.com/pell-chain/pellcore/relayer/common"
	"github.com/pell-chain/pellcore/relayer/metrics"
)

// WatchRPCStatus watches the RPC status of the EVM chain
func (ob *ChainClient) WatchRPCStatus(ctx context.Context) error {
	ob.Logger().Chain.Info().Msg("watchRPCStatus started")

	ticker := time.NewTicker(common.RPCStatusCheckInterval)
	for {
		select {
		case <-ticker.C:
			if !ob.GetChainParams().IsSupported {
				ob.Logger().Chain.Info().Msg("watchRPCStatus chain is not supported, skipping RPC status check")
				continue
			}

			ob.checkRPCStatus(ctx)
		case <-ob.StopChannel():
			return nil
		}
	}
}

// checkRPCStatus checks the RPC status of the EVM chain
func (ob *ChainClient) checkRPCStatus(ctx context.Context) {
	blockTime, err := rpc.CheckRPCStatus(ctx, ob.evmClient)
	if err != nil {
		metrics.RPCNodeStatus.WithLabelValues(fmt.Sprint(ob.Chain().Id)).Set(0)
		ob.Logger().Chain.Error().Err(err).Msg("watchRPCStatus checkRPCStatus failed")
		return
	}

	metrics.RPCNodeStatus.WithLabelValues(fmt.Sprint(ob.Chain().Id)).Set(1)

	// alert if RPC latency is too high
	ob.AlertOnRPCLatency(blockTime, rpc.RPCAlertLatency)
}
