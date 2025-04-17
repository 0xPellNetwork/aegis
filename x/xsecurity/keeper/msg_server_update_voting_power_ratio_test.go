package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	types "github.com/pell-chain/pellcore/x/xsecurity/types"
)

func TestUpdateVotingPowerRatio(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Test cases
	tests := []struct {
		name    string
		setup   func()
		msg     *types.MsgUpdateVotingPowerRatio
		wantErr error
	}{
		// Test Case 1: Successful update of voting power ratio
		// Verifies that an authorized signer can update the voting power ratio with valid parameters
		{
			name: "success",
			setup: func() {
				// Setup auth mock for authorized case
				ak.On("IsAuthorized",
					mock.Anything,
					"auth_signer",
					authoritytypes.PolicyType_GROUP_OPERATIONAL,
				).Return(true).Once()
			},
			msg: &types.MsgUpdateVotingPowerRatio{
				Signer:      "auth_signer",
				Numerator:   sdkmath.NewInt(30),
				Denominator: sdkmath.NewInt(100),
			},
			wantErr: nil,
		},

		// Test Case 2: Authorization failure
		// Verifies that an unauthorized signer cannot update the voting power ratio
		// and receives an authorization error
		{
			name: "unauthorized",
			setup: func() {
				// Setup auth mock for unauthorized case
				ak.On("IsAuthorized",
					mock.Anything,
					"unauth_signer",
					authoritytypes.PolicyType_GROUP_OPERATIONAL,
				).Return(false).Once()
			},
			msg: &types.MsgUpdateVotingPowerRatio{
				Signer:      "unauth_signer",
				Numerator:   sdkmath.NewInt(30),
				Denominator: sdkmath.NewInt(100),
			},
			wantErr: authoritytypes.ErrUnauthorized,
		},

		// Test Case 3: Invalid parameter - zero denominator
		// Verifies that the message validation rejects a zero denominator value
		// as it would cause division by zero errors
		{
			name: "zero denominator",
			msg: &types.MsgUpdateVotingPowerRatio{
				Signer:      "auth_signer",
				Numerator:   sdkmath.NewInt(30),
				Denominator: sdkmath.NewInt(0),
			},
			wantErr: types.ErrInvalidDenominator,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock before each test

			// Setup test case
			if tc.setup != nil {
				tc.setup()
			}

			// Execute
			_, err := mocks.UpdateVotingPowerRatio(ctx, tc.msg)

			// Verify results
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			// Verify mock expectations
			ak.AssertExpectations(t)
		})
	}
}
