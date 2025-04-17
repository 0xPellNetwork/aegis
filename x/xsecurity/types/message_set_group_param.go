package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

const TypMsgSetGroupParam = "create_add_pools"

var _ sdk.Msg = &MsgSetGroupParam{}

func NewMsgSetGroupParam(signer string, param *types.OperatorSetParam) *MsgSetGroupParam {
	return &MsgSetGroupParam{
		Signer:            signer,
		OperatorSetParams: param,
	}
}

func (msg *MsgSetGroupParam) Route() string {
	return RouterKey
}

func (msg *MsgSetGroupParam) Type() string {
	return TypMsgSetGroupParam
}

func (msg *MsgSetGroupParam) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (msg *MsgSetGroupParam) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetGroupParam) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
