package tokenwrapper

import (
	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"zigchain/testutil/sample"
	tokenwrappersimulation "zigchain/x/tokenwrapper/simulation"
	"zigchain/x/tokenwrapper/types"
)

// avoid unused import issue
var (
	_ = tokenwrappersimulation.FindAccount
	_ = rand.Rand{}
	_ = sample.AccAddress
	_ = sdk.AccAddress{}
	_ = simulation.MsgEntryKind
)

// #nosec G101 -- This is a simulation operation weight constant, not a credential
const (
	opWeightMsgFundModuleWallet = "op_weight_msg_fund_module_wallet"
	// TODO: Determine the simulation weight value
	defaultWeightMsgFundModuleWallet int = 100

	opWeightMsgRecoverZig = "op_weight_msg_recover_zig"
	// TODO: Determine the simulation weight value
	defaultWeightMsgRecoverZig int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	tokenwrapperGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&tokenwrapperGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgFundModuleWallet int
	simState.AppParams.GetOrGenerate(opWeightMsgFundModuleWallet, &weightMsgFundModuleWallet, nil,
		func(_ *rand.Rand) {
			weightMsgFundModuleWallet = defaultWeightMsgFundModuleWallet
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgFundModuleWallet,
		tokenwrappersimulation.SimulateMsgFundModuleWallet(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgRecoverZig int
	simState.AppParams.GetOrGenerate(opWeightMsgRecoverZig, &weightMsgRecoverZig, nil,
		func(_ *rand.Rand) {
			weightMsgRecoverZig = defaultWeightMsgRecoverZig
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRecoverZig,
		tokenwrappersimulation.SimulateMsgRecoverZig(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			opWeightMsgFundModuleWallet,
			defaultWeightMsgFundModuleWallet,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				tokenwrappersimulation.SimulateMsgFundModuleWallet(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgRecoverZig,
			defaultWeightMsgRecoverZig,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				tokenwrappersimulation.SimulateMsgRecoverZig(simState.TxConfig, am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		// this line is used by starport scaffolding # simapp/module/OpMsg
	}
}
