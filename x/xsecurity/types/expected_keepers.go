package types

import (
	"context"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/pell-chain/pellcore/pkg/proofs"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	restakingtypes "github.com/pell-chain/pellcore/x/restaking/types"
)

type StakingKeeper interface {
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)
	GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, err error)
	SetValidator(ctx context.Context, validator stakingtypes.Validator) error
	BlockValidatorUpdates(ctx context.Context) ([]abci.ValidatorUpdate, error)
	GetParams(ctx context.Context) (params stakingtypes.Params, err error)
	LastValidatorsIterator(ctx context.Context) (corestore.Iterator, error)
	ValidatorsPowerStoreIterator(ctx context.Context) (corestore.Iterator, error)
	PowerReduction(ctx context.Context) math.Int
	GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error)
}

type SlashingKeeper interface {
	IsTombstoned(ctx context.Context, addr sdk.ConsAddress) bool
	SetValidatorSigningInfo(ctx context.Context, address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) error
}

type StakingHooks interface {
	AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error
	AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error
	AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error
	BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction math.LegacyDec) error
}

type AuthorityKeeper interface {
	IsAuthorized(ctx sdk.Context, address string, policyType authoritytypes.PolicyType) bool

	// SetPolicies is solely used for the migration of policies from observer to authority
	SetPolicies(ctx sdk.Context, policies authoritytypes.Policies)
}

type RelayerKeeper interface {
}

type LightclientKeeper interface {
	CheckNewBlockHeader(
		ctx sdk.Context,
		chainID int64,
		blockHash []byte,
		height int64,
		header proofs.HeaderData,
	) ([]byte, error)
	AddBlockHeader(
		ctx sdk.Context,
		chainID int64,
		height int64,
		blockHash []byte,
		header proofs.HeaderData,
		parentHash []byte,
	)
}

type PevmKeeper interface {
	CallUpdateDestinationAddressOnPellGateway(
		ctx context.Context,
		chainId int64,
		destinationAddress string,
	) error
	CallUpdateSourceAddressOnPellGateway(
		ctx context.Context,
		chainId int64,
		sourceAddress string,
	) error
	CallUpdateDestinationAddressOnGasSwapPEVM(
		ctx context.Context,
		chainId int64,
		destinationAddress string,
	) error

	// --- LST Token staking methods ---

	CallRegistryRouterFactory(
		ctx context.Context,
		dvsChainApprover, churnApprover, ejector, pauser, unpauser ethcommon.Address,
		initialPausedStatus uint,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	CallRegistryRouterToCreateGroup(
		ctx sdk.Context,
		registryRouterAddress ethcommon.Address,
		operatorSetParams restakingtypes.OperatorSetParam,
		minimumStake int64,
		poolParams []restakingtypes.PoolParams,
		groupEjectionParams restakingtypes.GroupEjectionParam,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	CallRegistryRouterToRegisterOperator(
		ctx sdk.Context,
		registryRouterAddress ethcommon.Address,
		param RegisterOperatorParam,
		operatorAddress ethcommon.Address,
		groupNumbers uint64,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	CallStakelRegistryRouterToAddPools(
		ctx sdk.Context,
		stakeRegistryRouterAddress ethcommon.Address,
		groupNumbers uint64,
		poolParams []*restakingtypes.PoolParams,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	CallStakelRegistryRouterToRemovePools(
		ctx sdk.Context,
		stakeRegistryRouterAddress ethcommon.Address,
		groupNumbers uint64,
		indicesToRemove []uint,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
	CallRegistryRouterToSetOperatorSetParams(
		ctx sdk.Context,
		stakeRegistryRouterAddress ethcommon.Address,
		groupNumbers uint64,
		operatorSetParams *restakingtypes.OperatorSetParam,
	) (*evmtypes.MsgEthereumTxResponse, bool, error)
}

type RestakingKeeper interface {
	GetAllShares(ctx sdk.Context) []*restakingtypes.OperatorShares
}
