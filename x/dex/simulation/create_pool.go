package simulation

import (
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

func SimulateMsgCreatePool(
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// Get all balances for the account to find available denoms
		balances := bk.GetAllBalances(ctx, simAccount.Address)
		if len(balances) < 2 {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgCreatePool{}),
				"account does not have at least 2 different denoms",
			), nil, nil
		}

		// Find two different denoms with positive balances
		var baseCoin, quoteCoin sdk.Coin
		found := false

		for i := 0; i < len(balances) && !found; i++ {
			if !balances[i].IsPositive() {
				continue
			}
			for j := i + 1; j < len(balances); j++ {
				if !balances[j].IsPositive() {
					continue
				}
				if balances[i].Denom != balances[j].Denom {
					// Use a portion of each balance (between 1% and 50%)
					baseAmount := balances[i].Amount.QuoRaw(100).MulRaw(int64(simtypes.RandIntBetween(r, 1, 50)))
					if baseAmount.IsZero() {
						baseAmount = math.NewInt(1)
					}
					if baseAmount.GT(balances[i].Amount) {
						baseAmount = balances[i].Amount
					}

					quoteAmount := balances[j].Amount.QuoRaw(100).MulRaw(int64(simtypes.RandIntBetween(r, 1, 50)))
					if quoteAmount.IsZero() {
						quoteAmount = math.NewInt(1)
					}
					if quoteAmount.GT(balances[j].Amount) {
						quoteAmount = balances[j].Amount
					}

					baseCoin = sdk.NewCoin(balances[i].Denom, baseAmount)
					quoteCoin = sdk.NewCoin(balances[j].Denom, quoteAmount)
					found = true
					break
				}
			}
		}

		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgCreatePool{}),
				"could not find two different denoms with positive balances",
			), nil, nil
		}

		// Check if a pool already exists for this pair
		// We can't easily check this without the pool UIDs, so we'll let the message handler validate it

		// Optionally set a receiver (can be empty to use creator)
		var receiver string
		if r.Intn(2) == 0 {
			receiverAccount, _ := simtypes.RandomAcc(r, accs)
			receiver = receiverAccount.Address.String()
		}

		msg := &types.MsgCreatePool{
			Creator:  simAccount.Address.String(),
			Base:     baseCoin,
			Quote:    quoteCoin,
			Receiver: receiver,
		}

		// Validate the message
		if err := msg.ValidateBasic(); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), err.Error()), nil, nil
		}

		// Calculate coins spent
		coinsSpent := sdk.NewCoins(baseCoin, quoteCoin)

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
