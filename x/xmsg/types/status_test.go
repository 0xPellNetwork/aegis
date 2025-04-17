package types_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestStatus_ValidateTransition(t *testing.T) {
	tests := []struct {
		name          string
		oldStatus     types.XmsgStatus
		newStatus     types.XmsgStatus
		expectedValid bool
	}{
		{"Valid - PendingInbound to PendingOutbound", types.XmsgStatus_PENDING_INBOUND, types.XmsgStatus_PENDING_OUTBOUND, true},
		{"Valid - PendingInbound to Aborted", types.XmsgStatus_PENDING_INBOUND, types.XmsgStatus_ABORTED, true},
		{"Valid - PendingInbound to OutboundMined", types.XmsgStatus_PENDING_INBOUND, types.XmsgStatus_OUTBOUND_MINED, true},
		{"Valid - PendingInbound to PendingRevert", types.XmsgStatus_PENDING_INBOUND, types.XmsgStatus_PENDING_REVERT, true},

		{"Valid - PendingOutbound to Aborted", types.XmsgStatus_PENDING_OUTBOUND, types.XmsgStatus_ABORTED, true},
		{"Valid - PendingOutbound to PendingRevert", types.XmsgStatus_PENDING_OUTBOUND, types.XmsgStatus_PENDING_REVERT, true},
		{"Valid - PendingOutbound to OutboundMined", types.XmsgStatus_PENDING_OUTBOUND, types.XmsgStatus_OUTBOUND_MINED, true},
		{"Valid - PendingOutbound to Reverted", types.XmsgStatus_PENDING_OUTBOUND, types.XmsgStatus_REVERTED, true},

		{"Valid - PendingRevert to Aborted", types.XmsgStatus_PENDING_REVERT, types.XmsgStatus_ABORTED, true},
		{"Valid - PendingRevert to OutboundMined", types.XmsgStatus_PENDING_REVERT, types.XmsgStatus_OUTBOUND_MINED, true},
		{"Valid - PendingRevert to Reverted", types.XmsgStatus_PENDING_REVERT, types.XmsgStatus_REVERTED, true},

		{"Invalid - PendingInbound to Reverted", types.XmsgStatus_PENDING_INBOUND, types.XmsgStatus_REVERTED, false},
		{"Invalid - PendingInbound to PendingInbound", types.XmsgStatus_PENDING_INBOUND, types.XmsgStatus_PENDING_INBOUND, false},

		{"Invalid - PendingOutbound to PendingInbound", types.XmsgStatus_PENDING_OUTBOUND, types.XmsgStatus_PENDING_INBOUND, false},
		{"Invalid - PendingOutbound to PendingOutbound", types.XmsgStatus_PENDING_OUTBOUND, types.XmsgStatus_PENDING_OUTBOUND, false},

		{"Invalid - PendingRevert to PendingInbound", types.XmsgStatus_PENDING_REVERT, types.XmsgStatus_PENDING_INBOUND, false},
		{"Invalid - PendingRevert to PendingOutbound", types.XmsgStatus_PENDING_REVERT, types.XmsgStatus_PENDING_OUTBOUND, false},
		{"Invalid - PendingRevert to PendingRevert", types.XmsgStatus_PENDING_REVERT, types.XmsgStatus_PENDING_REVERT, false},

		{"Invalid old status - XmsgStatus_ABORTED", types.XmsgStatus_ABORTED, types.XmsgStatus_PENDING_REVERT, false},
		{"Invalid old status - XmsgStatus_REVERTED", types.XmsgStatus_REVERTED, types.XmsgStatus_PENDING_REVERT, false},
		{"Invalid old status - XmsgStatus_OUTBOUND_MINED", types.XmsgStatus_OUTBOUND_MINED, types.XmsgStatus_PENDING_REVERT, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := types.Status{Status: tc.oldStatus}
			valid := m.ValidateTransition(tc.newStatus)
			if valid != tc.expectedValid {
				t.Errorf("expected %v, got %v", tc.expectedValid, valid)
			}
		})
	}
}

func TestStatus_ChangeStatus(t *testing.T) {
	t.Run("should change status and msg if transition is valid", func(t *testing.T) {
		s := types.Status{Status: types.XmsgStatus_PENDING_INBOUND}

		s.ChangeStatus(types.XmsgStatus_PENDING_OUTBOUND, "msg")
		assert.Equal(t, s.Status, types.XmsgStatus_PENDING_OUTBOUND)
		assert.Equal(t, s.StatusMessage, "msg")
	})

	t.Run("should change status if transition is valid", func(t *testing.T) {
		s := types.Status{Status: types.XmsgStatus_PENDING_INBOUND}

		s.ChangeStatus(types.XmsgStatus_PENDING_OUTBOUND, "")
		assert.Equal(t, s.Status, types.XmsgStatus_PENDING_OUTBOUND)
		assert.Equal(t, s.StatusMessage, "")
	})

	t.Run("should change status to aborted and msg if transition is invalid", func(t *testing.T) {
		s := types.Status{Status: types.XmsgStatus_PENDING_OUTBOUND}

		s.ChangeStatus(types.XmsgStatus_PENDING_INBOUND, "msg")
		assert.Equal(t, s.Status, types.XmsgStatus_ABORTED)
		assert.Equal(t, fmt.Sprintf("Failed to transition : OldStatus %s , NewStatus %s , MSG : %s :", types.XmsgStatus_PENDING_OUTBOUND.String(), types.XmsgStatus_PENDING_INBOUND.String(), "msg"), s.StatusMessage)
	})
}
