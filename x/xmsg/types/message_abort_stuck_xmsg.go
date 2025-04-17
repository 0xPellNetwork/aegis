package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAbortStuckXmsg = "AbortStuckXmsg"

var _ sdk.Msg = &MsgAbortStuckXmsg{}

func NewMsgAbortStuckXmsg(creator string, xmsgIndex string) *MsgAbortStuckXmsg {
	return &MsgAbortStuckXmsg{
		Signer:    creator,
		XmsgIndex: xmsgIndex,
	}
}

func (msg *MsgAbortStuckXmsg) Route() string {
	return RouterKey
}

func (msg *MsgAbortStuckXmsg) Type() string {
	return TypeMsgAbortStuckXmsg
}

func (msg *MsgAbortStuckXmsg) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAbortStuckXmsg) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAbortStuckXmsg) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if len(msg.XmsgIndex) != PellIndexLength {
		return ErrInvalidIndexValue
	}
	return nil
}
