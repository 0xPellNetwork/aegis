package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/0xPellNetwork/aegis/pkg/chains"
)

const TypeMsgMigrateTssFunds = "MigrateTssFunds"

var _ sdk.Msg = &MsgMigrateTssFunds{}

func NewMsgMigrateTssFunds(creator string, chainID int64, amount sdkmath.Uint) *MsgMigrateTssFunds {
	return &MsgMigrateTssFunds{
		Signer:  creator,
		ChainId: chainID,
		Amount:  amount,
	}
}

func (msg *MsgMigrateTssFunds) Route() string {
	return RouterKey
}

func (msg *MsgMigrateTssFunds) Type() string {
	return TypeMsgMigrateTssFunds
}

func (msg *MsgMigrateTssFunds) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgMigrateTssFunds) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMigrateTssFunds) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if _, exist := chains.GetChainByChainId(msg.ChainId); !exist {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid chain id (%d)", msg.ChainId)
	}
	if msg.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "amount cannot be zero")
	}
	return nil
}
