package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

// UpdateChainParams updates chain parameters for a specific chain, or add a new one.
// Chain parameters include: confirmation count, outbound transaction schedule interval, PELL token,
// connector and ERC20 custody contract addresses, etc.
// Only the admin policy account is authorized to broadcast this message.
func (k msgServer) UpsertChainParams(goCtx context.Context, msg *types.MsgUpsertChainParams) (*types.MsgUpsertChainParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return &types.MsgUpsertChainParamsResponse{}, authoritytypes.ErrUnauthorized
	}

	tmpCtx, commit := ctx.CacheContext()

	if msg.ChainParams.GatewayEvmContractAddress != "" {
		chainId := msg.ChainParams.ChainId
		gatewayAddress := msg.ChainParams.GatewayEvmContractAddress

		// Because Pell's gateway needs to know the gateway address of the other chain to initiate a call.
		if err := k.pevmKeeper.CallUpdateDestinationAddressOnPellGateway(tmpCtx, chainId, gatewayAddress); err != nil {
			return &types.MsgUpsertChainParamsResponse{}, err
		}
		k.Logger(ctx).Info("Updated destination address on Pell Gateway", "chain_id", chainId, "gateway_address", gatewayAddress)

		// Since inbound listening has no whitelist, a whitelist is needed on the gateway contract to control the inbound PellSent events.
		if err := k.pevmKeeper.CallUpdateSourceAddressOnPellGateway(tmpCtx, chainId, gatewayAddress); err != nil {
			return &types.MsgUpsertChainParamsResponse{}, err
		}
		k.Logger(ctx).Info("Updated source address on Pell Gateway", "pell_chain_id", chainId, "gateway_address", gatewayAddress)
	}

	if msg.ChainParams.GasSwapContractAddress != "" {
		err := k.pevmKeeper.CallUpdateDestinationAddressOnGasSwapPEVM(
			tmpCtx, msg.ChainParams.ChainId, msg.ChainParams.GasSwapContractAddress,
		)
		if err != nil {
			return &types.MsgUpsertChainParamsResponse{}, err
		}
	}

	// find current chain params list or initialize a new one
	chainParamsList, found := k.GetChainParamsList(tmpCtx)
	if !found {
		chainParamsList = types.ChainParamsList{}
	}

	// find chain params for the chain
	for i, cp := range chainParamsList.ChainParams {
		if cp.ChainId == msg.ChainParams.ChainId {
			chainParamsList.ChainParams[i] = msg.ChainParams
			k.SetChainParamsList(tmpCtx, chainParamsList)

			commit()
			return &types.MsgUpsertChainParamsResponse{}, nil
		}
	}

	// add new chain params
	chainParamsList.ChainParams = append(chainParamsList.ChainParams, msg.ChainParams)
	k.SetChainParamsList(tmpCtx, chainParamsList)

	commit()
	return &types.MsgUpsertChainParamsResponse{}, nil
}
