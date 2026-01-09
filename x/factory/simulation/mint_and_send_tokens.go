package simulation

import (
	"fmt"
	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"

	cosmosmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgMintAndSendTokens(
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	fk keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// Get a random denom from existing denoms
		randomDenom, err := RandomExistingDenom(r, ctx, fk)

		commentGetDenom := fmt.Sprintf(
			"SimulateMsgMintAndSendTokens: Could not get all denoms: %s",
			err,
		)

		// Check for error
		if err != nil {
			// Log the error
			fk.Logger().Error(commentGetDenom)
			// Return a no-op message, meaning no operation was performed
			// This is recorded in the simulation stats as fail
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgMintAndSendTokens{}),
				commentGetDenom,
			), nil, nil
		}

		banker, err := GetBankerFromDenom(ctx, fk, accs, randomDenom.Denom)

		commentGetBanker := fmt.Sprintf(
			"SimulateMsgMintAndSendTokens: Could not get banker from denom: %s",
			err,
		)

		if err != nil {
			// Log the error
			fk.Logger().Error(commentGetBanker)
			// Return a no-op message, meaning no operation was performed
			// This is recorded in the simulation stats as fail
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgMintAndSendTokens{}),
				commentGetBanker,
			), nil, nil
		}

		commentBankAdminEmpty := fmt.Sprintf(
			"SimulateMsgMintAndSendTokens: Bank admin for Denom: %s is empty",
			randomDenom.Denom,
		)

		// Make sure bank admin is not empty, empty means no more minting allowed
		// This is a valid state, but we cannot execute minting in this case
		// Research how to mark this as ok in the simulation stats
		if banker.Address == nil {
			// Log the situation
			fk.Logger().Info(commentBankAdminEmpty)
			// Return a successful operation message without delivering a transaction
			// Create an OperationMsg directly
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgMintAndSendTokens",
				Comment: commentBankAdminEmpty,
				OK:      true,
			}

			return opMsg, nil, nil
		}

		// Send this to logger this info only if Debug mode is enabled
		fk.Logger().Debug(fmt.Sprintf(
			"SimulateMsgMintAndSendTokens: Bank admin for Denom: %s is %s",
			randomDenom.Denom,
			banker.Address.String(),
		))

		// Check if we've already reached the minting cap
		if randomDenom.Minted.GTE(randomDenom.MintingCap) {
			commentMintingCapReached := fmt.Sprintf(
				"SimulateMsgMintAndSendTokens: Minting cap reached for Denom: %s (Minted: %s, Cap: %s)",
				randomDenom.Denom,
				randomDenom.Minted.String(),
				randomDenom.MintingCap.String(),
			)
			fk.Logger().Info(commentMintingCapReached)
			// Return a successful operation message without delivering a transaction
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgMintAndSendTokens",
				Comment: commentMintingCapReached,
				OK:      true,
			}
			return opMsg, nil, nil
		}

		// Max mintable amount is the difference between max supply and current supply
		// We can mint only amount between max amount minus current supply and 1
		// First we calculate what is the max mintable amount
		maxMintableAmount := randomDenom.MintingCap.Sub(randomDenom.Minted)

		// Safety check: ensure maxMintableAmount is at least 1
		if maxMintableAmount.IsZero() {
			commentNoMintableAmount := fmt.Sprintf(
				"SimulateMsgMintAndSendTokens: No mintable amount for Denom: %s (Minted: %s, Cap: %s)",
				randomDenom.Denom,
				randomDenom.Minted.String(),
				randomDenom.MintingCap.String(),
			)
			fk.Logger().Info(commentNoMintableAmount)
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgMintAndSendTokens",
				Comment: commentNoMintableAmount,
				OK:      true,
			}
			return opMsg, nil, nil
		}

		// Generate a random amount to mint between 1 and maxMintableAmount
		// we need to pass r as a seed to RandUIntBetween - Rand is a source of random numbers.
		mintAmount := RandUIntBetween(r, cosmosmath.NewUint(1), maxMintableAmount)

		// Create a mint coin
		mintCoin := sdk.NewCoin(randomDenom.Denom, cosmosmath.Int(mintAmount))

		// Log the mintable amount
		fk.Logger().Info(fmt.Sprintf(
			"SimulateMsgMintAndSendTokens: maxMintableAmount: %s",
			maxMintableAmount.String(),
		))

		// Get a random recipient account
		recipientAccount, _ := simtypes.RandomAcc(r, accs)

		// Generate message for minting and sending tokens
		mintMsg := &types.MsgMintAndSendTokens{
			Signer:    banker.Address.String(),
			Token:     mintCoin,
			Recipient: recipientAccount.Address.String(),
		}

		// Generate transaction context
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txCfg,
			Cdc:             nil,
			Msg:             mintMsg,
			Context:         ctx,
			SimAccount:      banker,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}
		// Execute transaction
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
