package simulation

import (
	"fmt"
	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"
	ztests "zigchain/zutils/tests"
	"zigchain/zutils/validators"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgUpdateDenomURI(
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	fk keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// Set the random denom
		randomDenom, err := RandomExistingDenom(r, ctx, fk)

		// Check for error
		if err != nil {
			fk.Logger().Error(
				"SimulateMsgUpdateDenomURI: Could not get random denoms",
				"error",
				err,
			)
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgUpdateDenomURI{}),
				"SimulateMsgUpdateDenomURI: Could not get random denoms"), nil, nil
		}

		// Extract subdenom as everything after the second "/"
		subDenom := randomDenom.Denom[strings.Index(randomDenom.Denom, "/")+1:]

		// Get denom auth struct from factory keeper
		denomAuth, found := fk.GetDenomAuth(ctx, randomDenom.Denom)

		commentGetAuth := fmt.Sprintf(
			"SimulateMsgUpdateDenomURI: Could not get denom auth for %s",
			randomDenom.Denom,
		)

		// Check for missing denom auth
		if !found {
			fk.Logger().Error(commentGetAuth)
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgUpdateDenomURI{}),
				commentGetAuth,
			), nil, nil
		}

		commentBankAdminEmpty := fmt.Sprintf(
			"SimulateMsgUpdateDenomURI: Bank admin for Denom: %s is empty",
			randomDenom.Denom,
		)

		// Make sure bank admin is not empty
		if denomAuth.BankAdmin == "" {
			fk.Logger().Info(commentBankAdminEmpty)
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgUpdateDenomURI",
				Comment: commentBankAdminEmpty,
				OK:      true,
			}
			return opMsg, nil, nil
		}

		uri := "ipfs://" + ztests.RandomSubDenom(r, 46) + ".com/" + ztests.RandomAlphanumeric(r, 46)
		// Generate a random hash using URL, but in reality this is a hash of the metadata JSON
		uriHash := validators.SHA256HashOfURL(uri)

		// First set the metadata before updating the URI
		denomMetadata := banktypes.Metadata{
			Description: ztests.RandomSubDenom(r, 255),
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    randomDenom.Denom,
					Exponent: 0,
				},
			},
			Base:    randomDenom.Denom,
			Display: randomDenom.Denom,
			Name:    randomDenom.Denom,
			Symbol:  subDenom,
			URI:     uri,
			URIHash: uriHash,
		}

		bk.SetDenomMetaData(ctx, denomMetadata)

		// Now update the URI
		newURI := "https://" + ztests.RandomSubDenomRandomLength(r) + ".com/" + ztests.RandomSubDenomRandomLength(r)
		updatedMetadata := denomMetadata
		updatedMetadata.URI = newURI

		// Set the updated metadata with the new URI
		bk.SetDenomMetaData(ctx, updatedMetadata)

		// Find the banker account
		banker, found := FindAccount(accs, denomAuth.BankAdmin)

		// If banker not found, return no-op message
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgUpdateDenomURI{}),
				"SimulateMsgUpdateDenomURI: Banker not found",
			), nil, nil
		}

		updateURIMsg := &types.MsgUpdateDenomURI{
			Signer: banker.Address.String(),
			Denom:  randomDenom.Denom,
			URI:    newURI,
		}

		// Build the transaction context
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txCfg,
			Cdc:             nil,
			Msg:             updateURIMsg,
			Context:         ctx,
			SimAccount:      banker,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}

		// Return the operation message
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
