package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/authority/keeper"
	"github.com/0xPellNetwork/aegis/x/authority/types"
)

func TestMsgServer_UpdatePolicies(t *testing.T) {
	t.Run("can't update policies with invalid signer", func(t *testing.T) {
		k, ctx := keepertest.AuthorityKeeper(t)
		msgServer := keeper.NewMsgServerImpl(*k)

		policies := sample.Policies()

		_, err := msgServer.UpdatePolicies(sdk.WrapSDKContext(ctx), &types.MsgUpdatePolicies{
			Signer:   sample.AccAddress(),
			Policies: policies,
		})
		require.ErrorIs(t, err, govtypes.ErrInvalidSigner)
	})

	t.Run("can update policies", func(t *testing.T) {
		k, ctx := keepertest.AuthorityKeeper(t)
		msgServer := keeper.NewMsgServerImpl(*k)

		policies := sample.Policies()

		res, err := msgServer.UpdatePolicies(sdk.WrapSDKContext(ctx), &types.MsgUpdatePolicies{
			Signer:   keepertest.AuthorityGovAddress.String(),
			Policies: policies,
		})
		require.NotNil(t, res)
		require.NoError(t, err)

		// Check policy is set
		got, found := k.GetPolicies(ctx)
		require.True(t, found)
		require.Equal(t, policies, got)
	})
}
