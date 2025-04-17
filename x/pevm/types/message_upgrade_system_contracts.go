package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpgradeSystemContracts = "upgrade_system_contract"

var _ sdk.Msg = &MsgUpgradeSystemContracts{}

func NewMsgUpgradeSystemContracts(creator string) *MsgUpgradeSystemContracts {
	return &MsgUpgradeSystemContracts{
		Signer: creator,
	}
}

func (msg *MsgUpgradeSystemContracts) Route() string {
	return RouterKey
}

func (msg *MsgUpgradeSystemContracts) Type() string {
	return TypeMsgUpgradeSystemContracts
}

func (msg *MsgUpgradeSystemContracts) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpgradeSystemContracts) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpgradeSystemContracts) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
