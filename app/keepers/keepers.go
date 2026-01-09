package keepers

import (
	"sort"

	_ "cosmossdk.io/api/cosmos/tx/config/v1" // import for side-effects
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	_ "cosmossdk.io/x/circuit" // import for side effects
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	_ "cosmossdk.io/x/evidence" // import for side effects
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	_ "cosmossdk.io/x/feegrant/module" // import for side effects
	nftkeeper "cosmossdk.io/x/nft/keeper"
	_ "cosmossdk.io/x/nft/module" // import for side effects
	_ "cosmossdk.io/x/upgrade"    // import for side effects
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import for side effects
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/authz/module" // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/bank"         // import for side effects
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/consensus" // import for side effects
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/distribution" // import for side effects
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/group/module" // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/mint"         // import for side effects
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/params" // import for side effects
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/slashing" // import for side effects
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/staking" // import for side effects
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/keeper"
	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/keeper"
	ibcwasmkeeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/keeper"
	_ "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts" // import for side effects
	icacontrollerkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/keeper"
	icahostkeeper "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	dexmodulekeeper "zigchain/x/dex/keeper"
	factorymodulekeeper "zigchain/x/factory/keeper"

	tokenwrappermodulekeeper "zigchain/x/tokenwrapper/keeper"
	// this line is used by starport scaffolding # stargate/app/moduleImport
)

type AppKeepers struct {
	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	BankKeeperBase        bankkeeper.BaseKeeper
	StakingKeeper         *stakingkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	ConsensusParamsKeeper consensuskeeper.Keeper

	SlashingKeeper       slashingkeeper.Keeper
	MintKeeper           mintkeeper.Keeper
	GovKeeper            *govkeeper.Keeper
	UpgradeKeeper        *upgradekeeper.Keeper
	ParamsKeeper         paramskeeper.Keeper
	AuthzKeeper          authzkeeper.Keeper
	EvidenceKeeper       evidencekeeper.Keeper
	FeeGrantKeeper       feegrantkeeper.Keeper
	GroupKeeper          groupkeeper.Keeper
	NFTKeeper            nftkeeper.Keeper
	CircuitBreakerKeeper circuitkeeper.Keeper

	// IBC
	IBCKeeper           *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	TransferKeeper      ibctransferkeeper.Keeper // for cross-chain fungible token transfers
	PacketForwardKeeper *packetforwardkeeper.Keeper
	RatelimitKeeper     ratelimitkeeper.Keeper

	// Custom modules
	FactoryKeeper factorymodulekeeper.Keeper
	DexKeeper     dexmodulekeeper.Keeper

	// CosmWasm
	WasmKeeper       wasmkeeper.Keeper
	WasmClientKeeper ibcwasmkeeper.Keeper

	TokenwrapperKeeper tokenwrappermodulekeeper.Keeper
}

// KeepersConfig returns the default keepers config.
func KeepersConfig() depinject.Config {
	return depinject.Configs(
		keepersConfig,
		// Loads the app config from a YAML file.
		// appconfig.LoadYAML(AppConfigYAML),
		depinject.Supply(
			// supply custom module basics
			map[string]module.AppModuleBasic{
				genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
				govtypes.ModuleName:     gov.NewAppModuleBasic(getGovProposalHandlers()),
				// this line is used by starport scaffolding # stargate/appConfig/moduleBasic
			},
		),
	)
}

func NewAppKeepers(
	appBuilder **runtime.AppBuilder,
	appCodec *codec.Codec,
	legacyAmino **codec.LegacyAmino,
	txConfig *client.TxConfig,
	interfaceRegistry *codectypes.InterfaceRegistry,
	logger log.Logger,
	appOpts servertypes.AppOptions,
) AppKeepers {
	var (
		appKeepers = AppKeepers{}

		// merge the AppConfig and other configuration in one config
		appConfig = depinject.Configs(
			KeepersConfig(),
			depinject.Supply(
				appOpts, // supply app options
				logger,  // supply logger
				// Supply with IBC keeper getter for the IBC modules with App Wiring.
				// The IBC Keeper cannot be passed because it has not been initiated yet.
				// Passing the getter, the app IBC Keeper will always be accessible.
				// This needs to be removed after IBC supports App Wiring.
				appKeepers.GetIBCKeeper,

				// here alternative options can be supplied to the DI container.
				// those options can be used f.e to override the default behavior of some modules.
				// for instance supplying a custom address codec for not using bech32 addresses.
				// read the depinject documentation and depinject module wiring for more information
				// on available options and how to use them.
			),
		)
	)

	if err := depinject.Inject(
		appConfig,
		appBuilder,
		appCodec,
		legacyAmino,
		txConfig,
		interfaceRegistry,
		&appKeepers.AccountKeeper,
		&appKeepers.BankKeeper,
		&appKeepers.StakingKeeper,
		&appKeepers.DistrKeeper,
		&appKeepers.ConsensusParamsKeeper,
		&appKeepers.SlashingKeeper,
		&appKeepers.MintKeeper,
		&appKeepers.GovKeeper,
		&appKeepers.UpgradeKeeper,
		&appKeepers.ParamsKeeper,
		&appKeepers.AuthzKeeper,
		&appKeepers.EvidenceKeeper,
		&appKeepers.FeeGrantKeeper,
		&appKeepers.NFTKeeper,
		&appKeepers.GroupKeeper,
		&appKeepers.CircuitBreakerKeeper,
		&appKeepers.FactoryKeeper,
		&appKeepers.DexKeeper,
		&appKeepers.TokenwrapperKeeper,
		// this line is used by starport scaffolding # stargate/app/keeperDefinition
	); err != nil {
		panic(err)
	}

	return appKeepers
}

// GetIBCKeeper returns the IBC keeper.
func (appKeepers *AppKeepers) GetIBCKeeper() *ibckeeper.Keeper {
	return appKeepers.IBCKeeper
}

// getGovProposalHandlers return the chain proposal handlers.
func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler
	// this line is used by starport scaffolding # stargate/app/govProposalHandlers

	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
		// this line is used by starport scaffolding # stargate/app/govProposalHandler
	)

	return govProposalHandlers
}

// GetMaccPerms returns a copy of the module account permissions
//
// NOTE: This is solely to be used for testing purposes.
func GetMaccPerms() map[string][]string {
	dup := make(map[string][]string)
	for _, perms := range moduleAccPerms {
		dup[perms.Account] = perms.Permissions
	}
	return dup
}

// BlockedAddresses returns all the app's blocked account addresses in a deterministic manner.
func BlockedAddresses() map[string]bool {
	result := make(map[string]bool)
	var addrs []string

	// Collect addresses into a slice
	if len(blockAccAddrs) > 0 {
		addrs = append(addrs, blockAccAddrs...)
	} else {
		// this is range over a map, so it is not deterministic
		for addr := range GetMaccPerms() {
			addrs = append(addrs, addr)
		}
	}

	// Sort the addresses to ensure determinism
	sort.Strings(addrs)

	// Populate the result map
	for _, addr := range addrs {
		result[addr] = true
	}

	return result
}
