package observer

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/pkg/errors"

	"github.com/pell-chain/pellcore/relayer/metrics"
	clienttypes "github.com/pell-chain/pellcore/relayer/types"
)

// WatchGasToken watches evm chain for gas token balance and post to pellcore when it below threshold
func (ob *ChainClient) WatchGasToken(ctx context.Context) error {
	ticker, err := clienttypes.NewDynamicTicker(
		fmt.Sprintf("EVM_WatchGasToken_%d", ob.Chain().Id),
		ob.GetChainParams().WatchGasTokenTicker,
	)
	if err != nil {
		ob.Logger().Outbound.Error().Err(err).Msg("NewDynamicTicker error")
		return err
	}

	ob.Logger().Outbound.Info().Msgf("WatchGasToken started for chain %d with interval %d",
		ob.Chain().Id, ob.GetChainParams().WatchGasTokenTicker)

	defer ticker.Stop()
	for {
		select {
		case <-ticker.C():
			ob.Logger().Outbound.Info().Msgf("WatchGasToken for chain %d, gas token recharge enabled %v",
				ob.Chain().Id, ob.GetChainParams().GasTokenRechargeEnabled)
			if !ob.GetChainParams().IsSupported {
				continue
			}

			if err := ob.monitorGasTokenBalance(ctx); err != nil {
				ob.Logger().Outbound.Error().Err(err).Msg("monitorGasTokenBalance error")
			}

			ticker.UpdateInterval(ob.GetChainParams().WatchGasTokenTicker, ob.Logger().Outbound)
		case <-ob.StopChannel():
			ob.Logger().Outbound.Info().Msg("WatchGasTokenTicker stopped")
			return nil
		}
	}
}

// monitorGasTokenBalance checks and updates metrics for gas token balance
func (ob *ChainClient) monitorGasTokenBalance(ctx context.Context) error {
	balance, err := ob.evmClient.BalanceAt(ctx, ob.TSS().EVMAddress(), nil)
	if err != nil {
		return errors.Wrap(err, "unable to get gas token balance")
	}

	ob.Logger().Outbound.Info().Msgf("Gas token balance: %d", balance.Int64())

	// Update metrics
	metrics.TssAddressGasBalance.WithLabelValues(
		fmt.Sprint(ob.Chain().Id),
		ob.TSS().EVMAddress().String(),
	).Set(float64(balance.Int64()))

	ob.Logger().Outbound.Info().Msgf("Gas token recharge enabled: %v", ob.GetChainParams().GasTokenRechargeEnabled)

	if ob.GetChainParams().GasTokenRechargeEnabled {
		if err = ob.PostGasTokenVote(ctx, balance); err != nil {
			ob.Logger().Outbound.Error().Err(err).Msgf("PostGasTokenVote error for chain %d", ob.Chain().Id)
		}

		time.Sleep(time.Duration(ob.GetChainParams().GasTokenPostInterval) * time.Second)
	}

	return nil
}

// PostGasTokenVote posts vote for gas token recharge
func (ob *ChainClient) PostGasTokenVote(ctx context.Context, balance *big.Int) error {
	threshold := ob.GetChainParams().GasTokenRechargeThreshold.BigInt()
	if balance.Cmp(threshold) >= 0 {
		return nil
	}

	chain := ob.Chain()
	msg, err := ob.PellcoreClient().GetGasRechargeOperationIndex(ctx, chain.Id)
	if err != nil {
		return errors.Wrap(err, "unable to get gas token increment index")
	}

	voteIndex := msg.CurrIndex + 1
	txHash, err := ob.PellcoreClient().PostVoteOnGasRecharge(ctx, chain, voteIndex)
	if err != nil {
		return errors.Wrap(err, "unable to post vote for gas token recharge")
	}

	ob.Logger().Outbound.Info().Msgf("Vote for gas token recharge success, voteIndex: %d, txHash: %s", voteIndex, txHash)
	return nil
}
