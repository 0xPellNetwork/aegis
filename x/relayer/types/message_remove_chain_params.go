package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/0xPellNetwork/aegis/pkg/chains"
)

const TypeMsgRemoveChainParams = "remove_chain_params"

var _ sdk.Msg = &MsgRemoveChainParams{}

func NewMsgRemoveChainParams(creator string, chainID int64) *MsgRemoveChainParams {
	return &MsgRemoveChainParams{
		Signer:  creator,
		ChainId: chainID,
	}
}

func (msg *MsgRemoveChainParams) Route() string {
	return RouterKey
}

func (msg *MsgRemoveChainParams) Type() string {
	return TypeMsgRemoveChainParams
}

func (msg *MsgRemoveChainParams) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRemoveChainParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveChainParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Check if chain exists
	if _, exist := chains.GetChainByChainId(msg.ChainId); !exist {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidChainID, "invalid chain id (%d)", msg.ChainId)
	}

	return nil
}
