package types

const (
	ModuleName = "xsecurity"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName
)

var (
	KeyPrefixEpochNumber     = []byte{0x01}
	KeyBlocksPerEpoch        = []byte{0x02}
	LastNativeVotingPowerKey = []byte{0x03}

	LSTRegistryRouterAddressKey          = []byte{0x11}
	LSTGroupInfoKey                      = []byte{0x12}
	LSTOperatorRegistrationListKey       = []byte{0x13}
	LSTOperatorWeightedShareKey          = []byte{0x14}
	LSTLastRoundOperatorWeightedShareKey = []byte{0x15}
	LSTVotingPowerRatioKey               = []byte{0x16}
	LSTStakingEnabledKey                 = []byte{0x17}
)
