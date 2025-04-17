package context

import (
	goctx "context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/relayer/config"
	lightclienttypes "github.com/pell-chain/pellcore/x/lightclient/types"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
)

type appContextKey struct{}

// AppContext contains global app structs like config, core context and logger
type AppContext struct {
	coreContext *PellCoreContext

	// config is the config of the app
	config config.Config

	// logger is the logger of the app
	logger zerolog.Logger
}

// NewAppContext creates and returns new AppContext
func NewAppContext(
	config config.Config,
	logger zerolog.Logger,
) *AppContext {
	return &AppContext{
		coreContext: NewPellCoreContext(config),
		config:      config,
		logger:      logger.With().Str("module", "appcontext").Logger(),
	}
}

func (a *AppContext) Config() config.Config {
	return a.config
}

func (a *AppContext) PellCoreContext() *PellCoreContext {
	return a.coreContext
}

func (a *AppContext) GetEnabledChains() []chains.Chain {
	return a.coreContext.GetEnabledChains()
}

func (a *AppContext) GetKeygen() observertypes.Keygen {
	return a.coreContext.GetKeygen()
}

func (a *AppContext) Update(
	keygen observertypes.Keygen,
	newChains []chains.Chain,
	evmChainParams map[int64]*observertypes.ChainParams,
	tssPubKey string,
	crosschainFlags observertypes.CrosschainFlags,
	verificationFlags lightclienttypes.VerificationFlags,
	init bool,
	logger zerolog.Logger,
) error {
	return a.coreContext.Update(keygen, newChains, evmChainParams, tssPubKey, crosschainFlags, verificationFlags, init, logger)
}

// WithAppContext applied AppContext to standard Go context.Context.
func WithAppContext(ctx goctx.Context, app *AppContext) goctx.Context {
	return goctx.WithValue(ctx, appContextKey{}, app)
}

// FromContext extracts AppContext from context.Context
func FromContext(ctx goctx.Context) (*AppContext, error) {
	app, ok := ctx.Value(appContextKey{}).(*AppContext)
	if !ok || app == nil {
		return nil, errors.New("AppContext is not set in the context.Context")
	}

	return app, nil
}

// Copy copies AppContext from one context to another (is present).
// This is useful when you want to drop timeouts and deadlines from the context
// (e.g. run something in another goroutine).
func Copy(from, to goctx.Context) goctx.Context {
	app, err := FromContext(from)
	if err != nil {
		return to
	}

	return WithAppContext(to, app)
}
