package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	// Chain related errors (1101-1110)
	ErrInvalidChainID          = errorsmod.Register(ModuleName, 1101, "chain id cannot be negative")
	ErrUnsupportedChain        = errorsmod.Register(ModuleName, 1102, "chain parse error")
	ErrUnableToGetGasPrice     = errorsmod.Register(ModuleName, 1107, "unable to get gas price")
	ErrNotEnoughPellBurnt      = errorsmod.Register(ModuleName, 1109, "not enough pell burnt")
	ErrCannotFindReceiverNonce = errorsmod.Register(ModuleName, 1110, "cannot find receiver chain nonce")

	// Asset and coin related errors (1113-1119)
	ErrGasCoinNotFound         = errorsmod.Register(ModuleName, 1113, "gas coin not found for sender chain")
	ErrUnableToParseAddress    = errorsmod.Register(ModuleName, 1115, "cannot parse address and data")
	ErrCannotProcessWithdrawal = errorsmod.Register(ModuleName, 1116, "cannot process withdrawal event")
	ErrForeignCoinNotFound     = errorsmod.Register(ModuleName, 1118, "foreign coin not found for sender chain")

	// TSS and nonce related errors (1121-1123)
	ErrCannotFindPendingNonces = errorsmod.Register(ModuleName, 1121, "cannot find pending nonces")
	ErrCannotFindTSSKeys       = errorsmod.Register(ModuleName, 1122, "cannot find TSS keys")
	ErrNonceMismatch           = errorsmod.Register(ModuleName, 1123, "nonce mismatch")

	// Transaction and address related errors (1127-1132)
	ErrUnableToSendCoinType = errorsmod.Register(ModuleName, 1127, "unable to send this coin type to a receiver chain")
	ErrInvalidAddress       = errorsmod.Register(ModuleName, 1128, "invalid address")
	ErrDeployContract       = errorsmod.Register(ModuleName, 1129, "unable to deploy contract")
	ErrUnableToUpdateTss    = errorsmod.Register(ModuleName, 1130, "unable to update TSS address")
	ErrNotEnoughGas         = errorsmod.Register(ModuleName, 1131, "not enough gas")
	ErrNotEnoughFunds       = errorsmod.Register(ModuleName, 1132, "not enough funds")

	// Verification and status related errors (1133-1139)
	ErrProofVerificationFail = errorsmod.Register(ModuleName, 1133, "proof verification fail")
	ErrCannotFindXmsg        = errorsmod.Register(ModuleName, 1134, "cannot find xmsg")
	ErrStatusNotPending      = errorsmod.Register(ModuleName, 1135, "Status not pending")
	ErrCannotFindGasParams   = errorsmod.Register(ModuleName, 1136, "cannot find gas params")
	ErrInvalidGasAmount      = errorsmod.Register(ModuleName, 1137, "invalid gas amount")
	ErrNoLiquidityPool       = errorsmod.Register(ModuleName, 1138, "no liquidity pool")
	ErrInvalidCoinType       = errorsmod.Register(ModuleName, 1139, "invalid coin type")

	// TSS migration and transaction verification errors (1140-1146)
	ErrCannotMigrateTssFunds         = errorsmod.Register(ModuleName, 1140, "cannot migrate TSS funds")
	ErrTxBodyVerificationFail        = errorsmod.Register(ModuleName, 1141, "transaction body verification fail")
	ErrReceiverIsEmpty               = errorsmod.Register(ModuleName, 1142, "receiver is empty")
	ErrUnsupportedStatus             = errorsmod.Register(ModuleName, 1143, "unsupported status")
	ErrObservedTxAlreadyFinalized    = errorsmod.Register(ModuleName, 1144, "observed tx already finalized")
	ErrInsufficientFundsTssMigration = errorsmod.Register(ModuleName, 1145, "insufficient funds for TSS migration")
	ErrInvalidIndexValue             = errorsmod.Register(ModuleName, 1146, "invalid index hash")

	// Status and processing related errors (1147-1152)
	ErrInvalidStatus               = errorsmod.Register(ModuleName, 1147, "invalid xmsg status")
	ErrUnableProcessRefund         = errorsmod.Register(ModuleName, 1148, "unable to process refund")
	ErrUnableToFindPellAccounting  = errorsmod.Register(ModuleName, 1149, "unable to find pell accounting")
	ErrInsufficientPellAmount      = errorsmod.Register(ModuleName, 1150, "insufficient pell amount")
	ErrUnableToDecodeMessageString = errorsmod.Register(ModuleName, 1151, "unable to decode message string")
	ErrInvalidRateLimiterFlags     = errorsmod.Register(ModuleName, 1152, "invalid rate limiter flags")

	// Block and transaction tracking errors (1153-1155)
	ErrMaxTxOutTrackerHashesReached = errorsmod.Register(ModuleName, 1153, "max tx out tracker hashes reached")
	ErrInvalidXmsgBuilders          = errorsmod.Register(ModuleName, 1154, "invalid xmsg builders")
	ErrBlockProofAlreadyFinalized   = errorsmod.Register(ModuleName, 1155, "block proof already finalized")

	// inbound tx related errors (1156-1158)
	ErrInboundPrevEventNotFound = errorsmod.Register(ModuleName, 1156, "sequential error: inbound prev event not found")
	ErrInboundPrevBlockNotFound = errorsmod.Register(ModuleName, 1157, "sequential error: inbound prev block not found")
)
