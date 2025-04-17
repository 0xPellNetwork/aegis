package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDeployConnectorContract = "deploy_connector_contract"

var _ sdk.Msg = &MsgDeployConnectorContract{}

func NewMsgDeployConnectorContract(creator string) *MsgDeployConnectorContract {
	return &MsgDeployConnectorContract{
		Signer: creator,
	}
}

func (msg *MsgDeployConnectorContract) Route() string {
	return RouterKey
}

func (msg *MsgDeployConnectorContract) Type() string {
	return TypeMsgDeployConnectorContract
}

func (msg *MsgDeployConnectorContract) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeployConnectorContract) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeployConnectorContract) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
