package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	types "github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

func TestUpdateLSTStakingEnabled(t *testing.T) {
	// Initialize test environment
	mocks, ctx := keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	ak := keepertest.GetXSecurityAuthorityMock(t, mocks)

	// Test cases
	tests := []struct {
		name    string
		setup   func()
		msg     *types.MsgUpdateLSTStakingEnabled
		wantErr error
	}{
		// Test Case 1: Authorized signer successfully enables LST staking
		// Verifies that a user with GROUP_OPERATIONAL permission can set LST staking to enabled
		{
			name: "success - enable",
			setup: func() {
				// Setup auth mock for authorized case
				ak.On("IsAuthorized",
					mock.Anything,
					"auth_signer",
					authoritytypes.PolicyType_GROUP_OPERATIONAL,
				).Return(true).Once()
			},
			msg: &types.MsgUpdateLSTStakingEnabled{
				Signer:  "auth_signer",
				Enabled: true,
			},
			wantErr: nil,
		},

		// Test Case 2: Authorized signer successfully disables LST staking
		// Verifies that a user with GROUP_OPERATIONAL permission can set LST staking to disabled
		{
			name: "success - disable",
			setup: func() {
				// Setup auth mock for authorized case
				ak.On("IsAuthorized",
					mock.Anything,
					"auth_signer",
					authoritytypes.PolicyType_GROUP_OPERATIONAL,
				).Return(true).Once()
			},
			msg: &types.MsgUpdateLSTStakingEnabled{
				Signer:  "auth_signer",
				Enabled: false,
			},
			wantErr: nil,
		},

		// Test Case 3: Unauthorized signer attempts to modify LST staking settings
		// Verifies that a user without GROUP_OPERATIONAL permission cannot change LST staking status
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
			msg: &types.MsgUpdateLSTStakingEnabled{
				Signer:  "unauth_signer",
				Enabled: true,
			},
			wantErr: authoritytypes.ErrUnauthorized,
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
			_, err := mocks.UpdateLSTStakingEnabled(ctx, tc.msg)

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
