package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/0xPellNetwork/aegis/pkg/chains"
)

const TypeMsgAddBlameVote = "add_blame_vote"

var _ sdk.Msg = &MsgAddBlameVote{}

func NewMsgAddBlameVoteMsg(signer string, chainID int64, blameInfo Blame) *MsgAddBlameVote {
	return &MsgAddBlameVote{
		Signer:    signer,
		ChainId:   chainID,
		BlameInfo: blameInfo,
	}
}

func (m *MsgAddBlameVote) Route() string {
	return RouterKey
}

func (m *MsgAddBlameVote) Type() string {
	return TypeMsgAddBlameVote
}

func (m *MsgAddBlameVote) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if _, exist := chains.GetChainByChainId(m.ChainId); !exist {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidChainID, "chain id (%d)", m.ChainId)
	}

	return nil
}

func (m *MsgAddBlameVote) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (m *MsgAddBlameVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgAddBlameVote) Digest() string {
	msg := *m
	msg.Signer = ""
	// Generate an Identifier for the ballot corresponding to specific blame data
	hash := crypto.Keccak256Hash([]byte(msg.String()))
	return hash.Hex()
}
