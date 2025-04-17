package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Validate checks that the RateLimiterFlags is valid
func (r RateLimiterFlags) Validate() error {
	// window must not be negative
	if r.Window < 0 {
		return fmt.Errorf("window must be positive: %d", r.Window)
	}

	return nil
}

// GetConversionRate returns the conversion rate for the given zrc20
func (r RateLimiterFlags) GetConversionRate(zrc20 string) (math.LegacyDec, bool) {
	return math.LegacyNewDec(0), false
}
