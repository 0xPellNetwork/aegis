package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/0xPellNetwork/aegis/pkg/proofs"
)

const TypeMsgAddToOutTxTracker = "AddToTracker"

var _ sdk.Msg = &MsgAddToOutTxTracker{}

func NewMsgAddToOutTxTracker(
	creator string,
	chain int64,
	nonce uint64,
	txHash string,
	proof *proofs.Proof,
	blockHash string,
	txIndex int64,
) *MsgAddToOutTxTracker {
	return &MsgAddToOutTxTracker{
		Signer:    creator,
		ChainId:   chain,
		Nonce:     nonce,
		TxHash:    txHash,
		Proof:     proof,
		BlockHash: blockHash,
		TxIndex:   txIndex,
	}
}

func (msg *MsgAddToOutTxTracker) Route() string {
	return RouterKey
}

func (msg *MsgAddToOutTxTracker) Type() string {
	return TypeMsgAddToOutTxTracker
}

func (msg *MsgAddToOutTxTracker) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAddToOutTxTracker) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddToOutTxTracker) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if msg.ChainId < 0 {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidChainID, "chain id (%d)", msg.ChainId)
	}
	return nil
}
