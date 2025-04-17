package keeper

import (
	"context"
	"math/big"

	cosmoserror "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/pell-chain/pellcore/pkg/chains"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/pevm/types"
)

// DeployGatewayContract deploy new instances of the gateway contracts
//
// Authorized: admin policy group 2.
func (k msgServer) DeployGatewayContract(goCtx context.Context, msg *types.MsgDeployGatewayContract) (*types.MsgDeployGatewayContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return nil, cosmoserror.Wrap(authoritytypes.ErrUnauthorized, "System contract deployment can only be executed by the correct policy account")
	}

	sysContract, found := k.GetSystemContract(ctx)
	if !found {
		return nil, cosmoserror.Wrap(types.ErrSystemContractNotFound, "System contract not found")
	}

	systemContract := common.HexToAddress(sysContract.SystemContract)
	connector := common.HexToAddress(sysContract.Connector)
	wpell := common.HexToAddress(sysContract.WrappedPell)

	gatewayPEVM, err := k.DeployGatewayPEVM(ctx, connector, systemContract, wpell)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy GatewayEVM")
	}

	pellChainIDInt, err := chains.CosmosToEthChainID(ctx.ChainID())
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to get pell chain id on DeployGatewayContract")
	}

	pellChainID := big.NewInt(pellChainIDInt)
	// update gateway source address and destination address
	_, err = k.CallMethodOnGateway(ctx, gatewayPEVM, "updateDestinationAddress", pellChainID, connector.Bytes())
	if err != nil {
		k.Logger(ctx).Error("DeployGatewayContract failed to call updateDestinationAddress on gateway contract", "error", err)
		return nil, cosmoserror.Wrapf(err, "failed to call updateDestinationAddress on gateway contract")
	}

	_, err = k.CallMethodOnGateway(ctx, gatewayPEVM, "updateSourceAddress", pellChainID, connector.Bytes())
	if err != nil {
		k.Logger(ctx).Error("DeployGatewayContract failed to call updateSourceAddress on gateway contract", "error", err)
		return nil, cosmoserror.Wrapf(err, "failed to call updateSourceAddress on gateway contract")
	}

	return &types.MsgDeployGatewayContractResponse{
		Gateway: gatewayPEVM.Hex(),
	}, nil
}
