package app

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	xsecuritytypes "github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

const releaseVersion = "v1.4"

func (app *PellApp) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		releaseVersion,
		func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			// Debug: Log original fromVM
			app.Logger().Info("=== Upgrade Handler Debug ===")
			app.Logger().Info("Original fromVM", "count", len(fromVM))
			for name, version := range fromVM {
				app.Logger().Info("Original module version", "module", name, "version", version)
			}

			// Get current version map of all modules
			currentVersionMap := app.ModuleManager.GetVersionMap()
			app.Logger().Info("Current version map", "count", len(currentVersionMap))

			// Set all existing modules to their current consensus versions
			for moduleName, version := range currentVersionMap {
				fromVM[moduleName] = version
				app.Logger().Info("Set existing module version", "module", moduleName, "version", version)
			}

			// Only truly new modules need initialization
			fromVM[xsecuritytypes.ModuleName] = 1
			app.Logger().Info("Set new module version", "module", xsecuritytypes.ModuleName, "version", 1)

			app.Logger().Info("Final fromVM", "count", len(fromVM))

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
