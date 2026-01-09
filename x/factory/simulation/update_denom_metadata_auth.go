package simulation

import (
	"fmt"
	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgUpdateDenomMetadataAuth(
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	fk keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// select a random denom
		randomDenom, err := RandomExistingDenom(r, ctx, fk)

		if err != nil {
			fk.Logger().Error("SimulateMsgUpdateDenomMetadataAuth: Could not get random denoms", "error", err)
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgUpdateDenomMetadataAuth{}),
				"SimulateMsgUpdateDenomMetadataAuth: Could not get random denoms"), nil, nil
		}

		commentDenomAuth := fmt.Sprintf(
			"SimulateMsgUpdateDenomMetadataAuth: Could not get denom auth for denom: %s",
			randomDenom.Denom,
		)

		// check for missing denom auth
		denomAuth, found := fk.GetDenomAuth(ctx, randomDenom.Denom)
		if !found {
			fk.Logger().Error(commentDenomAuth)
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgUpdateDenomMetadataAuth{}),
				commentDenomAuth,
			), nil, nil
		}

		bankAdmin := denomAuth.BankAdmin

		commentBankAdmin := fmt.Sprintf(
			"SimulateMsgUpdateDenomMetadataAuth: Bank admin for Denom: %s is empty",
			randomDenom.Denom,
		)

		if bankAdmin == "" {
			// Log the situation
			fk.Logger().Info(commentBankAdmin)
			// Return a successful operation message without delivering a transaction
			// Create an OperationMsg directly
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgUpdateDenomMetadataAuth",
				Comment: commentBankAdmin,
				OK:      true,
			}

			return opMsg, nil, nil
		}

		// pick a random account from the list of sim accounts
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// create the MsgUpdateDenomMetadataAuth message
		metaAuthMsg := &types.MsgUpdateDenomMetadataAuth{
			Signer:        bankAdmin,
			Denom:         randomDenom.Denom,
			MetadataAdmin: simAccount.Address.String(),
		}

		banker, found := FindAccount(accs, bankAdmin)

		commentBankerAcc := fmt.Sprintf(
			"SimulateMsgUpdateDenomMetadataAuth: Could not get banker account: %s",
			bankAdmin,
		)

		if !found {
			fk.Logger().Error(commentBankerAcc)
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgUpdateDenomMetadataAuth{}),
				commentBankerAcc,
			), nil, nil
		}

		// build the transaction context
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txCfg,
			Cdc:             nil,
			Msg:             metaAuthMsg,
			Context:         ctx,
			SimAccount:      banker,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}

		// execute transaction
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
