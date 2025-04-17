package orchestrator

import (
	"context"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	ethrpc2 "github.com/onrik/ethrpc"
	"github.com/pkg/errors"

	evmobserver "github.com/pell-chain/pellcore/relayer/chains/evm/observer"
	evmsigner "github.com/pell-chain/pellcore/relayer/chains/evm/signer"
	"github.com/pell-chain/pellcore/relayer/chains/interfaces"
	"github.com/pell-chain/pellcore/relayer/config"
	pctx "github.com/pell-chain/pellcore/relayer/context"
	"github.com/pell-chain/pellcore/relayer/db"
	clientlogs "github.com/pell-chain/pellcore/relayer/logs"
	"github.com/pell-chain/pellcore/relayer/metrics"
)

// CreateSignerMap creates a map of interfaces.ChainSigner (by chainID) for all chains in the config.
// Note that signer construction failure for a chain does not prevent the creation of signers for other chains.
func CreateSignerMap(
	ctx context.Context,
	tss interfaces.TSSSigner,
	logger clientlogs.Logger,
	ts *metrics.TelemetryServer,
) (map[int64]interfaces.ChainSigner, error) {
	signers := make(map[int64]interfaces.ChainSigner)
	_, _, err := syncSignerMap(ctx, tss, logger, ts, &signers)
	if err != nil {
		return nil, err
	}

	return signers, nil
}

// syncSignerMap synchronizes the given signers map with the signers for all chains in the config.
// This semantic is used to allow dynamic updates to the signers map.
// Note that data race handling is the responsibility of the caller.
func syncSignerMap(
	ctx context.Context,
	tss interfaces.TSSSigner,
	logger clientlogs.Logger,
	ts *metrics.TelemetryServer,
	signers *map[int64]interfaces.ChainSigner,
) (int, int, error) {
	appContext, err := pctx.FromContext(ctx)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "failed to get app context")
	}

	var (
		added, removed int

		presentChainIDs = make([]int64, 0)

		onAfterAdd = func(chainID int64, _ interfaces.ChainSigner) {
			logger.Std.Info().Msgf("Added signer for chain %d", chainID)
			added++
		}

		addSigner = func(chainID int64, signer interfaces.ChainSigner) {
			mapSet[int64, interfaces.ChainSigner](signers, chainID, signer, onAfterAdd)
		}

		onBeforeRemove = func(chainID int64, _ interfaces.ChainSigner) {
			logger.Std.Info().Msgf("Removing signer for chain %d", chainID)
			removed++
		}
	)

	// EVM signers
	for _, evmConfig := range appContext.Config().GetAllEVMConfigs() {
		if evmConfig.Chain.IsPellChain() {
			continue
		}

		chainID := evmConfig.Chain.Id
		presentChainIDs = append(presentChainIDs, chainID)

		// noop for existing signers
		if mapHas(signers, chainID) {
			continue
		}

		evmChainParams, found := appContext.PellCoreContext().GetEVMChainParams(chainID)
		if !found {
			logger.Std.Warn().Msgf("Unable to find EVM config for chain %d", chainID)
			continue
		}

		pellConnectorAddress := ethcommon.HexToAddress(evmChainParams.ConnectorContractAddress)
		signer, err := evmsigner.NewEVMSigner(
			ctx,
			evmConfig.Chain,
			evmConfig.Endpoint,
			tss,
			pellConnectorAddress,
			logger,
			ts)
		if err != nil {
			logger.Std.Error().Err(err).Msgf("Unable to construct signer for EVM chain %d", chainID)
			continue
		}
		addSigner(chainID, signer)
	}

	// Remove all disabled signers
	mapDeleteMissingKeys(signers, presentChainIDs, onBeforeRemove)

	return added, removed, nil
}

// CreateChainObserverMap creates a map of interfaces.ChainObserver (by chainID) for all chains in the config.
// - Note (!) that it calls observer.Start() on creation
// - Note that data race handling is the responsibility of the caller.
func CreateChainObserverMap(
	ctx context.Context,
	pellcoreClient interfaces.PellCoreBridger,
	tss interfaces.TSSSigner,
	dbpath string,
	logger clientlogs.Logger,
	ts *metrics.TelemetryServer,
) (map[int64]interfaces.ChainClient, error) {
	observerMap := make(map[int64]interfaces.ChainClient)

	_, _, err := syncObserverMap(ctx, pellcoreClient, tss, dbpath, logger, ts, &observerMap)
	if err != nil {
		return nil, err
	}

	return observerMap, nil
}

// syncObserverMap synchronizes the given observer map with the observers for all chains in the config.
// This semantic is used to allow dynamic updates to the map.
// Note (!) that it calls observer.Start() on creation and observer.Stop() on deletion.
func syncObserverMap(
	ctx context.Context,
	pellcoreClient interfaces.PellCoreBridger,
	tss interfaces.TSSSigner,
	dbpath string,
	logger clientlogs.Logger,
	ts *metrics.TelemetryServer,
	observers *map[int64]interfaces.ChainClient,
) (int, int, error) {
	appContext, err := pctx.FromContext(ctx)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "failed to get app context")
	}

	var (
		added, removed int

		presentChainIDs = make([]int64, 0)

		onAfterAdd = func(_ int64, ob interfaces.ChainClient) {
			ob.Start(ctx)
			added++
		}

		addObserver = func(chainID int64, ob interfaces.ChainClient) {
			mapSet[int64, interfaces.ChainClient](observers, chainID, ob, onAfterAdd)
		}

		onBeforeRemove = func(_ int64, ob interfaces.ChainClient) {
			ob.Stop()
			removed++
		}
	)

	// EVM clients
	for _, evmConfig := range appContext.Config().GetAllEVMConfigs() {
		if evmConfig.Chain.IsPellChain() {
			continue
		}

		if evmConfig.MaxLatestIndexedBlockGap == 0 {
			evmConfig.MaxLatestIndexedBlockGap = config.DefaultMaxLatestIndexedBlockGap
		}

		chainID := evmConfig.Chain.Id
		presentChainIDs = append(presentChainIDs, chainID)

		// noop
		if mapHas(observers, chainID) {
			continue
		}

		chainParams, found := appContext.PellCoreContext().GetEVMChainParams(chainID)
		if !found {
			logger.Std.Warn().Msgf("Unable to find EVM config for chain %d", chainID)
			continue
		}

		httpClient, err := metrics.GetInstrumentedHTTPClient(evmConfig.Endpoint)
		if err != nil {
			logger.Std.Error().Err(err).Str("rpc.endpoint", evmConfig.Endpoint).Msgf("Unable to create HTTP client")
			continue
		}

		rpcClient, err := ethrpc.DialHTTPWithClient(evmConfig.Endpoint, httpClient)
		if err != nil {
			logger.Std.Error().Err(err).Str("rpc.endpoint", evmConfig.Endpoint).Msgf("Unable to dial EVM RPC")
			continue
		}
		evmClient := ethclient.NewClient(rpcClient)

		evmJSONRPCClient := ethrpc2.NewEthRPC(evmConfig.Endpoint, ethrpc2.WithHttpClient(httpClient))

		database, err := db.NewFromSqlite(dbpath, evmConfig.Chain.ChainName(), true)
		if err != nil {
			logger.Std.Error().Err(err).Msgf("Unable to open a database for EVM chain %q", evmConfig.Chain.ChainName())
			continue
		}

		observer, err := evmobserver.NewEVMChainClient(
			ctx,
			evmConfig,
			*chainParams,
			evmClient,
			evmJSONRPCClient,
			pellcoreClient,
			tss,
			database,
			logger,
			ts,
		)
		if err != nil {
			logger.Std.Error().Err(err).Msgf("NewEVMChainClient error for chain %s", evmConfig.Chain.String())
			continue
		}
		addObserver(chainID, observer)
	}

	// Remove all disabled observers
	mapDeleteMissingKeys(observers, presentChainIDs, onBeforeRemove)

	return added, removed, nil
}
