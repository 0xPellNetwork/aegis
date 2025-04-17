package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	"github.com/pell-chain/pellcore/x/xsecurity/types"
)

// CreateRegistryRouter creates a system level registry router for LST tokens.
// It will call the RegistryRouterFactory system contract in the pevm module
// to create a RegistryRouter contract. The transaction that creates the contract
// will return a RegistryRouterCreated event, which includes the address of the
// stakeRegistryRouter contract and needs to be recorded as well.
func (k Keeper) CreateRegistryRouter(goCtx context.Context, msg *types.MsgCreateRegistryRouter) (*types.MsgCreateRegistryRouterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := k.Logger(ctx)

	logger.Info(fmt.Sprintf("CreateRegistryRouter receive request msg: %v", msg))

	// check permission
	if !k.GetAuthorityKeeper().IsAuthorized(ctx, msg.Signer, authoritytypes.PolicyType_GROUP_OPERATIONAL) {
		return nil, authoritytypes.ErrUnauthorized
	}

	// check if registry router already exists
	if _, exist := k.GetLSTRegistryRouterAddress(ctx); exist {
		return nil, fmt.Errorf("registry router already exists")
	}

	// call pevm module to create registry router
	receipt, _, err := k.pevmKeeper.CallRegistryRouterFactory(
		ctx,
		common.HexToAddress(msg.ChainApprover),
		common.HexToAddress(msg.ChurnApprover),
		common.HexToAddress(msg.Ejector),
		common.HexToAddress(msg.Pauser),
		common.HexToAddress(msg.Unpauser),
		uint(msg.InitialPausedStatus),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("CreateRegistryRouter call pevm module err: %v", err))
		return nil, err
	}

	// get registry router address from receipt
	registryRouterAddress, stakeRegistryRouterAddress, err := k.GetRegistryRouterAddressFromReceipt(ctx, receipt)
	if err != nil {
		logger.Error(fmt.Sprintf("CreateRegistryRouter filter registry router err: %v", err))
		return nil, err
	}

	// save registry router address to store
	k.SetLSTRegistryRouterAddress(ctx, &types.LSTRegistryRouterAddress{
		RegistryRouterAddress:      registryRouterAddress,
		StakeRegistryRouterAddress: stakeRegistryRouterAddress,
	})

	logger.Info(fmt.Sprintf("Registry router created successfully: %s, stake registry: %s",
		registryRouterAddress, stakeRegistryRouterAddress))

	return &types.MsgCreateRegistryRouterResponse{}, nil
}

// GetRegistryRouterAddressFromReceipt filters the registry router address from the receipt
func (k Keeper) GetRegistryRouterAddressFromReceipt(ctx sdk.Context, receipt *evmtypes.MsgEthereumTxResponse) (registryRouterAddress string, stakeRegistryRouterAddress string, err error) {
	for _, log := range receipt.Logs {
		if log.Topics[0] == registryRouterFactoryMedaDataABI.Events["RegistryRouterCreated"].ID.String() {
			if len(log.Data) < 64 {
				return "", "", fmt.Errorf("insufficient data length: %d, expected at least 64", len(log.Data))
			}

			// Extract addresses
			registryRouterAddress = common.BytesToAddress(log.Data[12:32]).Hex()
			stakeRegistryRouterAddress = common.BytesToAddress(log.Data[44:64]).Hex()

			return
		}
	}

	return "", "", fmt.Errorf("RegistryRouterCreated event not found in receipt")
}
