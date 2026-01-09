package keeper

import (
	"context"

	"zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) AddPauserAddress(goCtx context.Context, msg *types.MsgAddPauserAddress) (*types.MsgAddPauserAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the signer is the current operator
	currentOperator := k.GetOperatorAddress(ctx)
	if msg.Signer != currentOperator {
		return nil, sdkerrors.ErrUnauthorized.Wrap("only the operator can add pauser addresses")
	}

	// Validate the new pauser address
	newPauser, err := sdk.AccAddressFromBech32(msg.NewPauser)
	if err != nil {
		return nil, err
	}

	// Add the new pauser address
	k.Keeper.AddPauserAddress(ctx, newPauser.String())

	// Emit event
	types.EmitPauserAddressAddedEvent(ctx, msg.NewPauser)

	return &types.MsgAddPauserAddressResponse{
		Signer:          msg.Signer,
		PauserAddresses: k.GetPauserAddresses(ctx),
	}, nil
}
