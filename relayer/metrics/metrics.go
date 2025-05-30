// Package metrics provides metrics functionalities for the pellclient
package metrics

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

// Metrics is a struct that contains the http server for metrics
type Metrics struct {
	s *http.Server
}

// PellClientNamespace is the namespace for the metrics
const PellClientNamespace = "pellclient"

var (
	// PendingTxsPerChain is a gauge that contains the number of pending transactions per chain
	PendingTxsPerChain = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "pending_txs_total",
		Help:      "Number of pending transactions per chain",
	}, []string{"chain"})

	// GetFilterLogsPerChain is a counter that contains the number of getLogs per chain
	GetFilterLogsPerChain = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: PellClientNamespace,
		Name:      "rpc_getFilterLogs_count",
		Help:      "Count of getLogs per chain",
	}, []string{"chain"})

	// GetBlockByNumberPerChain is a counter that contains the number of getBlockByNumber per chain
	GetBlockByNumberPerChain = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: PellClientNamespace,
		Name:      "rpc_getBlockByNumber_count",
		Help:      "Count of getLogs per chain",
	}, []string{"chain"})

	// TssNodeBlamePerPubKey is a counter that contains the number of tss node blame per pubkey
	TssNodeBlamePerPubKey = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: PellClientNamespace,
		Name:      "tss_node_blame_count",
		Help:      "Tss node blame counter per pubkey",
	}, []string{"pubkey"})

	// RelayerKeyBalance is a gauge that contains the relayer key balance of the chain
	RelayerKeyBalance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "relayer_key_balance",
		Help:      "Relayer key balance of the chain",
	}, []string{"chain"})

	// HotKeyBurnRate is a gauge that contains the fee burn rate of the hotkey
	HotKeyBurnRate = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "hotkey_burn_rate",
		Help:      "Fee burn rate of the hotkey",
	})

	// NumberOfUTXO is a gauge that contains the number of UTXOs
	NumberOfUTXO = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "utxo_number",
		Help:      "Number of UTXOs",
	})

	// LastScannedBlockNumber is a gauge that contains the last scanned block number per chain
	LastScannedBlockNumber = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "last_scanned_block_number",
		Help:      "Last scanned block number per chain",
	}, []string{"chain"})

	// LastScannedBlockNumber is a gauge that contains the last scanned block number per chain
	LastInboundBlockNumber = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "last_inbound_block_number",
		Help:      "Last inbound block number contract xmsg",
	}, []string{"chain"})

	// LastVoteInboundBlockNumber is a gauge that contains the last vote inbound block number per chain
	LastVoteInboundBlockNumber = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "last_vote_inbound_block_number",
		Help:      "Last vote inbound block number contract xmsg",
	}, []string{"chain"})

	// LastCoreBlockNumber is a gauge that contains the last core block number
	LastCoreBlockNumber = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "last_core_block_number",
		Help:      "Last core block number",
	})

	// Info is a gauge that contains information about the pellclient environment
	Info = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "info",
		Help:      "Information about Pellclient environment",
	}, []string{"version"})

	// LastStartTime is a gauge that contains the start time in Unix time
	LastStartTime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "last_start_timestamp_seconds",
		Help:      "Start time in Unix time",
	})

	// NumActiveMsgSigns is a gauge that contains the number of concurrent key signs
	NumActiveMsgSigns = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "num_active_message_signs",
		Help:      "Number of concurrent key signs",
	})

	// PercentageOfRateReached is a gauge that contains the percentage of the rate limiter rate reached
	PercentageOfRateReached = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "percentage_of_rate_reached",
		Help:      "Percentage of the rate limiter rate reached",
	})

	// SignLatency is a histogram of of the TSS keysign latency
	SignLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: PellClientNamespace,
		Name:      "sign_latency",
		Help:      "Histogram of the TSS keysign latency",
		Buckets:   []float64{1, 7, 15, 30, 60, 120, 240},
	}, []string{"result"})

	// RPCInProgress is a gauge that contains the number of RPCs requests in progress
	RPCInProgress = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "rpc_in_progress",
		Help:      "Number of RPC requests in progress",
	}, []string{"host"})

	// RPCCount is a counter that contains the number of total RPC requests
	RPCCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: PellClientNamespace,
			Name:      "rpc_count",
			Help:      "A counter for number of total RPC requests",
		},
		[]string{"host", "code"},
	)

	// RPCLatency is a histogram of the RPC latency
	RPCLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: PellClientNamespace,
			Name:      "rpc_duration_seconds",
			Help:      "A histogram of the RPC duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"host"},
	)

	// RPCNodeStatus is a gauge that indicates the status of RPC nodes
	RPCNodeStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "rpc_node_status",
		Help:      "Status of RPC nodes (1 = online, 0 = offline)",
	}, []string{"chain"})

	// TssAddressGasBalance is a gauge that tracks the balance of TSS addresses across different chains
	TssAddressGasBalance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "tss_address_balance",
		Help:      "Balance of TSS addresses per chain and address",
	}, []string{"chain", "address"})

	// TssAddressPellTokenBalance is a gauge that tracks the balance of TSS addresses across different chains
	TssAddressPellTokenBalance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: PellClientNamespace,
		Name:      "tss_address_pell_token_balance",
		Help:      "Balance of TSS addresses per chain and address",
	}, []string{"chain", "address"})
)

// NewMetrics creates a new Metrics instance
func NewMetrics() (*Metrics, error) {
	handler := promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer,
		promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
			Timeout: 30 * time.Second,
		}),
	)

	s := &http.Server{
		Addr:              ":8886",
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Metrics{
		s,
	}, nil
}

// Start starts the metrics server
func (m *Metrics) Start() {
	log.Info().Msg("metrics server starting")
	go func() {
		if err := m.s.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("fail to start metric server")
		}
	}()
}

// Stop stops the metrics server
func (m *Metrics) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	return m.s.Shutdown(ctx)
}

// GetInstrumentedHTTPClient sets up a http client that emits prometheus metrics
func GetInstrumentedHTTPClient(endpoint string) (*http.Client, error) {
	host := endpoint
	// try to parse as url (so that we do not expose auth uuid in metrics)
	endpointURL, err := url.Parse(endpoint)
	if err == nil {
		host = endpointURL.Host
	}
	labels := prometheus.Labels{"host": host}
	rpcCounterMetric, err := RPCCount.CurryWith(labels)
	if err != nil {
		return nil, err
	}
	rpcLatencyMetric, err := RPCLatency.CurryWith(labels)
	if err != nil {
		return nil, err
	}

	transport := http.DefaultTransport
	transport = promhttp.InstrumentRoundTripperDuration(rpcLatencyMetric, transport)
	transport = promhttp.InstrumentRoundTripperCounter(rpcCounterMetric, transport)
	transport = promhttp.InstrumentRoundTripperInFlight(RPCInProgress.With(labels), transport)

	return &http.Client{
		Transport: transport,
	}, nil
}
