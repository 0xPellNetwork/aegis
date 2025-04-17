package keeper

import (
	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

// GetDVSSupportedChainList returns the list of supported chains for a given registry router address
func (k *Keeper) GetDVSSupportedChainList(ctx sdk.Context, registryRouterAddress ethcommon.Address) (*types.DVSInfoList, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))

	// Get the DVSInfo array for the given registry router address
	key := types.SupportedChainKey(registryRouterAddress)
	data := store.Get(key)
	if data == nil {
		return nil, false
	}

	var dvsInfoList types.DVSInfoList
	k.cdc.MustUnmarshal(data, &dvsInfoList)

	return &dvsInfoList, true
}

// GetDVSSupportedChainListByStatus returns the list of supported chains for a given registry router address with normal status
func (k *Keeper) GetDVSSupportedChainListByStatus(ctx sdk.Context, registryRouterAddress ethcommon.Address, status types.OutboundStatus) ([]*types.DVSInfo, bool) {
	dvsInfoList, exists := k.GetDVSSupportedChainList(ctx, registryRouterAddress)
	if !exists {
		return nil, false
	}

	normalStatusChains := make([]*types.DVSInfo, 0)
	for _, dvsInfo := range dvsInfoList.DvsInfos {
		if dvsInfo.OutboundState == status {
			normalStatusChains = append(normalStatusChains, dvsInfo)
		}
	}

	if len(normalStatusChains) == 0 {
		return nil, false
	}

	return normalStatusChains, true
}

// GetDVSSupportedChain returns the DVS info for a given registry router address and chain id
func (k *Keeper) GetDVSSupportedChain(ctx sdk.Context, registryRouterAddress ethcommon.Address, chainId uint64) (*types.DVSInfo, bool) {
	dvsInfoList, exists := k.GetDVSSupportedChainList(ctx, registryRouterAddress)
	if !exists {
		return nil, false
	}

	for _, dvsInfo := range dvsInfoList.DvsInfos {
		if dvsInfo.ChainId == chainId {
			return dvsInfo, true
		}
	}
	return nil, false
}

// AddDVSSupportedChain adds a new DVS info to the list of supported chains for a given registry router address
func (k *Keeper) AddDVSSupportedChain(ctx sdk.Context, registryRouterAddress ethcommon.Address, dvsInfo *types.DVSInfo) error {
	// Get existing list if it exists
	existingList, exist := k.GetDVSSupportedChainList(ctx, registryRouterAddress)
	if !exist {
		// If not found, create a new list with the DVS info
		existingList = &types.DVSInfoList{
			DvsInfos: []*types.DVSInfo{dvsInfo},
		}
	} else {
		// Append new DVS info to existing list
		existingList.DvsInfos = append(existingList.DvsInfos, dvsInfo)
	}

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))
	key := types.SupportedChainKey(registryRouterAddress)
	store.Set(key, k.cdc.MustMarshal(existingList))
	return nil
}

// SetDVSSupportedChainStatus sets the outbound status for a given registry router address and chain id
func (k Keeper) SetDVSSupportedChainStatus(ctx sdk.Context, registryRouterAddress ethcommon.Address, chainId uint64, status types.OutboundStatus) error {
	dvsInfoList, exists := k.GetDVSSupportedChainList(ctx, registryRouterAddress)
	if !exists {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "dvs info list not found for registry router address: %s", registryRouterAddress.Hex())
	}

	found := false
	for _, dvsInfo := range dvsInfoList.DvsInfos {
		if dvsInfo.ChainId == chainId {
			dvsInfo.OutboundState = status
			found = true
			break
		}
	}

	if !found {
		return cosmoserrors.Wrapf(types.ErrInvalidData, "dvs info not found for chain id: %d", chainId)
	}

	// Store the updated list
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.GroupKey))
	key := types.SupportedChainKey(registryRouterAddress)
	store.Set(key, k.cdc.MustMarshal(dvsInfoList))
	return nil
}

// GetDVSSupportedChainStatus returns the outbound status for a given registry router address and chain id
func (k Keeper) GetDVSSupportedChainStatus(ctx sdk.Context, registryRouterAddress ethcommon.Address, chainId uint64) (types.OutboundStatus, error) {
	dvsInfoList, exists := k.GetDVSSupportedChainList(ctx, registryRouterAddress)
	if !exists {
		return types.OutboundStatus_OUTBOUND_STATUS_INITIALIZING, cosmoserrors.Wrapf(types.ErrContractNotFound, "registry router address: %s", registryRouterAddress.Hex())
	}

	for _, dvsInfo := range dvsInfoList.DvsInfos {
		if dvsInfo.ChainId == chainId {
			return dvsInfo.OutboundState, nil
		}
	}
	return types.OutboundStatus_OUTBOUND_STATUS_INITIALIZING, cosmoserrors.Wrapf(types.ErrContractNotFound, "chain id: %d", chainId)
}
