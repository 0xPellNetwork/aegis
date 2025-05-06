package authz

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/0xPellNetwork/aegis/pkg/authz"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

type Signer struct {
	KeyType        authz.KeyType
	GranterAddress string
	GranteeAddress sdk.AccAddress
}

func (a Signer) String() string {
	return a.KeyType.String() + " " + a.GranterAddress + " " + a.GranteeAddress.String()
}

var signers map[string]Signer

func init() {
	signersList := make(map[string]Signer)
	for _, tx := range xmsgtypes.GetAllAuthzPellclientTxTypes() {
		signersList[tx] = Signer{KeyType: authz.PellClientGranteeKey}
	}
	signers = signersList
}

func SetupAuthZSignerList(granter string, grantee sdk.AccAddress) {
	for k, v := range signers {
		v.GranterAddress = granter
		v.GranteeAddress = grantee
		signers[k] = v
	}
}

func GetSigner(msgURL string) Signer {
	return signers[msgURL]
}
