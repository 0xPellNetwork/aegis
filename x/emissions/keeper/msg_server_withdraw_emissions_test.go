package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/cmd/pellcored/config"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	"github.com/0xPellNetwork/aegis/x/emissions/keeper"
	"github.com/0xPellNetwork/aegis/x/emissions/types"
)

func TestMsgServer_WithdrawEmission(t *testing.T) {
	t.Run("successfully withdraw emissions", func(t *testing.T) {
		k, ctx, sk, _ := keepertest.EmissionsKeeper(t)

		msgServer := keeper.NewMsgServerImpl(*k)
		withdrawableEmission := sample.WithdrawableEmissions(t)
		k.SetWithdrawableEmission(ctx, withdrawableEmission)
		err := sk.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, withdrawableEmission.Amount)))
		require.NoError(t, err)
		err = sk.BankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.UndistributedObserverRewardsPool, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, withdrawableEmission.Amount)))
		require.NoError(t, err)

		_, err = msgServer.WithdrawEmission(ctx, &types.MsgWithdrawEmission{
			Signer: withdrawableEmission.Address,
			Amount: withdrawableEmission.Amount,
		})
		require.NoError(t, err)

		balance := k.GetBankKeeper().GetBalance(ctx, sdk.MustAccAddressFromBech32(withdrawableEmission.Address), config.BaseDenom).Amount.String()
		require.Equal(t, withdrawableEmission.Amount.String(), balance)
		balance = k.GetBankKeeper().GetBalance(ctx, types.UndistributedObserverRewardsPoolAddress, config.BaseDenom).Amount.String()
		require.Equal(t, math.ZeroInt().String(), balance)
	})

	t.Run("unable to withdraw emissions with invalid address", func(t *testing.T) {
		k, ctx, sk, _ := keepertest.EmissionsKeeper(t)

		msgServer := keeper.NewMsgServerImpl(*k)
		withdrawableEmission := sample.WithdrawableEmissions(t)
		k.SetWithdrawableEmission(ctx, withdrawableEmission)
		err := sk.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, withdrawableEmission.Amount)))
		require.NoError(t, err)
		err = sk.BankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.UndistributedObserverRewardsPool, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, withdrawableEmission.Amount)))
		require.NoError(t, err)

		_, err = msgServer.WithdrawEmission(ctx, &types.MsgWithdrawEmission{
			Signer: "invalid_address",
			Amount: withdrawableEmission.Amount,
		})
		require.ErrorIs(t, err, types.ErrInvalidAddress)
	})

	t.Run("unable to withdraw emissions if undistributed rewards pool does not have enough balance", func(t *testing.T) {
		k, ctx, _, _ := keepertest.EmissionsKeeper(t)

		msgServer := keeper.NewMsgServerImpl(*k)
		withdrawableEmission := sample.WithdrawableEmissions(t)
		k.SetWithdrawableEmission(ctx, withdrawableEmission)

		_, err := msgServer.WithdrawEmission(ctx, &types.MsgWithdrawEmission{
			Signer: withdrawableEmission.Address,
			Amount: withdrawableEmission.Amount,
		})
		require.ErrorIs(t, err, types.ErrRewardsPoolDoesNotHaveEnoughBalance)
	})

	t.Run("unable to withdraw emissions with invalid amount", func(t *testing.T) {
		k, ctx, _, _ := keepertest.EmissionsKeeper(t)
		msgServer := keeper.NewMsgServerImpl(*k)
		withdrawableEmission := sample.WithdrawableEmissions(t)
		k.SetWithdrawableEmission(ctx, withdrawableEmission)
		_, err := msgServer.WithdrawEmission(ctx, &types.MsgWithdrawEmission{
			Signer: withdrawableEmission.Address,
			Amount: sdkmath.NewInt(-1),
		})
		require.ErrorIs(t, err, types.ErrUnableToWithdrawEmissions)
	})

	t.Run("unable to withdraw emissions if SendCoinsFromModuleToAccount fails", func(t *testing.T) {
		k, ctx, sk, _ := keepertest.EmissionKeeperWithMockOptions(t, keepertest.EmissionMockOptions{
			UseBankMock: true,
		})
		bankMock := keepertest.GetEmissionsBankMock(t, k)
		msgServer := keeper.NewMsgServerImpl(*k)

		withdrawableEmission := sample.WithdrawableEmissions(t)
		k.SetWithdrawableEmission(ctx, withdrawableEmission)
		err := sk.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, withdrawableEmission.Amount)))
		require.NoError(t, err)
		err = sk.BankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.UndistributedObserverRewardsPool, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, withdrawableEmission.Amount)))
		require.NoError(t, err)
		address, err := sdk.AccAddressFromBech32(withdrawableEmission.Address)
		require.NoError(t, err)

		bankMock.On("SendCoinsFromModuleToAccount",
			ctx, types.UndistributedObserverRewardsPool, address, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, withdrawableEmission.Amount))).
			Return(types.ErrUnableToWithdrawEmissions).Once()
		bankMock.On("GetBalance",
			ctx, mock.Anything, config.BaseDenom).
			Return(sdk.NewCoin(config.BaseDenom, withdrawableEmission.Amount), nil).Once()
		_, err = msgServer.WithdrawEmission(ctx, &types.MsgWithdrawEmission{
			Signer: withdrawableEmission.Address,
			Amount: withdrawableEmission.Amount,
		})

		require.ErrorIs(t, err, types.ErrUnableToWithdrawEmissions)
		balance := sk.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(withdrawableEmission.Address), config.BaseDenom).Amount.String()
		require.Equal(t, math.ZeroInt().String(), balance)
	})
	t.Run("unable to withdraw emissions if amount requested is more that available", func(t *testing.T) {
		k, ctx, sk, _ := keepertest.EmissionsKeeper(t)

		msgServer := keeper.NewMsgServerImpl(*k)
		withdrawableEmission := sample.WithdrawableEmissions(t)
		k.SetWithdrawableEmission(ctx, withdrawableEmission)
		withdrawAmount := withdrawableEmission.Amount.Add(sdkmath.OneInt())
		err := sk.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, withdrawAmount)))
		require.NoError(t, err)
		err = sk.BankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.UndistributedObserverRewardsPool, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, withdrawAmount)))
		require.NoError(t, err)

		_, err = msgServer.WithdrawEmission(ctx, &types.MsgWithdrawEmission{
			Signer: withdrawableEmission.Address,
			Amount: withdrawAmount,
		})
		require.ErrorIs(t, err, types.ErrUnableToWithdrawEmissions)
		require.ErrorContains(t, err, "amount to be removed is greater than the available withdrawable emission")
	})

}
