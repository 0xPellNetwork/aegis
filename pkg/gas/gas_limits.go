package gas

import (
	sdkmath "cosmossdk.io/math"
)

const (
	// EVMSend is the gas limit required to transfer tokens on an EVM based chain
	EVMSend = 21000
	// TODO: Move gas limits from pellclient to this file
)

// MultiplyGasPrice multiplies the median gas price by the given multiplier and returns the truncated value
func MultiplyGasPrice(medianGasPrice sdkmath.Uint, multiplierString string) (sdkmath.Uint, error) {
	multiplier, err := sdkmath.LegacyNewDecFromStr(multiplierString)
	if err != nil {
		return sdkmath.ZeroUint(), err
	}

	gasPrice := sdkmath.LegacyNewDecFromBigInt(medianGasPrice.BigInt())
	return sdkmath.NewUintFromString(gasPrice.Mul(multiplier).TruncateInt().String()), nil
}
