package observer

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/tokens/pell.sol"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/0xPellNetwork/aegis/relayer/metrics"
	clienttypes "github.com/0xPellNetwork/aegis/relayer/types"
)

// WatchPellToken watches evm chain for pell token balance and post to pellcore when it below threshold
func (ob *ChainClient) WatchPellToken(ctx context.Context) error {
	ticker, err := clienttypes.NewDynamicTicker(
		fmt.Sprintf("EVM_WatchPellToken_%d", ob.Chain().Id),
		ob.GetChainParams().WatchPellTokenTicker,
	)
	if err != nil {
		ob.Logger().Outbound.Error().Err(err).Msg("NewDynamicTicker error")
		return err
	}

	ob.Logger().Outbound.Info().Msgf("WatchPellToken started with interval %d", ob.GetChainParams().WatchPellTokenTicker)

	defer ticker.Stop()
	for {
		select {
		case <-ticker.C():
			ob.Logger().Outbound.Info().Msgf("WatchPellToken pell token recharge enabled %v", ob.GetChainParams().PellTokenRechargeEnabled)

			if !ob.GetChainParams().IsSupported {
				continue
			}

			if err = ob.monitorPellTokenBalance(ctx); err != nil {
				ob.Logger().Outbound.Error().Err(err).Msg("monitorPellTokenBalance error")
			}

			ticker.UpdateInterval(ob.GetChainParams().WatchPellTokenTicker, ob.Logger().Outbound)
		case <-ob.StopChannel():
			ob.Logger().Outbound.Info().Msg("WatchPellToken stopped")
			return nil
		}
	}
}

// monitorPellTokenBalance checks and updates metrics for pell token balance
func (ob *ChainClient) monitorPellTokenBalance(ctx context.Context) error {
	pellTokenAddr := common.HexToAddress(ob.GetChainParams().PellTokenContractAddress)
	balance, err := ob.getERC20Balance(ctx, pellTokenAddr, ob.TSS().EVMAddress())
	if err != nil {
		return errors.Wrap(err, "unable to get pell token balance")
	}

	// Update metrics
	metrics.TssAddressPellTokenBalance.WithLabelValues(
		fmt.Sprint(ob.Chain().Id),
		ob.TSS().EVMAddress().String(),
	).Set(float64(balance.Int64()))

	if ob.GetChainParams().PellTokenRechargeEnabled {
		if err = ob.PostPellTokenVote(ctx, balance); err != nil {
			ob.Logger().Outbound.Error().Err(err).Msgf("PostPellTokenVote error for chain %d", ob.Chain().Id)
		}

		time.Sleep(time.Duration(ob.GetChainParams().PellTokenPostInterval) * time.Second)
	}

	return nil
}

func (ob *ChainClient) PostPellTokenVote(ctx context.Context, balance *big.Int) error {
	threshold := ob.GetChainParams().PellTokenRechargeThreshold.BigInt()
	if balance.Cmp(threshold) >= 0 {
		return nil
	}

	chain := ob.Chain()
	msg, err := ob.PellcoreClient().GetPellRechargeOperationIndex(ctx, chain.Id)
	if err != nil {
		return errors.Wrap(err, "unable to get pell token increment index")
	}

	voteIndex := msg.CurrIndex + 1
	txHash, err := ob.PellcoreClient().PostVoteOnPellRecharge(ctx, chain, voteIndex)
	if err != nil {
		return errors.Wrap(err, "unable to post vote for pell token recharge")
	}

	ob.Logger().Outbound.Info().Msgf("Vote for pell token recharge success, voteIndex: %d, txHash: %s", voteIndex, txHash)
	return nil
}

func (ob *ChainClient) getERC20Balance(ctx context.Context, tokenAddress, accountAddress common.Address) (*big.Int, error) {
	// Define the ABI of the ERC20 balanceOf function
	erc20ABI, err := pell.PELLMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %v", err)
	}

	// Pack the function call with the account address
	data, err := erc20ABI.Pack("balanceOf", accountAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to pack function call: %v", err)
	}

	// Prepare the call message
	msg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}

	// Execute the contract call, use nil for the block number to get the latest state
	result, err := ob.evmClient.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("contract call failed: %v", err)
	}

	// Unpack the result to get the balance
	var balance *big.Int
	if err = erc20ABI.UnpackIntoInterface(&balance, "balanceOf", result); err != nil {
		return nil, fmt.Errorf("failed to unpack result: %v", err)
	}

	return balance, nil
}
