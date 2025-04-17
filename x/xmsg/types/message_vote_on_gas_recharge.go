package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/pell-chain/pellcore/pkg/authz"
)

var _ sdk.Msg = &MsgVoteOnGasRecharge{}

func NewMsgVoteOnGasRecharge(creator string, chain int64, index uint64) *MsgVoteOnGasRecharge {
	return &MsgVoteOnGasRecharge{
		Signer:    creator,
		ChainId:   chain,
		VoteIndex: index,
	}
}

func (msg *MsgVoteOnGasRecharge) Route() string {
	return RouterKey
}

func (msg *MsgVoteOnGasRecharge) Type() string {
	return authz.GasPriceVoter.String()
}

func (msg *MsgVoteOnGasRecharge) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgVoteOnGasRecharge) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgVoteOnGasRecharge) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if msg.ChainId < 0 {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidChainID, "chain id (%d)", msg.ChainId)
	}
	return nil
}
