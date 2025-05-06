package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/0xPellNetwork/aegis/x/restaking/types"
)

const TypMsgRemovePools = "create_remove_pools"

var _ sdk.Msg = &MsgRemovePools{}

func NewMsgRemovePools(signer string, poolParams []*types.PoolParams) *MsgRemovePools {
	return &MsgRemovePools{
		Signer: signer,
		Pools:  poolParams,
	}
}

func (msg *MsgRemovePools) Route() string {
	return RouterKey
}

func (msg *MsgRemovePools) Type() string {
	return TypMsgRemovePools
}

func (msg *MsgRemovePools) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (msg *MsgRemovePools) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemovePools) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
