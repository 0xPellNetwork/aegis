package keeper_test

import (
	"math/rand"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/relayer/keeper"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestMsgServer_AddBlameVote(t *testing.T) {
	t.Run("should error if supported chain not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		srv := keeper.NewMsgServerImpl(*k)

		res, err := srv.AddBlameVote(ctx, &types.MsgAddBlameVote{
			ChainId: 1,
		})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should error if not tombstoned observer", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		srv := keeper.NewMsgServerImpl(*k)

		chainId := getValidEthChainIDWithIndex(t, 0)
		setSupportedChain(ctx, *k, chainId)

		res, err := srv.AddBlameVote(ctx, &types.MsgAddBlameVote{
			ChainId: chainId,
		})
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should return response and set blame if finalizing vote", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		srv := keeper.NewMsgServerImpl(*k)

		chainId := getValidEthChainIDWithIndex(t, 0)
		setSupportedChain(ctx, *k, chainId)

		r := rand.New(rand.NewSource(9))
		// Set validator in the store
		validator := sample.Validator(t, r)
		k.GetStakingKeeper().SetValidator(ctx, validator)
		consAddress, err := validator.GetConsAddr()
		require.NoError(t, err)
		k.GetSlashingKeeper().SetValidatorSigningInfo(ctx, consAddress, slashingtypes.ValidatorSigningInfo{
			Address:             string(consAddress),
			StartHeight:         0,
			JailedUntil:         ctx.BlockHeader().Time.Add(1000000 * time.Second),
			Tombstoned:          false,
			MissedBlocksCounter: 1,
		})

		accAddressOfValidator, err := types.GetAccAddressFromOperatorAddress(validator.OperatorAddress)
		require.NoError(t, err)

		k.SetObserverSet(ctx, types.RelayerSet{
			RelayerList: []string{accAddressOfValidator.String()},
		})

		blameInfo := sample.BlameRecord_pell(t, "index")
		res, err := srv.AddBlameVote(ctx, &types.MsgAddBlameVote{
			Signer:    accAddressOfValidator.String(),
			ChainId:   chainId,
			BlameInfo: blameInfo,
		})
		require.NoError(t, err)
		require.Equal(t, &types.MsgAddBlameVoteResponse{}, res)

		blame, found := k.GetBlame(ctx, blameInfo.Index)
		require.True(t, found)
		require.Equal(t, blameInfo, blame)
	})

	t.Run("should error if add vote fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		srv := keeper.NewMsgServerImpl(*k)

		chainId := getValidEthChainIDWithIndex(t, 1)
		setSupportedChain(ctx, *k, chainId)

		r := rand.New(rand.NewSource(9))
		// Set validator in the store
		validator := sample.Validator(t, r)
		k.GetStakingKeeper().SetValidator(ctx, validator)
		consAddress, err := validator.GetConsAddr()
		require.NoError(t, err)
		k.GetSlashingKeeper().SetValidatorSigningInfo(ctx, consAddress, slashingtypes.ValidatorSigningInfo{
			Address:             string(consAddress),
			StartHeight:         0,
			JailedUntil:         ctx.BlockHeader().Time.Add(1000000 * time.Second),
			Tombstoned:          false,
			MissedBlocksCounter: 1,
		})

		accAddressOfValidator, err := types.GetAccAddressFromOperatorAddress(validator.OperatorAddress)
		require.NoError(t, err)

		k.SetObserverSet(ctx, types.RelayerSet{
			RelayerList: []string{accAddressOfValidator.String(), "Observer2"},
		})
		blameInfo := sample.BlameRecord_pell(t, "index")
		vote := &types.MsgAddBlameVote{
			Signer:    accAddressOfValidator.String(),
			ChainId:   chainId,
			BlameInfo: blameInfo,
		}
		ballot := types.Ballot{
			Index:            vote.Digest(),
			BallotIdentifier: vote.Digest(),
			VoterList:        []string{accAddressOfValidator.String()},
			Votes:            []types.VoteType{types.VoteType_SUCCESS_OBSERVATION},
			BallotStatus:     types.BallotStatus_BALLOT_IN_PROGRESS,
			BallotThreshold:  sdkmath.LegacyNewDec(2),
		}
		k.SetBallot(ctx, &ballot)

		_, err = srv.AddBlameVote(ctx, vote)
		require.Error(t, err)
	})

	t.Run("should return response and not set blame if not finalizing vote", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		srv := keeper.NewMsgServerImpl(*k)

		chainId := getValidEthChainIDWithIndex(t, 1)
		setSupportedChain(ctx, *k, chainId)

		r := rand.New(rand.NewSource(9))
		// Set validator in the store
		validator := sample.Validator(t, r)
		k.GetStakingKeeper().SetValidator(ctx, validator)
		consAddress, err := validator.GetConsAddr()
		require.NoError(t, err)
		k.GetSlashingKeeper().SetValidatorSigningInfo(ctx, consAddress, slashingtypes.ValidatorSigningInfo{
			Address:             string(consAddress),
			StartHeight:         0,
			JailedUntil:         ctx.BlockHeader().Time.Add(1000000 * time.Second),
			Tombstoned:          false,
			MissedBlocksCounter: 1,
		})

		accAddressOfValidator, err := types.GetAccAddressFromOperatorAddress(validator.OperatorAddress)
		require.NoError(t, err)

		k.SetObserverSet(ctx, types.RelayerSet{
			RelayerList: []string{accAddressOfValidator.String(), "Observer2"},
		})
		blameInfo := sample.BlameRecord_pell(t, "index")
		vote := &types.MsgAddBlameVote{
			Signer:    accAddressOfValidator.String(),
			ChainId:   chainId,
			BlameInfo: blameInfo,
		}
		ballot := types.Ballot{
			Index:            vote.Digest(),
			BallotIdentifier: vote.Digest(),
			VoterList:        []string{accAddressOfValidator.String()},
			Votes:            []types.VoteType{types.VoteType_NOT_YET_VOTED},
			BallotStatus:     types.BallotStatus_BALLOT_IN_PROGRESS,
			BallotThreshold:  sdkmath.LegacyNewDec(2),
		}
		k.SetBallot(ctx, &ballot)

		res, err := srv.AddBlameVote(ctx, vote)
		require.NoError(t, err)
		require.Equal(t, &types.MsgAddBlameVoteResponse{}, res)

		_, found := k.GetBlame(ctx, blameInfo.Index)
		require.False(t, found)
	})
}
