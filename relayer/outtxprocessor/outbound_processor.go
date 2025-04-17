package outtxprocessor

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Processor is a struct that contains data about outbound being processed
// TODO(revamp): rename this struct as it is not used to process outbound but track their processing
// We can also consider removing it once we refactor chain client to contains common logic to sign outbounds
type Processor struct {
	outTxStartTime     map[string]time.Time
	outTxEndTime       map[string]time.Time
	outTxActive        map[string]struct{}
	mu                 sync.Mutex
	Logger             zerolog.Logger
	numActiveProcessor int64
}

// NewProcessor creates a new Processor
func NewOutTxProcessor(logger zerolog.Logger) *Processor {
	return &Processor{
		outTxStartTime:     make(map[string]time.Time),
		outTxEndTime:       make(map[string]time.Time),
		outTxActive:        make(map[string]struct{}),
		mu:                 sync.Mutex{},
		Logger:             logger.With().Str("module", "OutTxProcessor").Logger(),
		numActiveProcessor: 0,
	}
}

func (outTxMan *Processor) StartTryProcess(outTxID string) {
	outTxMan.mu.Lock()
	defer outTxMan.mu.Unlock()
	outTxMan.outTxStartTime[outTxID] = time.Now()
	outTxMan.outTxActive[outTxID] = struct{}{}
	outTxMan.numActiveProcessor++
	outTxMan.Logger.Info().
		Str("outTxID", outTxID).
		Int64("numActiveProcessor", outTxMan.numActiveProcessor).
		Msg("StartTryProcess")
}

func (outTxMan *Processor) EndTryProcess(outTxID string) {
	outTxMan.mu.Lock()
	defer outTxMan.mu.Unlock()
	outTxMan.outTxEndTime[outTxID] = time.Now()
	delete(outTxMan.outTxActive, outTxID)
	outTxMan.numActiveProcessor--
	outTxMan.Logger.Info().
		Str("outTxID", outTxID).
		Int64("numActiveProcessor", outTxMan.numActiveProcessor).
		Dur("timeElapsed", time.Since(outTxMan.outTxStartTime[outTxID])).
		Msg("EndTryProcess")
}

func (outTxMan *Processor) IsOutTxActive(outTxID string) bool {
	outTxMan.mu.Lock()
	defer outTxMan.mu.Unlock()
	_, found := outTxMan.outTxActive[outTxID]
	return found
}

func (outTxMan *Processor) TimeInTryProcess(outTxID string) time.Duration {
	outTxMan.mu.Lock()
	defer outTxMan.mu.Unlock()
	if _, found := outTxMan.outTxActive[outTxID]; found {
		return time.Since(outTxMan.outTxStartTime[outTxID])
	}
	return 0
}

// ToOutTxID returns the outTxID for OutTxProcessorManager to track
func ToOutTxID(index string, receiverChainID int64, nonce uint64) string {
	return fmt.Sprintf("%s-%d-%d", index, receiverChainID, nonce)
}
