package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	observerTypes "github.com/pell-chain/pellcore/x/relayer/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// SetXmsgAndNonceToXmsgAndInTxHashToXmsg does the following things in one function:
// 1. set the xmsg in the store
// 2. set the mapping inTxHash -> xmsgIndex , one inTxHash can be connected to multiple xmsgindex
// 3. set the mapping nonce => xmsg
// 4. update the pell accounting
func (k Keeper) SetXmsgAndNonceToXmsgAndInTxHashToXmsg(ctx sdk.Context, xmsg types.Xmsg) {
	tss, found := k.relayerKeeper.GetTSS(ctx)
	if !found {
		return
	}
	// set mapping nonce => xmsgIndex
	if xmsg.XmsgStatus.Status == types.XmsgStatus_PENDING_OUTBOUND || xmsg.XmsgStatus.Status == types.XmsgStatus_PENDING_REVERT {
		k.GetRelayerKeeper().SetNonceToXmsg(ctx, observerTypes.NonceToXmsg{
			ChainId: xmsg.GetCurrentOutTxParam().ReceiverChainId,
			// #nosec G701 always in range
			Nonce:     int64(xmsg.GetCurrentOutTxParam().OutboundTxTssNonce),
			XmsgIndex: xmsg.Index,
			Tss:       tss.TssPubkey,
		})
	}
	//outbound
	k.SetXmsg(ctx, xmsg)

	// set mapping inTxHash -> xmsgIndex
	in, _ := k.GetInTxHashToXmsg(ctx, xmsg.InboundTxParams.InboundTxHash)
	in.InTxHash = xmsg.InboundTxParams.InboundTxHash
	found = false
	for _, xmsgIndex := range in.XmsgIndices {
		if xmsgIndex == xmsg.Index {
			found = true
			break
		}
	}
	if !found {
		in.XmsgIndices = append(in.XmsgIndices, xmsg.Index)
	}
	k.SetInTxHashToXmsg(ctx, in)
}

// SetXmsg set a specific send in the store from its index
func (k Keeper) SetXmsg(ctx sdk.Context, xmsg types.Xmsg) {
	p := types.KeyPrefix(types.SendKey)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), p)
	b := k.cdc.MustMarshal(&xmsg)
	store.Set(types.KeyPrefix(xmsg.Index), b)
}

// GetXmsg returns a send from its index
func (k Keeper) GetXmsg(ctx sdk.Context, index string) (val types.Xmsg, found bool) {
	p := types.KeyPrefix(types.SendKey)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), p)

	b := store.Get(types.KeyPrefix(index))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// SetXmsg set a specific send in the store from its index
func (k Keeper) SetXmsgByEventIndex(ctx sdk.Context, eventIndex string, xmsg types.Xmsg) {
	p := types.KeyPrefix(types.SendKey)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), p)
	b := k.cdc.MustMarshal(&xmsg)
	store.Set(types.KeyPrefix(eventIndex), b)
}

// GetXmsg returns a send from its index
func (k Keeper) GetXmsgByEventIndex(ctx sdk.Context, eventIndex string) (val types.Xmsg, found bool) {
	p := types.KeyPrefix(types.SendKey)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), p)

	b := store.Get(types.KeyPrefix(eventIndex))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetXmsg returns a send from its index
func (k Keeper) DeleteXmsgByEventIndex(ctx sdk.Context, eventIndex string) {
	p := types.KeyPrefix(types.SendKey)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), p)

	store.Delete(types.KeyPrefix(eventIndex))
}

func (k Keeper) GetAllXmsg(ctx sdk.Context) (list []types.Xmsg) {
	p := types.KeyPrefix(types.SendKey)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), p)

	iterator := store.Iterator(nil, nil)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Xmsg
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return list
}

// RemoveXmsg removes a send from the store
func (k Keeper) RemoveXmsg(ctx sdk.Context, index string) {
	p := types.KeyPrefix(types.SendKey)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), p)
	store.Delete(types.KeyPrefix(index))
}
