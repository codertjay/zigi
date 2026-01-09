package simulation

import (
	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"zigchain/testutil/sample"
	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"
)

// SimulateMsgFundModuleWallet generates a MsgFundModuleWallet with random values
func SimulateMsgFundModuleWallet(txCfg client.TxConfig, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get a random account that has enough balance
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// Set this account as the operator
		k.SetOperatorAddress(ctx, simAccount.Address.String())

		// Generate a random amount to fund
		amount := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(int64(simtypes.RandIntBetween(r, 100, 1000)))))

		// Create the message
		msg := &types.MsgFundModuleWallet{
			Signer: simAccount.Address.String(),
			Amount: amount,
		}

		// Create a tx
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txCfg,
			Cdc:             nil,
			Msg:             msg,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: amount,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgWithdrawFromModuleWallet generates a MsgWithdrawFromModuleWallet with random values
func SimulateMsgWithdrawFromModuleWallet(txCfg client.TxConfig, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get a random account that has enough balance
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// Set this account as the operator
		k.SetOperatorAddress(ctx, simAccount.Address.String())

		// Generate a random amount to withdraw
		amount := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(int64(simtypes.RandIntBetween(r, 100, 1000)))))

		// Create the message
		msg := &types.MsgWithdrawFromModuleWallet{
			Signer: simAccount.Address.String(),
			Amount: amount,
		}

		// Create a tx
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txCfg,
			Cdc:             nil,
			Msg:             msg,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: amount,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgEnableTokenWrapper generates a MsgEnableTokenWrapper
func SimulateMsgEnableTokenWrapper(txCfg client.TxConfig, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get a random account
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// Set this account as the operator
		k.SetOperatorAddress(ctx, simAccount.Address.String())

		// Create the message
		msg := &types.MsgEnableTokenWrapper{
			Signer: simAccount.Address.String(),
		}

		// Create a tx
		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         txCfg,
			Cdc:           nil,
			Msg:           msg,
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgDisableTokenWrapper generates a MsgDisableTokenWrapper
func SimulateMsgDisableTokenWrapper(txCfg client.TxConfig, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get a random account
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// Set this account as the operator
		k.SetOperatorAddress(ctx, simAccount.Address.String())

		// Create the message
		msg := &types.MsgDisableTokenWrapper{
			Signer: simAccount.Address.String(),
		}

		// Create a tx
		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         txCfg,
			Cdc:           nil,
			Msg:           msg,
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgUpdateIbcSettings generates a MsgUpdateIbcSettings with random values
func SimulateMsgUpdateIbcSettings(txCfg client.TxConfig, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get a random account
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// Set this account as the operator
		k.SetOperatorAddress(ctx, simAccount.Address.String())

		// Create the message with random values
		msg := &types.MsgUpdateIbcSettings{
			Signer:               simAccount.Address.String(),
			NativeClientId:       sample.AccAddress(),
			CounterpartyClientId: sample.AccAddress(),
			NativePort:           "transfer",
			CounterpartyPort:     "transfer",
			NativeChannel:        "channel-0",
			CounterpartyChannel:  "channel-0",
			Denom:                "uzig",
			// #nosec G115 -- This is a simulation file, random number is bounded between 0 and 18
			DecimalDifference: uint32(simtypes.RandIntBetween(r, 0, 18)),
		}

		// Create a tx
		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         txCfg,
			Cdc:           nil,
			Msg:           msg,
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
