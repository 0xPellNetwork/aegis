package types

import (
	context "context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/0xPellNetwork/aegis/pkg/proofs"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
)

type StakingKeeper interface {
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)
	GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, err error)
	SetValidator(ctx context.Context, validator stakingtypes.Validator) error
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
}
