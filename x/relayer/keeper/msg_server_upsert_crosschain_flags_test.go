package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/relayer/keeper"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func TestMsgServer_UpsertCrosschainFlags(t *testing.T) {
	t.Run("can update crosschain flags", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		srv := keeper.NewMsgServerImpl(*k)
		admin := sample.AccAddress()

		// mock the authority keeper for authorization
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		_, err := srv.UpsertCrosschainFlags(sdk.WrapSDKContext(ctx), &types.MsgUpsertCrosschainFlags{
			Signer:            admin,
			IsInboundEnabled:  false,
			IsOutboundEnabled: false,
			GasPriceIncreaseFlags: &types.GasPriceIncreaseFlags{
				EpochLength:             42,
				RetryInterval:           time.Minute * 42,
				GasPriceIncreasePercent: 42,
			},
			BlockHeaderVerificationFlags: &types.BlockHeaderVerificationFlags{
				IsEthTypeChainEnabled: true,
				IsBtcTypeChainEnabled: false,
			},
		})
		require.NoError(t, err)

		flags, found := k.GetCrosschainFlags(ctx)
		require.True(t, found)
		require.False(t, flags.IsInboundEnabled)
		require.False(t, flags.IsOutboundEnabled)
		require.Equal(t, int64(42), flags.GasPriceIncreaseFlags.EpochLength)
		require.Equal(t, time.Minute*42, flags.GasPriceIncreaseFlags.RetryInterval)
		require.Equal(t, uint32(42), flags.GasPriceIncreaseFlags.GasPriceIncreasePercent)
		require.True(t, flags.BlockHeaderVerificationFlags.IsEthTypeChainEnabled)
		require.False(t, flags.BlockHeaderVerificationFlags.IsBtcTypeChainEnabled)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		// can update flags again
		_, err = srv.UpsertCrosschainFlags(sdk.WrapSDKContext(ctx), &types.MsgUpsertCrosschainFlags{
			Signer:            admin,
			IsInboundEnabled:  true,
			IsOutboundEnabled: true,
			GasPriceIncreaseFlags: &types.GasPriceIncreaseFlags{
				EpochLength:             43,
				RetryInterval:           time.Minute * 43,
				GasPriceIncreasePercent: 43,
			},
			BlockHeaderVerificationFlags: &types.BlockHeaderVerificationFlags{
				IsEthTypeChainEnabled: false,
				IsBtcTypeChainEnabled: true,
			},
		})
		require.NoError(t, err)

		flags, found = k.GetCrosschainFlags(ctx)
		require.True(t, found)
		require.True(t, flags.IsInboundEnabled)
		require.True(t, flags.IsOutboundEnabled)
		require.Equal(t, int64(43), flags.GasPriceIncreaseFlags.EpochLength)
		require.Equal(t, time.Minute*43, flags.GasPriceIncreaseFlags.RetryInterval)
		require.Equal(t, uint32(43), flags.GasPriceIncreaseFlags.GasPriceIncreasePercent)
		require.False(t, flags.BlockHeaderVerificationFlags.IsEthTypeChainEnabled)
		require.True(t, flags.BlockHeaderVerificationFlags.IsBtcTypeChainEnabled)

		// group 1 should be able to disable inbound and outbound
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, true)

		// if gas price increase flags is nil, it should not be updated
		_, err = srv.UpsertCrosschainFlags(sdk.WrapSDKContext(ctx), &types.MsgUpsertCrosschainFlags{
			Signer:            admin,
			IsInboundEnabled:  false,
			IsOutboundEnabled: false,
		})
		require.NoError(t, err)

		flags, found = k.GetCrosschainFlags(ctx)
		require.True(t, found)
		require.False(t, flags.IsInboundEnabled)
		require.False(t, flags.IsOutboundEnabled)
		require.Equal(t, int64(43), flags.GasPriceIncreaseFlags.EpochLength)
		require.Equal(t, time.Minute*43, flags.GasPriceIncreaseFlags.RetryInterval)
		require.Equal(t, uint32(43), flags.GasPriceIncreaseFlags.GasPriceIncreasePercent)
		require.False(t, flags.BlockHeaderVerificationFlags.IsEthTypeChainEnabled)
		require.True(t, flags.BlockHeaderVerificationFlags.IsBtcTypeChainEnabled)

		// group 1 should be able to disable header verification
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, true)

		// if gas price increase flags is nil, it should not be updated
		_, err = srv.UpsertCrosschainFlags(sdk.WrapSDKContext(ctx), &types.MsgUpsertCrosschainFlags{
			Signer:            admin,
			IsInboundEnabled:  false,
			IsOutboundEnabled: false,
			BlockHeaderVerificationFlags: &types.BlockHeaderVerificationFlags{
				IsEthTypeChainEnabled: false,
				IsBtcTypeChainEnabled: false,
			},
		})
		require.NoError(t, err)

		flags, found = k.GetCrosschainFlags(ctx)
		require.True(t, found)
		require.False(t, flags.IsInboundEnabled)
		require.False(t, flags.IsOutboundEnabled)
		require.Equal(t, int64(43), flags.GasPriceIncreaseFlags.EpochLength)
		require.Equal(t, time.Minute*43, flags.GasPriceIncreaseFlags.RetryInterval)
		require.Equal(t, uint32(43), flags.GasPriceIncreaseFlags.GasPriceIncreasePercent)
		require.False(t, flags.BlockHeaderVerificationFlags.IsEthTypeChainEnabled)
		require.False(t, flags.BlockHeaderVerificationFlags.IsBtcTypeChainEnabled)

		// if flags are not defined, default should be used
		k.RemoveCrosschainFlags(ctx)
		_, found = k.GetCrosschainFlags(ctx)
		require.False(t, found)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, true)

		_, err = srv.UpsertCrosschainFlags(sdk.WrapSDKContext(ctx), &types.MsgUpsertCrosschainFlags{
			Signer:            admin,
			IsInboundEnabled:  false,
			IsOutboundEnabled: true,
		})
		require.NoError(t, err)

		flags, found = k.GetCrosschainFlags(ctx)
		require.True(t, found)
		require.False(t, flags.IsInboundEnabled)
		require.True(t, flags.IsOutboundEnabled)
		require.Equal(t, types.DefaultGasPriceIncreaseFlags.EpochLength, flags.GasPriceIncreaseFlags.EpochLength)
		require.Equal(t, types.DefaultGasPriceIncreaseFlags.RetryInterval, flags.GasPriceIncreaseFlags.RetryInterval)
		require.Equal(t, types.DefaultGasPriceIncreaseFlags.GasPriceIncreasePercent, flags.GasPriceIncreaseFlags.GasPriceIncreasePercent)
	})

	t.Run("cannot update crosschain flags if not authorized", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMockOptions{
			UseAuthorityMock: true,
		})
		srv := keeper.NewMsgServerImpl(*k)

		admin := sample.AccAddress()
		authorityMock := keepertest.GetRelayerAuthorityMock(t, k)
		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_EMERGENCY, false)

		_, err := srv.UpsertCrosschainFlags(sdk.WrapSDKContext(ctx), &types.MsgUpsertCrosschainFlags{
			Signer:            admin,
			IsInboundEnabled:  false,
			IsOutboundEnabled: false,
		})
		require.Error(t, err)
		require.Equal(t, authoritytypes.ErrUnauthorized, err)

		keepertest.MockIsAuthorized(&authorityMock.Mock, admin, authoritytypes.PolicyType_GROUP_OPERATIONAL, false)

		_, err = srv.UpsertCrosschainFlags(sdk.WrapSDKContext(ctx), &types.MsgUpsertCrosschainFlags{
			Signer:            admin,
			IsInboundEnabled:  false,
			IsOutboundEnabled: true,
		})
		require.Error(t, err)
		require.Equal(t, authoritytypes.ErrUnauthorized, err)
	})
}
