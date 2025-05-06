package context

import (
	"sort"
	"sync"

	"github.com/rs/zerolog"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/relayer/config"
	lightclienttypes "github.com/0xPellNetwork/aegis/x/lightclient/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

// PellCoreContext contains core context params
// these are initialized and updated at runtime at every height
type PellCoreContext struct {
	coreContextLock    *sync.RWMutex
	keygen             relayertypes.Keygen
	chainsEnabled      []chains.Chain
	evmChainParams     map[int64]*relayertypes.ChainParams
	bitcoinChainParams *relayertypes.ChainParams
	currentTssPubkey   string
	crossChainFlags    relayertypes.CrosschainFlags

	// verificationFlags is used to store the verification flags for the lightclient module to enable header/proof verification
	verificationFlags lightclienttypes.VerificationFlags
}

// NewPellCoreContext creates and returns new PellCoreContext
// it is initializing chain params from provided config
func NewPellCoreContext(cfg config.Config) *PellCoreContext {
	evmChainParams := make(map[int64]*relayertypes.ChainParams)
	for _, e := range cfg.EVMChainConfigs {
		evmChainParams[e.Chain.Id] = &relayertypes.ChainParams{}
	}
	var bitcoinChainParams *relayertypes.ChainParams
	_, found := cfg.GetBTCConfig()
	if found {
		bitcoinChainParams = &relayertypes.ChainParams{}
	}
	return &PellCoreContext{
		coreContextLock:    new(sync.RWMutex),
		chainsEnabled:      []chains.Chain{},
		evmChainParams:     evmChainParams,
		bitcoinChainParams: bitcoinChainParams,
		crossChainFlags:    relayertypes.CrosschainFlags{},
		verificationFlags:  lightclienttypes.VerificationFlags{},
	}
}

func (c *PellCoreContext) GetKeygen() relayertypes.Keygen {
	c.coreContextLock.RLock()
	defer c.coreContextLock.RUnlock()

	var copiedPubkeys []string
	if c.keygen.GranteePubkeys != nil {
		copiedPubkeys = make([]string, len(c.keygen.GranteePubkeys))
		copy(copiedPubkeys, c.keygen.GranteePubkeys)
	}
	return relayertypes.Keygen{
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

func (c *PellCoreContext) GetEVMChainParams(chainID int64) (*relayertypes.ChainParams, bool) {
	c.coreContextLock.RLock()
	defer c.coreContextLock.RUnlock()
	evmChainParams, found := c.evmChainParams[chainID]
	return evmChainParams, found
}

func (c *PellCoreContext) GetAllEVMChainParams() map[int64]*relayertypes.ChainParams {
	c.coreContextLock.RLock()
	defer c.coreContextLock.RUnlock()

	// deep copy evm chain params
	copied := make(map[int64]*relayertypes.ChainParams, len(c.evmChainParams))
	for chainID, evmConfig := range c.evmChainParams {
		copied[chainID] = &relayertypes.ChainParams{}
		*copied[chainID] = *evmConfig
	}
	return copied
}

func (c *PellCoreContext) GetCrossChainFlags() relayertypes.CrosschainFlags {
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
func (c *PellCoreContext) IsOutboundObservationEnabled(chainParams relayertypes.ChainParams) bool {
	return c.isObservationEnabled(chainParams, func(f relayertypes.CrosschainFlags) bool {
		return f.IsOutboundEnabled
	})
}

// IsInboundObservationEnabled returns true if the chain is supported and inbound flag is enabled
func (c *PellCoreContext) IsInboundObservationEnabled(chainParams relayertypes.ChainParams) bool {
	return c.isObservationEnabled(chainParams, func(f relayertypes.CrosschainFlags) bool {
		return f.IsInboundEnabled
	})
}

// Update updates core context and params for all chains
// this must be the ONLY function that writes to core context
func (c *PellCoreContext) Update(
	keygen relayertypes.Keygen,
	newChains []chains.Chain,
	evmChainParams map[int64]*relayertypes.ChainParams,
	tssPubKey string,
	crosschainFlags relayertypes.CrosschainFlags,
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

// 私有辅助函数，减少代码重复
func (c *PellCoreContext) isObservationEnabled(
	chainParams relayertypes.ChainParams,
	flagCheck func(relayertypes.CrosschainFlags) bool,
) bool {
	if !chainParams.IsSupported {
		return false
	}
	flags := c.GetCrossChainFlags()
	return flagCheck(flags)
}
