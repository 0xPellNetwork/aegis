package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCreateRegistryRouter = "create_registry_router"

var _ sdk.Msg = &MsgCreateRegistryRouter{}

func NewMsgCreateRegistryRouter(signer, chainApprover, churnApprover, ejector, pauser, unpauser string, initialPausedStatus int64) *MsgCreateRegistryRouter {
	return &MsgCreateRegistryRouter{
		Signer:              signer,
		ChainApprover:       chainApprover,
		ChurnApprover:       churnApprover,
		Ejector:             ejector,
		Pauser:              pauser,
		Unpauser:            unpauser,
		InitialPausedStatus: initialPausedStatus,
	}
}

func (msg *MsgCreateRegistryRouter) Route() string {
	return RouterKey
}

func (msg *MsgCreateRegistryRouter) Type() string {
	return TypeMsgCreateRegistryRouter
}

func (msg *MsgCreateRegistryRouter) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (msg *MsgCreateRegistryRouter) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateRegistryRouter) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
