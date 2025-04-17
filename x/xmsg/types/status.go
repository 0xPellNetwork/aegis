package types

import (
	"fmt"
)

// empty msg does not overwrite old status message
func (m *Status) ChangeStatus(newStatus XmsgStatus, msg string) {
	if len(msg) > 0 {
		if m.StatusMessage != "" {
			m.StatusMessage = fmt.Sprintf("%s : %s", m.StatusMessage, msg)
		} else {
			m.StatusMessage = msg
		}
	}
	if !m.ValidateTransition(newStatus) {
		m.StatusMessage = fmt.Sprintf("Failed to transition : OldStatus %s , NewStatus %s , MSG : %s :", m.Status.String(), newStatus.String(), msg)
		m.Status = XmsgStatus_ABORTED
		return
	}
	m.Status = newStatus

} //nolint:typecheck

func (m *Status) ValidateTransition(newStatus XmsgStatus) bool {
	stateTransitionMap := stateTransitionMap()
	oldStatus := m.Status
	nextStatusList, isOldStatusValid := stateTransitionMap[oldStatus]
	if !isOldStatusValid {
		return false
	}
	for _, status := range nextStatusList {
		if status == newStatus {
			return true
		}
	}
	return false
}

func stateTransitionMap() map[XmsgStatus][]XmsgStatus {
	stateTransitionMap := make(map[XmsgStatus][]XmsgStatus)
	stateTransitionMap[XmsgStatus_PENDING_INBOUND] = []XmsgStatus{
		XmsgStatus_PENDING_OUTBOUND,
		XmsgStatus_ABORTED,
		XmsgStatus_OUTBOUND_MINED, // EVM Deposit
		XmsgStatus_PENDING_REVERT, // EVM Deposit contract call reverted; should refund
	}
	stateTransitionMap[XmsgStatus_PENDING_OUTBOUND] = []XmsgStatus{
		XmsgStatus_ABORTED,
		XmsgStatus_PENDING_REVERT,
		XmsgStatus_OUTBOUND_MINED,
		XmsgStatus_REVERTED,
	}

	stateTransitionMap[XmsgStatus_PENDING_REVERT] = []XmsgStatus{
		XmsgStatus_ABORTED,
		XmsgStatus_OUTBOUND_MINED,
		XmsgStatus_REVERTED,
	}
	return stateTransitionMap
}
