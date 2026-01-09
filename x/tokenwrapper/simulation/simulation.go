package simulation

import (
	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"
)

// #nosec G101 -- These are simulation operation weight constants, not credentials
// Simulation operation weights constants
const (
	OpWeightMsgFundModuleWallet         = "op_weight_msg_fund_module_wallet"
	OpWeightMsgWithdrawFromModuleWallet = "op_weight_msg_withdraw_from_module_wallet"
	OpWeightMsgEnableTokenWrapper       = "op_weight_msg_enable_token_wrapper"
	OpWeightMsgDisableTokenWrapper      = "op_weight_msg_disable_token_wrapper"
	OpWeightMsgUpdateIbcSettings        = "op_weight_msg_update_ibc_settings"

	DefaultWeightFundModuleWallet         = 100
	DefaultWeightWithdrawFromModuleWallet = 100
	DefaultWeightEnableTokenWrapper       = 50
	DefaultWeightDisableTokenWrapper      = 50
	DefaultWeightUpdateIbcSettings        = 30
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var weightFundModuleWallet int
	appParams.GetOrGenerate(OpWeightMsgFundModuleWallet, &weightFundModuleWallet, nil,
		func(_ *rand.Rand) {
			weightFundModuleWallet = DefaultWeightFundModuleWallet
		},
	)

	var weightWithdrawFromModuleWallet int
	appParams.GetOrGenerate(OpWeightMsgWithdrawFromModuleWallet, &weightWithdrawFromModuleWallet, nil,
		func(_ *rand.Rand) {
			weightWithdrawFromModuleWallet = DefaultWeightWithdrawFromModuleWallet
		},
	)

	var weightEnableTokenWrapper int
	appParams.GetOrGenerate(OpWeightMsgEnableTokenWrapper, &weightEnableTokenWrapper, nil,
		func(_ *rand.Rand) {
			weightEnableTokenWrapper = DefaultWeightEnableTokenWrapper
		},
	)

	var weightDisableTokenWrapper int
	appParams.GetOrGenerate(OpWeightMsgDisableTokenWrapper, &weightDisableTokenWrapper, nil,
		func(_ *rand.Rand) {
			weightDisableTokenWrapper = DefaultWeightDisableTokenWrapper
		},
	)

	var weightUpdateIbcSettings int
	appParams.GetOrGenerate(OpWeightMsgUpdateIbcSettings, &weightUpdateIbcSettings, nil,
		func(_ *rand.Rand) {
			weightUpdateIbcSettings = DefaultWeightUpdateIbcSettings
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightFundModuleWallet,
			SimulateMsgFundModuleWallet(txCfg, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightWithdrawFromModuleWallet,
			SimulateMsgWithdrawFromModuleWallet(txCfg, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightEnableTokenWrapper,
			SimulateMsgEnableTokenWrapper(txCfg, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightDisableTokenWrapper,
			SimulateMsgDisableTokenWrapper(txCfg, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightUpdateIbcSettings,
			SimulateMsgUpdateIbcSettings(txCfg, ak, bk, k),
		),
	}
}

// ProposalMsgs returns all the tokenwrapper module msgs used to simulate governance proposals
func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}

// RegisterStoreDecoder registers a decoder for tokenwrapper module's types
func RegisterStoreDecoder(registry simtypes.StoreDecoderRegistry) {
}

// GenerateGenesisState creates a randomized GenState of the tokenwrapper module
func GenerateGenesisState(simState *module.SimulationState) {
	// TODO: Implement genesis state generation if needed
}
