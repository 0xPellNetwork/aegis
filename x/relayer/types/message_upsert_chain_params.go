package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpsertChainParams = "update_chain_params"

var _ sdk.Msg = &MsgUpsertChainParams{}

func NewMsgUpsertChainParams(creator string, chainParams *ChainParams) *MsgUpsertChainParams {
	return &MsgUpsertChainParams{
		Signer:      creator,
		ChainParams: chainParams,
	}
}

func (msg *MsgUpsertChainParams) Route() string {
	return RouterKey
}

func (msg *MsgUpsertChainParams) Type() string {
	return TypeMsgUpsertChainParams
}

func (msg *MsgUpsertChainParams) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpsertChainParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpsertChainParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if err := ValidateChainParams(msg.ChainParams); err != nil {
		return cosmoserrors.Wrapf(ErrInvalidChainParams, err.Error())
	}

	return nil
}
