package simulation

import (
	"math/rand"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"zigchain/x/tokenwrapper/keeper"
	"zigchain/x/tokenwrapper/types"
	"zigchain/zutils/constants"
)

func SimulateMsgRecoverZig(
	txCfg client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Check if module is enabled - if not, skip this operation
		if !k.IsEnabled(ctx) {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgRecoverZig{}), "module is disabled"), nil, nil
		}

		// Get module denom and IBC settings - if not configured, skip this operation
		moduleDenom := k.GetDenom(ctx)
		if moduleDenom == "" {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgRecoverZig{}), "module denom not configured"), nil, nil
		}

		// Get IBC receive denom
		recvDenom := k.GetIBCRecvDenom(ctx, moduleDenom)

		// Get a random account for the signer
		simAccount, _ := simtypes.RandomAcc(r, accs)

		// Get a random account to recover from (prefer using simulation accounts which may have tokens)
		// This increases the chance of having IBC vouchers available
		recoverAccount, _ := simtypes.RandomAcc(r, accs)
		recoverAddr := recoverAccount.Address

		// Generate a random IBC amount that will produce a positive native amount after scaling
		// The amount should be at least the conversion factor to ensure positive native tokens
		conversionFactor := k.GetDecimalConversionFactor(ctx)
		// Generate IBC amount between 1x and 10x the conversion factor to ensure positive native amount
		minIBCAmount := conversionFactor
		maxIBCAmount := conversionFactor.Mul(sdkmath.NewInt(10))
		ibcAmount := sdkmath.NewInt(int64(simtypes.RandIntBetween(r, int(minIBCAmount.Int64()), int(maxIBCAmount.Int64()))))

		// Calculate the converted native amount
		convertedAmount, err := k.ScaleDownTokenPrecision(ctx, ibcAmount)
		if err != nil || convertedAmount.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgRecoverZig{}), "failed to calculate converted amount"), nil, nil
		}

		// Try to find an account that has IBC vouchers to transfer from
		// If found, transfer IBC vouchers to the recover address
		ibcCoins := sdk.NewCoins(sdk.NewCoin(recvDenom, ibcAmount))
		ibcVoucherSource := recoverAddr // Default to recover address itself
		foundSource := false

		// Check if recover address already has enough IBC vouchers
		recoverBalance := bk.GetBalance(ctx, recoverAddr, recvDenom)
		if recoverBalance.Amount.GTE(ibcAmount) {
			foundSource = true
		} else {
			// Try to find another account with IBC vouchers
			for _, acc := range accs {
				balance := bk.GetBalance(ctx, acc.Address, recvDenom)
				if balance.Amount.GTE(ibcAmount) {
					ibcVoucherSource = acc.Address
					foundSource = true
					break
				}
			}
		}

		// If we found a source, transfer IBC vouchers to recover address
		if foundSource && ibcVoucherSource.String() != recoverAddr.String() {
			if err := bk.SendCoins(ctx, ibcVoucherSource, recoverAddr, ibcCoins); err != nil {
				// If transfer fails, continue anyway - the address might already have vouchers
			}
		} else if !foundSource {
			// If no source found, we can't set up IBC vouchers, but we'll still try the operation
			// This will exercise the failure path, which is also valuable for simulation
		}

		// Ensure module wallet has enough native ZIG tokens for the conversion
		// We check the balance and if insufficient, we skip (can't mint through interface)
		moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
		moduleBalance := bk.GetBalance(ctx, moduleAddr, constants.BondDenom)
		if moduleBalance.Amount.LT(convertedAmount) {
			// Try to find an account with enough native tokens to fund the module
			foundNativeSource := false
			for _, acc := range accs {
				balance := bk.GetBalance(ctx, acc.Address, constants.BondDenom)
				if balance.Amount.GTE(convertedAmount) {
					nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, convertedAmount))
					// Transfer to module via SendCoinsFromAccountToModule
					if err := bk.SendCoinsFromAccountToModule(ctx, acc.Address, types.ModuleName, nativeCoins); err == nil {
						foundNativeSource = true
						break
					}
				}
			}
			if !foundNativeSource {
				// If we can't fund the module, skip this operation
				return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(&types.MsgRecoverZig{}), "insufficient module wallet balance and no funding source"), nil, nil
			}
		}

		// Create the message
		msg := &types.MsgRecoverZig{
			Signer:  simAccount.Address.String(),
			Address: recoverAddr.String(),
		}

		// Create a tx
		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         txCfg,
			Cdc:           nil,
			Msg:           msg,
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
