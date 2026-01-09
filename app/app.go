package app

import (
	"io"
	"math/rand"

	"zigchain/app/keepers"

	_ "cosmossdk.io/api/cosmos/tx/config/v1" // import for side-effects
	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	_ "cosmossdk.io/x/circuit"         // import for side effects
	_ "cosmossdk.io/x/evidence"        // import for side effects
	_ "cosmossdk.io/x/feegrant/module" // import for side effects
	_ "cosmossdk.io/x/nft/module"      // import for side effects
	_ "cosmossdk.io/x/upgrade"         // import for side effects
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import for side effects
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting" // import for side effects
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	_ "github.com/cosmos/cosmos-sdk/x/authz/module" // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/bank"         // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/consensus"    // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/distribution" // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/group/module" // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/mint"         // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/params"       // import for side effects
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	_ "github.com/cosmos/cosmos-sdk/x/slashing"                          // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/staking"                           // import for side effects
	_ "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts" // import for side effects

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"zigchain/zutils/constants"

	// this line is used by starport scaffolding # stargate/app/moduleImport

	"zigchain/docs"
)

// Using external constants, so we can use them in tests without the need to load app.go
const (
	// Name is the name of the application.
	Name = constants.BlockChainName
	// AccountAddressPrefix is the prefix for accounts addresses.
	AccountAddressPrefix = constants.AddressPrefix
	// ChainCoinType is the coin type of the chain.
	ChainCoinType = constants.CoinType
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string
)

var (
	_ runtime.AppI            = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*runtime.App
	keepers.AppKeepers

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// this line is used by starport scaffolding # stargate/app/keeperDeclaration

	// module manager
	ModuleBasics module.BasicManager

	// simulation manager
	sm *module.SimulationManager
}

func init() {
	var err error
	clienthelpers.EnvPrefix = Name
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory("." + Name)
	if err != nil {
		panic(err)
	}
}

// wasmModuleSimulationWrapper wraps wasm.AppModule to filter out MsgStoreCode operations
// from simulation to avoid address codec errors.
//
// This wrapper filters out:
// 1. Direct MsgStoreCode messages
// 2. MsgExec messages from authz module that may contain MsgStoreCode as inner messages
//
// Note: Due to limitations in accessing message content from simtypes.OperationMsg,
// MsgExec messages are filtered conservatively when detected. This ensures that
// MsgStoreCode operations wrapped in MsgExec (via authz grants) are also filtered.
type wasmModuleSimulationWrapper struct {
	module.AppModuleSimulation
}

// newWasmModuleSimulationWrapper creates a new wrapper that filters out MsgStoreCode operations
func newWasmModuleSimulationWrapper(wasmModule module.AppModuleSimulation) module.AppModuleSimulation {
	return &wasmModuleSimulationWrapper{AppModuleSimulation: wasmModule}
}

// containsMsgStoreCode checks if a message or its inner messages (in case of MsgExec) contain MsgStoreCode.
// Since OperationMsg only contains metadata, we check for MsgExec and filter it conservatively.
func containsMsgStoreCode(msg simtypes.OperationMsg) bool {
	// Check for direct MsgStoreCode messages
	isStoreCode := msg.Name == "MsgStoreCode" ||
		msg.Name == "StoreCode" ||
		msg.Name == "/cosmwasm.wasm.v1.MsgStoreCode" ||
		msg.Route == "/cosmwasm.wasm.v1.MsgStoreCode"

	if isStoreCode {
		return true
	}

	// Check if this is a MsgExec message from authz module
	// MsgExec can contain MsgStoreCode as an inner message, so we filter it conservatively
	isMsgExec := msg.Name == "MsgExec" ||
		msg.Name == "/cosmos.authz.v1beta1.MsgExec" ||
		msg.Route == "/cosmos.authz.v1beta1.MsgExec" ||
		msg.Route == authztypes.ModuleName

	if isMsgExec {
		// Conservatively filter out MsgExec messages to prevent MsgStoreCode wrapped inside
		// from bypassing the filter. This is necessary because OperationMsg doesn't provide
		// access to the actual message bytes to inspect inner messages.
		return true
	}

	return false
}

// WeightedOperations filters out operations that generate MsgStoreCode messages
func (w *wasmModuleSimulationWrapper) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := w.AppModuleSimulation.WeightedOperations(simState)
	filtered := make([]simtypes.WeightedOperation, 0, len(operations))

	_ = len(operations) // placeholder for debug info

	for i, op := range operations {
		opIndex := i // capture for closure
		// Wrap the operation to check if it generates MsgStoreCode
		wrappedOp := simulation.NewWeightedOperation(
			op.Weight(),
			func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (
				simtypes.OperationMsg, []simtypes.FutureOperation, error,
			) {
				msg, futureOps, err := op.Op()(r, app, ctx, accs, chainID)

				// Check if the message contains MsgStoreCode (directly or wrapped in MsgExec)
				if containsMsgStoreCode(msg) {
					// Log when we filter out MsgStoreCode or MsgExec
					if app != nil && app.Logger() != nil {
						app.Logger().Debug(
							"wasmModuleSimulationWrapper: filtering out operation containing MsgStoreCode",
							"msg_name", msg.Name,
							"msg_route", msg.Route,
							"msg_ok", msg.OK,
							"msg_comment", msg.Comment,
							"operation_index", opIndex,
						)
					}
					// Return a no-op message instead
					return simtypes.NoOpMsg(wasmtypes.ModuleName, "/cosmwasm.wasm.v1.MsgStoreCode", "MsgStoreCode simulation disabled (including when wrapped in MsgExec)"), nil, nil
				}

				// Debug: log other wasm operations that pass through (only for non-empty messages)
				if app != nil && app.Logger() != nil && msg.Name != "" {
					app.Logger().Debug(
						"wasmModuleSimulationWrapper: allowing wasm operation",
						"msg_name", msg.Name,
						"msg_route", msg.Route,
						"operation_index", opIndex,
					)
				}

				return msg, futureOps, err
			},
		)
		filtered = append(filtered, wrappedOp)
	}

	return filtered
}

// New returns a reference to an initialized App.
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) (*App, error) {
	var (
		app        = &App{}
		appBuilder *runtime.AppBuilder
	)

	// initialize keepers
	app.AppKeepers = keepers.NewAppKeepers(
		&appBuilder,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		logger,
		appOpts,
	)

	// add to default baseapp options
	// enable optimistic execution
	baseAppOptions = append(baseAppOptions, baseapp.SetOptimisticExecution())

	// build app
	app.App = appBuilder.Build(db, traceStore, baseAppOptions...)

	// Register legacy modules
	if err := app.registerIBCModules(appOpts); err != nil {
		return nil, err
	}

	// register streaming services
	if err := app.RegisterStreamingServices(appOpts, app.kvStoreKeys()); err != nil {
		return nil, err
	}

	/****  Module Options ****/

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing transactions
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(
			app.appCodec,
			app.AccountKeeper,
			authsims.RandomGenesisAccounts,
			app.GetSubspace(authtypes.ModuleName),
		),
		// Override wasm module to use correct TxConfig with bech32 prefixes
		// The wasm module's simulation code creates its own interface registry without address codec
		// Filter out MsgStoreCode operations to avoid address codec errors
		// Also filters out MsgExec messages that may contain MsgStoreCode as inner messages
		wasmtypes.ModuleName: newWasmModuleSimulationWrapper(
			wasm.NewAppModule(
				app.AppCodec(),
				&app.WasmKeeper,
				app.StakingKeeper,
				app.AccountKeeper,
				app.BankKeeper,
				app.MsgServiceRouter(),
				app.GetSubspace(wasmtypes.ModuleName),
			),
		),
	}
	app.sm = module.NewSimulationManagerFromAppModules(
		app.ModuleManager.Modules,
		overrideModules,
	)
	app.sm.RegisterStoreDecoders()

	// A custom InitChainer sets if extra pre-init-genesis logic is required.
	// This is necessary for manually registered modules that do not support app wiring.
	// Manually set the module version map as shown below.
	// The upgrade module will automatically handle de-duplication of the module version map.
	app.SetInitChainer(func(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
		if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap()); err != nil {
			return nil, err
		}
		return app.App.InitChainer(ctx, req)
	})

	app.SetPreBlocker(app.PreBlocker)

	// Set upgrade handlers and store loaders
	app.setupUpgradeHandlers()
	app.setupUpgradeStoreLoaders()

	if err := app.Load(loadLatest); err != nil {
		return nil, err
	}

	// configure the wasm variables
	configureWasmVariables()

	return app, app.WasmKeeper.
		InitializePinnedCodes(app.NewUncachedContext(true, tmproto.Header{}))
}

// PreBlocker application updates every pre block
func (app *App) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	PreBlockForks(ctx, app)
	return app.App.PreBlocker(ctx, req)
}

// LegacyAmino returns App's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns App's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns App's interfaceRegistry.
func (app *App) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

// TxConfig returns App's tx config.
func (app *App) TxConfig() client.TxConfig {
	return app.txConfig
}

// GetKey returns the KVStoreKey for the provided store key.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	kvStoreKey, ok := app.UnsafeFindStoreKey(storeKey).(*storetypes.KVStoreKey)
	if !ok {
		return nil
	}
	return kvStoreKey
}

// GetMemKey returns the MemoryStoreKey for the provided store key.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	key, ok := app.UnsafeFindStoreKey(storeKey).(*storetypes.MemoryStoreKey)
	if !ok {
		return nil
	}

	return key
}

// kvStoreKeys returns all the kv store keys registered inside App.
func (app *App) kvStoreKeys() map[string]*storetypes.KVStoreKey {
	keys := make(map[string]*storetypes.KVStoreKey)
	// checked: valid use
	for _, k := range app.GetStoreKeys() {
		if kv, ok := k.(*storetypes.KVStoreKey); ok {
			keys[kv.Name()] = kv
		}
	}

	return keys
}

// GetSubspace returns a param subspace for a given module name.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface.
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	app.App.RegisterAPIRoutes(apiSvr, apiConfig)
	// register swagger API in app.go so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}

	// register app's OpenAPI routes.
	docs.RegisterOpenAPIService(Name, apiSvr.Router)
}
