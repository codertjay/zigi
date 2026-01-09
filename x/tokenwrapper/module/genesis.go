package tokenwrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	// Initialize total transfer amounts
	k.SetTotalTransferredIn(ctx, genState.TotalTransferredIn)
	k.SetTotalTransferredOut(ctx, genState.TotalTransferredOut)

	// Initialize operator address
	k.SetOperatorAddress(ctx, genState.OperatorAddress)
	k.SetProposedOperatorAddress(ctx, genState.ProposedOperatorAddress)

	// Initialize pauser addresses
	k.SetPauserAddresses(ctx, genState.PauserAddresses)

	// Initialize enabled status
	k.SetEnabled(ctx, genState.Enabled)

	// Initialize IBC settings
	k.SetNativeClientId(ctx, genState.NativeClientId)
	k.SetCounterpartyClientId(ctx, genState.CounterpartyClientId)
	k.SetNativePort(ctx, genState.NativePort)
	k.SetCounterpartyPort(ctx, genState.CounterpartyPort)
	k.SetNativeChannel(ctx, genState.NativeChannel)
	k.SetCounterpartyChannel(ctx, genState.CounterpartyChannel)
	k.SetDenom(ctx, genState.Denom)

	if err := k.SetDecimalDifference(ctx, genState.DecimalDifference); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.Params = k.GetParams(ctx)

	genesis.TotalTransferredIn = k.GetTotalTransferredIn(ctx)
	genesis.TotalTransferredOut = k.GetTotalTransferredOut(ctx)

	genesis.OperatorAddress = k.GetOperatorAddress(ctx)
	genesis.ProposedOperatorAddress = k.GetProposedOperatorAddress(ctx)
	genesis.PauserAddresses = k.GetPauserAddresses(ctx)
	genesis.Enabled = k.IsEnabled(ctx)

	genesis.NativeClientId = k.GetNativeClientId(ctx)
	genesis.CounterpartyClientId = k.GetCounterpartyClientId(ctx)
	genesis.NativePort = k.GetNativePort(ctx)
	genesis.CounterpartyPort = k.GetCounterpartyPort(ctx)
	genesis.NativeChannel = k.GetNativeChannel(ctx)
	genesis.CounterpartyChannel = k.GetCounterpartyChannel(ctx)
	genesis.Denom = k.GetDenom(ctx)
	genesis.DecimalDifference = k.GetDecimalDifference(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
