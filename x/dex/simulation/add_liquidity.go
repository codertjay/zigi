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

func SimulateMsgAddLiquidity(
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
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgAddLiquidity{}), "no pools available"), nil, nil
		}

		// Select a random pool
		pool := pools[r.Intn(len(pools))]

		// Ensure pool has at least 2 coins
		if len(pool.Coins) < 2 {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgAddLiquidity{}), "pool does not have enough coins"), nil, nil
		}

		// Select a random account
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// Get base and quote coins from pool
		baseDenom := pool.Coins[0].Denom
		quoteDenom := pool.Coins[1].Denom

		// Generate random amounts for base and quote
		// At least one must be positive
		baseAmount := math.NewInt(int64(simtypes.RandIntBetween(r, 1, 1000)))
		quoteAmount := math.NewInt(int64(simtypes.RandIntBetween(r, 0, 1000)))

		// If both are zero, make at least one positive
		if baseAmount.IsZero() && quoteAmount.IsZero() {
			baseAmount = math.NewInt(1)
		}

		baseCoin := sdk.NewCoin(baseDenom, baseAmount)
		quoteCoin := sdk.NewCoin(quoteDenom, quoteAmount)

		// Check if account has sufficient balance for base coin
		if !bk.HasBalance(ctx, simAccount.Address, baseCoin) {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgAddLiquidity{}),
				fmt.Sprintf("account does not have sufficient balance for base coin: %s", baseCoin),
			), nil, nil
		}

		// Check if account has sufficient balance for quote coin (if amount > 0)
		if quoteAmount.IsPositive() && !bk.HasBalance(ctx, simAccount.Address, quoteCoin) {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgAddLiquidity{}),
				fmt.Sprintf("account does not have sufficient balance for quote coin: %s", quoteCoin),
			), nil, nil
		}

		// Optionally set a receiver (can be empty to use creator)
		var receiver string
		if r.Intn(2) == 0 {
			receiverAccount, _ := simtypes.RandomAcc(r, accs)
			receiver = receiverAccount.Address.String()
		}

		msg := &types.MsgAddLiquidity{
			Creator:  simAccount.Address.String(),
			PoolId:   pool.PoolId,
			Base:     baseCoin,
			Quote:    quoteCoin,
			Receiver: receiver,
		}

		// Validate the message
		if err := msg.ValidateBasic(); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), err.Error()), nil, nil
		}

		// Calculate coins spent
		coinsSpent := sdk.NewCoins(baseCoin)
		if quoteAmount.IsPositive() {
			coinsSpent = coinsSpent.Add(quoteCoin)
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
			CoinsSpentInMsg: coinsSpent,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
