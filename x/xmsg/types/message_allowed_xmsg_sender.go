package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/pell-chain/pellcore/pkg/authz"
)

// add allowed xmsg sender
func NewMsgAddAllowedXmsgSender(signer string, builders []string) *MsgAddAllowedXmsgSender {
	return &MsgAddAllowedXmsgSender{
		Signer:   signer,
		Builders: builders,
	}
}

// route of add allowed xmsg sender
func (msg *MsgAddAllowedXmsgSender) Route() string {
	return RouterKey
}

// type of add allowed xmsg sender
func (msg *MsgAddAllowedXmsgSender) Type() string {
	return string(authz.XmsgSender)
}

// get signers of add allowed xmsg sender
func (msg *MsgAddAllowedXmsgSender) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// get sign bytes of add allowed xmsg sender
func (msg *MsgAddAllowedXmsgSender) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// validate basic of add allowed xmsg sender
func (msg *MsgAddAllowedXmsgSender) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid signer address (%s)", err)
	}

	if len(msg.Builders) == 0 {
		return cosmoserrors.Wrapf(ErrInvalidXmsgBuilders, "builders cannot be empty")
	}

	return nil
}

// digest of add allowed xmsg sender
func (msg *MsgAddAllowedXmsgSender) Digest() string {
	m := *msg
	m.Signer = ""

	hash := crypto.Keccak256Hash([]byte(m.String()))
	return hash.Hex()
}

// remove allowed xmsg sender
func NewMsgRemoveAllowedXmsgSender(signer string, builders []string) *MsgRemoveAllowedXmsgSender {
	return &MsgRemoveAllowedXmsgSender{
		Signer:   signer,
		Builders: builders,
	}
}

// route of remove allowed xmsg sender
func (msg *MsgRemoveAllowedXmsgSender) Route() string {
	return RouterKey
}

// type of remove allowed xmsg sender
func (msg *MsgRemoveAllowedXmsgSender) Type() string {
	return string(authz.XmsgSender)
}

// get signers of remove allowed xmsg sender
func (msg *MsgRemoveAllowedXmsgSender) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// get sign bytes of remove allowed xmsg sender
func (msg *MsgRemoveAllowedXmsgSender) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// validate basic of remove xmsg builders
func (msg *MsgRemoveAllowedXmsgSender) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid signer address (%s)", err)
	}

	if len(msg.Builders) == 0 {
		return cosmoserrors.Wrapf(ErrInvalidXmsgBuilders, "builders cannot be empty")
	}

	return nil
}

// digest of remove xmsg builders
func (msg *MsgRemoveAllowedXmsgSender) Digest() string {
	m := *msg
	m.Signer = ""

	hash := crypto.Keccak256Hash([]byte(m.String()))
	return hash.Hex()
}
