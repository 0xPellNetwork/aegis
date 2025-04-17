package types

import (
	"fmt"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// ModuleName defines the module name
	ModuleName = "pevm"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_pevm"

	OperatorShareStoreKey = "operator_share"

	EpochOperatorSharesSnapshotKey = "epoch_operator_shares_snapshot"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

var (
	ModuleAddress = authtypes.NewModuleAddress(ModuleName)
	//ModuleAddressEVM common.EVMAddress
	ModuleAddressEVM = common.BytesToAddress(ModuleAddress.Bytes())
	AdminAddress     = "pell1rx9r8hff0adaqhr5tuadkzj4e7ns2ntg446vtt"
)

func init() {
	//fmt.Printf("ModuleAddressEVM of %s: %s\n", ModuleName, ModuleAddressEVM.String())
	// 0x735b14BB79463307AAcBED86DAf3322B1e6226aB
}

const (
	SystemContractKey = "SystemContract-value-"
)

func RegistryRouterKey() []byte {
	return KeyPrefix(fmt.Sprintf("%s", "registry_router"))
}

func SupportedChainKey(registryRouter common.Address) []byte {
	return KeyPrefix(fmt.Sprintf("%s-%s", "supported_chain", registryRouter.Hex()))
}

func QuorumKey(registryRouter common.Address) []byte {
	return KeyPrefix(fmt.Sprintf("%s-%s", "quorum", registryRouter.Hex()))
}

func QuorumOperatorKey(registryRouter, operator common.Address) []byte {
	return KeyPrefix(fmt.Sprintf("%s-%s-%s", "quorum_operator", registryRouter.Hex(), operator.Hex()))
}
