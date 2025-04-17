package observer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	clienttypes "github.com/0xPellNetwork/aegis/relayer/types"
)

// WatchGasPrice watches evm chain for gas prices and post to pellcore
// TODO(revamp): move inner logic to a separate function
func (ob *ChainClient) WatchGasPrice(ctx context.Context) error {
	// report gas price right away as the ticker takes time to kick in
	err := ob.PostGasPrice(ctx)
	if err != nil {
		ob.Logger().GasPrice.Error().Err(err).Msgf("PostGasPrice error for chain %d", ob.Chain().Id)
	}

	// start gas price ticker
	ticker, err := clienttypes.NewDynamicTicker(
		fmt.Sprintf("EVM_WatchGasPrice_%d", ob.Chain().Id),
		ob.GetChainParams().GasPriceTicker,
	)
	if err != nil {
		ob.Logger().GasPrice.Error().Err(err).Msg("NewDynamicTicker error")
		return err
	}
	ob.Logger().GasPrice.Info().Msgf("WatchGasPrice started for chain %d with interval %d",
		ob.Chain().Id, ob.GetChainParams().GasPriceTicker)

	defer ticker.Stop()
	for {
		select {
		case <-ticker.C():
			if !ob.GetChainParams().IsSupported {
				continue
			}
			if err = ob.PostGasPrice(ctx); err != nil {
				ob.Logger().GasPrice.Error().Err(err).Msgf("PostGasPrice error for chain %d", ob.Chain().Id)
			}

			ticker.UpdateInterval(ob.GetChainParams().GasPriceTicker, ob.Logger().GasPrice)
		case <-ob.StopChannel():
			ob.Logger().GasPrice.Info().Msg("WatchGasPrice stopped")
			return nil
		}
	}
}

func (ob *ChainClient) PostGasPrice(ctx context.Context) error {
	// GAS PRICE
	gasPrice, err := ob.evmClient.SuggestGasPrice(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to suggest gas price")
	}

	blockNum, err := ob.evmClient.BlockNumber(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get block number")
	}

	// SUPPLY
	supply := "100" // lockedAmount on ETH, totalSupply on other chains

	_, err = ob.PellcoreClient().PostGasPrice(ctx, ob.Chain(), gasPrice.Uint64(), supply, blockNum)
	if err != nil {
		return errors.Wrap(err, "unable to post vote for gas price")
	}

	return nil
}
