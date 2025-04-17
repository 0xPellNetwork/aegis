package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/pell-chain/pellcore/x/restaking/types"
)

const TypMsgCreateGroup = "create_dvs_group"

var _ sdk.Msg = &MsgCreateGroup{}

func NewMsgCreateGroup(signer string, operatorSetParams *types.OperatorSetParam, poolParams []*types.PoolParams, groupEjectionParams *types.GroupEjectionParam, minStake sdkmath.Int) *MsgCreateGroup {
	return &MsgCreateGroup{
		Signer:              signer,
		OperatorSetParams:   operatorSetParams,
		PoolParams:          poolParams,
		GroupEjectionParams: groupEjectionParams,
		MinStake:            minStake,
	}
}

func (msg *MsgCreateGroup) Route() string {
	return RouterKey
}

func (msg *MsgCreateGroup) Type() string {
	return TypMsgCreateGroup
}

func (msg *MsgCreateGroup) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

func (msg *MsgCreateGroup) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateGroup) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
