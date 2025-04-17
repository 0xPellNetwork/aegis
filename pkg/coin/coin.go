package coin

import (
	"fmt"
	"strconv"

	"cosmossdk.io/math"
)

func GetCoinType(coin string) (CoinType, error) {
	coinInt, err := strconv.ParseInt(coin, 10, 32)
	if err != nil {
		return CoinType_CMD, err
	}
	if coinInt < 0 || coinInt > 3 {
		return CoinType_CMD, fmt.Errorf("invalid coin type %d", coinInt)
	}
	// #nosec G701 always in range
	return CoinType(coinInt), nil
}

func GetApellDecFromAmountInPell(pellAmount string) (math.LegacyDec, error) {
	pellDec, err := math.LegacyNewDecFromStr(pellAmount)
	if err != nil {
		return math.LegacyDec{}, err
	}

	return pellDec.Mul(math.LegacyNewDec(1e18)), nil
}
