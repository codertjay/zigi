package keeper

import (
	"context"

	"zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DisableTokenWrapper disables the token wrapper functionality
func (k msgServer) DisableTokenWrapper(goCtx context.Context, msg *types.MsgDisableTokenWrapper) (*types.MsgDisableTokenWrapperResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the signer is either a pauser or the operator
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	currentOperator := k.GetOperatorAddress(ctx)
	if msg.Signer != currentOperator && !k.IsPauserAddress(ctx, signer.String()) {
		return nil, sdkerrors.ErrUnauthorized.Wrap("signer is neither a pauser nor the operator")
	}

	// Disable the token wrapper
	k.SetEnabled(ctx, false)

	// Emit event for disabling token wrapper
	types.EmitTokenWrapperDisabledEvent(ctx, msg.Signer)

	return &types.MsgDisableTokenWrapperResponse{
		Signer:  msg.Signer,
		Enabled: false,
	}, nil
}
