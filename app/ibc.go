package app

import (
	"fmt"
	"path/filepath"

	tokenwrappermodule "zigchain/x/tokenwrapper/module"

	"cosmossdk.io/core/appmodule"
	storetypes "cosmossdk.io/store/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvm "github.com/CosmWasm/wasmvm/v2"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	packetforward "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/types"
	ratelimit "github.com/cosmos/ibc-apps/modules/rate-limiting/v10"
	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/keeper"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
	ratelimitv2 "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/v2"
	ibcwasm "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10"
	"github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/blsverifier"
	ibcwasmkeeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/keeper"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
	icamodule "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	ibccallbacks "github.com/cosmos/ibc-go/v10/modules/apps/callbacks"
	ibccallbacksv2 "github.com/cosmos/ibc-go/v10/modules/apps/callbacks/v2"
	ibctransfer "github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	transferv2 "github.com/cosmos/ibc-go/v10/modules/apps/transfer/v2"
	ibc "github.com/cosmos/ibc-go/v10/modules/core"
	ibcclienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcapi "github.com/cosmos/ibc-go/v10/modules/core/api"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"
	solomachine "github.com/cosmos/ibc-go/v10/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

	// this line is used by starport scaffolding # ibc/app/import

	"zigchain/zutils/constants"
)

// registerIBCModules register IBC keepers and non dependency inject modules.
func (app *App) registerIBCModules(appOpts servertypes.AppOptions) error {
	// set up non depinject support modules store keys
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(ibcexported.StoreKey),
		storetypes.NewKVStoreKey(ibctransfertypes.StoreKey),
		storetypes.NewKVStoreKey(icahosttypes.StoreKey),
		storetypes.NewKVStoreKey(icacontrollertypes.StoreKey),
		storetypes.NewTransientStoreKey(paramstypes.TStoreKey),
		storetypes.NewKVStoreKey(packetforwardtypes.StoreKey),
		storetypes.NewKVStoreKey(ratelimittypes.StoreKey),
	); err != nil {
		return err
	}

	// register the key tables for legacy param subspaces
	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	app.ParamsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(keyTable)
	app.ParamsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())
	app.ParamsKeeper.Subspace(icacontrollertypes.SubModuleName).WithKeyTable(icacontrollertypes.ParamKeyTable())
	app.ParamsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())
	app.ParamsKeeper.Subspace(ratelimittypes.ModuleName).WithKeyTable(ratelimittypes.ParamKeyTable())

	// Create IBC keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(ibcexported.StoreKey)),
		app.GetSubspace(ibcexported.ModuleName),
		app.UpgradeKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://docs.cosmos.network/main/modules/gov#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper))

	// Set legacy router for backwards compatibility with gov v1beta1
	app.GovKeeper.SetLegacyRouter(govRouter)

	// Create custom ICS4Wrapper so that we can use it in the transfer stack
	ics4Wrapper := tokenwrappermodule.NewICS4Wrapper(
		app.IBCKeeper.ChannelKeeper,
		app.BankKeeper,
		nil, // Will be zero-value here. Reference is set later on with SetTransferKeeper.
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ConnectionKeeper,
		app.TokenwrapperKeeper,
	)

	// Create RateLimit keeper
	app.RatelimitKeeper = *ratelimitkeeper.NewKeeper(
		app.appCodec, // BinaryCodec
		runtime.NewKVStoreService(app.GetKey(ratelimittypes.StoreKey)), // StoreKey
		app.GetSubspace(ratelimittypes.ModuleName),                     // param Subspace
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),       // authority
		app.BankKeeper,
		app.IBCKeeper.ChannelKeeper, // ChannelKeeper
		app.IBCKeeper.ClientKeeper,
		ics4Wrapper, // ICS4Wrapper
	)

	// PFMRouterKeeper must be created before TransferKeeper
	app.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(packetforwardtypes.StoreKey)),
		nil, // Will be zero-value here. Reference is set later on with SetTransferKeeper.
		app.IBCKeeper.ChannelKeeper,
		app.BankKeeper,
		app.RatelimitKeeper, // ICS4Wrapper
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create IBC transfer keeper
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(ibctransfertypes.StoreKey)),
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.PacketForwardKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Set the TransferKeeper reference in PacketForwardKeeper
	app.PacketForwardKeeper.SetTransferKeeper(app.TransferKeeper)

	// Set the TransferKeeper reference in ICS4Wrapper
	ics4Wrapper.SetTransferKeeper(app.TransferKeeper)

	// Create interchain account keepers
	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(icahosttypes.StoreKey)),
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.AccountKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(icacontrollertypes.StoreKey)),
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.MsgServiceRouter(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	wasmStack, err := app.registerWasmModules(appOpts)
	if err != nil {
		return err
	}

	// Create Transfer Stack (from bottom to top of stack)
	// - core IBC
	// - ratelimit
	// - pfm
	// - tokenwrapper
	// - callbacks
	// - transfer
	//
	// This is how transfer stack will work in the end:
	// * RecvPacket -> IBC core -> RateLimit -> PFM -> TokenWrapper -> Callbacks -> Transfer (AddRoute)
	// * SendPacket -> Transfer -> Callbacks -> TokenWrapper -> PFM -> RateLimit -> IBC core (ICS4Wrapper)

	// Create the transfer stack
	var transferStack porttypes.IBCModule
	transferStack = ibctransfer.NewIBCModule(app.TransferKeeper)

	// Create the callbacks stack
	cbStack := ibccallbacks.NewIBCMiddleware(transferStack, app.PacketForwardKeeper, wasmStack, constants.MaxIBCCallbackGas)

	// Now wrap the callbacks stack with the tokenwrapper middleware
	transferStack = tokenwrappermodule.NewIBCModule(
		app.TokenwrapperKeeper,
		app.TransferKeeper,
		app.BankKeeper,
		cbStack,
		app.IBCKeeper.ChannelKeeper, // Pass ChannelKeeper for channel state checks
		app.IBCKeeper.ConnectionKeeper,
	)

	// Set the PacketForwardKeeper for the transfer stack
	transferStack = packetforward.NewIBCMiddleware(
		transferStack,
		app.PacketForwardKeeper,
		0,
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
	)

	// Set the RatelimitKeeper for the transfer stack
	transferStack = ratelimit.NewIBCMiddleware(app.RatelimitKeeper, transferStack)

	// Create IBCv2 Router & seal
	var ibcv2TransferStack ibcapi.IBCModule
	ibcv2TransferStack = transferv2.NewIBCModule(app.TransferKeeper)
	ibcv2TransferStack = ibccallbacksv2.NewIBCMiddleware(
		ibcv2TransferStack,
		app.IBCKeeper.ChannelKeeperV2,
		wasmStack,
		app.IBCKeeper.ChannelKeeperV2,
		constants.MaxIBCCallbackGas,
	)
	ibcv2TransferStack = ratelimitv2.NewIBCMiddleware(app.RatelimitKeeper, ibcv2TransferStack)

	// Create ICAHost Stack
	var icaHostStack porttypes.IBCModule = icahost.NewIBCModule(app.ICAHostKeeper)

	// integration point for custom authentication modules
	var icaControllerStack porttypes.IBCModule = icacontroller.NewIBCMiddleware(app.ICAControllerKeeper)
	icaControllerStack = ibccallbacks.NewIBCMiddleware(
		icaControllerStack,
		app.IBCKeeper.ChannelKeeper,
		wasmStack,
		constants.MaxIBCCallbackGas,
	)
	icaICS4Wrapper := icaControllerStack.(porttypes.ICS4Wrapper)
	app.ICAControllerKeeper.WithICS4Wrapper(icaICS4Wrapper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter().
		AddRoute(icahosttypes.SubModuleName, icaHostStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		AddRoute(wasmtypes.ModuleName, wasmStack)

	// Create IBCv2 Router & seal
	ibcv2Router := ibcapi.NewRouter().
		AddRoute(ibctransfertypes.PortID, ibcv2TransferStack)
	app.IBCKeeper.SetRouterV2(ibcv2Router)

	// this line is used by starport scaffolding # ibc/app/module

	app.IBCKeeper.SetRouter(ibcRouter)

	wasmLightClientQuerier := ibcwasmkeeper.QueryPlugins{
		// Custom: MyCustomQueryPlugin(),
		// `myAcceptList` is a `[]string` containing the list of gRPC query paths that the chain wants to allow for the `08-wasm` module to query.
		// These queries must be registered in the chain's gRPC query router, be deterministic, and track their gas usage.
		// The `AcceptListStargateQuerier` function will return a query plugin that will only allow queries for the paths in the `myAcceptList`.
		// The query responses are encoded in protobuf unlike the implementation in `x/wasm`.
		Stargate: ibcwasmkeeper.AcceptListStargateQuerier([]string{
			"/ibc.core.client.v1.Query/ClientState",
			"/ibc.core.client.v1.Query/ConsensusState",
			"/ibc.core.connection.v1.Query/Connection",
		}, app.GRPCQueryRouter()),
		Custom: blsverifier.CustomQuerier(),
	}

	// Get home directory from appOpts to support unique test directories
	homePath := DefaultNodeHome
	if homeOpt := appOpts.Get(flags.FlagHome); homeOpt != nil {
		if homeStr, ok := homeOpt.(string); ok && homeStr != "" {
			homePath = homeStr
		}
	}
	dataDir := filepath.Join(homePath, "data")

	var memCacheSizeMB uint32 = 100
	lc08, err := wasmvm.NewVM(filepath.Join(dataDir, "08-light-client"), wasmkeeper.BuiltInCapabilities(), 32, false, memCacheSizeMB)
	if err != nil {
		panic(fmt.Errorf("failed to create VM for 08 light client: %w", err))
	}

	app.WasmClientKeeper = ibcwasmkeeper.NewKeeperWithVM(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(ibcwasmtypes.StoreKey)),
		app.IBCKeeper.ClientKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		lc08,
		app.GRPCQueryRouter(),
		ibcwasmkeeper.WithQueryPlugins(&wasmLightClientQuerier),
	)

	// Light client modules
	clientKeeper := app.IBCKeeper.ClientKeeper
	storeProvider := app.IBCKeeper.ClientKeeper.GetStoreProvider()

	tmLightClientModule := ibctm.NewLightClientModule(app.appCodec, storeProvider)
	clientKeeper.AddRoute(ibctm.ModuleName, &tmLightClientModule)

	wasmLightClientModule := ibcwasm.NewLightClientModule(app.WasmClientKeeper, storeProvider)
	clientKeeper.AddRoute(ibcwasmtypes.ModuleName, &wasmLightClientModule)

	// register IBC modules
	if err := app.RegisterModules(
		ibc.NewAppModule(app.IBCKeeper),
		ibctransfer.NewAppModule(app.TransferKeeper),
		icamodule.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper),
		packetforward.NewAppModule(app.PacketForwardKeeper, app.GetSubspace(packetforwardtypes.ModuleName)),
		ratelimit.NewAppModule(app.appCodec, app.RatelimitKeeper),
		ibctm.AppModule{},
		solomachine.AppModule{},
	); err != nil {
		return err
	}

	return nil
}

// RegisterIBC Since the IBC modules don't support dependency injection,
// we need to manually register the modules on the client side.
// This needs to be removed after IBC supports App Wiring.
func RegisterIBC(registry cdctypes.InterfaceRegistry) map[string]appmodule.AppModule {
	modules := map[string]appmodule.AppModule{
		ibcexported.ModuleName:        ibc.AppModule{},
		ibctransfertypes.ModuleName:   ibctransfer.AppModule{},
		icatypes.ModuleName:           icamodule.AppModule{},
		packetforwardtypes.ModuleName: packetforward.AppModule{},
		ratelimittypes.ModuleName:     ratelimit.AppModule{},
		ibctm.ModuleName:              ibctm.AppModule{},
		solomachine.ModuleName:        solomachine.AppModule{},
		wasmtypes.ModuleName:          wasm.AppModule{},
	}

	for name, m := range modules {
		module.CoreAppModuleBasicAdaptor(name, m).RegisterInterfaces(registry)
	}

	return modules
}
