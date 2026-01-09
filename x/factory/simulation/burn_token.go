package simulation

import (
	"sync"

	"fmt"

	cosmosmath "cosmossdk.io/math"

	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"
	d "zigchain/zutils/debug"

	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Debug info
var invocationCount int

// BurnMeta These vars we need to calculate only once as they don't change so we can optimize a bit
type BurnMeta struct {
	firstDenom types.Denom
	banker     simtypes.Account
}

var burnMetaInstance *BurnMeta
var once sync.Once

// GetBurnMeta returns burn meta from genesis state
func GetBurnMeta(ctx sdk.Context, fk keeper.Keeper, accs []simtypes.Account) *BurnMeta {

	// Initialize the burnMetaInstance only once
	once.Do(func() {
		// Get first denom
		firstDenom, found := getFirstDenom(ctx, fk)
		if !found {
			commentGetFirstDenom := "SimulateMsgBurnTokens: Could not find first denom in genesis state"
			// Log the error
			fk.Logger().Error(commentGetFirstDenom)

			// Panic in case we can't find the first denom in genesis state,
			// This might indicate a problem with the genesis state setup in the test environment.
			// This should never happen in a real-world application, just the way we set up genesis.
			panic("could not find first denom in genesis state")
		}

		// Get banker from first denom
		banker, err := GetBankerFromDenom(ctx, fk, accs, firstDenom.Denom)
		if err != nil {
			commentGetBanker := fmt.Sprintf(
				"SimulateMsgBurnTokens: Could not get banker from denom: %s",
				err,
			)
			// Log the error
			fk.Logger().Error(commentGetBanker)

			// THis should never happen as we used that in genesis to fund the first 10 accounts
			panic(err)
		}

		burnMetaInstance = &BurnMeta{
			firstDenom: firstDenom,
			banker:     banker,
		}

	})
	return burnMetaInstance

}

// getFirstDenom extract first Denom from genesis state
func getFirstDenom(ctx sdk.Context, fk keeper.Keeper) (denom types.Denom, found bool) {

	// Get a first denom, as we used that in genesis to fund the first 10 accounts
	allDenom := fk.GetAllDenom(ctx)

	// check if there are denoms in allDenom list
	if len(allDenom) == 0 {
		return types.Denom{}, false
	}

	// get first denom as we have funded that to banker in genesis state
	return allDenom[0], true

}

func SimulateMsgBurnTokens(
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	fk keeper.Keeper,
) simtypes.Operation {
	invocationCount++
	fk.Logger().Error(fmt.Sprintf("SimulateMsgBurnTokens invocation #%d", invocationCount))
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		burnMeta := GetBurnMeta(ctx, fk, accs)

		// Make sure bank admin is not empty, empty means no more minting allowed
		// This is a valid state, but we cannot execute minting in this case
		// Research how to mark this as ok in the simulation stats
		if burnMeta.banker.Address == nil {

			commentBankAdminEmpty := fmt.Sprintf(
				"SimulateMsgBurnTokens: Bank admin for Denom: %s is empty",
				burnMeta.firstDenom.Denom,
			)
			// Log the situation
			fk.Logger().Info(commentBankAdminEmpty)
			// Return a successful operation message without delivering a transaction
			// Create an OperationMsg directly
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgBurnTokens",
				Comment: commentBankAdminEmpty,
				OK:      true,
			}

			return opMsg, nil, nil
		}

		// Send this to logger this info only if Debug mode is enabled
		fk.Logger().Debug(fmt.Sprintf(
			"SimulateMsgBurnTokens: Bank admin for Denom: %s is %s",
			burnMeta.firstDenom.Denom,
			burnMeta.banker.Address.String(),
		))

		// how much coin banker has
		bankerCoin := bk.GetBalance(ctx, burnMeta.banker.Address, burnMeta.firstDenom.Denom)

		if bankerCoin.Amount.IsZero() {

			commentNoDenoms := fmt.Sprintf(
				"SimulateMsgBurnTokens: Banker %s has no denoms %s",
				burnMeta.banker.Address.String(),
				bankerCoin.Denom,
			)

			fk.Logger().Info(commentNoDenoms)
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgBurnTokens",
				Comment: commentNoDenoms,
				OK:      true,
			}

			return opMsg, nil, nil
		} else {
			fk.Logger().Info(
				fmt.Sprintf(
					"SimulateMsgBurnTokens: Banker %s has %s",
					burnMeta.banker.Address.String(),
					bankerCoin,
				),
			)
		}

		// TODO delete after debug
		fk.Logger().Error(d.L(ctx, fmt.Sprintf("Banker %s Coin: %s", burnMeta.banker.Address, bankerCoin)))

		// Calculate max burn amount (half of current balance, but at least 1)
		maxBurnAmount := bankerCoin.Amount.Quo(cosmosmath.NewInt(2))
		if maxBurnAmount.IsZero() {
			// If balance is 1, we can't burn half, so return early
			commentInsufficientBalance := fmt.Sprintf(
				"SimulateMsgBurnTokens: Banker %s has insufficient balance to burn (balance: %s, need at least 2)",
				burnMeta.banker.Address.String(),
				bankerCoin,
			)
			fk.Logger().Info(commentInsufficientBalance)
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgBurnTokens",
				Comment: commentInsufficientBalance,
				OK:      true,
			}
			return opMsg, nil, nil
		}

		// Generate a random amount to burn between 1 and maxBurnAmount (so we have some left for next time)
		// we need to pass r as a seed to RandUIntBetween - Rand is a source of random numbers.
		burnAmount := RandUIntBetween(
			r,
			cosmosmath.NewUint(1),
			cosmosmath.Uint(maxBurnAmount),
		)

		// TODO not needed remove after debug
		if burnAmount.LTE(cosmosmath.Uint(cosmosmath.ZeroInt())) {
			msg := fmt.Sprintf(
				"SimulateMsgBurnTokens: Banker %s has no denoms %s",
				burnMeta.banker.Address.String(),
				bankerCoin.Denom,
			)
			fk.Logger().Warn(msg)
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgBurnTokens",
				Comment: msg,
				OK:      true,
			}
			return opMsg, nil, nil
		}

		// Re-check balance right before creating the burn coin to ensure we have current balance
		// Balance might have changed due to other operations in the same block
		currentBalance := bk.GetBalance(ctx, burnMeta.banker.Address, burnMeta.firstDenom.Denom)
		if currentBalance.Amount.IsZero() {
			commentNoBalance := fmt.Sprintf(
				"SimulateMsgBurnTokens: Banker %s has no balance for denom %s",
				burnMeta.banker.Address.String(),
				burnMeta.firstDenom.Denom,
			)
			fk.Logger().Info(commentNoBalance)
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgBurnTokens",
				Comment: commentNoBalance,
				OK:      true,
			}
			return opMsg, nil, nil
		}

		// Adjust burn amount if it exceeds current balance
		// Use at most half of current balance, or the calculated amount, whichever is smaller
		maxBurnAmountFromCurrent := currentBalance.Amount.Quo(cosmosmath.NewInt(2))
		if maxBurnAmountFromCurrent.IsZero() {
			// If current balance is 1 or less, we can't burn anything
			commentInsufficientBalance := fmt.Sprintf(
				"SimulateMsgBurnTokens: Banker %s has insufficient balance to burn (current balance: %s, need at least 2)",
				burnMeta.banker.Address.String(),
				currentBalance,
			)
			fk.Logger().Info(commentInsufficientBalance)
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgBurnTokens",
				Comment: commentInsufficientBalance,
				OK:      true,
			}
			return opMsg, nil, nil
		}
		if burnAmount.GT(cosmosmath.Uint(maxBurnAmountFromCurrent)) {
			burnAmount = cosmosmath.Uint(maxBurnAmountFromCurrent)
		}

		// Ensure we have at least 1 to burn
		if burnAmount.LTE(cosmosmath.NewUint(0)) {
			commentNoBalance := fmt.Sprintf(
				"SimulateMsgBurnTokens: Banker %s has insufficient balance for denom %s (current: %s)",
				burnMeta.banker.Address.String(),
				burnMeta.firstDenom.Denom,
				currentBalance,
			)
			fk.Logger().Info(commentNoBalance)
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgBurnTokens",
				Comment: commentNoBalance,
				OK:      true,
			}
			return opMsg, nil, nil
		}

		burnCoin := sdk.NewCoin(burnMeta.firstDenom.Denom, cosmosmath.Int(burnAmount))

		fk.Logger().Info(
			fmt.Sprintf(
				"SimulateMsgBurnTokens: Burn amount: %s (current balance: %s)",
				burnCoin,
				currentBalance,
			),
		)

		// Final balance check before transaction
		if !bk.HasBalance(ctx, burnMeta.banker.Address, burnCoin) {
			fk.Logger().Warn(
				fmt.Sprintf(
					"SimulateMsgBurnTokens: Banker %s does not have enough denoms %s %s (current balance: %s)",
					burnMeta.banker.Address.String(),
					burnCoin.Amount,
					burnCoin.Denom,
					currentBalance,
				))

			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgBurnTokens{}),
				fmt.Sprintf(
					"SimulateMsgBurnTokens: Banker %s does not have enough denoms %s %s (current balance: %s)",
					burnMeta.banker.Address.String(),
					burnCoin.Amount,
					burnCoin.Denom,
					currentBalance,
				),
			), nil, nil
		}

		// TODO remove after debug
		log(ctx, fk, fmt.Sprintf(
			"SimulateMsgBurnTokens: Signer: %s \n balance: %s \n burn   : %s",
			burnMeta.banker.Address.String(),
			bankerCoin,
			burnCoin,
		))

		// generate the burn message
		burnMsg := &types.MsgBurnTokens{
			Signer: burnMeta.banker.Address.String(),
			Token:  burnCoin,
		}

		// transaction context
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txCfg,
			Cdc:             nil,
			Msg:             burnMsg,
			Context:         ctx,
			SimAccount:      burnMeta.banker,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(burnCoin),
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}

		// execute transaction
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
