package types

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

func (m Xmsg) IsCrossChainPellTx() bool {
	rc := false

	if m.InboundTxParams.InboundPellTx != nil {
		rc = m.InboundTxParams.InboundPellTx.isAvailable()
	}
	return rc
}

// GetCurrentOutTxParam returns the current outbound tx params.
// There can only be one active outtx.
// OutboundTxParams[0] is the original outtx, if it reverts, then
// OutboundTxParams[1] is the new outtx.
func (m Xmsg) GetCurrentOutTxParam() *OutboundTxParams {
	if len(m.OutboundTxParams) == 0 {
		return &OutboundTxParams{}
	}
	return m.OutboundTxParams[len(m.OutboundTxParams)-1]
}

// IsCurrentOutTxRevert returns true if the current outbound tx is the revert tx.
func (m Xmsg) IsCurrentOutTxRevert() bool {
	return len(m.OutboundTxParams) >= 2
}

// OriginalDestinationChainID returns the original destination of the outbound tx, reverted or not
// If there is no outbound tx, return -1
func (m Xmsg) OriginalDestinationChainID() int64 {
	if len(m.OutboundTxParams) == 0 {
		return -1
	}
	return m.OutboundTxParams[0].ReceiverChainId
}

// Validate checks if the Xmsg is valid.
func (m Xmsg) Validate() error {
	if m.InboundTxParams == nil {
		return fmt.Errorf("inbound tx params cannot be nil")
	}
	if m.OutboundTxParams == nil {
		return fmt.Errorf("outbound tx params cannot be nil")
	}
	if m.XmsgStatus == nil {
		return fmt.Errorf("xmsg status cannot be nil")
	}
	if len(m.OutboundTxParams) > 2 {
		return fmt.Errorf("outbound tx params cannot be more than 2")
	}
	if m.Index != "" {
		err := ValidatePellIndex(m.Index)
		if err != nil {
			return err
		}
	}
	err := m.InboundTxParams.Validate()
	if err != nil {
		return err
	}
	for _, outboundTxParam := range m.OutboundTxParams {
		err = outboundTxParam.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

/*
AddRevertOutbound does the following things in one function:

	1. create a new OutboundTxParams for the revert

	2. append the new OutboundTxParams to the current OutboundTxParams

	3. update the TxFinalizationStatus of the current OutboundTxParams to Executed.
*/

func (m *Xmsg) AddRevertOutbound(gasLimit uint64) error {
	if m.IsCurrentOutTxRevert() {
		return fmt.Errorf("cannot revert a revert tx")
	}
	if len(m.OutboundTxParams) == 0 {
		return fmt.Errorf("cannot revert before trying to process an outbound tx")
	}

	revertTxParams := &OutboundTxParams{
		Receiver:           m.InboundTxParams.Sender,
		ReceiverChainId:    m.InboundTxParams.SenderChainId,
		OutboundTxGasLimit: gasLimit,
		TssPubkey:          m.GetCurrentOutTxParam().TssPubkey,
	}
	// The original outbound has been finalized, the new outbound is pending
	m.GetCurrentOutTxParam().TxFinalizationStatus = TxFinalizationStatus_EXECUTED
	m.OutboundTxParams = append(m.OutboundTxParams, revertTxParams)
	return nil
}

// AddOutbound adds a new outbound tx to the Xmsg.
func (m *Xmsg) AddOutbound(ctx sdk.Context, msg MsgVoteOnObservedOutboundTx, ballotStatus relayertypes.BallotStatus) error {
	if ballotStatus != relayertypes.BallotStatus_BALLOT_FINALIZED_FAILURE_OBSERVATION {
		// todo  result == expect?
	}
	// Update Xmsg values
	m.GetCurrentOutTxParam().OutboundTxHash = msg.ObservedOutTxHash
	m.GetCurrentOutTxParam().OutboundTxGasUsed = msg.ObservedOutTxGasUsed
	m.GetCurrentOutTxParam().OutboundTxEffectiveGasPrice = msg.ObservedOutTxEffectiveGasPrice
	m.GetCurrentOutTxParam().OutboundTxEffectiveGasLimit = msg.ObservedOutTxEffectiveGasLimit
	m.GetCurrentOutTxParam().OutboundTxExternalHeight = msg.ObservedOutTxBlockHeight
	m.XmsgStatus.LastUpdateTimestamp = ctx.BlockHeader().Time.Unix()
	m.XmsgStatus.StatusMessage = msg.ObservedOutTxFailedReasonMsg

	return nil
}

// SetAbort sets the Xmsg status to Aborted with the given error message.
func (m Xmsg) SetAbort(message string) {
	m.XmsgStatus.ChangeStatus(XmsgStatus_ABORTED, message)
}

// SetPendingRevert sets the Xmsg status to PendingRevert with the given error message.
func (m Xmsg) SetPendingRevert(message string) {
	m.XmsgStatus.ChangeStatus(XmsgStatus_PENDING_REVERT, message)
}

// SetPendingOutbound sets the Xmsg status to PendingOutbound with the given error message.
func (m Xmsg) SetPendingOutbound(message string) {
	m.XmsgStatus.ChangeStatus(XmsgStatus_PENDING_OUTBOUND, message)
}

// SetOutBoundMined sets the Xmsg status to OutboundMined with the given error message.
func (m Xmsg) SetOutBoundMined(message string) {
	m.XmsgStatus.ChangeStatus(XmsgStatus_OUTBOUND_MINED, message)
}

// SetReverted sets the Xmsg status to Reverted with the given error message.
func (m Xmsg) SetReverted(message string) {
	m.XmsgStatus.ChangeStatus(XmsgStatus_REVERTED, message)
}

func (m Xmsg) GetXmsgIndicesBytes() ([32]byte, error) {
	sendHash := [32]byte{}
	if len(m.Index) < 2 {
		return [32]byte{}, fmt.Errorf("decode Xmsg %s index too short", m.Index)
	}
	decodedIndex, err := hex.DecodeString(m.Index[2:]) // remove the leading 0x
	if err != nil || len(decodedIndex) != 32 {
		return [32]byte{}, fmt.Errorf("decode Xmsg %s error", m.Index)
	}
	copy(sendHash[:32], decodedIndex[:32])
	return sendHash, nil
}

func GetXmsgIndicesFromBytes(sendHash [32]byte) string {
	return fmt.Sprintf("0x%s", hex.EncodeToString(sendHash[:]))
}

// NewXmsg creates a new Xmsg.From a MsgVoteOnObservedInboundTx message and a TSS pubkey.
// It also validates the created xmsg
func NewXmsg(ctx sdk.Context, msg MsgVoteOnObservedInboundTx, tssPubkey string) (Xmsg, error) {
	index := msg.Digest()

	if msg.TxOrigin == "" {
		msg.TxOrigin = msg.Sender
	}
	inboundParams := &InboundTxParams{
		Sender:                       msg.Sender,
		SenderChainId:                msg.SenderChainId,
		TxOrigin:                     msg.TxOrigin,
		InboundTxHash:                msg.InTxHash,
		InboundTxBlockHeight:         msg.InBlockHeight,
		InboundTxFinalizedPellHeight: 0,
		InboundTxBallotIndex:         index,
		InboundTxEventIndex:          msg.EventIndex,
		InboundPellTx:                msg.PellTx,
	}

	outBoundParams := &OutboundTxParams{
		Receiver:                 msg.Receiver,
		ReceiverChainId:          msg.ReceiverChain,
		OutboundTxHash:           "",
		OutboundTxTssNonce:       0,
		OutboundTxGasLimit:       msg.GasLimit,
		OutboundTxGasPrice:       "",
		OutboundTxBallotIndex:    "",
		OutboundTxExternalHeight: 0,
		TssPubkey:                tssPubkey,
	}
	status := &Status{
		Status:              XmsgStatus_PENDING_INBOUND,
		StatusMessage:       "",
		LastUpdateTimestamp: ctx.BlockHeader().Time.Unix(),
	}
	xmsg := Xmsg{
		Signer:           msg.Signer,
		Index:            index,
		XmsgStatus:       status,
		InboundTxParams:  inboundParams,
		OutboundTxParams: []*OutboundTxParams{outBoundParams},
	}
	err := xmsg.Validate()
	if err != nil {
		return Xmsg{}, err
	}
	return xmsg, nil
}
