package simulation

import (
	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand"
	"strconv"

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

func SimulateMsgProposeDenomAuth(
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	fk keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var (
			simAccount   = simtypes.Account{}
			denomAuth    = types.DenomAuth{}
			msg          = &types.MsgProposeDenomAdmin{}
			allDenomAuth = fk.GetAllDenomAuth(ctx)
			found        = false
		)
		for _, obj := range allDenomAuth {
			if obj.BankAdmin == "" {
				fk.Logger().Info("SimulateMsgUpdateDenomAuth: Denom", obj.Denom, "has no bank admin")
				continue
			}
			simAccount, found = FindAccount(accs, obj.BankAdmin)
			if found {
				denomAuth = obj
				break
			}
		}
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "SimulateMsgUpdateDenomAuth: denomAuth admin not found"), nil, nil
		}
		msg.Signer = simAccount.Address.String()
		msg.Denom = denomAuth.Denom

		// Pick a random account for the new BankAdmin that's different from the current one
		var newBankAdminAccount simtypes.Account
		maxAttempts := 10
		foundNewAdmin := false
		for i := 0; i < maxAttempts; i++ {
			newBankAdminAccount, _ = simtypes.RandomAcc(r, accs)
			if newBankAdminAccount.Address.String() != denomAuth.BankAdmin {
				foundNewAdmin = true
				break
			}
		}
		if !foundNewAdmin {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "SimulateMsgProposeDenomAuth: could not find different admin account"), nil, nil
		}
		msg.BankAdmin = newBankAdminAccount.Address.String()

		// Optionally set MetadataAdmin (can be empty, but let's use a random account sometimes)
		if r.Intn(2) == 0 {
			metaAdminAccount, _ := simtypes.RandomAcc(r, accs)
			msg.MetadataAdmin = metaAdminAccount.Address.String()
		} else {
			msg.MetadataAdmin = ""
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
			CoinsSpentInMsg: sdk.NewCoins(),
			AccountKeeper:   ak,
			Bankkeeper:      bk,
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
