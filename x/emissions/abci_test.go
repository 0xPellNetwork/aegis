package emissions_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/cmd/pellcored/config"
	"github.com/0xPellNetwork/aegis/pkg/coin"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	emissionsModule "github.com/0xPellNetwork/aegis/x/emissions"
	emissionstypes "github.com/0xPellNetwork/aegis/x/emissions/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

func TestBeginBlocker(t *testing.T) {
	t.Run("no observer distribution happens if emissions module account is empty", func(t *testing.T) {
		k, ctx, _, zk := keepertest.EmissionsKeeper(t)
		var ballotIdentifiers []string

		observerSet := sample.ObserverSet_pell(10)
		zk.ObserverKeeper.SetObserverSet(ctx, observerSet)

		ballotList := sample.BallotList_pell(10, observerSet.RelayerList)
		for _, ballot := range ballotList {
			zk.ObserverKeeper.SetBallot(ctx, &ballot)
			ballotIdentifiers = append(ballotIdentifiers, ballot.BallotIdentifier)
		}
		zk.ObserverKeeper.SetBallotList(ctx, &relayertypes.BallotListForHeight{
			Height:           0,
			BallotsIndexList: ballotIdentifiers,
		})
		for i := 0; i < 100; i++ {
			emissionsModule.BeginBlocker(ctx, *k)
			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
		}
		for _, observer := range observerSet.RelayerList {
			_, found := k.GetWithdrawableEmission(ctx, observer)
			require.False(t, found)
		}
	})
	t.Run("no validator distribution happens if emissions module account is empty", func(t *testing.T) {
		k, ctx, sk, _ := keepertest.EmissionsKeeper(t)
		feeCollectorAddress := sk.AuthKeeper.GetModuleAccount(ctx, types.FeeCollectorName).GetAddress()
		for i := 0; i < 100; i++ {
			emissionsModule.BeginBlocker(ctx, *k)
			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
		}
		require.True(t, sk.BankKeeper.GetBalance(ctx, feeCollectorAddress, config.BaseDenom).Amount.IsZero())
	})
	t.Run("tmp ctx is not committed if any of the distribution fails", func(t *testing.T) {
		k, ctx, sk, _ := keepertest.EmissionsKeeper(t)
		// Fund the emission pool to start the emission process
		err := sk.BankKeeper.MintCoins(ctx, emissionstypes.ModuleName, sdk.NewCoins(sdk.NewCoin(config.BaseDenom, sdkmath.NewInt(1000000000000))))
		require.NoError(t, err)
		// Setup module accounts for emission pools except for observer pool , so that the observer distribution fails
		_ = sk.AuthKeeper.GetModuleAccount(ctx, emissionstypes.UndistributedTssRewardsPool).GetAddress()
		feeCollectorAddress := sk.AuthKeeper.GetModuleAccount(ctx, types.FeeCollectorName).GetAddress()
		_ = sk.AuthKeeper.GetModuleAccount(ctx, emissionstypes.ModuleName).GetAddress()

		for i := 0; i < 100; i++ {
			// produce a block
			emissionsModule.BeginBlocker(ctx, *k)
			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
		}
		require.True(t, sk.BankKeeper.GetBalance(ctx, feeCollectorAddress, config.BaseDenom).Amount.IsZero())
		require.True(t, sk.BankKeeper.GetBalance(ctx, emissionstypes.EmissionsModuleAddress, config.BaseDenom).Amount.Equal(sdkmath.NewInt(1000000000000)))
	})
	t.Run("begin blocker returns early if validator distribution fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.EmissionKeeperWithMockOptions(t, keepertest.EmissionMockOptions{
			UseBankMock: true,
		})

		// Total block rewards is the fixed amount of rewards that are distributed
		totalBlockRewards, err := coin.GetApellDecFromAmountInPell(emissionstypes.BlockRewardsInPell)
		totalRewardCoins := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, totalBlockRewards.TruncateInt()))
		require.NoError(t, err)

		bankMock := keepertest.GetEmissionsBankMock(t, k)
		bankMock.On("GetBalance",
			ctx, mock.Anything, config.BaseDenom).
			Return(totalRewardCoins[0], nil).Once()

		// fail first distribution
		bankMock.On("SendCoinsFromModuleToModule",
			mock.Anything, emissionstypes.ModuleName, k.GetFeeCollector(), mock.Anything).
			Return(emissionstypes.ErrUnableToWithdrawEmissions).Once()
		emissionsModule.BeginBlocker(ctx, *k)

		bankMock.AssertNumberOfCalls(t, "SendCoinsFromModuleToModule", 1)
	})

	t.Run("begin blocker returns early if observer distribution fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.EmissionKeeperWithMockOptions(t, keepertest.EmissionMockOptions{
			UseBankMock: true,
		})
		// Total block rewards is the fixed amount of rewards that are distributed
		totalBlockRewards, err := coin.GetApellDecFromAmountInPell(emissionstypes.BlockRewardsInPell)
		totalRewardCoins := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, totalBlockRewards.TruncateInt()))
		require.NoError(t, err)

		bankMock := keepertest.GetEmissionsBankMock(t, k)
		bankMock.On("GetBalance",
			ctx, mock.Anything, config.BaseDenom).
			Return(totalRewardCoins[0], nil).Once()

		// allow first distribution
		bankMock.On("SendCoinsFromModuleToModule",
			mock.Anything, emissionstypes.ModuleName, k.GetFeeCollector(), mock.Anything).
			Return(nil).Once()

		// fail second distribution
		bankMock.On("SendCoinsFromModuleToModule",
			mock.Anything, emissionstypes.ModuleName, emissionstypes.UndistributedObserverRewardsPool, mock.Anything).
			Return(emissionstypes.ErrUnableToWithdrawEmissions).Once()
		emissionsModule.BeginBlocker(ctx, *k)

		bankMock.AssertNumberOfCalls(t, "SendCoinsFromModuleToModule", 2)
	})

	t.Run("begin blocker returns early if tss distribution fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.EmissionKeeperWithMockOptions(t, keepertest.EmissionMockOptions{
			UseBankMock: true,
		})
		// Total block rewards is the fixed amount of rewards that are distributed
		totalBlockRewards, err := coin.GetApellDecFromAmountInPell(emissionstypes.BlockRewardsInPell)
		totalRewardCoins := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, totalBlockRewards.TruncateInt()))
		require.NoError(t, err)

		bankMock := keepertest.GetEmissionsBankMock(t, k)
		bankMock.On("GetBalance",
			ctx, mock.Anything, config.BaseDenom).
			Return(totalRewardCoins[0], nil).Once()

		// allow first distribution
		bankMock.On("SendCoinsFromModuleToModule",
			mock.Anything, emissionstypes.ModuleName, k.GetFeeCollector(), mock.Anything).
			Return(nil).Once()

		// allow second distribution
		bankMock.On("SendCoinsFromModuleToModule",
			mock.Anything, emissionstypes.ModuleName, emissionstypes.UndistributedObserverRewardsPool, mock.Anything).
			Return(nil).Once()

		// fail third distribution
		bankMock.On("SendCoinsFromModuleToModule",
			mock.Anything, emissionstypes.ModuleName, emissionstypes.UndistributedTssRewardsPool, mock.Anything).
			Return(emissionstypes.ErrUnableToWithdrawEmissions).Once()
		emissionsModule.BeginBlocker(ctx, *k)

		bankMock.AssertNumberOfCalls(t, "SendCoinsFromModuleToModule", 3)
	})

	t.Run("successfully distribute rewards", func(t *testing.T) {
		numberOfTestBlocks := 100
		k, ctx, sk, zk := keepertest.EmissionsKeeper(t)
		observerSet := sample.ObserverSet_pell(10)
		zk.ObserverKeeper.SetObserverSet(ctx, observerSet)
		ballotList := sample.BallotList_pell(10, observerSet.RelayerList)

		// set the ballot list
		ballotIdentifiers := []string{}
		for _, ballot := range ballotList {
			zk.ObserverKeeper.SetBallot(ctx, &ballot)
			ballotIdentifiers = append(ballotIdentifiers, ballot.BallotIdentifier)
		}
		zk.ObserverKeeper.SetBallotList(ctx, &relayertypes.BallotListForHeight{
			Height:           0,
			BallotsIndexList: ballotIdentifiers,
		})

		// Total block rewards is the fixed amount of rewards that are distributed
		totalBlockRewards, err := coin.GetApellDecFromAmountInPell(emissionstypes.BlockRewardsInPell)
		totalRewardCoins := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, totalBlockRewards.TruncateInt()))
		require.NoError(t, err)
		// Fund the emission pool to start the emission process
		err = sk.BankKeeper.MintCoins(ctx, emissionstypes.ModuleName, totalRewardCoins)
		require.NoError(t, err)

		// Setup module accounts for emission pools
		undistributedObserverPoolAddress := sk.AuthKeeper.GetModuleAccount(ctx, emissionstypes.UndistributedObserverRewardsPool).GetAddress()
		undistributedTssPoolAddress := sk.AuthKeeper.GetModuleAccount(ctx, emissionstypes.UndistributedTssRewardsPool).GetAddress()
		feeCollecterAddress := sk.AuthKeeper.GetModuleAccount(ctx, types.FeeCollectorName).GetAddress()
		emissionPool := sk.AuthKeeper.GetModuleAccount(ctx, emissionstypes.ModuleName).GetAddress()

		blockRewards := emissionstypes.BlockReward
		observerRewardsForABlock := blockRewards.Mul(sdkmath.LegacyMustNewDecFromStr(k.GetParamsIfExists(ctx).ObserverEmissionPercentage)).TruncateInt()
		validatorRewardsForABlock := blockRewards.Mul(sdkmath.LegacyMustNewDecFromStr(k.GetParamsIfExists(ctx).ValidatorEmissionPercentage)).TruncateInt()
		tssSignerRewardsForABlock := blockRewards.Mul(sdkmath.LegacyMustNewDecFromStr(k.GetParamsIfExists(ctx).TssSignerEmissionPercentage)).TruncateInt()
		tssGasReserveForABlock := blockRewards.Mul(sdkmath.LegacyMustNewDecFromStr(k.GetParamsIfExists(ctx).TssGasEmissionPercentage)).TruncateInt()
		distributedRewards := observerRewardsForABlock.Add(validatorRewardsForABlock).Add(tssSignerRewardsForABlock).Add(tssGasReserveForABlock)

		require.True(t, blockRewards.TruncateInt().GT(distributedRewards))

		for i := 0; i < numberOfTestBlocks; i++ {
			emissionPoolBeforeBlockDistribution := sk.BankKeeper.GetBalance(ctx, emissionPool, config.BaseDenom).Amount
			// produce a block
			emissionsModule.BeginBlocker(ctx, *k)

			// require distribution amount
			emissionPoolBalanceAfterBlockDistribution := sk.BankKeeper.GetBalance(ctx, emissionPool, config.BaseDenom).Amount
			require.True(t, emissionPoolBeforeBlockDistribution.Sub(emissionPoolBalanceAfterBlockDistribution).Equal(distributedRewards))

			// totalDistributedTillCurrentBlock is the net amount of rewards distributed till the current block, this works in a unit test as the fees are not being collected by validators
			totalDistributedTillCurrentBlock := sk.BankKeeper.GetBalance(ctx, feeCollecterAddress, config.BaseDenom).Amount.
				Add(sk.BankKeeper.GetBalance(ctx, undistributedObserverPoolAddress, config.BaseDenom).Amount).
				Add(sk.BankKeeper.GetBalance(ctx, undistributedTssPoolAddress, config.BaseDenom).Amount)
			// require we are always under the max limit of block rewards
			require.True(t, totalRewardCoins.AmountOf(config.BaseDenom).
				Sub(totalDistributedTillCurrentBlock).GTE(sdkmath.ZeroInt()))

			ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
		}

		// We can simplify the calculation as the rewards are distributed equally among all the observers
		rewardPerUnit := observerRewardsForABlock.Quo(sdkmath.NewInt(int64(len(ballotList) * len(observerSet.RelayerList))))
		emissionAmount := rewardPerUnit.Mul(sdkmath.NewInt(int64(len(ballotList))))

		// Check if the rewards are distributed equally among all the observers
		for _, observer := range observerSet.RelayerList {
			observerEmission, found := k.GetWithdrawableEmission(ctx, observer)
			require.True(t, found)
			require.Equal(t, emissionAmount, observerEmission.Amount)
		}

		// Check pool balances after the distribution
		feeCollectorBalance := sk.BankKeeper.GetBalance(ctx, feeCollecterAddress, config.BaseDenom).Amount
		require.Equal(t, feeCollectorBalance, validatorRewardsForABlock.Mul(sdkmath.NewInt(int64(numberOfTestBlocks))))

		tssPoolBalances := sk.BankKeeper.GetBalance(ctx, undistributedTssPoolAddress, config.BaseDenom).Amount
		require.Equal(t, tssSignerRewardsForABlock.Mul(sdkmath.NewInt(int64(numberOfTestBlocks))).String(), tssPoolBalances.String())

		observerPoolBalances := sk.BankKeeper.GetBalance(ctx, undistributedObserverPoolAddress, config.BaseDenom).Amount
		require.Equal(t, observerRewardsForABlock.Mul(sdkmath.NewInt(int64(numberOfTestBlocks))).String(), observerPoolBalances.String())
	})
}

func TestDistributeObserverRewards(t *testing.T) {
	keepertest.SetConfig(false)
	k, ctx, _, _ := keepertest.EmissionsKeeper(t)
	observerSet := sample.ObserverSet_pell(4)

	tt := []struct {
		name                 string
		votes                [][]relayertypes.VoteType
		totalRewardsForBlock sdkmath.Int
		expectedRewards      map[string]int64
		ballotStatus         relayertypes.BallotStatus
		slashAmount          sdkmath.Int
	}{
		{
			name:  "all observers rewarded correctly",
			votes: [][]relayertypes.VoteType{{relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION}},
			// total reward units would be 4 as all votes match the ballot status
			totalRewardsForBlock: sdkmath.NewInt(100),
			expectedRewards: map[string]int64{
				observerSet.RelayerList[0]: 125,
				observerSet.RelayerList[1]: 125,
				observerSet.RelayerList[2]: 125,
				observerSet.RelayerList[3]: 125,
			},
			ballotStatus: relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION,
			slashAmount:  sdkmath.NewInt(25),
		},
		{
			name:  "one observer slashed",
			votes: [][]relayertypes.VoteType{{relayertypes.VoteType_FAILURE_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION}},
			// total reward units would be 3 as 3 votes match the ballot status
			totalRewardsForBlock: sdkmath.NewInt(75),
			expectedRewards: map[string]int64{
				observerSet.RelayerList[0]: 75,
				observerSet.RelayerList[1]: 125,
				observerSet.RelayerList[2]: 125,
				observerSet.RelayerList[3]: 125,
			},
			ballotStatus: relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION,
			slashAmount:  sdkmath.NewInt(25),
		},
		{
			name:  "all observer slashed",
			votes: [][]relayertypes.VoteType{{relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION}},
			// total reward units would be 0 as no votes match the ballot status
			totalRewardsForBlock: sdkmath.NewInt(100),
			expectedRewards: map[string]int64{
				observerSet.RelayerList[0]: 75,
				observerSet.RelayerList[1]: 75,
				observerSet.RelayerList[2]: 75,
				observerSet.RelayerList[3]: 75,
			},
			ballotStatus: relayertypes.BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION,
			slashAmount:  sdkmath.NewInt(25),
		},
		{
			name:  "slashed to zero if slash amount is greater than available emissions",
			votes: [][]relayertypes.VoteType{{relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION}},
			// total reward units would be 0 as no votes match the ballot status
			totalRewardsForBlock: sdkmath.NewInt(100),
			expectedRewards: map[string]int64{
				observerSet.RelayerList[0]: 0,
				observerSet.RelayerList[1]: 0,
				observerSet.RelayerList[2]: 0,
				observerSet.RelayerList[3]: 0,
			},
			ballotStatus: relayertypes.BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION,
			slashAmount:  sdkmath.NewInt(2500),
		},
		{
			name: "withdraw able emissions unchanged if rewards and slashes are equal",
			votes: [][]relayertypes.VoteType{
				{relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION},
				{relayertypes.VoteType_FAILURE_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION, relayertypes.VoteType_SUCCESS_OBSERVATION},
			},
			// total reward units would be 7 as 7 votes match the ballot status, including both ballots
			totalRewardsForBlock: sdkmath.NewInt(70),
			expectedRewards: map[string]int64{
				observerSet.RelayerList[0]: 100,
				observerSet.RelayerList[1]: 120,
				observerSet.RelayerList[2]: 120,
				observerSet.RelayerList[3]: 120,
			},
			ballotStatus: relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION,
			slashAmount:  sdkmath.NewInt(25),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			for _, observer := range observerSet.RelayerList {
				k.SetWithdrawableEmission(ctx, emissionstypes.WithdrawableEmissions{
					Address: observer,
					Amount:  sdkmath.NewInt(100),
				})
			}

			// Keeper initialization
			k, ctx, sk, zk := keepertest.EmissionsKeeper(t)
			zk.ObserverKeeper.SetObserverSet(ctx, observerSet)

			// Total block rewards is the fixed amount of rewards that are distributed
			totalBlockRewards, err := coin.GetApellDecFromAmountInPell(emissionstypes.BlockRewardsInPell)
			totalRewardCoins := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, totalBlockRewards.TruncateInt()))
			require.NoError(t, err)

			// Fund the emission pool to start the emission process
			err = sk.BankKeeper.MintCoins(ctx, emissionstypes.ModuleName, totalRewardCoins)
			require.NoError(t, err)

			// Set starting emission for all observers to 100 so that we can calculate the rewards and slashes
			for _, observer := range observerSet.RelayerList {
				k.SetWithdrawableEmission(ctx, emissionstypes.WithdrawableEmissions{
					Address: observer,
					Amount:  sdkmath.NewInt(100),
				})
			}

			// Set the params
			params := emissionstypes.DefaultParams()
			params.ObserverSlashAmount = tc.slashAmount
			k.SetParams(ctx, params)

			// Set the ballot list
			ballotIdentifiers := []string{}
			for i, votes := range tc.votes {
				ballot := relayertypes.Ballot{
					BallotIdentifier: "ballot" + string(rune(i)),
					BallotStatus:     tc.ballotStatus,
					VoterList:        observerSet.RelayerList,
					Votes:            votes,
				}
				zk.ObserverKeeper.SetBallot(ctx, &ballot)
				ballotIdentifiers = append(ballotIdentifiers, ballot.BallotIdentifier)
			}
			zk.ObserverKeeper.SetBallotList(ctx, &relayertypes.BallotListForHeight{
				Height:           0,
				BallotsIndexList: ballotIdentifiers,
			})
			ctx = ctx.WithBlockHeight(100)

			// Distribute the rewards and check if the rewards are distributed correctly
			err = emissionsModule.DistributeObserverRewards(ctx, tc.totalRewardsForBlock, *k, tc.slashAmount)
			require.NoError(t, err)

			for i, observer := range observerSet.RelayerList {
				observerEmission, found := k.GetWithdrawableEmission(ctx, observer)
				require.True(t, found, "withdrawable emission not found for observer %d", i)
				require.Equal(t, tc.expectedRewards[observer], observerEmission.Amount.Int64(), "invalid withdrawable emission for observer %d", i)
			}
		})
	}
}
