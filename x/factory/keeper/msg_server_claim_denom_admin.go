// ClaimDenomAdmin claims the denom admin role
package keeper

import (
	"context"
	"zigchain/x/factory/events"

	"zigchain/x/factory/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) ClaimDenomAdmin(goCtx context.Context, msg *types.MsgClaimDenomAdmin) (*types.MsgClaimDenomAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	currentDenomAuth, isFound := k.GetDenomAuth(
		ctx,
		msg.Denom,
	)
	if !isFound {
		return nil, errorsmod.Wrapf(
			types.ErrDenomAuthNotFound,
			"Denom: (%s)",
			msg.Denom,
		)
	}

	// if currentDenomAuth.BankAdmin is empty return an error
	if currentDenomAuth.BankAdmin == "" {
		return nil, errorsmod.Wrapf(
			types.ErrDenomLocked,
			"denom admin was permanently disabled for denom: %s",
			msg.Denom,
		)
	}

	// Get the proposed denom admin
	proposedAdmin, found := k.GetProposedDenomAuth(ctx, msg.Denom)
	if !found {
		return nil, errorsmod.Wrapf(
			types.ErrNoAdminProposal,
			"no admin has been proposed for denom %s",
			msg.Denom,
		)
	}

	// Check if the signer is the proposed admin
	if msg.Signer != proposedAdmin.BankAdmin {
		return nil, errorsmod.Wrapf(
			types.ErrUnauthorizedAdminClaim,
			"only the proposed admin can claim the role",
		)
	}

	claimedDenomAuth := types.DenomAuth{
		Denom:         msg.Denom,
		BankAdmin:     proposedAdmin.BankAdmin,
		MetadataAdmin: proposedAdmin.MetadataAdmin,
	}

	// Update the denom admin
	k.SetDenomAuth(ctx, claimedDenomAuth)

	// Clear the proposed denom admin
	k.DeleteProposedDenomAuth(ctx, msg.Denom)

	// Remove the previous bank and metadata admin from the admin denom auth list
	k.RemoveDenomFromAdminDenomAuthList(ctx, currentDenomAuth.BankAdmin, msg.Denom)
	k.RemoveDenomFromAdminDenomAuthList(ctx, currentDenomAuth.MetadataAdmin, msg.Denom)

	// Add the new bank and metadata admin to the admin denom auth list
	k.AddDenomToAdminDenomAuthList(ctx, claimedDenomAuth.BankAdmin, msg.Denom)
	k.AddDenomToAdminDenomAuthList(ctx, claimedDenomAuth.MetadataAdmin, msg.Denom)

	// Emit event for denom admin claim
	events.EmitDenomAuthClaimed(ctx, msg.Signer, &claimedDenomAuth)

	return &types.MsgClaimDenomAdminResponse{
		Denom: msg.Denom,
	}, nil
}
