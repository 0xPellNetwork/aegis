package keeper

import (
	"math/big"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	"github.com/ethereum/go-ethereum/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

// encodeSyncRegisterOperator encodes the sync register operator event
func encodeSyncRegisterOperator(operator ethcommon.Address, groupNumbers []byte, params registryrouter.IRegistryRouterSyncPubkeyRegistrationParams) ([]byte, error) {
	data, err := centralschedulerMetaDataABI.Pack("syncRegisterOperator", operator, groupNumbers, params)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// encodeSyncRegisterOperatorWithChurn encodes the sync register operator with churn event
func encodeSyncRegisterOperatorWithChurn(operator ethcommon.Address, groupNumbers []byte, params registryrouter.IRegistryRouterSyncPubkeyRegistrationParams, kickParam []registryrouter.IRegistryRouterOperatorKickParam) ([]byte, error) {
	data, err := centralschedulerMetaDataABI.Pack("syncRegisterOperatorWithChurn", operator, groupNumbers, params, kickParam)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// encodeSyncCreateGroup encodes the sync create group event
func encodeSyncCreateGroup(groupNumber uint8, operatorSetParam registryrouter.IRegistryRouterOperatorSetParam, minimumStake *big.Int, poolParams []registryrouter.IStakeRegistryRouterPoolParams) ([]byte, error) {
	data, err := centralschedulerMetaDataABI.Pack("syncCreateGroup", groupNumber, operatorSetParam, minimumStake, poolParams)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// encodeSyncSetOperatorSetParams encodes the sync set operator set params event
func encodeSyncSetOperatorSetParams(groupNumber uint8, params registryrouter.IRegistryRouterOperatorSetParam) ([]byte, error) {
	data, err := centralschedulerMetaDataABI.Pack("syncSetOperatorSetParams", groupNumber, params)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// encodeSyncSetGroupEjectionParams encodes the sync set group ejection params event
func encodeSyncSetGroupEjectionParams(groupNumber uint8, params registryrouter.IRegistryRouterGroupEjectionParams) ([]byte, error) {
	data, err := ejectionManagerMetaDataABI.Pack("syncSetGroupEjectionParams", groupNumber, params)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// encodeSyncEjectionCooldown encodes the sync ejection cooldown event
func encodeSyncEjectionCooldown(ejectionCooldown *big.Int) ([]byte, error) {
	data, err := centralschedulerMetaDataABI.Pack("syncEjectionCooldown", ejectionCooldown)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// encodeSyncAddPools encodes the sync add pools event
func encodeSyncAddPools(groupNumber uint8, poolParams []stakeregistryrouter.IStakeRegistryRouterPoolParams) ([]byte, error) {
	data, err := operatorstakemanagerMetaDataABI.Pack("syncAddPools", groupNumber, poolParams)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// encodeSyncRemovePools encodes the sync remove pools event
func encodeSyncRemovePools(groupNumber uint8, indices []*big.Int) ([]byte, error) {
	data, err := operatorstakemanagerMetaDataABI.Pack("syncRemovePools", groupNumber, indices)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func encodeSyncModifyPoolParams(groupNumber uint8, indices, multipliers []*big.Int) ([]byte, error) {
	data, err := operatorstakemanagerMetaDataABI.Pack("syncModifyPoolParams", groupNumber, indices, multipliers)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// encodeSyncDeregisterOperator encodes the sync deregister operator event
func encodeSyncDeregisterOperator(operator ethcommon.Address, groupNumbers []byte) ([]byte, error) {
	data, err := centralschedulerMetaDataABI.Pack("syncDeregisterOperator", operator, groupNumbers)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// encodeSyncUpdateOperators encodes the sync update operators event
func encodeSyncUpdateOperators(operators []common.Address) ([]byte, error) {
	data, err := centralschedulerMetaDataABI.Pack("syncUpdateOperators", operators)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// encodeSyncUpdateOperatorsForGroup encodes the sync update operators for group event
func encodeSyncUpdateOperatorsForGroup(operatorsPerGroup [][]common.Address, groupNumbers []byte) ([]byte, error) {
	data, err := centralschedulerMetaDataABI.Pack("syncUpdateOperatorsForGroup", operatorsPerGroup, groupNumbers)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// encodeSyncEjectOperators encodes the sync eject operators event
func encodeSyncEjectOperators(operatorIds [][][32]byte) ([]byte, error) {
	data, err := ejectionManagerMetaDataABI.Pack("syncEjectOperators", operatorIds)
	if err != nil {
		return nil, err
	}

	return data, nil
}
