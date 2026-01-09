package simulation

import (

	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"

	"fmt"
	"math"

	cosmosmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgUpdateDenomMintingCap(
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	fk keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand,
		app *baseapp.BaseApp,
		ctx sdk.Context,
		accs []simtypes.Account,
		chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// Get random denom
		randomDenom, err := RandomExistingDenom(r, ctx, fk)

		commentGetDenom := fmt.Sprintf(
			"SimulateMsgUpdateDenomMintingCap: Could not get all denoms: %s",
			err,
		)

		if err != nil {
			fk.Logger().Error(commentGetDenom)
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgUpdateDenomMintingCap{}),
				commentGetDenom,
			), nil, nil
		}

		// Log the random denom
		fk.Logger().Info(fmt.Sprintf(
			"SimulateMsgUpdateDenomMintingCap: randomDenom: %s can change supply: %t",
			randomDenom.Denom,
			randomDenom.CanChangeMintingCap,
		))

		if !randomDenom.CanChangeMintingCap {
			fk.Logger().Info(fmt.Sprintf(
				"SimulateMsgUpdateDenomMintingCap: Denom: %s cannot change max supply",
				randomDenom.Denom,
			))
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgUpdateDenomMintingCap",
				Comment: "SimulateMsgUpdateDenomMintingCap: can not change max supply",
				OK:      true,
			}

			return opMsg, nil, nil
		}

		// Get denom auth from keeper
		denomAuth, found := fk.GetDenomAuth(ctx, randomDenom.Denom)

		commentGetAuth := fmt.Sprintf(
			"SimulateMsgUpdateDenomMintingCap: Could not get denom auth: %s",
			err,
		)

		if !found {
			fk.Logger().Error(commentGetAuth)
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgUpdateDenomMintingCap{}),
				commentGetAuth,
			), nil, nil
		}

		bankAdmin := denomAuth.BankAdmin

		if bankAdmin == "" {
			fk.Logger().Debug(fmt.Sprintf(
				"SimulateMsgUpdateDenomMintingCap: Bank admin for Denom: %s is empty",
				randomDenom.Denom,
			))
			// Mark operation as OK - since this is a normal condition - bank admin can be nothing
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgUpdateDenomMintingCap",
				Comment: "SimulateMsgUpdateDenomMintingCap: Can not change max supply",
				OK:      true,
			}

			return opMsg, nil, nil
		}

		// Check if there is an error
		fk.Logger().Debug(fmt.Sprintf(
			"SimulateMsgUpdateDenomMintingCap: Bank admin for denom: %s is: %s",
			randomDenom.Denom,
			bankAdmin,
		))

		// Get random max supply between current supply and max uint
		newMintingCap := RandUIntBetween(r, randomDenom.Minted, cosmosmath.NewUint(math.MaxUint64))

		updateMintingCapMsg := &types.MsgUpdateDenomMintingCap{
			Signer:              bankAdmin,
			Denom:               randomDenom.Denom,
			MintingCap:          newMintingCap,
			CanChangeMintingCap: r.Intn(2) == 0,
		}

		simAccount, found := FindAccount(accs, bankAdmin)

		commentBankAdmin := fmt.Sprintf(
			"SimulateMsgUpdateDenomMintingCap: Bank Admin not found: %s",
			bankAdmin,
		)

		if !found {
			fk.Logger().Error(commentBankAdmin)
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(updateMintingCapMsg),
				commentBankAdmin), nil, nil
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txCfg,
			Cdc:             nil,
			Msg:             updateMintingCapMsg,
			Context:         ctx,
			SimAccount:      simAccount,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
