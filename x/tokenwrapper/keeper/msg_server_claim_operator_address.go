package keeper

import (
	"context"
	"fmt"

	"zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ClaimOperatorAddress claims the operator role
func (k msgServer) ClaimOperatorAddress(goCtx context.Context, msg *types.MsgClaimOperatorAddress) (*types.MsgClaimOperatorAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the proposed operator address
	proposedOperator := k.GetProposedOperatorAddress(ctx)
	if proposedOperator == "" {
		return nil, fmt.Errorf("no operator address has been proposed")
	}

	// Check if the signer is the proposed operator
	if msg.Signer != proposedOperator {
		return nil, fmt.Errorf("only the proposed operator can claim the role")
	}

	// Get the current operator address
	oldOperator := k.GetOperatorAddress(ctx)

	// Update the operator address
	k.SetOperatorAddress(ctx, proposedOperator)

	// Clear the proposed operator address
	k.SetProposedOperatorAddress(ctx, "")

	// Emit event for operator address claim
	types.EmitOperatorAddressClaimedEvent(ctx, oldOperator, proposedOperator)

	return &types.MsgClaimOperatorAddressResponse{
		Signer:          msg.Signer,
		OperatorAddress: proposedOperator,
	}, nil
}
