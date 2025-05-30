package keeper

import (
	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/x/lightclient/types"
)

// SetVerificationFlags set the verification flags in the store
func (k Keeper) SetVerificationFlags(ctx sdk.Context, crosschainFlags types.VerificationFlags) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.VerificationFlagsKey))
	b := k.cdc.MustMarshal(&crosschainFlags)
	store.Set([]byte{0}, b)
}

// GetVerificationFlags returns the verification flags
func (k Keeper) GetVerificationFlags(ctx sdk.Context) (val types.VerificationFlags, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.VerificationFlagsKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// CheckVerificationFlagsEnabled checks for a specific chain if the verification flags are enabled
// It returns an error if the chain is not enabled
func (k Keeper) CheckVerificationFlagsEnabled(ctx sdk.Context, chainID int64) error {
	verificationFlags, found := k.GetVerificationFlags(ctx)
	if !found {
		return types.ErrVerificationFlagsNotFound
	}

	// check if the chain is enabled for the specific type
	if chains.IsEVMChain(chainID) {
		if !verificationFlags.EthTypeChainEnabled {
			return cosmoserrors.Wrapf(
				types.ErrBlockHeaderVerificationDisabled,
				"proof verification not enabled for evm ,chain id: %d",
				chainID,
			)
		}
	} else {
		return cosmoserrors.Wrapf(
			types.ErrChainNotSupported,
			"chain ID %d doesn't support block header verification",
			chainID,
		)
	}

	return nil
}
