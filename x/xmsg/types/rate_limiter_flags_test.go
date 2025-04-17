package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestRateLimiterFlags_Validate(t *testing.T) {
	tt := []struct {
		name  string
		flags types.RateLimiterFlags
		isErr bool
	}{
		{
			name: "valid flags",
			flags: types.RateLimiterFlags{
				Enabled: true,
				Window:  42,
				Rate:    math.NewUint(42),
			},
		},
		{
			name:  "empty is valid",
			flags: types.RateLimiterFlags{},
		},
		{
			name: "negative window",
			flags: types.RateLimiterFlags{
				Enabled: true,
				Window:  -1,
			},
			isErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.flags.Validate()
			if tc.isErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

}
