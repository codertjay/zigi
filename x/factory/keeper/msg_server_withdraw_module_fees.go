package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/x/factory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) WithdrawModuleFees(goCtx context.Context, msg *types.MsgWithdrawModuleFees) (*types.MsgWithdrawModuleFeesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)

	if params.Beneficiary == "" {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrUnauthorized,
			"beneficiary address is not set, cannot withdraw",
		)
	}

	if params.Beneficiary != msg.Signer {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrUnauthorized,
			"signer: %s is not the beneficiary: %s, only the beneficiary can withdraw",
			msg.Signer,
			params.Beneficiary,
		)
	}

	// Retrieve the module account
	moduleAccount := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAccount == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "module account not found")
	}

	// Get the module account's entire balance
	moduleBalance := k.bankKeeper.GetAllBalances(ctx, moduleAccount.GetAddress())
	if moduleBalance.IsZero() {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"module %s account has no balance to withdraw",
			types.ModuleName,
		)
	}

	// Determine the receiver address
	var receiverAddress sdk.AccAddress
	if msg.Receiver == "" {
		// If the receiver is empty, use the signer address
		// This is checked in validators but better to be safe
		signerAddress, err := sdk.AccAddressFromBech32(msg.Signer)
		if err != nil {
			return nil,
				errorsmod.Wrapf(
					sdkerrors.ErrInvalidAddress,
					"invalid signer address: %s (%s)",
					msg.Signer,
					err,
				)
		}
		receiverAddress = signerAddress
	} else {
		// Use the provided receiver address
		var err error
		receiverAddress, err = sdk.AccAddressFromBech32(msg.Receiver)
		if err != nil {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrInvalidAddress,
				"invalid receiver address: %s (%s)",
				msg.Receiver,
				err,
			)
		}
	}

	// Transfer logic: Take only the first 100 denominations
	// In case amount is expected to be large, consider paginating the transfer
	//maxDenoms := 100
	//denomCount := len(moduleBalance)
	//transferBalance := sdk.NewCoins()
	//
	//if denomCount > maxDenoms {
	//	// Select the first 100 denominations
	//	transferBalance = moduleBalance[:maxDenoms]
	//} else {
	//	// Transfer all denominations if less than or equal to 100
	//	transferBalance = moduleBalance
	//}
	//
	//// Transfer the selected balance to the receiver
	//if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiverAddress, transferBalance); err != nil {
	//	return nil, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "failed to transfer funds: %v", err)
	//}

	// Transfer all funds from the module account to the receiver
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiverAddress, moduleBalance); err != nil {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"failed to transfer funds: %v",
			err,
		)
	}

	// Log the withdrawal event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdrawModuleFees,
			sdk.NewAttribute(types.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAmount, moduleBalance.String()),
			sdk.NewAttribute(types.AttributeKeyReceiver, receiverAddress.String()),
		),
	})

	return &types.MsgWithdrawModuleFeesResponse{
		Signer:   msg.Signer,
		Receiver: receiverAddress.String(),
		Amounts:  moduleBalance,
	}, nil
}
