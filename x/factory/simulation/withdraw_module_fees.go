package simulation

import (
	"fmt"
	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"

	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

func SimulateMsgWithdrawModuleFees(
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get params to check beneficiary
		params := k.GetParams(ctx)
		if params.Beneficiary == "" {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgWithdrawModuleFees{}),
				"beneficiary address is not set",
			), nil, nil
		}

		// Find the beneficiary account
		beneficiaryAccount, found := FindAccount(accs, params.Beneficiary)
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgWithdrawModuleFees{}),
				fmt.Sprintf("beneficiary account not found in simulation accounts: %s", params.Beneficiary),
			), nil, nil
		}

		// Check if module account has any balance
		moduleAccount := ak.GetModuleAccount(ctx, types.ModuleName)
		if moduleAccount == nil {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgWithdrawModuleFees{}),
				"module account not found",
			), nil, nil
		}

		moduleBalance := bk.GetAllBalances(ctx, moduleAccount.GetAddress())
		if moduleBalance.IsZero() {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgWithdrawModuleFees{}),
				"module account has no balance to withdraw",
			), nil, nil
		}

		// Optionally set a receiver (can be empty to use signer)
		var receiver string
		if r.Intn(2) == 0 {
			receiverAccount, _ := simtypes.RandomAcc(r, accs)
			receiver = receiverAccount.Address.String()
		}

		msg := &types.MsgWithdrawModuleFees{
			Signer:   beneficiaryAccount.Address.String(),
			Receiver: receiver,
		}

		// Validate the message
		if err := msg.ValidateBasic(); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), err.Error()), nil, nil
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txCfg,
			Cdc:             nil,
			Msg:             msg,
			Context:         ctx,
			SimAccount:      beneficiaryAccount,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(), // No coins spent, just withdrawing from module
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
