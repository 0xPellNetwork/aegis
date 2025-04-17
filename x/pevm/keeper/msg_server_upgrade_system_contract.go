package keeper

import (
	"context"

	cosmoserror "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/pevm/types"
)

// UpgradeSystemContracts upgrades the system contracts.
//
// Authorized: admin policy group 2.
func (k msgServer) UpgradeSystemContracts(goCtx context.Context, msg *types.MsgUpgradeSystemContracts) (*types.MsgUpgradeSystemContractsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return nil, cosmoserror.Wrap(authoritytypes.ErrUnauthorized, "System contract deployment can only be executed by the correct policy account")
	}

	sysContract, found := k.GetSystemContract(ctx)
	if !found {
		return nil, cosmoserror.Wrap(types.ErrSystemContractNotFound, "System contract not found")
	}

	// registry router contract
	dvsProxy := common.HexToAddress(sysContract.DvsDirectoryProxy)
	systemContract := common.HexToAddress(sysContract.SystemContract)
	registryRouter, err := k.DeployPellRegistryRouter(ctx, dvsProxy, systemContract)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy new pellRegistryRouter")
	}

	registryRouterBeacon := common.HexToAddress(sysContract.RegistryRouterBeacon)
	res, err := k.UpgradeRegistryRouterBeacon(ctx, registryRouterBeacon, registryRouter)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to upgrade registryRouterBeacon")
	}
	k.Logger(ctx).Info("Upgrade system contract result", "result", res)

	delegationManagerProxy := common.HexToAddress(sysContract.DelegationManagerProxy)
	stakeRegistryRouter, err := k.DeployPellStakeRegistryRouter(ctx, delegationManagerProxy)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to deploy new pellStakeRegistryRouter")
	}

	// stakeRegistryRouterBeacon contract
	stakeRegistryRouterBeacon := common.HexToAddress(sysContract.StakeRegistryRouterBeacon)
	res, err = k.UpgradeRegistryRouterBeacon(ctx, stakeRegistryRouterBeacon, stakeRegistryRouter)
	if err != nil {
		return nil, cosmoserror.Wrapf(err, "failed to upgrade registryRouterBeacon")
	}
	k.Logger(ctx).Info("Upgrade system contract result", "result", res)

	return &types.MsgUpgradeSystemContractsResponse{
		RegistryRouter:      registryRouter.String(),
		StakeRegistryRouter: stakeRegistryRouter.String(),
	}, nil
}
