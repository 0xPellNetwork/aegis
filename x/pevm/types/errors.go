package types

// DONTCOVER

import (
	cosmoserrors "cosmossdk.io/errors"
)

// x/pevm module sentinel errors
var (
	ErrABIPack                = cosmoserrors.Register(ModuleName, 1101, "failed to pack abi")
	ErrABIGet                 = cosmoserrors.Register(ModuleName, 1102, "failed to get abi")
	ErrABIUnpack              = cosmoserrors.Register(ModuleName, 1104, "failed to unpack abi")
	ErrContractNotFound       = cosmoserrors.Register(ModuleName, 1107, "contract not found")
	ErrContractCall           = cosmoserrors.Register(ModuleName, 1109, "contract call error")
	ErrSystemContractNotFound = cosmoserrors.Register(ModuleName, 1110, "system contract not found")
	ErrInvalidAddress         = cosmoserrors.Register(ModuleName, 1111, "invalid address")
	ErrStateVariableNotFound  = cosmoserrors.Register(ModuleName, 1112, "state variable not found")
	ErrEmitEvent              = cosmoserrors.Register(ModuleName, 1114, "emit event error")
	ErrInvalidDecimals        = cosmoserrors.Register(ModuleName, 1115, "invalid decimals")
	ErrInvalidGasLimit        = cosmoserrors.Register(ModuleName, 1118, "invalid gas limit")
	ErrSetBytecode            = cosmoserrors.Register(ModuleName, 1119, "set bytecode error")
	ErrInvalidContract        = cosmoserrors.Register(ModuleName, 1120, "invalid contract")
	ErrCallNonContract        = cosmoserrors.Register(ModuleName, 1124, "can't call a non-contract address")
	ErrNilGasPrice            = cosmoserrors.Register(ModuleName, 1127, "nil gas price")
)
