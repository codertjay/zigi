package keeper

import (
	"context"

	"zigchain/x/factory/events"
	"zigchain/x/factory/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) DisableDenomAdmin(goCtx context.Context, msg *types.MsgDisableDenomAdmin) (*types.MsgDisableDenomAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	denomAuth, isFound := k.GetDenomAuth(
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

	// Disable the denom admin
	if err := k.DisableDenomAuth(ctx, msg.Denom); err != nil {
		return nil, err
	}

	// Remove the denom from the admin denom auth list
	k.RemoveDenomFromAdminDenomAuthList(ctx, denomAuth.BankAdmin, msg.Denom)

	// Emit event for denom admin disabled
	events.EmitDenomAuthDisabled(ctx, msg.Signer, &denomAuth)

	return &types.MsgDisableDenomAdminResponse{Denom: msg.Denom}, nil
}
