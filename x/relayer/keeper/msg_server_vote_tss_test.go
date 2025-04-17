package keeper_test

import (
	"math"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/relayer/keeper"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestMsgServer_VoteTSS(t *testing.T) {
	t.Run("fail if node account not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		srv := keeper.NewMsgServerImpl(*k)

		_, err := srv.VoteTSS(ctx, &types.MsgVoteTSS{
			Signer:           sample.AccAddress(),
			TssPubkey:        sample.Tss_pell().TssPubkey,
			KeygenPellHeight: 42,
			Status:           chains.ReceiveStatus_SUCCESS,
		})
		require.ErrorIs(t, err, sdkerrors.ErrorInvalidSigner)
	})

	t.Run("fail if keygen is not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		srv := keeper.NewMsgServerImpl(*k)

		// setup state
		nodeAcc := sample.NodeAccount_pell()
		k.SetNodeAccount(ctx, *nodeAcc)

		_, err := srv.VoteTSS(ctx, &types.MsgVoteTSS{
			Signer:           nodeAcc.Operator,
			TssPubkey:        sample.Tss_pell().TssPubkey,
			KeygenPellHeight: 42,
			Status:           chains.ReceiveStatus_SUCCESS,
		})
		require.ErrorIs(t, err, types.ErrKeygenNotFound)
	})

	t.Run("fail if keygen already completed ", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		srv := keeper.NewMsgServerImpl(*k)

		// setup state
		nodeAcc := sample.NodeAccount_pell()
		keygen := sample.Keygen_pell(t)
		keygen.Status = types.KeygenStatus_SUCCESS
		k.SetNodeAccount(ctx, *nodeAcc)
		k.SetKeygen(ctx, *keygen)

		_, err := srv.VoteTSS(ctx, &types.MsgVoteTSS{
			Signer:           nodeAcc.Operator,
			TssPubkey:        sample.Tss_pell().TssPubkey,
			KeygenPellHeight: 42,
			Status:           chains.ReceiveStatus_SUCCESS,
		})
		require.ErrorIs(t, err, types.ErrKeygenCompleted)
	})

	t.Run("can create a new ballot, vote success and finalize", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		ctx = ctx.WithBlockHeight(42)
		srv := keeper.NewMsgServerImpl(*k)

		// setup state
		nodeAcc := sample.NodeAccount_pell()
		keygen := sample.Keygen_pell(t)
		keygen.Status = types.KeygenStatus_PENDING
		k.SetNodeAccount(ctx, *nodeAcc)
		k.SetKeygen(ctx, *keygen)

		// there is a single node account, so the ballot will be created and finalized in a single vote
		res, err := srv.VoteTSS(ctx, &types.MsgVoteTSS{
			Signer:           nodeAcc.Operator,
			TssPubkey:        sample.Tss_pell().TssPubkey,
			KeygenPellHeight: 42,
			Status:           chains.ReceiveStatus_SUCCESS,
		})
		require.NoError(t, err)

		// check response
		require.True(t, res.BallotCreated)
		require.True(t, res.VoteFinalized)
		require.True(t, res.KeygenSuccess)

		// check keygen updated
		newKeygen, found := k.GetKeygen(ctx)
		require.True(t, found)
		require.EqualValues(t, types.KeygenStatus_SUCCESS, newKeygen.Status)
		require.EqualValues(t, ctx.BlockHeight(), newKeygen.BlockNumber)
	})

	t.Run("can create a new ballot, vote failure and finalize", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		ctx = ctx.WithBlockHeight(42)
		srv := keeper.NewMsgServerImpl(*k)

		// setup state
		nodeAcc := sample.NodeAccount_pell()
		keygen := sample.Keygen_pell(t)
		keygen.Status = types.KeygenStatus_PENDING
		k.SetNodeAccount(ctx, *nodeAcc)
		k.SetKeygen(ctx, *keygen)

		// there is a single node account, so the ballot will be created and finalized in a single vote
		res, err := srv.VoteTSS(ctx, &types.MsgVoteTSS{
			Signer:           nodeAcc.Operator,
			TssPubkey:        sample.Tss_pell().TssPubkey,
			KeygenPellHeight: 42,
			Status:           chains.ReceiveStatus_FAILED,
		})
		require.NoError(t, err)

		// check response
		require.True(t, res.BallotCreated)
		require.True(t, res.VoteFinalized)
		require.False(t, res.KeygenSuccess)

		// check keygen updated
		newKeygen, found := k.GetKeygen(ctx)
		require.True(t, found)
		require.EqualValues(t, types.KeygenStatus_FAILED, newKeygen.Status)
		require.EqualValues(t, math.MaxInt64, newKeygen.BlockNumber)
	})

	t.Run("can create a new ballot, vote without finalizing, then add vote and finalizing", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		ctx = ctx.WithBlockHeight(42)
		srv := keeper.NewMsgServerImpl(*k)

		// setup state with 3 node accounts
		nodeAcc1 := sample.NodeAccount_pell()
		nodeAcc2 := sample.NodeAccount_pell()
		nodeAcc3 := sample.NodeAccount_pell()
		keygen := sample.Keygen_pell(t)
		keygen.Status = types.KeygenStatus_PENDING
		tss := sample.Tss_pell()
		k.SetNodeAccount(ctx, *nodeAcc1)
		k.SetNodeAccount(ctx, *nodeAcc2)
		k.SetNodeAccount(ctx, *nodeAcc3)
		k.SetKeygen(ctx, *keygen)

		// 1st vote: created ballot, but not finalized
		res, err := srv.VoteTSS(ctx, &types.MsgVoteTSS{
			Signer:           nodeAcc1.Operator,
			TssPubkey:        tss.TssPubkey,
			KeygenPellHeight: 42,
			Status:           chains.ReceiveStatus_SUCCESS,
		})
		require.NoError(t, err)

		// check response
		require.True(t, res.BallotCreated)
		require.False(t, res.VoteFinalized)
		require.False(t, res.KeygenSuccess)

		// check keygen not updated
		newKeygen, found := k.GetKeygen(ctx)
		require.True(t, found)
		require.EqualValues(t, types.KeygenStatus_PENDING, newKeygen.Status)

		// 2nd vote: already created ballot, and not finalized
		res, err = srv.VoteTSS(ctx, &types.MsgVoteTSS{
			Signer:           nodeAcc2.Operator,
			TssPubkey:        tss.TssPubkey,
			KeygenPellHeight: 42,
			Status:           chains.ReceiveStatus_SUCCESS,
		})
		require.NoError(t, err)

		// check response
		require.False(t, res.BallotCreated)
		require.False(t, res.VoteFinalized)
		require.False(t, res.KeygenSuccess)

		// check keygen not updated
		newKeygen, found = k.GetKeygen(ctx)
		require.True(t, found)
		require.EqualValues(t, types.KeygenStatus_PENDING, newKeygen.Status)

		// 3rd vote: finalize the ballot
		res, err = srv.VoteTSS(ctx, &types.MsgVoteTSS{
			Signer:           nodeAcc3.Operator,
			TssPubkey:        tss.TssPubkey,
			KeygenPellHeight: 42,
			Status:           chains.ReceiveStatus_SUCCESS,
		})
		require.NoError(t, err)

		// check response
		require.False(t, res.BallotCreated)
		require.True(t, res.VoteFinalized)
		require.True(t, res.KeygenSuccess)

		// check keygen not updated
		newKeygen, found = k.GetKeygen(ctx)
		require.True(t, found)
		require.EqualValues(t, types.KeygenStatus_SUCCESS, newKeygen.Status)
		require.EqualValues(t, ctx.BlockHeight(), newKeygen.BlockNumber)
	})

	t.Run("fail if voting fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		ctx = ctx.WithBlockHeight(42)
		srv := keeper.NewMsgServerImpl(*k)

		// setup state with two node accounts to not finalize the ballot
		nodeAcc := sample.NodeAccount_pell()
		keygen := sample.Keygen_pell(t)
		keygen.Status = types.KeygenStatus_PENDING
		k.SetNodeAccount(ctx, *nodeAcc)
		k.SetNodeAccount(ctx, *sample.NodeAccount_pell())
		k.SetKeygen(ctx, *keygen)

		// add a first vote
		res, err := srv.VoteTSS(ctx, &types.MsgVoteTSS{
			Signer:           nodeAcc.Operator,
			TssPubkey:        sample.Tss_pell().TssPubkey,
			KeygenPellHeight: 42,
			Status:           chains.ReceiveStatus_SUCCESS,
		})
		require.NoError(t, err)
		require.False(t, res.VoteFinalized)

		// vote again: voting should fail
		_, err = srv.VoteTSS(ctx, &types.MsgVoteTSS{
			Signer:           nodeAcc.Operator,
			TssPubkey:        sample.Tss_pell().TssPubkey,
			KeygenPellHeight: 42,
			Status:           chains.ReceiveStatus_SUCCESS,
		})
		require.ErrorIs(t, err, types.ErrUnableToAddVote)
	})
}
