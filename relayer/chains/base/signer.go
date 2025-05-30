package base

import (
	"sync"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	corecontext "github.com/0xPellNetwork/aegis/relayer/context"
	clientlogs "github.com/0xPellNetwork/aegis/relayer/logs"
	"github.com/0xPellNetwork/aegis/relayer/metrics"
)

// Signer is the base structure for grouping the common logic between chain signers.
// The common logic includes: chain, chainParams, contexts, tss, metrics, loggers etc.
type Signer struct {
	// chain contains static information about the external chain
	chain *chains.Chain

	// tss is the TSS signer
	tssSigner interfaces.TSSSigner

	// ts is the telemetry server for metrics
	ts *metrics.TelemetryServer

	// logger contains the loggers used by signer
	logger clientlogs.Logger

	// outboundBeingReported is a map of outbound being reported to tracker
	outboundBeingReported map[string]bool

	coreContext *corecontext.PellCoreContext

	// mu protects fields from concurrent access
	// Note: base signer simply provides the mutex. It's the sub-struct's responsibility to use it to be thread-safe
	mu *sync.Mutex
}

// NewSigner creates a new base signer.
func NewSigner(
	chain chains.Chain,
	tssSigner interfaces.TSSSigner,
	coreContext *corecontext.PellCoreContext,
	ts *metrics.TelemetryServer,
	logger clientlogs.Logger,
) *Signer {
	return &Signer{
		chain:       &chain,
		tssSigner:   tssSigner,
		coreContext: coreContext,
		ts:          ts,
		logger: clientlogs.Logger{
			Std: logger.Std.With().
				Int64(clientlogs.FieldChain, chain.Id).
				Str(clientlogs.FieldModule, "signer").
				Logger(),
			Compliance: logger.Compliance,
		},
		mu:                    &sync.Mutex{},
		outboundBeingReported: make(map[string]bool),
	}
}

// Chain returns the chain for the signer.
func (s *Signer) Chain() *chains.Chain {
	return s.chain
}

// WithChain attaches a new chain to the signer.
func (s *Signer) WithChain(chain chains.Chain) *Signer {
	s.chain = &chain
	return s
}

// Tss returns the tss signer for the signer.
func (s *Signer) TSS() interfaces.TSSSigner {
	return s.tssSigner
}

// WithTSS attaches a new tss signer to the signer.
func (s *Signer) WithTSS(tssSigner interfaces.TSSSigner) *Signer {
	s.tssSigner = tssSigner
	return s
}

// TelemetryServer returns the telemetry server for the signer.
func (s *Signer) TelemetryServer() *metrics.TelemetryServer {
	return s.ts
}

// WithTelemetryServer attaches a new telemetry server to the signer.
func (s *Signer) WithTelemetryServer(ts *metrics.TelemetryServer) *Signer {
	s.ts = ts
	return s
}

// Logger returns the logger for the signer.
func (s *Signer) Logger() *clientlogs.Logger {
	return &s.logger
}

// CoreContext returns the coreclient
func (s *Signer) CoreContext() *corecontext.PellCoreContext {
	return s.coreContext
}

// SetBeingReportedFlag sets the outbound as being reported if not already set.
// Returns true if the outbound is already being reported.
// This method is used by outbound tracker reporter to avoid repeated reporting of same hash.
func (s *Signer) SetBeingReportedFlag(hash string) (alreadySet bool) {
	s.Lock()
	defer s.Unlock()

	alreadySet = s.outboundBeingReported[hash]
	if !alreadySet {
		// mark as being reported
		s.outboundBeingReported[hash] = true
	}
	return
}

// ClearBeingReportedFlag clears the being reported flag for the outbound.
func (s *Signer) ClearBeingReportedFlag(hash string) {
	s.Lock()
	defer s.Unlock()
	delete(s.outboundBeingReported, hash)
}

// Exported for unit tests

// GetReportedTxList returns a list of outboundHash being reported.
// TODO: investigate pointer usage
func (s *Signer) GetReportedTxList() *map[string]bool {
	return &s.outboundBeingReported
}

func (s *Signer) GetBeingReportedFlag(hash string) bool {
	return s.outboundBeingReported[hash]
}

// Lock locks the signer.
func (s *Signer) Lock() {
	s.mu.Lock()
}

// Unlock unlocks the signer.
func (s *Signer) Unlock() {
	s.mu.Unlock()
}
