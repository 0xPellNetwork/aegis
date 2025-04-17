package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/pell-chain/pellcore/pkg/authz"
)

var _ sdk.Msg = &MsgVoteOnPellRecharge{}

func NewMsgVoteOnPellRecharge(creator string, chain int64, index uint64) *MsgVoteOnPellRecharge {
	return &MsgVoteOnPellRecharge{
		Signer:    creator,
		ChainId:   chain,
		VoteIndex: index,
	}
}

func (msg *MsgVoteOnPellRecharge) Route() string {
	return RouterKey
}

func (msg *MsgVoteOnPellRecharge) Type() string {
	return authz.GasPriceVoter.String()
}

func (msg *MsgVoteOnPellRecharge) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgVoteOnPellRecharge) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgVoteOnPellRecharge) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if msg.ChainId < 0 {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidChainID, "chain id (%d)", msg.ChainId)
	}
	return nil
}
