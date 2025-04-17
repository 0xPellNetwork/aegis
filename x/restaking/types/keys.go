package types

import (
	fmt "fmt"
	"strings"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// ModuleName defines the module name
	ModuleName = "restaking"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_restaking"

	KeyOperatorShareStore          = "operator_share"
	KeyPrefixEpochSharesChange     = "epoch_shares_change"
	KeyPrefixEpochNumber           = "epoch_number"
	KeyBlocksPerEpoch              = "blocks_per_epoch"
	KeyPrefixOutboundState         = "outbound_state"
	KeyEpochOperatorSharesSyncTxs  = "epoch_operator_shares_sync_txs"
	KeyEpochOperatorSharesSnapshot = "epoch_operator_shares_snapshot"
	KeyOperator                    = "operator"

	GroupKey = "group_data"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

var (
	ModuleAddress = authtypes.NewModuleAddress(ModuleName)
	//ModuleAddressEVM common.EVMAddress
	ModuleAddressEVM = common.BytesToAddress(ModuleAddress.Bytes())
)

func RegistryRouterKey() []byte {
	return KeyPrefix("registry_router")
}

func StakeRegistryRouterKey(stakeRegistryRouter common.Address) []byte {
	return KeyPrefix(fmt.Sprintf("%s-%s", "stake_registry_router", stakeRegistryRouter.Hex()))
}

func SupportedChainKey(registryRouter common.Address) []byte {
	return KeyPrefix(fmt.Sprintf("%s-%s", "supported_chain", registryRouter.Hex()))
}

func GroupDataKey(registryRouter common.Address) []byte {
	return KeyPrefix(fmt.Sprintf("%s-%s", "group_data", registryRouter.Hex()))
}

func GroupOperatorKey(registryRouter common.Address) []byte {
	return KeyPrefix(fmt.Sprintf("%s-%s", "group_operator", registryRouter.Hex()))
}

func GroupSyncKey(txHash string) []byte {
	txHash = strings.ToLower(txHash)
	return KeyPrefix(fmt.Sprintf("%s-%s", "group_data_sync", txHash))
}

func init() {
	//fmt.Printf("ModuleAddressEVM of %s: %s\n", ModuleName, ModuleAddressEVM.String())
	// 0x735b14BB79463307AAcBED86DAf3322B1e6226aB
}
