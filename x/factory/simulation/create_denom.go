package simulation

import (
	"fmt"
	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"
	"strconv"

	"zigchain/zutils/constants"
	ztests "zigchain/zutils/tests"
	"zigchain/zutils/validators"

	cosmosmath "cosmossdk.io/math"

	"zigchain/x/factory/keeper"
	"zigchain/x/factory/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func SimulateMsgCreateDenom(
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
		// Randomly select an account to create the denomination
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// get number between 3 and 44
		subDenomLength := r.Intn(
			constants.MaxSubDenomLength-constants.MinSubDenomLength-1) + constants.MinSubDenomLength

		fk.Logger().Info(fmt.Sprintf("SimulateMsgCreateDenom: subDenomLength: %d", subDenomLength))
		// Generate a random from 3 to 44-character long sub denom example: btc
		subDenom := ztests.RandomSubDenom(r, subDenomLength)
		fullDenom, err := types.GetTokenDenom(simAccount.Address.String(), subDenom)
		if err != nil {
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(&types.MsgCreateDenom{}),
				fmt.Sprintf("SimulateMsgCreateDenom: Unable to get full denom from: %s", subDenom),
			), nil, err
		}
		// Generate a random max supply
		mintingCap := cosmosmath.NewUint(r.Uint64())
		// Generate random boolean
		canChangeMintingCap := r.Intn(2) == 0
		// Generate a random URI
		uri := "ipfs://" + ztests.RandomSubDenom(r, 46) + ".com/" + ztests.RandomAlphanumeric(r, 46)

		uriHash := validators.SHA256HashOfURL(uri)

		msg := &types.MsgCreateDenom{
			Creator:             simAccount.Address.String(),
			SubDenom:            subDenom,
			MintingCap:          mintingCap,
			CanChangeMintingCap: canChangeMintingCap,
			URI:                 uri,
			URIHash:             uriHash,
		}

		_, found := fk.GetDenom(ctx, fullDenom)
		comment := fmt.Sprintf("SimulateMsgCreateDenom: Denom name taken (fullDenom): %s", fullDenom)

		if found {
			fk.Logger().Error(comment)
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(msg),
				comment,
			), nil, nil
		}

		// Get params
		params := fk.GetParams(ctx)

		feeAmount := sdk.NewInt64Coin(params.CreateFeeDenom, int64(params.CreateFeeAmount))
		hasBalance := bk.HasBalance(ctx, simAccount.Address, feeAmount)

		if !hasBalance {
			fk.Logger().Info(fmt.Sprintf("SimulateMsgCreateDenom: No balance feeAmount: %s", feeAmount.String()))
			fk.Logger().Info(fmt.Sprintf("SimulateMsgCreateDenom: No balance hasBalance: %t", hasBalance))
			return simtypes.NoOpMsg(
				types.ModuleName,
				sdk.MsgTypeURL(msg),
				fmt.Sprintf(
					"SimulateMsgCreateDenom: Account %s does not have sufficient funds to pay create denom fee: %s",
					simAccount.Address,
					feeAmount,
				),
			), nil, nil
		}

		// Send coins using BankKeeper
		err = bk.SendCoinsFromAccountToModule(
			ctx,
			simAccount.Address,
			types.ModuleName,
			sdk.NewCoins(feeAmount),
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "SimulateMsgCreateDenom: unable to send coins"), nil, err
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
			CoinsSpentInMsg: sdk.NewCoins(feeAmount),
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
