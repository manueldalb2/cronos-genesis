package app

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"

	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"

	e2eetypes "github.com/crypto-org-chain/cronos/v2/x/e2ee/types"

	// OLD (Genesis v1.0.0) store keys you must DELETE at upgrade height
	//authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	//capabilitytypes "github.com/cosmos/ibc-go/v5/modules/capability/types"
	//ibchost "github.com/cosmos/ibc-go/v5/modules/core/24-host"
	//ibcfeetypes "github.com/cosmos/ibc-go/v5/modules/apps/29-fee/types"
)

// Must match the on-chain upgrade plan name exactly.
const MegaPlanName = "genesis-v1.6.2"

// RegisterUpgradeHandlers returns if store loader is overridden.
func (app *App) RegisterUpgradeHandlers(cdc codec.BinaryCodec, maxVersion int64) bool {
	app.UpgradeKeeper.SetUpgradeHandler(MegaPlanName, func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Just run module migrations; chain-specific patches (e.g. Cronos contract storage edits) are intentionally omitted.
		return app.ModuleManager.RunMigrations(ctx, app.configurator, fromVM)
	})

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk: %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return false
	}
	if upgradeInfo.Name != MegaPlanName {
		return false
	}

	storeUpgrades := storetypes.StoreUpgrades{
		// New stores present in your v1.6.x app.go StoreKeys()
		Added: []string{
			consensusparamtypes.StoreKey,
			crisistypes.StoreKey,
			ibcexported.StoreKey,
			icacontrollertypes.StoreKey,
			icahosttypes.StoreKey,
			e2eetypes.StoreKey,
		},
		// Stores present in Genesis v1.0.0 StoreKeys() but NOT present in v1.6.x StoreKeys()
		Deleted: []string{
			"capability",
			"authz",
			"ibc",
			"feeibc",
			// If your Genesis v1.0.0 ever had ICA auth store, delete it too:
			// "icaauth",
			// If you ever enabled experimental gravity store on mainnet, delete it too:
			// gravitytypes.StoreKey,
		},
	}

	// app.go uses MaxVersionStoreLoader(qmsVersion) by default; follow Cronos pattern and override it at upgrade height.
	app.SetStoreLoader(MaxVersionUpgradeStoreLoader(maxVersion, upgradeInfo.Height, &storeUpgrades))
	return true
}
