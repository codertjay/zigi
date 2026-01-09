package app_test

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/stretchr/testify/require"

	"zigchain/app"
	"zigchain/app/keepers"
)

// Profile with:
// `go test -benchmem -run=^$ -bench ^BenchmarkFullAppSimulation ./app -Commit=true -cpuprofile cpu.out`
func BenchmarkFullAppSimulation(b *testing.B) {
	b.ReportAllocs()

	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "goleveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	if err != nil {
		b.Fatalf("simulation setup failed: %s", err.Error())
	}

	if skip {
		b.Skip("skipping benchmark application simulation")
	}

	defer func() {
		require.NoError(b, db.Close())
		require.NoError(b, os.RemoveAll(dir))
	}()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = app.DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	bApp, err := app.New(logger, db, nil, true, appOptions, interBlockCacheOpt())
	require.NoError(b, err)
	require.Equal(b, app.Name, bApp.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		bApp.BaseApp,
		simtestutil.AppStateFn(bApp.AppCodec(), bApp.SimulationManager(), bApp.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		BuildSimulationOperations(bApp, bApp.AppCodec().(codec.JSONCodec), config, MakeTestTxConfig()),
		keepers.BlockedAddresses(),
		config,
		bApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	if err = simtestutil.CheckExportSimulation(bApp, config, simParams); err != nil {
		b.Fatal(err)
	}

	if simErr != nil {
		b.Fatal(simErr)
	}

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}

func BenchmarkInvariants(b *testing.B) {
	b.ReportAllocs()

	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "leveldb-app-invariant-bench", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	if err != nil {
		b.Fatalf("simulation setup failed: %s", err.Error())
	}

	if skip {
		b.Skip("skipping benchmark application simulation")
	}

	defer func() {
		require.NoError(b, db.Close())
		require.NoError(b, os.RemoveAll(dir))
	}()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = app.DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	bApp, err := app.New(logger, db, nil, true, appOptions, interBlockCacheOpt())
	require.NoError(b, err)
	require.Equal(b, app.Name, bApp.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		bApp.BaseApp,
		simtestutil.AppStateFn(bApp.AppCodec(), bApp.SimulationManager(), bApp.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		BuildSimulationOperations(bApp, bApp.AppCodec().(codec.JSONCodec), config, MakeTestTxConfig()),
		keepers.BlockedAddresses(),
		config,
		bApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	if err = simtestutil.CheckExportSimulation(bApp, config, simParams); err != nil {
		b.Fatal(err)
	}

	if simErr != nil {
		b.Fatal(simErr)
	}

	if config.Commit {
		simtestutil.PrintStats(db)
	}

	// NOTE: Invariant benchmarking is disabled since the crisis module has been removed.
	// The crisis module was previously used to access all registered invariants via CrisisKeeper.Routes().
	// Without the crisis module, there's no direct way to iterate through all invariants for benchmarking.
	// Invariants are still registered and checked during normal operation, but individual benchmarking
	// is no longer available.

	// ctx := bApp.NewContextLegacy(true, cmtproto.Header{Height: bApp.LastBlockHeight() + 1})

	// 3. Benchmark each invariant separately
	//
	// NOTE: We use the crisis keeper as it has all the invariants registered with
	// their respective metadata which makes it useful for testing/benchmarking.
	// for _, cr := range bApp.CrisisKeeper.Routes() {
	// 	cr := cr
	// 	b.Run(fmt.Sprintf("%s/%s", cr.ModuleName, cr.Route), func(b *testing.B) {
	// 		if res, stop := cr.Invar(ctx); stop {
	// 			b.Fatalf(
	// 				"broken invariant at block %d of %d\n%s",
	// 				ctx.BlockHeight()-1, config.NumBlocks, res,
	// 			)
	// 		}
	// 	})
	// }

	// ### Invariant benchmarking is disabled since the crisis module has been removed.
}
