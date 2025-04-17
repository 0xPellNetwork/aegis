package signer

import (
	"fmt"
	"math/big"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

const (
	MIN_GAS_LIMIT = 100_000
	MAX_GAS_LIMIT = 3_000_000
)

// These chains' gas_limit ignores checks
var UNCHECKED_CHAIN_IDS []int64 = []int64{5000, 5003}

// Gas represents gas parameters for EVM transactions.
//
// This is pretty interesting because all EVM chains now support EIP-1559, but some chains do it in a specific way
// https://eips.ethereum.org/EIPS/eip-1559
// https://www.blocknative.com/blog/eip-1559-fees
// https://github.com/binance-chain/BEPs/blob/master/BEPs/BEP226.md (tl;dr: baseFee is always zero)
//
// However, this doesn't affect tx creation nor broadcasting
type Gas struct {
	Limit uint64

	// This is a "total" gasPrice per 1 unit of gas.
	// GasPrice for pre EIP-1559 transactions or maxFeePerGas for EIP-1559.
	Price *big.Int

	// PriorityFee a fee paid directly to validators for EIP-1559.
	PriorityFee *big.Int
}

func (g Gas) validate() error {
	switch {
	case g.Limit == 0:
		return errors.New("gas limit is zero")
	case g.Price == nil:
		return errors.New("max fee per unit is nil")
	case g.PriorityFee == nil:
		return errors.New("priority fee per unit is nil")
	case g.Price.Cmp(g.PriorityFee) == -1:
		return fmt.Errorf(
			"max fee per unit (%d) is less than priority fee per unit (%d)",
			g.Price.Int64(),
			g.PriorityFee.Int64(),
		)
	default:
		return nil
	}
}

// isLegacy determines whether the gas is meant for LegacyTx{} (pre EIP-1559)
// or DynamicFeeTx{} (post EIP-1559).
//
// Returns true if priority fee is <= 0.
func (g Gas) isLegacy() bool {
	return g.PriorityFee.Sign() < 1
}

// unchecked gas limit chain
func uncheckedGaslimitChain(chainId int64) bool {
	for _, v := range UNCHECKED_CHAIN_IDS {
		if chainId == v {
			return true
		}
	}

	return false
}

// Ensure that the chain's gas limit is within the specified range.
func clampGasLimit(logger zerolog.Logger, chainId int64, gasLimit *uint64) {
	if !uncheckedGaslimitChain(chainId) {
		if *gasLimit < MIN_GAS_LIMIT {
			logger.Warn().
				Uint64("xmsg.initial_gas_limit", *gasLimit).
				Uint64("xmsg.gas_limit", MIN_GAS_LIMIT).
				Msgf("Gas limit is too low. Setting to the minimum (%d)", MIN_GAS_LIMIT)
			*gasLimit = MIN_GAS_LIMIT
			return
		}

		if *gasLimit > MAX_GAS_LIMIT {
			logger.Warn().
				Uint64("xmsg.initial_gas_limit", *gasLimit).
				Uint64("xmsg.gas_limit", MAX_GAS_LIMIT).
				Msgf("Gas limit is too high; Setting to the maximum (%d)", MAX_GAS_LIMIT)
			*gasLimit = MAX_GAS_LIMIT
		}
	}
}

// calc gas from xmsg
func gasFromXmsg(logger zerolog.Logger, xmsg *types.Xmsg) (Gas, error) {
	params := xmsg.GetCurrentOutTxParam()
	limit := params.OutboundTxGasLimit

	// ensure gas limit ∈ (MIN_GAS_LIMIT， MAX_GAS_LIMIT)
	clampGasLimit(logger, params.ReceiverChainId, &limit)

	gasPrice, valid := new(big.Int).SetString(params.OutboundTxGasPrice, 10)
	if !valid || gasPrice.Sign() == -1 {
		return Gas{}, errors.New(fmt.Sprintf("unable to parse gasPrice: %s", params.OutboundTxGasPrice))
	}

	// TODO: fix me
	priorityFee := big.NewInt(0)
	// priorityFee, err := bigIntFromString(params.GasPriorityFee)
	// switch {
	// case err != nil:
	// 	return Gas{}, errors.Wrap(err, "unable to parse priorityFee")
	// case gasPrice.Cmp(priorityFee) == -1:
	// 	return Gas{}, fmt.Errorf("gasPrice (%d) is less than priorityFee (%d)", gasPrice.Int64(), priorityFee.Int64())
	// }

	logger.Info().Uint64("xmsg.gas_limit", limit).Uint64("xmsg.gas_price", gasPrice.Uint64())

	return Gas{
		Limit:       limit,
		Price:       gasPrice,
		PriorityFee: priorityFee,
	}, nil
}
