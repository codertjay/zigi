package keeper

import (
	"context"
	"fmt"
	"zigchain/x/factory/events"

	"zigchain/x/factory/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProposeDenomAdmin proposes a new denom bank admin and metadata admin.
func (k msgServer) ProposeDenomAdmin(goCtx context.Context, msg *types.MsgProposeDenomAdmin) (*types.MsgProposeDenomAdminResponse, error) {
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

	// Check if the signer has permission to update the denom auth
	if err := k.Auth(ctx, msg.Denom, "bank", msg.Signer); err != nil {
		return nil, err
	}

	// Validate the new admin address
	if _, err := sdk.AccAddressFromBech32(msg.BankAdmin); err != nil {
		return nil, fmt.Errorf("invalid new admin address: %s", msg.BankAdmin)
	}

	// Check if the new bank admin address is different from the current one
	if currentDenomAuth.BankAdmin == msg.BankAdmin {
		return nil, errorsmod.Wrapf(
			types.ErrDuplicateBankAdminProposal,
			"current=%s, proposed=%s",
			currentDenomAuth.BankAdmin,
			msg.BankAdmin,
		)
	}

	proposedDenomAuth := types.DenomAuth{
		Denom:         msg.Denom,
		BankAdmin:     msg.BankAdmin,
		MetadataAdmin: msg.MetadataAdmin,
	}

	// Store the proposed denom admin
	k.SetProposedDenomAuth(ctx, proposedDenomAuth)

	// Emit event for denom admin proposal
	events.EmitDenomAuthProposed(ctx, msg.Signer, &proposedDenomAuth)

	return &types.MsgProposeDenomAdminResponse{
		Denom:         msg.Denom,
		BankAdmin:     msg.BankAdmin,
		MetadataAdmin: msg.MetadataAdmin,
	}, nil
}
