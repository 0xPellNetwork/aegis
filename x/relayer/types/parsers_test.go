package types_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestConvertReceiveStatusToVoteType(t *testing.T) {
	tests := []struct {
		name     string
		status   chains.ReceiveStatus
		expected types.VoteType
	}{
		{"TestSuccessStatus", chains.ReceiveStatus_SUCCESS, types.VoteType_SUCCESS_OBSERVATION},
		{"TestFailedStatus", chains.ReceiveStatus_FAILED, types.VoteType_FAILURE_OBSERVATION},
		{"TestDefaultStatus", chains.ReceiveStatus_CREATED, types.VoteType_NOT_YET_VOTED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := types.ConvertReceiveStatusToVoteType(tt.status)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestParseStringToObservationType(t *testing.T) {
	tests := []struct {
		name            string
		observationType string
		expected        types.ObservationType
	}{
		{"TestEmptyObserverType", "EMPTY_OBSERVER_TYPE", types.ObservationType(0)},
		{"TestInBoundTx", "IN_BOUND_TX", types.ObservationType(1)},
		{"TestOutBoundTx", "OUT_BOUND_TX", types.ObservationType(2)},
		{"TestTSSKeyGen", "TSS_KEY_GEN", types.ObservationType(3)},
		{"TestTSSKeySign", "TSS_KEY_SIGN", types.ObservationType(4)},
		{"TestInBoundBlock", "IN_BOUND_BLOCK", types.ObservationType(5)},
		{"TestPellTokenRecharge", "PELL_TOKEN_RECHARGE", types.ObservationType(6)},
		{"TestGasTokenRecharge", "GAS_TOKEN_RECHARGE", types.ObservationType(7)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := types.ParseStringToObservationType(tt.observationType)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGetOperatorAddressFromAccAddress(t *testing.T) {
	tests := []struct {
		name    string
		accAddr string
		wantErr bool
	}{
		{"TestValidAccAddress", sample.AccAddress(), false},
		{"TestInvalidAccAddress", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := types.GetOperatorAddressFromAccAddress(tt.accAddr)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetAccAddressFromOperatorAddress(t *testing.T) {
	// #nosec G404 test purpose - weak randomness is not an issue here
	r := rand.New(rand.NewSource(1))
	tests := []struct {
		name       string
		valAddress string
		wantErr    bool
	}{
		{"TestValidValAddress", sample.ValAddress(r).String(), false},
		{"TestInvalidValAddress", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := types.GetAccAddressFromOperatorAddress(tt.valAddress)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
