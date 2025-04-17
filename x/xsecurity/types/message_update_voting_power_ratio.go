package types

import (
	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypMsgUpdateVotingPowerRatio = "update_voting_power_ratio"

var _ sdk.Msg = &MsgUpdateVotingPowerRatio{}

func NewMsgUpdateVotingPowerRatio(signer string, numerator, denominator math.Int) *MsgUpdateVotingPowerRatio {
	return &MsgUpdateVotingPowerRatio{
		Signer:      signer,
		Numerator:   numerator,
		Denominator: denominator,
	}
}

func (msg *MsgUpdateVotingPowerRatio) Route() string {
	return RouterKey
}

func (msg *MsgUpdateVotingPowerRatio) Type() string {
	return TypMsgUpdateVotingPowerRatio
}

func (msg *MsgUpdateVotingPowerRatio) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (msg *MsgUpdateVotingPowerRatio) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateVotingPowerRatio) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
