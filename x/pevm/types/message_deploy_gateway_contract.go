package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDeployGatewayContract = "deploy_gateway_contract"

var _ sdk.Msg = &MsgDeployGatewayContract{}

func NewMsgDeployGatewayContract(creator string) *MsgDeployGatewayContract {
	return &MsgDeployGatewayContract{
		Signer: creator,
	}
}

func (msg *MsgDeployGatewayContract) Route() string {
	return RouterKey
}

func (msg *MsgDeployGatewayContract) Type() string {
	return TypeMsgDeployGatewayContract
}

func (msg *MsgDeployGatewayContract) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeployGatewayContract) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeployGatewayContract) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
