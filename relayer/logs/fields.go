package logs

// A group of predefined field keys and module names for pellclient logs
const (
	// field keys
	FieldModule = "module"
	FieldMethod = "method"
	FieldChain  = "chain"
	FieldNonce  = "nonce"
	FieldTx     = "tx"
	FieldXmsg   = "xmsg"

	// module names
	ModNameInbound  = "inbound"
	ModNameOutbound = "outbound"
	ModNameGasPrice = "gasprice"
	ModNameHeaders  = "headers"
)
