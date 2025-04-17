package coin

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func Test_GetApellDecFromAmountInPell(t *testing.T) {
	tt := []struct {
		name        string
		pellAmount  string
		err         require.ErrorAssertionFunc
		apellAmount math.LegacyDec
	}{
		{
			name:        "valid pell amount",
			pellAmount:  "210000000",
			err:         require.NoError,
			apellAmount: math.LegacyMustNewDecFromStr("210000000000000000000000000"),
		},
		{
			name:        "very high pell amount",
			pellAmount:  "21000000000000000000",
			err:         require.NoError,
			apellAmount: math.LegacyMustNewDecFromStr("21000000000000000000000000000000000000"),
		},
		{
			name:        "very low pell amount",
			pellAmount:  "1",
			err:         require.NoError,
			apellAmount: math.LegacyMustNewDecFromStr("1000000000000000000"),
		},
		{
			name:        "zero pell amount",
			pellAmount:  "0",
			err:         require.NoError,
			apellAmount: math.LegacyMustNewDecFromStr("0"),
		},
		{
			name:        "decimal pell amount",
			pellAmount:  "0.1",
			err:         require.NoError,
			apellAmount: math.LegacyMustNewDecFromStr("100000000000000000"),
		},
		{
			name:        "invalid pell amount",
			pellAmount:  "%%%%%$#",
			err:         require.Error,
			apellAmount: math.LegacyMustNewDecFromStr("0"),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			apell, err := GetApellDecFromAmountInPell(tc.pellAmount)
			tc.err(t, err)
			if err == nil {
				require.Equal(t, tc.apellAmount, apell)
			}
		})
	}

}

func TestGetCoinType(t *testing.T) {
	tests := []struct {
		name    string
		coin    string
		want    CoinType
		wantErr bool
	}{
		{
			name:    "valid coin type 0",
			coin:    "0",
			want:    CoinType(0),
			wantErr: false,
		},
		{
			name:    "valid coin type 1",
			coin:    "1",
			want:    CoinType(1),
			wantErr: false,
		},
		{
			name:    "valid coin type 2",
			coin:    "2",
			want:    CoinType(2),
			wantErr: false,
		},
		{
			name:    "valid coin type 3",
			coin:    "3",
			want:    CoinType(3),
			wantErr: false,
		},
		{
			name:    "invalid coin type negative",
			coin:    "-1",
			want:    CoinType_CMD,
			wantErr: true,
		},
		{
			name:    "invalid coin type large number",
			coin:    "4",
			want:    CoinType_CMD,
			wantErr: true,
		},
		{
			name:    "invalid coin type non-integer",
			coin:    "abc",
			want:    CoinType_CMD,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCoinType(tt.coin)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
