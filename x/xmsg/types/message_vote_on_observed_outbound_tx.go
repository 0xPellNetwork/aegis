package types

import (
	cosmoserrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/0xPellNetwork/aegis/pkg/authz"
	"github.com/0xPellNetwork/aegis/pkg/chains"
)

var _ sdk.Msg = &MsgVoteOnObservedOutboundTx{}

func NewMsgVoteOnObservedOutboundTx(
	creator,
	sendHash,
	outTxHash string,
	outBlockHeight,
	outTxGasUsed uint64,
	outTxEffectiveGasPrice math.Int,
	outTxEffectiveGasLimit uint64,
	status chains.ReceiveStatus,
	failedReasonMsg string,
	chain int64,
	nonce uint64,
) *MsgVoteOnObservedOutboundTx {
	return &MsgVoteOnObservedOutboundTx{
		Signer:                         creator,
		XmsgHash:                       sendHash,
		ObservedOutTxHash:              outTxHash,
		ObservedOutTxBlockHeight:       outBlockHeight,
		ObservedOutTxGasUsed:           outTxGasUsed,
		ObservedOutTxEffectiveGasPrice: outTxEffectiveGasPrice,
		ObservedOutTxEffectiveGasLimit: outTxEffectiveGasLimit,
		Status:                         status,
		ObservedOutTxFailedReasonMsg:   failedReasonMsg,
		OutTxChain:                     chain,
		OutTxTssNonce:                  nonce,
	}
}

func (msg *MsgVoteOnObservedOutboundTx) Route() string {
	return RouterKey
}

func (msg *MsgVoteOnObservedOutboundTx) Type() string {
	return authz.OutboundVoter.String()
}

func (msg *MsgVoteOnObservedOutboundTx) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgVoteOnObservedOutboundTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgVoteOnObservedOutboundTx) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if msg.OutTxChain < 0 {
		return cosmoserrors.Wrapf(ErrInvalidChainID, "chain id (%d)", msg.OutTxChain)
	}

	return nil
}

func (msg *MsgVoteOnObservedOutboundTx) Digest() string {
	m := *msg
	m.Signer = ""

	// Set status to ReceiveStatus_Created to make sure both successful and failed votes are added to the same ballot
	m.Status = chains.ReceiveStatus_CREATED

	// Outbound and reverted txs have different digest as ObservedOutTxHash is different so they are stored in different ballots
	hash := crypto.Keccak256Hash([]byte(m.String()))
	return hash.Hex()
}
