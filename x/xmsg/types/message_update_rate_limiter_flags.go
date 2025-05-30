package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateRateLimiterFlags = "UpdateRateLimiterFlags"

var _ sdk.Msg = &MsgUpdateRateLimiterFlags{}

func NewMsgUpdateRateLimiterFlags(creator string, flags RateLimiterFlags) *MsgUpdateRateLimiterFlags {
	return &MsgUpdateRateLimiterFlags{
		Signer:           creator,
		RateLimiterFlags: flags,
	}
}

func (msg *MsgUpdateRateLimiterFlags) Route() string {
	return RouterKey
}

func (msg *MsgUpdateRateLimiterFlags) Type() string {
	return TypeMsgUpdateRateLimiterFlags
}

func (msg *MsgUpdateRateLimiterFlags) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateRateLimiterFlags) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateRateLimiterFlags) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := msg.RateLimiterFlags.Validate(); err != nil {
		return errorsmod.Wrapf(ErrInvalidRateLimiterFlags, err.Error())
	}
	return nil
}
