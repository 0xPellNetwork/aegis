package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	types2 "github.com/cosmos/cosmos-sdk/codec/types"
	sdked25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	mocks "github.com/pell-chain/pellcore/testutil/keeper/mocks/xsecurity"
	restakingtypes "github.com/pell-chain/pellcore/x/restaking/types"
	"github.com/pell-chain/pellcore/x/xsecurity/keeper"
	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

const (
	BlocksPerEpoch       = int64(10)
	PowerReduction       = int64(1000000)
	OneLSTToken          = int64(1000000000000000000)
	DefaultMaxValidators = 3 // Number of active validators
	TotalValidators      = 3 // Total number of validators
	DefaultTokens        = 1000 * PowerReduction
)

// ValidatorInfo holds validator data
type ValidatorInfo struct {
	PrivKey *sdked25519.PrivKey
	PubKey  cryptotypes.PubKey
	ValAddr sdk.ValAddress
}

// setupTestEnv creates a fresh environment for each test case: new mocks, context,
// default validators and keeper mocks.
func setupTestEnv(t *testing.T, blockHeight int64) (
	mocks *keeper.Keeper,
	ctx sdk.Context,
	validators []stakingtypes.Validator,
	validatorInfos []ValidatorInfo,
	restakingKeeper *mocks.XSecurityRestakingKeeper,
	stakingKeeper *mocks.XSecurityStakingKeeper,
) {
	// Initialize test environment
	mocks, ctx = keepertest.XSecurityKeeperWithMocks(t, keepertest.XSecurityMocksAll)
	restakingKeeper = keepertest.GetXSecurityRestakingMock(t, mocks)
	stakingKeeper = keepertest.GetXSecurityStakingMock(t, mocks)

	// Setup block height and enable LST staking
	ctx = setupBlockHeight(ctx, mocks, blockHeight)

	// Set LST voting power ratio
	mocks.SetLSTVotingPowerRatio(ctx, sdkmath.NewInt(50), sdkmath.NewInt(100))

	// Create validators slice and validatorInfos slice
	validators = make([]stakingtypes.Validator, TotalValidators)
	validatorInfos = make([]ValidatorInfo, TotalValidators)

	return
}

// setupBlockHeight sets block height and enables LST staking
func setupBlockHeight(ctx sdk.Context, mocks *keeper.Keeper, blockHeight int64) sdk.Context {
	mocks.SetLSTStakingEnabled(ctx, true)
	mocks.SetBlocksPerEpoch(ctx, uint64(BlocksPerEpoch))
	mocks.SetEpochNumber(ctx, uint64(blockHeight/BlocksPerEpoch))

	return ctx.WithBlockHeight(blockHeight)
}

// initStakingValidatorsInfo initializes a default validator set:
// total 3 active validators with a total native voting power of 3000
func initStakingValidatorsInfo(t *testing.T, ctx sdk.Context, validators []stakingtypes.Validator, validatorInfos []ValidatorInfo, stakingKeeper *mocks.XSecurityStakingKeeper) {
	for i := 0; i < TotalValidators; i++ {
		privKey := sdked25519.GenPrivKey()
		sdkPubKey := privKey.PubKey()
		valAddr := sdk.ValAddress(sdkPubKey.Address())

		anyPk, err := types2.NewAnyWithValue(sdkPubKey)
		require.NoError(t, err)

		validator := stakingtypes.Validator{
			OperatorAddress: valAddr.String(),
			ConsensusPubkey: anyPk,
			Status:          stakingtypes.Bonded,
			Tokens:          sdkmath.NewInt(DefaultTokens),
			DelegatorShares: sdkmath.LegacyNewDec(DefaultTokens),
		}
		validators[i] = validator

		validatorInfos[i] = ValidatorInfo{
			PrivKey: privKey,
			PubKey:  sdkPubKey,
			ValAddr: valAddr,
		}

		// Only the first DefaultMaxValidators are set as active
		if i < DefaultMaxValidators {
			stakingKeeper.On("GetValidator", ctx, valAddr).Return(validator, nil).Maybe()
		}
	}
	stakingKeeper.On("GetBondedValidatorsByPower", ctx).
		Return(validators[:DefaultMaxValidators], nil).Maybe()
	stakingKeeper.On("PowerReduction", ctx).
		Return(sdkmath.NewInt(PowerReduction)).Maybe()
}

// setupStakingValidatorsInfo updates mock returns to reflect a changed active validator set
func setupStakingValidatorsInfo(ctx sdk.Context, validators []stakingtypes.Validator, validatorInfos []ValidatorInfo, max int64, stakingKeeper *mocks.XSecurityStakingKeeper) {
	for k := range validators {
		valAddr := validatorInfos[k].ValAddr
		stakingKeeper.On("GetValidator", ctx, valAddr).Return(validators[k], nil).Maybe()
	}
	stakingKeeper.On("GetBondedValidatorsByPower", ctx).
		Return(validators[:max], nil).Maybe()
	stakingKeeper.On("PowerReduction", ctx).
		Return(sdkmath.NewInt(PowerReduction)).Maybe()
}

// mockRestakingSharesInfo updates existing operator shares info
func mockRestakingSharesInfo(ctx sdk.Context, restakingKeeper *mocks.XSecurityRestakingKeeper, allShares []*restakingtypes.OperatorShares, operatorAddress string, shares int64) {
	for i, share := range allShares {
		if share.Operator == operatorAddress {
			allShares[i].Shares = sdkmath.NewInt(shares)
		}
	}
	restakingKeeper.On("GetAllShares", ctx).Return(allShares)
}

// setupGroupInfo initializes group info with a pool address
func setupGroupInfo(ctx sdk.Context, mocks *keeper.Keeper, poolAddress string) {
	mocks.SetGroupInfo(ctx, &types.LSTGroupInfo{
		GroupNumber: 0,
		PoolParams: []*restakingtypes.PoolParams{{
			ChainId:    0,
			Pool:       poolAddress,
			Multiplier: 1,
		}},
		OperatorSetParam:   nil,
		MinimumStake:       sdkmath.Int{},
		GroupEjectionParam: nil,
	})
}

// addStakingSharesInfo adds new operator shares info
func addStakingSharesInfo(allShares []*restakingtypes.OperatorShares, operatorAddress, poolAddress string, shares int64) []*restakingtypes.OperatorShares {
	return append(allShares, &restakingtypes.OperatorShares{
		Operator: operatorAddress,
		Strategy: poolAddress,
		Shares:   sdkmath.NewInt(shares),
	})
}

// setupOperatorRegistration sets up operator registration in xsecurity module
func setupOperatorRegistration(ctx sdk.Context, mocks *keeper.Keeper, operatorAddress, validatorAddress string) {
	mocks.AddOperatorRegistration(ctx, &types.LSTOperatorRegistration{
		OperatorAddress:  operatorAddress,
		ValidatorAddress: validatorAddress,
	})
}

// generateEVMAddress creates a random EVM address (placeholder implementation)
func generateEVMAddress() string {
	return "0x" + ed25519.GenPrivKey().PubKey().Address().String()
}

func TestLSTStaking(t *testing.T) {
	t.Run("should return nil validatorUpdates when LST staking is disabled", func(t *testing.T) {
		// Get a fresh environment
		mocks, ctx, validators, validatorInfos, _, stakingKeeper := setupTestEnv(t, 10)
		// Setup block height and enable LST staking
		ctx = setupBlockHeight(ctx, mocks, 10)
		// Initialize a default validator set: total 3 active validators with a total native voting power of 3000
		initStakingValidatorsInfo(t, ctx, validators, validatorInfos, stakingKeeper)

		mocks.SetLSTStakingEnabled(ctx, false)

		_, exist := mocks.GetLSTStakingEnabled(ctx)
		require.False(t, exist) // 'exist' defaults to false, as Protobuf omits serialization of fields set to their default values.

		number, exists := mocks.GetEpochNumber(ctx)
		require.True(t, exists)
		require.Equal(t, uint64(1), number)

		validatorUpdates, err := mocks.ProcessEpoch(ctx)
		require.NoError(t, err)
		require.Nil(t, validatorUpdates)
	})

	t.Run("should return empty validatorUpdates when no bond operator to validator", func(t *testing.T) {
		// Get a fresh environment
		mocks, ctx, validators, _, restakingKeeper, stakingKeeper := setupTestEnv(t, 10)
		ctx = setupBlockHeight(ctx, mocks, 10)
		// Initialize a default validator set: total 3 active validators with a total native voting power of 3000
		initStakingValidatorsInfo(t, ctx, validators, make([]ValidatorInfo, TotalValidators), stakingKeeper)

		operatorAddress := generateEVMAddress()
		poolAddress := generateEVMAddress()
		validatorAddress := generateEVMAddress()

		setupGroupInfo(ctx, mocks, poolAddress)
		var allShares []*restakingtypes.OperatorShares
		allShares = addStakingSharesInfo(allShares, operatorAddress, poolAddress, 1*OneLSTToken)
		setupOperatorRegistration(ctx, mocks, operatorAddress, validatorAddress)
		restakingKeeper.On("GetAllShares", ctx).Return(allShares)

		err := mocks.SnapshotSharesFromRestakingModuleByEpoch(ctx)
		require.NoError(t, err)

		changedShare, err := mocks.GetChangedSharesByEpoch(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, len(changedShare))

		votingPower, err := mocks.CalcNativeVotingPower(ctx, validators)
		require.NoError(t, err)
		require.Equal(t, int64(TotalValidators*1000), votingPower.Int64())

		power, exists := mocks.GetLastNativeVotingPower(ctx)
		require.False(t, exists)
		require.Equal(t, int64(0), power)

		validatorUpdates, err := mocks.ProcessEpoch(ctx)
		require.NoError(t, err)
		require.Equal(t, 0, len(validatorUpdates))
	})

	t.Run("should return non-empty validatorUpdates", func(t *testing.T) {
		// Declare a variable to store all operator shares across epochs
		var allShares []*restakingtypes.OperatorShares

		//
		// Epoch 1:
		//   - Native validators: v1, v2, v3
		//   - LST operators: v2, v3, v4
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		// Action: Default status
		// Expected voting power:
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*1/2 lst = 1750
		//   - v3 = 1000 native + 1500*1/2 lst = 1750
		//   - v4 = 0
		//
		mocks, ctx, validators, validatorInfos, restakingKeeper, stakingKeeper := setupTestEnv(t, 10)
		ctx = setupBlockHeight(ctx, mocks, 10)
		initStakingValidatorsInfo(t, ctx, validators, validatorInfos, stakingKeeper)

		poolAddress := generateEVMAddress()
		setupGroupInfo(ctx, mocks, poolAddress)

		v2OperatorAddress := generateEVMAddress()
		v3OperatorAddress := generateEVMAddress()
		v4OperatorAddress := generateEVMAddress()

		// v2 operator lst stake and bind
		allShares = addStakingSharesInfo(allShares, v2OperatorAddress, poolAddress, 1*OneLSTToken)
		setupOperatorRegistration(ctx, mocks, v2OperatorAddress, validators[1].OperatorAddress)
		// v3 operator lst stake and bind
		allShares = addStakingSharesInfo(allShares, v3OperatorAddress, poolAddress, 1*OneLSTToken)
		setupOperatorRegistration(ctx, mocks, v3OperatorAddress, validators[2].OperatorAddress)
		// 4 operator just stake not bind
		allShares = addStakingSharesInfo(allShares, v4OperatorAddress, poolAddress, 1*OneLSTToken)
		restakingKeeper.On("GetAllShares", ctx).Return(allShares)

		t.Log("----- epoch1 -----")
		validatorUpdates, err := mocks.ProcessEpoch(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, len(validatorUpdates))
		require.Equal(t, int64(1750), validatorUpdates[0].GetPower()) // v2
		require.Equal(t, int64(1750), validatorUpdates[1].GetPower()) // v3

		//
		// Epoch 2:
		//   - Native validators: v1, v2, v3
		//   - LST operators: v2, v3, v4
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		// After epoch1 status:
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*1/2 lst = 1750
		//   - v3 = 1000 native + 1500*1/2 lst = 1750
		//   - v4 = 0
		// Action: Increase v2 lst shares from 1 to 2
		// After action:
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		// Expected voting power:
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*2/3 lst = 2000
		//   - v3 = 1000 native + 1500*1/3 lst = 1500
		//   - v4 = 0
		//
		ctx = setupBlockHeight(ctx, mocks, 20)
		setupStakingValidatorsInfo(ctx, validators, validatorInfos, DefaultMaxValidators, stakingKeeper)
		// increase v2 shares
		mockRestakingSharesInfo(ctx, restakingKeeper, allShares, v2OperatorAddress, 2*OneLSTToken)
		restakingKeeper.On("GetAllShares", ctx).Return(allShares)

		t.Log("----- epoch2 -----")
		validatorUpdatesEpoch2, err := mocks.ProcessEpoch(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, len(validatorUpdatesEpoch2))
		require.Equal(t, int64(2000), validatorUpdatesEpoch2[0].GetPower()) // v2
		require.Equal(t, int64(1500), validatorUpdatesEpoch2[1].GetPower()) // v3

		//
		// Epoch 3:
		//   - Native validators: v1, v2, v3
		//   - LST operators: v2, v3, v4
		// After epoch2 status:
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*2/3 lst = 2000
		//   - v3 = 1000 native + 1500*1/3 lst = 1500
		//   - v4 = 0
		// Action: Increase v2 native shares from 1000 to 2000
		// Expected voting power:
		//   - Native voting power = 4000
		//   - LST voting power = 2000
		//   - v1 = 1000 native
		//   - v2 = 2000 native + 2000*2/3 lst = 3333
		//   - v3 = 1000 native + 2000*1/3 lst = 1666
		//   - v4 = 0
		//
		ctx = setupBlockHeight(ctx, mocks, 30)
		// change v2's native shares from 1000 to 2000
		validators[1].Tokens = sdkmath.NewInt(2000 * PowerReduction)
		validators[1].DelegatorShares = sdkmath.LegacyDec(sdkmath.NewInt(2000 * PowerReduction))
		setupStakingValidatorsInfo(ctx, validators, validatorInfos, DefaultMaxValidators, stakingKeeper)
		restakingKeeper.On("GetAllShares", ctx).Return(allShares)

		t.Log("----- epoch3 -----")
		validatorUpdatesEpoch3, err := mocks.ProcessEpoch(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, len(validatorUpdatesEpoch3))
		require.Equal(t, int64(3333), validatorUpdatesEpoch3[0].GetPower())
		require.Equal(t, int64(1666), validatorUpdatesEpoch3[1].GetPower())

		//
		// Epoch 4:
		//   - Native validators: v1, v2, v3
		//   - LST operators: v2, v3, v4
		// After epoch3 status:
		//   - Native voting power = 4000
		//   - LST voting power = 2000
		//   - v1 = 1000 native
		//   - v2 = 2000 native + 2000*2/3 lst = 3333
		//   - v3 = 1000 native + 2000*1/3 lst = 1666
		//   - v4 = 0
		// Action: Decrease v2 lst shares from 2 to 1
		// Expected voting power:
		//   - Native voting power = 4000
		//   - LST voting power = 2000
		//   - v1 = 1000 native
		//   - v2 = 2000 native + 2000*1/2 lst = 3000
		//   - v3 = 1000 native + 2000*1/2 lst = 2000
		//   - v4 = 0
		//
		ctx = setupBlockHeight(ctx, mocks, 40)
		setupStakingValidatorsInfo(ctx, validators, validatorInfos, DefaultMaxValidators, stakingKeeper)
		// decrease v2 shares
		mockRestakingSharesInfo(ctx, restakingKeeper, allShares, v2OperatorAddress, 1*OneLSTToken)
		restakingKeeper.On("GetAllShares", ctx).Return(allShares)

		validatorUpdatesEpoch4, err := mocks.ProcessEpoch(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, len(validatorUpdatesEpoch4))
		require.Equal(t, int64(3000), validatorUpdatesEpoch4[0].GetPower())
		require.Equal(t, int64(2000), validatorUpdatesEpoch4[1].GetPower())

		//
		// Epoch 5:
		//   - Native validators: v1, v2, v3
		//   - LST operators: v2, v3, v4
		// After epoch3 status:
		//   - Native voting power = 4000
		//   - LST voting power = 2000
		//   - v1 = 1000 native
		//   - v2 = 2000 native + 2000*1/2 lst = 3000
		//   - v3 = 1000 native + 2000*1/2 lst = 2000
		//   - v4 = 0
		// Action: Decrease v2 native shares from 2000 to 1000
		// Expected voting power:
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*1/2 lst = 1750
		//   - v3 = 1000 native + 1500*1/2 lst = 1750
		//   - v4 = 0
		//
		ctx = setupBlockHeight(ctx, mocks, 50)
		// change v2's native shares from 1000 to 2000
		validators[1].Tokens = sdkmath.NewInt(1000 * PowerReduction)
		validators[1].DelegatorShares = sdkmath.LegacyDec(sdkmath.NewInt(1000 * PowerReduction))
		setupStakingValidatorsInfo(ctx, validators, validatorInfos, DefaultMaxValidators, stakingKeeper)
		restakingKeeper.On("GetAllShares", ctx).Return(allShares)

		t.Log("----- epoch5 -----")
		validatorUpdatesEpoch5, err := mocks.ProcessEpoch(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, len(validatorUpdatesEpoch3))
		require.Equal(t, int64(1750), validatorUpdatesEpoch5[0].GetPower())
		require.Equal(t, int64(1750), validatorUpdatesEpoch5[1].GetPower())

		//
		// Epoch 6:
		//   - Native validators: v1, v2, v3
		//   - LST operators: v2, v3, v4
		// After epoch3 status:
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*1/2 lst = 1750
		//   - v3 = 1000 native + 1500*1/2 lst = 1750
		//   - v4 = 0
		// Action: Increase v4 lst shares from 1 to 2
		// Expected voting power:
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*1/2 lst = 1750
		//   - v3 = 1000 native + 1500*1/2 lst = 1750
		//   - v4 = 0 native + any = 0
		//
		ctx = setupBlockHeight(ctx, mocks, 60)
		// change v2's native shares from 1000 to 2000
		setupStakingValidatorsInfo(ctx, validators, validatorInfos, DefaultMaxValidators, stakingKeeper)
		// increase v4 shares
		mockRestakingSharesInfo(ctx, restakingKeeper, allShares, v4OperatorAddress, 2*OneLSTToken)
		restakingKeeper.On("GetAllShares", ctx).Return(allShares)

		t.Log("----- epoch6 -----")
		validatorUpdatesEpoch6, err := mocks.ProcessEpoch(ctx)
		require.NoError(t, err)
		require.Equal(t, 0, len(validatorUpdatesEpoch6))

		//
		// Epoch 7:
		//   - Native validators: v1, v2, v3
		//   - LST operators: v2, v3, v4
		// After epoch3 status:
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*1/2 lst = 1750
		//   - v3 = 1000 native + 1500*1/2 lst = 1750
		//   - v4 = 0
		// Action: Decrease v4 lst shares from 2 to 0
		// Expected voting power:
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*1/2 lst = 1750
		//   - v3 = 1000 native + 1500*1/2 lst = 1750
		//   - v4 = 0 native + any = 0
		//
		ctx = setupBlockHeight(ctx, mocks, 70)
		// change v2's native shares from 1000 to 2000
		setupStakingValidatorsInfo(ctx, validators, validatorInfos, DefaultMaxValidators, stakingKeeper)
		// decrease v4 lst shares
		mockRestakingSharesInfo(ctx, restakingKeeper, allShares, v4OperatorAddress, 0*OneLSTToken)
		restakingKeeper.On("GetAllShares", ctx).Return(allShares)

		t.Log("----- epoch7 -----")
		validatorUpdatesEpoch7, err := mocks.ProcessEpoch(ctx)
		require.NoError(t, err)
		require.Equal(t, 0, len(validatorUpdatesEpoch7))

		//
		// Epoch 8:
		//   - Native validators: v1, v2, v3
		//   - LST operators: v2, v3, v4
		// After epoch3 status:
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*1/2 lst = 1750
		//   - v3 = 1000 native + 1500*1/2 lst = 1750
		//   - v4 = 0
		// Action: Decrease v2 lst shares from 1 to 0
		// Expected voting power:
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*0 lst = 1000
		//   - v3 = 1000 native + 1500*1 lst = 2500
		//   - v4 = 0
		//
		ctx = setupBlockHeight(ctx, mocks, 80)
		// change v2's native shares from 1000 to 2000
		setupStakingValidatorsInfo(ctx, validators, validatorInfos, DefaultMaxValidators, stakingKeeper)
		// decrease v2 lst shares
		mockRestakingSharesInfo(ctx, restakingKeeper, allShares, v2OperatorAddress, 0*OneLSTToken)
		restakingKeeper.On("GetAllShares", ctx).Return(allShares)

		t.Log("----- epoch8 -----")
		validatorUpdatesEpoch8, err := mocks.ProcessEpoch(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, len(validatorUpdatesEpoch8))
		require.Equal(t, int64(1000), validatorUpdatesEpoch8[0].GetPower())
		require.Equal(t, int64(2500), validatorUpdatesEpoch8[1].GetPower())

		//
		// Epoch 9:
		//   - Native validators: v1, v2, v3
		//   - LST operators: v2, v3, v4
		// After epoch3 status:
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*1/2 lst = 1750
		//   - v3 = 1000 native + 1500*1/2 lst = 1750
		//   - v4 = 0
		// Action: Remain v2 lst shares 1, and decrease v2 native shares from 1000 to 0
		// Expected voting power:
		//   - Native voting power = 2000
		//   - LST voting power = 1000
		//   - v1 = 1000 native
		//   - v2 = 0 native + 1000*1/2 lst = 500
		//   - v3 = 1000 native + 1000*1 lst = 1500
		//   - v4 = 0
		//
		ctx = setupBlockHeight(ctx, mocks, 90)
		// decrease v2 native token shares to 0
		validators[1].Tokens = sdkmath.NewInt(0 * PowerReduction)
		validators[1].DelegatorShares = sdkmath.LegacyDec(sdkmath.NewInt(0 * PowerReduction))
		setupStakingValidatorsInfo(ctx, validators, validatorInfos, DefaultMaxValidators, stakingKeeper)
		// remain v2 lst shares
		mockRestakingSharesInfo(ctx, restakingKeeper, allShares, v2OperatorAddress, 1*OneLSTToken)
		restakingKeeper.On("GetAllShares", ctx).Return(allShares)

		t.Log("----- epoch9 -----")
		validatorUpdatesEpoch9, err := mocks.ProcessEpoch(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, len(validatorUpdatesEpoch9))
		require.Equal(t, int64(500), validatorUpdatesEpoch9[0].GetPower())
		require.Equal(t, int64(1500), validatorUpdatesEpoch9[1].GetPower())

		//
		// Epoch 10:
		//   - Native validators: v1, v2, v3
		//   - LST operators: v2, v3, v4
		// After epoch3 status:
		//   - Native voting power = 2000
		//   - LST voting power = 1000
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1000*1/2 lst = 500
		//   - v3 = 1000 native + 1000*1 lst = 1500
		//   - v4 = 0
		// Action: Remain v2 lst shares 1, and remain v2 native shares 1000
		// Expected voting power:
		//   - Native voting power = 3000
		//   - LST voting power = 1500
		//   - v1 = 1000 native
		//   - v2 = 1000 native + 1500*1/2 lst = 1750
		//   - v3 = 1000 native + 1500*1 lst = 1750
		//   - v4 = 0
		//
		ctx = setupBlockHeight(ctx, mocks, 100)
		// remain v2's native shares 1000
		validators[1].Tokens = sdkmath.NewInt(1000 * PowerReduction)
		validators[1].DelegatorShares = sdkmath.LegacyDec(sdkmath.NewInt(1000 * PowerReduction))
		setupStakingValidatorsInfo(ctx, validators, validatorInfos, DefaultMaxValidators, stakingKeeper)
		// remain v2 lst shares
		mockRestakingSharesInfo(ctx, restakingKeeper, allShares, v2OperatorAddress, 1*OneLSTToken)
		restakingKeeper.On("GetAllShares", ctx).Return(allShares)

		t.Log("----- epoch10 -----")
		validatorUpdatesEpoch10, err := mocks.ProcessEpoch(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, len(validatorUpdatesEpoch10))
		require.Equal(t, int64(1750), validatorUpdatesEpoch10[0].GetPower())
		require.Equal(t, int64(1750), validatorUpdatesEpoch10[1].GetPower())
	})
}
