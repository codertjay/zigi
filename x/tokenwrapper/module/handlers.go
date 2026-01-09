package tokenwrapper

import (
	"fmt"

	"zigchain/zutils/constants"

	"zigchain/x/tokenwrapper/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
)

// handleUnescrowToken handles the unescrow process when a token is sent
func (w ICS4Wrapper) handleUnescrowToken(ctx sdk.Context, escrowAddress, receiver sdk.AccAddress, token sdk.Coin) error {
	if err := w.bankKeeper.SendCoins(ctx, escrowAddress, receiver, sdk.NewCoins(token)); err != nil {
		return errorsmod.Wrap(err, "unable to unescrow tokens")
	}

	// track the total amount in escrow keyed by denomination to allow for efficient iteration
	currentTotalEscrow := w.transferKeeper.GetTotalEscrowForDenom(ctx, token.GetDenom())
	newTotalEscrow := currentTotalEscrow.Sub(token)
	w.transferKeeper.SetTotalEscrowForDenom(ctx, newTotalEscrow)

	return nil
}

// handleRefund handles the refund process when an IBC transfer fails
func (im IBCModule) handleRefund(ctx sdk.Context, sender sdk.AccAddress, amount sdkmath.Int, denom string) error {
	im.keeper.Logger().Info(fmt.Sprintf("Handling refund for sender: %s, amount: %s, denom: %s", sender.String(), amount.String(), denom))

	// Get the IBC denom for the refunded tokens
	ibcDenom := transfertypes.ExtractDenomFromPath(denom).IBCDenom()

	// Create IBC coins with the refunded amount
	ibcCoins := sdk.NewCoins(sdk.NewCoin(ibcDenom, amount))

	// Convert the amount back to uzig decimals (from 18 to 6 decimals)
	convertedAmount, err := im.keeper.ScaleDownTokenPrecision(ctx, amount)
	if err != nil {
		return err
	}

	// Check if account has enough balance to lock the IBC tokens
	if err := im.keeper.CheckAccountBalance(ctx, sender, ibcCoins); err != nil {
		return fmt.Errorf("failed to check account balance: %w", err)
	}

	// Create native coins with the converted amount
	nativeCoins := sdk.NewCoins(sdk.NewCoin(constants.BondDenom, convertedAmount))

	// Check if module has enough balance to unlock the uzig tokens
	if err := im.keeper.CheckModuleBalance(ctx, nativeCoins); err != nil {
		return fmt.Errorf("failed to check module balance: %w", err)
	}

	// Lock the refunded IBC vouchers in the module
	if err := im.keeper.LockTokens(ctx, sender, ibcCoins); err != nil {
		return fmt.Errorf("failed to lock refunded IBC tokens: %w", err)
	}

	im.keeper.Logger().Info(fmt.Sprintf("Locked refunded IBC tokens: %s for sender: %s", ibcCoins.String(), sender.String()))

	// Unlock uzig tokens to the user
	if err := im.keeper.UnlockTokens(ctx, sender, nativeCoins); err != nil {
		// If unlocking fails, we need to unlock the previously locked IBC tokens
		if unlockErr := im.keeper.UnlockTokens(ctx, sender, ibcCoins); unlockErr != nil {
			im.keeper.Logger().Error(fmt.Sprintf("failed to unlock previously locked IBC tokens: %v", unlockErr))
		}
		return fmt.Errorf("failed to unlock native tokens: %w", err)
	}

	im.keeper.Logger().Info(fmt.Sprintf("Unlocked native tokens: %s for sender: %s", nativeCoins.String(), sender.String()))

	// Emit refund event
	types.EmitTokenWrapperRefundEvent(
		ctx,
		sender.String(),
		amount,
		ibcDenom,
		convertedAmount,
		constants.BondDenom,
	)

	return nil
}
