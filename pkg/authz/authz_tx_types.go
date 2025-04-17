package authz

type TxType string

const (
	InboundVoter      TxType = "InboundVoter"
	InboundBlockVoter TxType = "InboundBlockVoter"
	OutboundVoter     TxType = "OutboundVoter"
	NonceVoter        TxType = "NonceVoter"
	GasPriceVoter     TxType = "GasPriceVoter"
	XmsgSender        TxType = "XmsgSender"
)

func (t TxType) String() string {
	return string(t)
}

type KeyType string

const (
	TssSignerKey         KeyType = "tss_signer"
	ValidatorGranteeKey  KeyType = "validator_grantee"
	PellClientGranteeKey KeyType = "pellclient_grantee"
)

func GetAllKeyTypes() []KeyType {
	return []KeyType{ValidatorGranteeKey, PellClientGranteeKey, TssSignerKey}
}

func (k KeyType) String() string {
	return string(k)
}
