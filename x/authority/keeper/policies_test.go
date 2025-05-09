package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/authority/types"
)

func TestKeeper_SetPolicies(t *testing.T) {
	k, ctx := keepertest.AuthorityKeeper(t)
	policies := sample.Policies()

	_, found := k.GetPolicies(ctx)
	require.False(t, found)

	k.SetPolicies(ctx, policies)

	// Check policy is set
	got, found := k.GetPolicies(ctx)
	require.True(t, found)
	require.Equal(t, policies, got)

	// Can set policies again
	newPolicies := sample.Policies()
	require.NotEqual(t, policies, newPolicies)
	k.SetPolicies(ctx, newPolicies)
	got, found = k.GetPolicies(ctx)
	require.True(t, found)
	require.Equal(t, newPolicies, got)
}

func TestKeeper_IsAuthorized(t *testing.T) {
	k, ctx := keepertest.AuthorityKeeper(t)

	// Not authorized if no policies
	require.False(t, k.IsAuthorized(ctx, sample.AccAddress(), types.PolicyType_GROUP_ADMIN))
	require.False(t, k.IsAuthorized(ctx, sample.AccAddress(), types.PolicyType_GROUP_EMERGENCY))

	policies := sample.Policies()
	k.SetPolicies(ctx, policies)

	// Check policy is set
	got, found := k.GetPolicies(ctx)
	require.True(t, found)
	require.Equal(t, policies, got)

	// Check policy is authorized
	for _, policy := range policies.Items {
		require.True(t, k.IsAuthorized(ctx, policy.Address, policy.PolicyType))
	}

	// Check policy is not authorized
	require.False(t, k.IsAuthorized(ctx, sample.AccAddress(), types.PolicyType_GROUP_ADMIN))
	require.False(t, k.IsAuthorized(ctx, sample.AccAddress(), types.PolicyType_GROUP_EMERGENCY))
}
