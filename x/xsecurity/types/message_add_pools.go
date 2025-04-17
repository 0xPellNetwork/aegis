package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

const TypMsgAddPools = "create_add_pools"

var _ sdk.Msg = &MsgAddPools{}

func NewMsgAddPools(signer string, poolParams []*types.PoolParams) *MsgAddPools {
	return &MsgAddPools{
		Signer: signer,
		Pools:  poolParams,
	}
}

func (msg *MsgAddPools) Route() string {
	return RouterKey
}

func (msg *MsgAddPools) Type() string {
	return TypMsgAddPools
}

func (msg *MsgAddPools) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (msg *MsgAddPools) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddPools) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
