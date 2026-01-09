package keeper

import (
	"context"

	"zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// EnableTokenWrapper enables the token wrapper functionality
func (k msgServer) EnableTokenWrapper(goCtx context.Context, msg *types.MsgEnableTokenWrapper) (*types.MsgEnableTokenWrapperResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the signer is the current operator
	currentOperator := k.GetOperatorAddress(ctx)
	if msg.Signer != currentOperator {
		return nil, sdkerrors.ErrUnauthorized.Wrap("only the operator can enable the token wrapper")
	}

	// Enable the token wrapper
	k.SetEnabled(ctx, true)

	// Emit event for enabling token wrapper
	types.EmitTokenWrapperEnabledEvent(ctx, msg.Signer)

	return &types.MsgEnableTokenWrapperResponse{
		Signer:  msg.Signer,
		Enabled: true,
	}, nil
}
