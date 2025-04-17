package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/pkg/coin"
	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestKeeper_VoteOnInboundBallot(t *testing.T) {

	t.Run("fail if inbound not enabled", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)

		k.SetCrosschainFlags(ctx, types.CrosschainFlags{
			IsInboundEnabled: false,
		})

		_, _, err := k.VoteOnInboundBallot(
			ctx,
			getValidEthChainIDWithIndex(t, 0),
			chains.PellPrivnetChain().Id,
			coin.CoinType_ERC20,
			sample.AccAddress(),
			"index",
			"inTxHash",
		)

		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInboundDisabled)
	})

	t.Run("fail if sender chain not supported", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)

		k.SetCrosschainFlags(ctx, types.CrosschainFlags{
			IsInboundEnabled: true,
		})
		k.SetChainParamsList(ctx, types.ChainParamsList{})

		_, _, err := k.VoteOnInboundBallot(
			ctx,
			getValidEthChainIDWithIndex(t, 0),
			chains.PellPrivnetChain().Id,
			coin.CoinType_ERC20,
			sample.AccAddress(),
			"index",
			"inTxHash",
		)
		require.NotNil(t, err)
		require.ErrorContains(t, err, types.ErrSupportedChains.Error())

		// set the chain but not supported
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				{
					ChainId:     getValidEthChainIDWithIndex(t, 0),
					IsSupported: false,
				},
			},
		})

		_, _, err = k.VoteOnInboundBallot(
			ctx,
			getValidEthChainIDWithIndex(t, 0),
			chains.PellPrivnetChain().Id,
			coin.CoinType_ERC20,
			sample.AccAddress(),
			"index",
			"inTxHash",
		)
		require.Error(t, err)
		require.ErrorContains(t, err, types.ErrSupportedChains.Error())
	})

	t.Run("fail if not authorized", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)

		k.SetCrosschainFlags(ctx, types.CrosschainFlags{
			IsInboundEnabled: true,
		})
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				{
					ChainId:     getValidEthChainIDWithIndex(t, 0),
					IsSupported: true,
				},
			},
		})
		k.SetObserverSet(ctx, types.RelayerSet{})

		_, _, err := k.VoteOnInboundBallot(
			ctx,
			getValidEthChainIDWithIndex(t, 0),
			chains.PellPrivnetChain().Id,
			coin.CoinType_ERC20,
			sample.AccAddress(),
			"index",
			"inTxHash",
		)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrNotObserver)
	})

	t.Run("fail if receiver chain not supported", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMocksAll)

		observer := sample.AccAddress()
		stakingMock := keepertest.GetRelayerStakingMock(t, k)
		slashingMock := keepertest.GetRelayerSlashingMock(t, k)

		k.SetCrosschainFlags(ctx, types.CrosschainFlags{
			IsInboundEnabled: true,
		})
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				{
					ChainId:     getValidEthChainIDWithIndex(t, 0),
					IsSupported: true,
				},
			},
		})
		k.SetObserverSet(ctx, types.RelayerSet{
			RelayerList: []string{observer},
		})
		stakingMock.MockGetValidator(sample.Validator(t, sample.Rand()))
		slashingMock.MockIsTombstoned(false)

		_, _, err := k.VoteOnInboundBallot(
			ctx,
			getValidEthChainIDWithIndex(t, 0),
			chains.PellPrivnetChain().Id,
			coin.CoinType_ERC20,
			observer,
			"index",
			"inTxHash",
		)
		require.Error(t, err)
		require.ErrorContains(t, err, types.ErrSupportedChains.Error())

		// set the chain but not supported
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				{
					ChainId:     getValidEthChainIDWithIndex(t, 0),
					IsSupported: true,
				},
				{
					ChainId:     chains.PellPrivnetChain().Id,
					IsSupported: false,
				},
			},
		})
		stakingMock.MockGetValidator(sample.Validator(t, sample.Rand()))
		slashingMock.MockIsTombstoned(false)

		_, _, err = k.VoteOnInboundBallot(
			ctx,
			getValidEthChainIDWithIndex(t, 0),
			chains.PellPrivnetChain().Id,
			coin.CoinType_ERC20,
			observer,
			"index",
			"inTxHash",
		)
		require.Error(t, err)
		require.ErrorContains(t, err, types.ErrSupportedChains.Error())
	})

	t.Run("can add vote and create ballot", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMocksAll)

		observer := sample.AccAddress()
		stakingMock := keepertest.GetRelayerStakingMock(t, k)
		slashingMock := keepertest.GetRelayerSlashingMock(t, k)

		k.SetCrosschainFlags(ctx, types.CrosschainFlags{
			IsInboundEnabled: true,
		})
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				{
					ChainId:     getValidEthChainIDWithIndex(t, 0),
					IsSupported: true,
				},
				{
					ChainId:     getValidEthChainIDWithIndex(t, 1),
					IsSupported: true,
				},
			},
		})
		k.SetObserverSet(ctx, types.RelayerSet{
			RelayerList: []string{observer},
		})
		stakingMock.MockGetValidator(sample.Validator(t, sample.Rand()))
		slashingMock.MockIsTombstoned(false)

		isFinalized, isNew, err := k.VoteOnInboundBallot(
			ctx,
			getValidEthChainIDWithIndex(t, 0),
			getValidEthChainIDWithIndex(t, 1),
			coin.CoinType_ERC20,
			observer,
			"index",
			"inTxHash",
		)
		require.NoError(t, err)

		// ballot should be finalized since there is only one observer
		require.True(t, isFinalized)
		require.True(t, isNew)
	})

	t.Run("fail if can not add vote", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMocksAll)

		observer := sample.AccAddress()
		stakingMock := keepertest.GetRelayerStakingMock(t, k)
		slashingMock := keepertest.GetRelayerSlashingMock(t, k)

		k.SetCrosschainFlags(ctx, types.CrosschainFlags{
			IsInboundEnabled: true,
		})
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				{
					ChainId:     getValidEthChainIDWithIndex(t, 0),
					IsSupported: true,
				},
				{
					ChainId:     getValidEthChainIDWithIndex(t, 1),
					IsSupported: true,
				},
			},
		})
		k.SetObserverSet(ctx, types.RelayerSet{
			RelayerList: []string{observer},
		})
		stakingMock.MockGetValidator(sample.Validator(t, sample.Rand()))
		slashingMock.MockIsTombstoned(false)
		ballot := types.Ballot{
			Index:            "index",
			BallotIdentifier: "index",
			VoterList:        []string{observer},
			// already voted
			Votes:           []types.VoteType{types.VoteType_SUCCESS_OBSERVATION},
			BallotStatus:    types.BallotStatus_BALLOT_IN_PROGRESS,
			BallotThreshold: sdkmath.LegacyNewDec(2),
		}
		k.SetBallot(ctx, &ballot)
		isFinalized, isNew, err := k.VoteOnInboundBallot(
			ctx,
			getValidEthChainIDWithIndex(t, 0),
			getValidEthChainIDWithIndex(t, 1),
			coin.CoinType_ERC20,
			observer,
			"index",
			"inTxHash",
		)
		require.Error(t, err)
		require.False(t, isFinalized)
		require.False(t, isNew)
	})

	t.Run("can add vote and create ballot without finalizing ballot", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMocksAll)

		observer := sample.AccAddress()
		stakingMock := keepertest.GetRelayerStakingMock(t, k)
		slashingMock := keepertest.GetRelayerSlashingMock(t, k)

		// threshold high enough to not finalize ballot
		threshold, err := sdkmath.LegacyNewDecFromStr("0.7")
		require.NoError(t, err)

		k.SetCrosschainFlags(ctx, types.CrosschainFlags{
			IsInboundEnabled: true,
		})
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				{
					ChainId:         getValidEthChainIDWithIndex(t, 0),
					IsSupported:     true,
					BallotThreshold: threshold,
				},
				{
					ChainId:         getValidEthChainIDWithIndex(t, 1),
					IsSupported:     true,
					BallotThreshold: threshold,
				},
			},
		})
		k.SetObserverSet(ctx, types.RelayerSet{
			RelayerList: []string{
				observer,
				sample.AccAddress(),
			},
		})
		stakingMock.MockGetValidator(sample.Validator(t, sample.Rand()))
		slashingMock.MockIsTombstoned(false)

		isFinalized, isNew, err := k.VoteOnInboundBallot(
			ctx,
			getValidEthChainIDWithIndex(t, 0),
			getValidEthChainIDWithIndex(t, 1),
			coin.CoinType_ERC20,
			observer,
			"index",
			"inTxHash",
		)
		require.NoError(t, err)

		// ballot should be finalized since there is only one observer
		require.False(t, isFinalized)
		require.True(t, isNew)
	})

	t.Run("can add vote to an existing ballot", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMocksAll)

		observer := sample.AccAddress()
		stakingMock := keepertest.GetRelayerStakingMock(t, k)
		slashingMock := keepertest.GetRelayerSlashingMock(t, k)

		k.SetCrosschainFlags(ctx, types.CrosschainFlags{
			IsInboundEnabled: true,
		})
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				{
					ChainId:     getValidEthChainIDWithIndex(t, 0),
					IsSupported: true,
				},
				{
					ChainId:     getValidEthChainIDWithIndex(t, 1),
					IsSupported: true,
				},
			},
		})
		k.SetObserverSet(ctx, types.RelayerSet{
			RelayerList: []string{observer},
		})
		stakingMock.MockGetValidator(sample.Validator(t, sample.Rand()))
		slashingMock.MockIsTombstoned(false)

		// set a ballot
		threshold, err := sdkmath.LegacyNewDecFromStr("0.7")
		require.NoError(t, err)
		ballot := types.Ballot{
			Index:            "index",
			BallotIdentifier: "index",
			VoterList: []string{
				sample.AccAddress(),
				sample.AccAddress(),
				observer,
				sample.AccAddress(),
				sample.AccAddress(),
			},
			Votes:           types.CreateVotes(5),
			ObservationType: types.ObservationType_IN_BOUND_TX,
			BallotThreshold: threshold,
			BallotStatus:    types.BallotStatus_BALLOT_IN_PROGRESS,
		}
		k.SetBallot(ctx, &ballot)

		isFinalized, isNew, err := k.VoteOnInboundBallot(
			ctx,
			getValidEthChainIDWithIndex(t, 0),
			getValidEthChainIDWithIndex(t, 1),
			coin.CoinType_ERC20,
			observer,
			"index",
			"inTxHash",
		)
		require.NoError(t, err)

		// ballot should not be finalized as the threshold is not reached
		require.False(t, isFinalized)
		require.False(t, isNew)
	})

	t.Run("can add vote to an existing ballot and finalize ballot", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeperWithMocks(t, keepertest.RelayerMocksAll)

		observer := sample.AccAddress()
		stakingMock := keepertest.GetRelayerStakingMock(t, k)
		slashingMock := keepertest.GetRelayerSlashingMock(t, k)

		k.SetCrosschainFlags(ctx, types.CrosschainFlags{
			IsInboundEnabled: true,
		})
		k.SetChainParamsList(ctx, types.ChainParamsList{
			ChainParams: []*types.ChainParams{
				{
					ChainId:     getValidEthChainIDWithIndex(t, 0),
					IsSupported: true,
				},
				{
					ChainId:     getValidEthChainIDWithIndex(t, 1),
					IsSupported: true,
				},
			},
		})
		k.SetObserverSet(ctx, types.RelayerSet{
			RelayerList: []string{observer},
		})
		stakingMock.MockGetValidator(sample.Validator(t, sample.Rand()))
		slashingMock.MockIsTombstoned(false)

		// set a ballot
		threshold, err := sdkmath.LegacyNewDecFromStr("0.1")
		require.NoError(t, err)
		ballot := types.Ballot{
			Index:            "index",
			BallotIdentifier: "index",
			VoterList: []string{
				observer,
				sample.AccAddress(),
				sample.AccAddress(),
			},
			Votes:           types.CreateVotes(3),
			ObservationType: types.ObservationType_IN_BOUND_TX,
			BallotThreshold: threshold,
			BallotStatus:    types.BallotStatus_BALLOT_IN_PROGRESS,
		}
		k.SetBallot(ctx, &ballot)

		isFinalized, isNew, err := k.VoteOnInboundBallot(
			ctx,
			getValidEthChainIDWithIndex(t, 0),
			getValidEthChainIDWithIndex(t, 1),
			coin.CoinType_ERC20,
			observer,
			"index",
			"inTxHash",
		)
		require.NoError(t, err)

		// ballot should not be finalized as the threshold is not reached
		require.True(t, isFinalized)
		require.False(t, isNew)
	})
}
