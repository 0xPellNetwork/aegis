package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypMsgUpdateLSTStakingEnabled = "update_lst_staking_enabled"

var _ sdk.Msg = &MsgUpdateLSTStakingEnabled{}

func NewMsgUpdateLSTStakingEnabled(signer string, enabled bool) *MsgUpdateLSTStakingEnabled {
	return &MsgUpdateLSTStakingEnabled{
		Signer:  signer,
		Enabled: enabled,
	}
}

func (msg *MsgUpdateLSTStakingEnabled) Route() string {
	return RouterKey
}

func (msg *MsgUpdateLSTStakingEnabled) Type() string {
	return TypMsgUpdateLSTStakingEnabled
}

func (msg *MsgUpdateLSTStakingEnabled) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (msg *MsgUpdateLSTStakingEnabled) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateLSTStakingEnabled) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
