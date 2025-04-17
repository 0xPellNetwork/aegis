package types

func (p *InboundPellEvent) isAvailable() bool {
	rc := true

	switch p.PellData.(type) {
	case *InboundPellEvent_StakerDelegated:
	case *InboundPellEvent_StakerDeposited:
	case *InboundPellEvent_WithdrawalQueued:
	case *InboundPellEvent_StakerUndelegated:
	case *InboundPellEvent_PellSent:
	case *InboundPellEvent_RegisterChainDvsToPell:
	default:
		rc = false
	}

	return rc
}
