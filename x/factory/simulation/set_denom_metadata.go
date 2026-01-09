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

	ztests "zigchain/zutils/tests"
	"zigchain/zutils/validators"

	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgSetDenomMetadata(
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	fk keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		randomDenom, err := RandomExistingDenom(r, ctx, fk)

		if err != nil {
			fk.Logger().Error("SimulateMsgSetDenomMetadata: Could not get random denoms", "error", err)
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgSetDenomMetadata{}),
				"SimulateMsgSetDenomMetadata: Could not get random denoms"), nil, nil
		}

		// extract subdenom as everything after the second "/"
		subDenom := randomDenom.Denom[strings.Index(randomDenom.Denom, "/")+1:]

		// Get denom auth struct from factory keeper
		denomAuth, found := fk.GetDenomAuth(ctx, randomDenom.Denom)

		commentGetAuth := fmt.Sprintf(
			"SimulateMsgSetDenomMetadata: Could not get denom auth for %s",
			randomDenom.Denom,
		)

		// Check for missing denom auth
		if !found {
			fk.Logger().Error(commentGetAuth)
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgSetDenomMetadata{}),
				commentGetAuth,
			), nil, nil
		}

		commentBankAdminEmpty := fmt.Sprintf(
			"SimulateMsgSetDenomMetadata: Bank admin for Denom: %s is empty",
			randomDenom.Denom,
		)

		// Make sure bank admin is not empty, empty means no more minting allowed
		// This is a valid state, but we cannot execute minting in this case
		// Research how to mark this as ok in the simulation stats
		if denomAuth.BankAdmin == "" {
			// Log the situation
			fk.Logger().Info(commentBankAdminEmpty)
			// Return a successful operation message without delivering a transaction
			// Create an OperationMsg directly
			opMsg := simtypes.OperationMsg{
				Route:   types.ModuleName,
				Name:    "MsgSetDenomMetadata",
				Comment: commentBankAdminEmpty,
				OK:      true,
			}

			return opMsg, nil, nil
		}

		uri := "ipfs://" + ztests.RandomSubDenom(r, 46) + ".com/" + ztests.RandomAlphanumeric(r, 46)
		// Generate a random hash using URL, but in reality this is a hash of the metadata JSON
		uriHash := validators.SHA256HashOfURL(uri)

		metaMsg := &types.MsgSetDenomMetadata{
			Signer: denomAuth.BankAdmin,
			Metadata: banktypes.Metadata{

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
			},
		}

		banker, found := FindAccount(accs, denomAuth.BankAdmin)

		// If banker not found, return no-op message
		if !found {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(metaMsg),
				"SimulateMsgSetDenomMetadata: Banker not found",
			), nil, nil
		}

		// Generate transaction context
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txCfg,
			Cdc:             nil,
			Msg:             metaMsg,
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
