package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

// SetOperator sets an operator in the store
func (k Keeper) SetOperator(ctx sdk.Context, operator types.Operator) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.KeyOperator))

	store.Set(types.KeyPrefix(operator.Operator), k.cdc.MustMarshal(&operator))
}

// GetOperator returns an operator from the store
func (k Keeper) GetOperator(ctx sdk.Context, operator string) (types.Operator, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.KeyOperator))
	b := store.Get(types.KeyPrefix(operator))
	if b == nil {
		return types.Operator{}, false
	}

	var val types.Operator
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveOperator removes an operator from the store
func (k Keeper) RemoveOperator(ctx sdk.Context, operator string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.KeyOperator))
	store.Delete(types.KeyPrefix(operator))
}

// GetAllOperators returns all operators from the store
func (k Keeper) GetAllOperators(ctx sdk.Context) []types.Operator {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.KeyOperator))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var operators []types.Operator
	for ; iterator.Valid(); iterator.Next() {
		var val types.Operator
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		operators = append(operators, val)
	}

	return operators
}
