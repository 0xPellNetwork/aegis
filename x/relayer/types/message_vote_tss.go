package types

import (
	"fmt"

	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/pell-chain/pellcore/pkg/chains"
)

const TypeMsgVoteTSS = "VoteTSS"

var _ sdk.Msg = &MsgVoteTSS{}

func NewMsgVoteTSS(signer string, pubkey string, keygenPellHeight int64, status chains.ReceiveStatus) *MsgVoteTSS {
	return &MsgVoteTSS{
		Signer:           signer,
		TssPubkey:        pubkey,
		KeygenPellHeight: keygenPellHeight,
		Status:           status,
	}
}

func (msg *MsgVoteTSS) Route() string {
	return RouterKey
}

func (msg *MsgVoteTSS) Type() string {
	return TypeMsgVoteTSS
}

func (msg *MsgVoteTSS) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgVoteTSS) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgVoteTSS) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// either success or observation failure
	if msg.Status != chains.ReceiveStatus_SUCCESS && msg.Status != chains.ReceiveStatus_FAILED {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid status: %s", msg.Status)
	}

	return nil
}

func (msg *MsgVoteTSS) Digest() string {
	// We support only 1 keygen at a particular height
	return fmt.Sprintf("%d-%s", msg.KeygenPellHeight, "tss-keygen")
}
