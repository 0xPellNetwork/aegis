package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

const (
	// ModuleName defines the module name
	ModuleName = "relayer"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_relayer"

	GroupID1Address = "pell14r8nqy53kuruf7pp6aau3d8029ncxnwer54weg"

	MinObserverDelegation = "1000000000000000000"
)

func GetMinObserverDelegation() (sdkmath.Int, bool) {
	return sdkmath.NewIntFromString(MinObserverDelegation)
}

func GetMinObserverDelegationDec() (sdkmath.LegacyDec, error) {
	return sdkmath.LegacyNewDecFromStr(MinObserverDelegation)

}

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func BallotListKeyPrefix(p int64) []byte {
	return []byte(fmt.Sprint(p))
}

const (
	BlameKey = "Blame-"
	// TODO change identifier for VoterKey to something more descriptive
	VoterKey = "Voter-value-"

	// AllChainParamsKey is the ke prefix for all chain params
	// NOTE: CoreParams is old name for AllChainParams we keep it as key value for backward compatibility
	AllChainParamsKey = "CoreParams"

	ObserverMapperKey             = "Observer-value-"
	RelayerSetKey                 = "ObserverSet-value-"
	ObserverParamsKey             = "ObserverParams"
	AdminPolicyParamsKey          = "AdminParams"
	BallotMaturityBlocksParamsKey = "BallotMaturityBlocksParams"

	// CrosschainFlagsKey is the key for the crosschain flags
	// NOTE: PermissionFlags is old name for CrosschainFlags we keep it as key value for backward compatibility
	CrosschainFlagsKey    = "PermissionFlags-value-"
	CrosschainEventFeeKey = "CrosschainEventFee-value"

	LastBlockObserverCountKey = "ObserverCount-value-"
	NodeAccountKey            = "NodeAccount-value-"
	KeygenKey                 = "Keygen-value-"
	BlockHeaderKey            = "BlockHeader-value-"
	BlockHeaderStateKey       = "BlockHeaderState-value-"

	BallotListKey      = "BallotList-value-"
	TSSKey             = "TSS-value-"
	TSSHistoryKey      = "TSS-History-value-"
	TssFundMigratorKey = "FundsMigrator-value-"

	PendingNoncesKeyPrefix = "PendingNonces-value-"
	ChainNoncesKey         = "ChainNonces-value-"
	NonceToXmsgKeyPrefix   = "NonceToXmsg-value-"

	AddPellTokenBallotPrefix = "AddPellTokenBallot-"
	AddGasTokenBallotPrefix  = "AddGasTokenBallot-"
)

func GetBlameIndex(chainID int64, nonce uint64, digest string, height uint64) string {
	return fmt.Sprintf("%d-%d-%s-%d", chainID, nonce, digest, height)
}

func GetBlamePrefix(chainID int64, nonce int64) string {
	return fmt.Sprintf("%d-%d", chainID, nonce)
}
