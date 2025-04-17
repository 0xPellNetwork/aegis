package types

import (
	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/pell-chain/pellcore/pkg/authz"
)

// MaxMessageLength is the maximum length of a message in a xmsg
// TODO: should parameterize the hardcoded max len
// FIXME: should allow this observation and handle errors in the state machine
const MaxMessageLength = 10240

var _ sdk.Msg = &MsgVoteOnObservedInboundTx{}

func NewMsgVoteOnObservedInboundTx(
	creator,
	sender string,
	senderChain int64,
	txOrigin,
	receiver string,
	receiverChain int64,
	inTxHash string,
	inBlockHeight,
	gasLimit uint64,
	eventIndex uint,
	pellTx InboundPellEvent,
) *MsgVoteOnObservedInboundTx {
	return &MsgVoteOnObservedInboundTx{
		Signer:        creator,
		Sender:        sender,
		SenderChainId: senderChain,
		TxOrigin:      txOrigin,
		Receiver:      receiver,
		ReceiverChain: receiverChain,
		InTxHash:      inTxHash,
		InBlockHeight: inBlockHeight,
		GasLimit:      gasLimit,
		EventIndex:    uint64(eventIndex),
		PellTx:        &pellTx,
	}
}

func (msg *MsgVoteOnObservedInboundTx) Route() string {
	return RouterKey
}

func (msg *MsgVoteOnObservedInboundTx) Type() string {
	return authz.InboundVoter.String()
}

func (msg *MsgVoteOnObservedInboundTx) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgVoteOnObservedInboundTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgVoteOnObservedInboundTx) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s): %s", err, msg.Signer)
	}

	if msg.SenderChainId < 0 {
		return cosmoserrors.Wrapf(ErrInvalidChainID, "chain id (%d)", msg.SenderChainId)
	}

	if msg.ReceiverChain < 0 {
		return cosmoserrors.Wrapf(ErrInvalidChainID, "chain id (%d)", msg.ReceiverChain)
	}

	return nil
}

func (msg *MsgVoteOnObservedInboundTx) Digest() string {
	m := *msg
	m.Signer = ""
	m.InBlockHeight = 0
	hash := crypto.Keccak256Hash([]byte(m.String()))
	return hash.Hex()
}

func (msg *MsgVoteOnObservedInboundTx) IsPellMsg() bool {
	return msg.PellTx != nil && msg.PellTx.isAvailable()
}

func (msg *MsgVoteInboundBlock) Digest() string {
	m := *msg
	m.Signer = ""

	hash := crypto.Keccak256Hash([]byte(m.String()))
	return hash.Hex()
}

// msg vote block proof
func NewMsgVoteInboundBlock(
	creator string,
	chainId, prevBlockHeight, blockHeight uint64,
	blockHash string,
	events []*Event,
) *MsgVoteInboundBlock {
	return &MsgVoteInboundBlock{
		Signer: creator,
		BlockProof: &BlockProof{
			ChainId:         chainId,
			PrevBlockHeight: prevBlockHeight,
			BlockHeight:     blockHeight,
			BlockHash:       blockHash,
			Events:          events,
		},
	}
}

func (msg *MsgVoteInboundBlock) Route() string {
	return RouterKey
}

func (msg *MsgVoteInboundBlock) Type() string {
	return authz.InboundVoter.String()
}

func (msg *MsgVoteInboundBlock) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgVoteInboundBlock) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)

	return sdk.MustSortJSON(bz)
}

func (msg *MsgVoteInboundBlock) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s): %s", err, msg.Signer)
	}

	if msg.BlockProof.PrevBlockHeight >= msg.BlockProof.BlockHeight {
		return ErrInvalidLengthLastBlockHeight
	}

	return nil
}
