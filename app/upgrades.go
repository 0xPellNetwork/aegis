package app

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	xsecuritytypes "github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

const releaseVersion = "v1.4"

func (app *PellApp) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		releaseVersion,
		func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)

			// Check if capability module is already initialized
			if app.CapabilityKeeper.IsInitialized(sdkCtx) {
				// If already initialized, set its version to skip InitGenesis
				fromVM[capabilitytypes.ModuleName] = capability.AppModule{}.ConsensusVersion()
			}

			// Set initial version for new modules only
			fromVM[xsecuritytypes.ModuleName] = 1

			return app.ModuleManager.RunMigrations(ctx, app.configurator, fromVM)
		},
	)

	// Configure store upgrades for new modules
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == releaseVersion && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{xsecuritytypes.StoreKey},
		}

		// Use upgrade store loader for the initial loading of all stores when app starts,
		// it checks if version == upgradeHeight and applies store upgrades before loading the stores,
		// so that new stores start with the correct version (the current height of chain),
		// instead the default which is the latest version that store last committed i.e 0 for new stores.
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
