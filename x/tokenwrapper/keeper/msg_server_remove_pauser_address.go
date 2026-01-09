package keeper

import (
	"context"

	"zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) RemovePauserAddress(goCtx context.Context, msg *types.MsgRemovePauserAddress) (*types.MsgRemovePauserAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the signer is the current operator
	currentOperator := k.GetOperatorAddress(ctx)
	if msg.Signer != currentOperator {
		return nil, sdkerrors.ErrUnauthorized.Wrap("only the operator can remove pauser addresses")
	}

	// Validate the new pauser address
	pauser, err := sdk.AccAddressFromBech32(msg.Pauser)
	if err != nil {
		return nil, err
	}

	// Remove the pauser address
	k.Keeper.RemovePauserAddress(ctx, pauser.String())

	// Emit event
	types.EmitPauserAddressRemovedEvent(ctx, msg.Pauser)

	return &types.MsgRemovePauserAddressResponse{
		Signer:          msg.Signer,
		PauserAddresses: k.GetPauserAddresses(ctx),
	}, nil
}
