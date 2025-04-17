package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpsertCrosschainFeeParams = "upsert_crosschain_fee_params"

var _ sdk.Msg = &MsgUpsertCrosschainFeeParams{}

func NewMsgUpsertCrosschainFeeParams(creator string, crosschainEventFees []*CrosschainFeeParam) *MsgUpsertCrosschainFeeParams {
	return &MsgUpsertCrosschainFeeParams{
		Signer:              creator,
		CrosschainFeeParams: crosschainEventFees,
	}
}

func (msg *MsgUpsertCrosschainFeeParams) Route() string {
	return RouterKey
}

func (msg *MsgUpsertCrosschainFeeParams) Type() string {
	return TypeMsgUpsertCrosschainFeeParams
}

func (msg *MsgUpsertCrosschainFeeParams) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpsertCrosschainFeeParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpsertCrosschainFeeParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	return nil
}
