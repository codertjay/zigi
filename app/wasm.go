package app

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cast"

	owasm "zigchain/wasmbinding"
	"zigchain/zutils/constants"
)

var (
	// customWasmCapabilities defines chain-specific WASM capabilities.
	// These are in addition to the built-in capabilities provided by CosmWasm.
	customWasmCapabilities = []string{
		constants.BlockChainName, // "zigchain" - required for chain-specific WASM contracts
	}
)

// registerWasmModules register CosmWasm keepers and non dependency inject modules.
func (app *App) registerWasmModules(
	appOpts servertypes.AppOptions,
	wasmOpts ...wasmkeeper.Option,
) (wasm.IBCHandler, error) {
	// set up non depinject support modules store keys
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(wasmtypes.StoreKey),
	); err != nil {
		panic(err)
	}

	homePath := cast.ToString(appOpts.Get(flags.FlagHome))

	wasmDir := homePath
	wasmConfig, err := wasm.ReadNodeConfig(appOpts)
	if err != nil {
		panic("error while reading wasm config: " + err.Error())
	}

	// Register custom plugins, pass our module's keepers to wasm bindings to interact with them
	wasmOpts = append(
		owasm.RegisterCustomPlugins(
			&app.BankKeeperBase,
			&app.FactoryKeeper,
			&app.DexKeeper,
		),
		wasmOpts...,
	)

	// register stargate queries,
	wasmOpts = append(
		owasm.RegisterStargateQueries(
			*app.GRPCQueryRouter(),
			app.appCodec,
		),
		wasmOpts...,
	)

	app.WasmKeeper = wasmkeeper.NewKeeper(
		app.AppCodec(),
		runtime.NewKVStoreService(app.GetKey(wasmtypes.StoreKey)),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		wasmtypes.VMConfig{},
		getWasmCapabilities(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	// register IBC modules
	if err := app.RegisterModules(
		wasm.NewAppModule(
			app.AppCodec(),
			&app.WasmKeeper,
			app.StakingKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			app.MsgServiceRouter(),
			app.GetSubspace(wasmtypes.ModuleName),
		)); err != nil {
		return wasm.IBCHandler{}, err
	}

	if err := app.setAnteHandler(app.txConfig, wasmConfig, app.GetKey(wasmtypes.StoreKey)); err != nil {
		return wasm.IBCHandler{}, err
	}

	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
		)
		if err != nil {
			return wasm.IBCHandler{}, fmt.Errorf("failed to register snapshot extension: %w", err)
		}
	}

	if err := app.setPostHandler(); err != nil {
		return wasm.IBCHandler{}, err
	}

	// At startup, after all modules have been registered, check that all proto
	// annotations are correct.
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		return wasm.IBCHandler{}, err
	}
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		return wasm.IBCHandler{}, err
	}

	// Create fee enabled wasm ibc Stack
	wasmStack := wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper)

	return wasmStack, nil
}

func (app *App) setPostHandler() error {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		return err
	}
	app.SetPostHandler(postHandler)
	return nil
}

func (app *App) setAnteHandler(txConfig client.TxConfig, wasmConfig wasmtypes.NodeConfig, txCounterStoreKey *storetypes.KVStoreKey) error {
	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				SignModeHandler: txConfig.SignModeHandler(),
				FeegrantKeeper:  app.FeeGrantKeeper,
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},
			IBCKeeper:             app.IBCKeeper,
			WasmConfig:            &wasmConfig,
			WasmKeeper:            &app.WasmKeeper,
			TXCounterStoreService: runtime.NewKVStoreService(txCounterStoreKey),
			CircuitKeeper:         &app.CircuitBreakerKeeper,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create AnteHandler: %w", err)
	}

	// Set the AnteHandler for the app
	app.SetAnteHandler(anteHandler)
	return nil
}

// getWasmCapabilities returns all WASM capabilities (built-in + custom).
// This combines CosmWasm's built-in capabilities with chain-specific custom capabilities.
func getWasmCapabilities() []string {
	return append(wasmkeeper.BuiltInCapabilities(), customWasmCapabilities...)
}

// configureWasmVariables configures the wasm variables
func configureWasmVariables() {
	// default values
	maxLabelSize := 256            // to set the maximum label size on instantiation (default 128)
	maxWasmSize := 1638400         //  to set the max size of compiled wasm to be accepted (default 819200)
	maxProposalWasmSize := 6291456 // to set the max size of gov proposal compiled wasm to be accepted (default 3145728)

	// increase wasm size limit
	wasmtypes.MaxLabelSize = maxLabelSize
	wasmtypes.MaxWasmSize = maxWasmSize
	wasmtypes.MaxProposalWasmSize = maxProposalWasmSize
}
