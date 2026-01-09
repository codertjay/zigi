package keeper

import (
	"context"
	"zigchain/x/factory/events"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/x/factory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UpdateDenomMetadataAuth updates the metadata admin of a specific denom, needed in case bank admin is disabled.
func (k msgServer) UpdateDenomMetadataAuth(goCtx context.Context, msg *types.MsgUpdateDenomMetadataAuth) (*types.MsgUpdateDenomMetadataAuthResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	currentDenomAuth, isFound := k.GetDenomAuth(
		ctx,
		msg.Denom,
	)
	if !isFound {
		return nil, errorsmod.Wrapf(
			types.ErrDenomAuthNotFound,
			"denom (%s) not found",
			msg.Denom,
		)
	}

	// Checks if the msg admin is the same as the current admin
	if msg.Signer != currentDenomAuth.MetadataAdmin && msg.Signer != currentDenomAuth.BankAdmin {
		if currentDenomAuth.BankAdmin == "" {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrUnauthorized,
				"Incorrect admin for denom (%s), only metadata admin (%s) can update the denom admins",
				msg.Denom,
				currentDenomAuth.MetadataAdmin,
			)
		}
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrUnauthorized,
			"incorrect admin for denom (%s), only meta admin (%s) or bank admin (%s) can update the denom admins",
			msg.Denom,
			currentDenomAuth.MetadataAdmin,
			currentDenomAuth.BankAdmin,
		)
	}

	var denomAuth = types.DenomAuth{
		Denom:         msg.Denom,
		BankAdmin:     currentDenomAuth.BankAdmin,
		MetadataAdmin: msg.MetadataAdmin,
	}

	k.SetDenomAuth(ctx, denomAuth)

	// Remove the previous metadata admin from the admin denom auth list
	k.RemoveDenomFromAdminDenomAuthList(ctx, currentDenomAuth.MetadataAdmin, msg.Denom)

	// Add the new metadata admin to the admin denom auth list
	k.AddDenomToAdminDenomAuthList(ctx, denomAuth.MetadataAdmin, msg.Denom)

	// Make sure the bank admin is also in the admin denom auth list
	k.AddDenomToAdminDenomAuthList(ctx, denomAuth.BankAdmin, msg.Denom)

	events.EmitDenomAuthUpdated(ctx, msg.Signer, &denomAuth)

	return &types.MsgUpdateDenomMetadataAuthResponse{
		Denom:         msg.Denom,
		MetadataAdmin: msg.MetadataAdmin,
	}, nil
}
