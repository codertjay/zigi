package simulation

import (
	"fmt"
	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"

	"cosmossdk.io/math"

	"zigchain/x/dex/keeper"
	"zigchain/x/dex/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

func SimulateMsgRemoveLiquidity(
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get all existing pools
		pools := k.GetAllPool(ctx)
		if len(pools) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgRemoveLiquidity{}), "no pools available"), nil, nil
		}

		// Select a random pool
		pool := pools[r.Intn(len(pools))]

		// Select a random account
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// Get LP token denom from pool
		lpTokenDenom := pool.LpToken.Denom

		// Check account's LP token balance
		allBalances := bk.GetAllBalances(ctx, simAccount.Address)
		var lpTokenBalance sdk.Coin
		found := false
		for _, balance := range allBalances {
			if balance.Denom == lpTokenDenom {
				lpTokenBalance = balance
				found = true
				break
			}
		}
		if !found || !lpTokenBalance.IsPositive() {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgRemoveLiquidity{}),
				fmt.Sprintf("account does not have LP tokens for pool %s", pool.PoolId),
			), nil, nil
		}

		// Generate a random amount to remove (between 1 and available balance)
		maxAmount := lpTokenBalance.Amount
		if maxAmount.IsZero() {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgRemoveLiquidity{}),
				"account has zero LP token balance",
			), nil, nil
		}

		// Use a portion of the balance (between 1% and 100%)
		removeAmount := maxAmount.QuoRaw(100).MulRaw(int64(simtypes.RandIntBetween(r, 1, 100)))
		if removeAmount.IsZero() {
			removeAmount = math.NewInt(1)
		}
		if removeAmount.GT(maxAmount) {
			removeAmount = maxAmount
		}

		lpToken := sdk.NewCoin(lpTokenDenom, removeAmount)

		// Optionally set a receiver (can be empty to use creator)
		var receiver string
		if r.Intn(2) == 0 {
			receiverAccount, _ := simtypes.RandomAcc(r, accs)
			receiver = receiverAccount.Address.String()
		}

		msg := &types.MsgRemoveLiquidity{
			Creator:  simAccount.Address.String(),
			Lptoken:  lpToken,
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
			SimAccount:      simAccount,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(lpToken),
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
