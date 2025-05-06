package keeper

import (
	"slices"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

var BASE_ALLOWED_XMSG_SENDER_KEY = []byte{0}

// SaveAllowedXmsgSenders saves a list of allowed xmsg senders to the store
// If no senders exist, it creates a new entry
// If senders already exist, it appends the new senders to the existing list
func (k Keeper) SaveAllowedXmsgSenders(ctx sdk.Context, builders []string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.XmsgAllowedSenderKey))

	b := store.Get(BASE_ALLOWED_XMSG_SENDER_KEY)

	var val types.AllowedXmsgSenders
	// Create new builders if not exist
	if b == nil {
		val = types.AllowedXmsgSenders{
			AllowedSenders: []string{},
		}
	} else {
		k.cdc.MustUnmarshal(b, &val)
	}

	// Append builders if exist. May result in duplicate builders
	val.AllowedSenders = append(val.AllowedSenders, builders...)

	store.Set(BASE_ALLOWED_XMSG_SENDER_KEY, k.cdc.MustMarshal(&val))
}

// GetAllowedXmsgSenders retrieves the list of allowed xmsg senders from the store
func (k Keeper) GetAllowedXmsgSenders(ctx sdk.Context) ([]string, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.XmsgAllowedSenderKey))

	b := store.Get(BASE_ALLOWED_XMSG_SENDER_KEY)
	if b == nil {
		return nil, false
	}

	var val types.AllowedXmsgSenders
	k.cdc.MustUnmarshal(b, &val)

	return val.AllowedSenders, true
}

// DeleteAllowedXmsgSenders removes specified allowed xmsg senders from the store
func (k Keeper) DeleteAllowedXmsgSenders(ctx sdk.Context, builders []string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.XmsgAllowedSenderKey))
	b := store.Get(BASE_ALLOWED_XMSG_SENDER_KEY)
	if b == nil {
		return
	}

	var val types.AllowedXmsgSenders
	k.cdc.MustUnmarshal(b, &val)

	removeAllowedXmsgSenders(&val, builders)

	store.Set(BASE_ALLOWED_XMSG_SENDER_KEY, k.cdc.MustMarshal(&val))
}

// IsAllowedXmsgSender checks if the given allowed xmsg sender is in the list of registered allowed xmsg senders.
func (k Keeper) IsAllowedXmsgSender(ctx sdk.Context, builder string) bool {
	builders, found := k.GetAllowedXmsgSenders(ctx)
	if !found {
		return false
	}

	return slices.Contains(builders, builder)
}

// removeAllowedXmsgSenders removes all occurrences of the specified allowed xmsg senders from val.AllowedSenders.
// It modifies val.AllowedSenders in-place and handles potential duplicates.
//
// Parameters:
//   - val: A pointer to the AllowedXmsgSenders struct containing the slice to be modified.
//   - builders: A slice of strings representing the allowed xmsg senders to be removed.
//
// The function iterates through each builder in the 'builders' slice and removes
// all of its occurrences from val.Builders. It uses a nested loop approach to
// ensure all matches are removed, even if there are duplicates.
func removeAllowedXmsgSenders(val *types.AllowedXmsgSenders, builders []string) {
	for _, builder := range builders {
		for i := 0; i < len(val.AllowedSenders); {
			if val.AllowedSenders[i] == builder {
				// Remove the element by slicing
				val.AllowedSenders = append(val.AllowedSenders[:i], val.AllowedSenders[i+1:]...)
				// Don't increment i, as we need to check the new element at this index
			} else {
				// Only increment i if we didn't remove an element
				i++
			}
		}
	}
}
