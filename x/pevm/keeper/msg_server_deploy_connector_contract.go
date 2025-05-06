package keeper

import (
	"context"
	"math/big"

	cosmoserror "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	"github.com/0xPellNetwork/aegis/x/pevm/types"
)

// DeployConnectorContract deploy new instances of the gateway contracts
//
// Authorized: admin policy group 2.
func (k msgServer) DeployConnectorContract(goCtx context.Context, msg *types.MsgDeployConnectorContract) (*types.MsgDeployConnectorContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return nil, cosmoserror.Wrap(authoritytypes.ErrUnauthorized, "System contract deployment can only be executed by the correct policy account")
	}

	sysContract, found := k.GetSystemContract(ctx)
	if !found {
		return nil, cosmoserror.Wrap(types.ErrSystemContractNotFound, "System contract not found")
	}

	systemContract := common.HexToAddress(sysContract.SystemContract)
	gatewayPEVM := common.HexToAddress(sysContract.Gateway)

	// deploy connector contract
	connector, err := k.DeployPellConnector(ctx, systemContract, types.ModuleAddressEVM)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy new PellConnector")
	}

	pellChainIDInt, err := chains.CosmosToEthChainID(ctx.ChainID())
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to get pell chain id on DeployConnectorContract")
	}

	pellChainID := big.NewInt(pellChainIDInt)
	// update gateway source address and destination address
	_, err = k.CallMethodOnGateway(ctx, gatewayPEVM, "updateDestinationAddress", pellChainID, connector.Bytes())
	if err != nil {
		k.Logger(ctx).Error("failed to call updateDestinationAddress on gateway contract", "error", err)
		return nil, cosmoserror.Wrapf(err, "failed to call updateDestinationAddress on gateway contract")
	}

	_, err = k.CallMethodOnGateway(ctx, gatewayPEVM, "updateSourceAddress", pellChainID, connector.Bytes())
	if err != nil {
		k.Logger(ctx).Error("failed to call updateSourceAddress on gateway contract", "error", err)
		return nil, cosmoserror.Wrapf(err, "failed to call updateSourceAddress on gateway contract")
	}

	return &types.MsgDeployConnectorContractResponse{
		Connector: connector.Hex(),
	}, nil
}
