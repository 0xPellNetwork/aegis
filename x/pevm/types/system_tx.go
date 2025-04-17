package types

// System tx type occupies the first byte (8 bits)
const (
	SystemTxTypeSyncDelegationShares   uint8 = 1
	SystemTxTypeSyncOperatorRegistered uint8 = 2
	SystemTxTypeSyncDVSGroup           uint8 = 3
)
