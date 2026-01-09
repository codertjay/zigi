package simulation

import (
	"fmt"
	// nosem: math-random-used
	"math/rand" // checked: used for simulation,
	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests

	"cosmossdk.io/math"

	"zigchain/x/dex/keeper"
	"zigchain/x/dex/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

func SimulateMsgSwap(
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accounts []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get all existing pools
		pools := k.GetAllPool(ctx)
		if len(pools) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgSwapExactIn{}), "no pools available"), nil, nil
		}

		// Select a random pool
		pool := pools[r.Intn(len(pools))]

		// Ensure pool has at least 2 coins
		if len(pool.Coins) < 2 {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgSwapExactIn{}), "pool does not have enough coins"), nil, nil
		}

		// Select a random account
		simAccount, _ := simtypes.RandomAcc(r, accounts)

		// Randomly select which coin to swap (base or quote)
		incomingDenom := pool.Coins[r.Intn(2)].Denom

		// Check account's balance for the incoming coin
		allBalances := bk.GetAllBalances(ctx, simAccount.Address)
		var incomingBalance sdk.Coin
		found := false
		for _, balance := range allBalances {
			if balance.Denom == incomingDenom {
				incomingBalance = balance
				found = true
				break
			}
		}
		if !found || !incomingBalance.IsPositive() {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgSwapExactIn{}),
				fmt.Sprintf("account does not have balance for incoming coin: %s", incomingDenom),
			), nil, nil
		}

		// Generate a random amount to swap (between 1 and available balance)
		maxAmount := incomingBalance.Amount
		// Use a portion of the balance (between 1% and 50% to leave some for fees)
		swapAmount := maxAmount.QuoRaw(100).MulRaw(int64(simtypes.RandIntBetween(r, 1, 50)))
		if swapAmount.IsZero() {
			swapAmount = math.NewInt(1)
		}
		if swapAmount.GT(maxAmount) {
			swapAmount = maxAmount
		}

		incomingCoin := sdk.NewCoin(incomingDenom, swapAmount)

		// Optionally set a receiver (can be empty to use signer)
		var receiver string
		if r.Intn(2) == 0 {
			receiverAccount, _ := simtypes.RandomAcc(r, accounts)
			receiver = receiverAccount.Address.String()
		}

		// Optionally set outgoing_min (minimum amount to receive)
		var outgoingMin *sdk.Coin
		if r.Intn(2) == 0 {
			// Set a minimum that's reasonable (e.g., 50% of expected)
			// This is a simplified calculation - in reality, we'd need to calculate based on pool reserves
			minAmount := swapAmount.QuoRaw(2)
			// Determine the outgoing denom (the other coin in the pool)
			var outgoingDenom string
			if incomingDenom == pool.Coins[0].Denom {
				outgoingDenom = pool.Coins[1].Denom
			} else {
				outgoingDenom = pool.Coins[0].Denom
			}
			outgoingMinCoin := sdk.NewCoin(outgoingDenom, minAmount)
			outgoingMin = &outgoingMinCoin
		}

		msg := &types.MsgSwapExactIn{
			Signer:      simAccount.Address.String(),
			Incoming:    incomingCoin,
			PoolId:      pool.PoolId,
			Receiver:    receiver,
			OutgoingMin: outgoingMin,
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
			SimAccount:      simAccount,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(incomingCoin),
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
