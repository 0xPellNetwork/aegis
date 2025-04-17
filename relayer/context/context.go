package context

import (
	"sort"
	"sync"

	"github.com/rs/zerolog"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/relayer/config"
	lightclienttypes "github.com/pell-chain/pellcore/x/lightclient/types"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
)

// PellCoreContext contains core context params
// these are initialized and updated at runtime at every height
type PellCoreContext struct {
	coreContextLock    *sync.RWMutex
	keygen             observertypes.Keygen
	chainsEnabled      []chains.Chain
	evmChainParams     map[int64]*observertypes.ChainParams
	bitcoinChainParams *observertypes.ChainParams
	currentTssPubkey   string
	crossChainFlags    observertypes.CrosschainFlags

	// verificationFlags is used to store the verification flags for the lightclient module to enable header/proof verification
	verificationFlags lightclienttypes.VerificationFlags
}

// NewPellCoreContext creates and returns new PellCoreContext
// it is initializing chain params from provided config
func NewPellCoreContext(cfg config.Config) *PellCoreContext {
	evmChainParams := make(map[int64]*observertypes.ChainParams)
	for _, e := range cfg.EVMChainConfigs {
		evmChainParams[e.Chain.Id] = &observertypes.ChainParams{}
	}
	var bitcoinChainParams *observertypes.ChainParams
	_, found := cfg.GetBTCConfig()
	if found {
		bitcoinChainParams = &observertypes.ChainParams{}
	}
	return &PellCoreContext{
		coreContextLock:    new(sync.RWMutex),
		chainsEnabled:      []chains.Chain{},
		evmChainParams:     evmChainParams,
		bitcoinChainParams: bitcoinChainParams,
		crossChainFlags:    observertypes.CrosschainFlags{},
		verificationFlags:  lightclienttypes.VerificationFlags{},
	}
}

func (c *PellCoreContext) GetKeygen() observertypes.Keygen {
	c.coreContextLock.RLock()
	defer c.coreContextLock.RUnlock()

	var copiedPubkeys []string
	if c.keygen.GranteePubkeys != nil {
		copiedPubkeys = make([]string, len(c.keygen.GranteePubkeys))
		copy(copiedPubkeys, c.keygen.GranteePubkeys)
	}
	return observertypes.Keygen{
		Status:         c.keygen.Status,
		GranteePubkeys: copiedPubkeys,
		BlockNumber:    c.keygen.BlockNumber,
	}
}

func (c *PellCoreContext) GetCurrentTssPubkey() string {
	c.coreContextLock.RLock()
	defer c.coreContextLock.RUnlock()
	return c.currentTssPubkey
}

func (c *PellCoreContext) GetEnabledChains() []chains.Chain {
	c.coreContextLock.RLock()
	defer c.coreContextLock.RUnlock()
	copiedChains := make([]chains.Chain, len(c.chainsEnabled))
	copy(copiedChains, c.chainsEnabled)
	return copiedChains
}

func (c *PellCoreContext) GetEVMChainParams(chainID int64) (*observertypes.ChainParams, bool) {
	c.coreContextLock.RLock()
	defer c.coreContextLock.RUnlock()
	evmChainParams, found := c.evmChainParams[chainID]
	return evmChainParams, found
}

func (c *PellCoreContext) GetAllEVMChainParams() map[int64]*observertypes.ChainParams {
	c.coreContextLock.RLock()
	defer c.coreContextLock.RUnlock()

	// deep copy evm chain params
	copied := make(map[int64]*observertypes.ChainParams, len(c.evmChainParams))
	for chainID, evmConfig := range c.evmChainParams {
		copied[chainID] = &observertypes.ChainParams{}
		*copied[chainID] = *evmConfig
	}
	return copied
}

func (c *PellCoreContext) GetCrossChainFlags() observertypes.CrosschainFlags {
	c.coreContextLock.RLock()
	defer c.coreContextLock.RUnlock()
	return c.crossChainFlags
}

func (c *PellCoreContext) GetVerificationFlags() lightclienttypes.VerificationFlags {
	c.coreContextLock.RLock()
	defer c.coreContextLock.RUnlock()
	return c.verificationFlags
}

// IsOutboundObservationEnabled returns true if the chain is supported and outbound flag is enabled
func (c *PellCoreContext) IsOutboundObservationEnabled(chainParams observertypes.ChainParams) bool {
	flags := c.GetCrossChainFlags()
	return chainParams.IsSupported && flags.IsOutboundEnabled
}

// IsInboundObservationEnabled returns true if the chain is supported and inbound flag is enabled
func (c *PellCoreContext) IsInboundObservationEnabled(chainParams observertypes.ChainParams) bool {
	flags := c.GetCrossChainFlags()
	return chainParams.IsSupported && flags.IsInboundEnabled
}

// Update updates core context and params for all chains
// this must be the ONLY function that writes to core context
func (c *PellCoreContext) Update(
	keygen observertypes.Keygen,
	newChains []chains.Chain,
	evmChainParams map[int64]*observertypes.ChainParams,
	tssPubKey string,
	crosschainFlags observertypes.CrosschainFlags,
	verificationFlags lightclienttypes.VerificationFlags,
	init bool,
	logger zerolog.Logger,
) error {
	// Ignore whatever order pellbridge organizes chain list in state
	sort.SliceStable(newChains, func(i, j int) bool {
		return newChains[i].Id < newChains[j].Id
	})
	if len(newChains) == 0 {
		logger.Warn().Msg("UpdateChainParams: No chains enabled in PellCore")
	}

	// Add some warnings if chain list changes at runtime
	if !init {
		if len(c.chainsEnabled) != len(newChains) {
			logger.Warn().Msgf(
				"UpdateChainParams: ChainsEnabled changed at runtime!! current: %v, new: %v",
				c.chainsEnabled,
				newChains,
			)
		} else {
			for i, chain := range newChains {
				if chain != c.chainsEnabled[i] {
					logger.Warn().Msgf(
						"UpdateChainParams: ChainsEnabled changed at runtime!! current: %v, new: %v",
						c.chainsEnabled,
						newChains,
					)
				}
			}
		}
	}

	c.coreContextLock.Lock()
	defer c.coreContextLock.Unlock()

	c.keygen = keygen
	c.chainsEnabled = newChains
	c.crossChainFlags = crosschainFlags
	c.verificationFlags = verificationFlags

	// update core params for evm chains we have configs in file
	for _, params := range evmChainParams {
		_, found := c.evmChainParams[params.ChainId]
		if !found {
			continue
		}
		c.evmChainParams[params.ChainId] = params
	}

	if tssPubKey != "" {
		c.currentTssPubkey = tssPubKey
	}

	return nil
}
