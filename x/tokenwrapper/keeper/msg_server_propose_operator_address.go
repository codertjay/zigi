package keeper

import (
	"context"
	"fmt"

	"zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProposeOperatorAddress proposes a new operator address
func (k msgServer) ProposeOperatorAddress(goCtx context.Context, msg *types.MsgProposeOperatorAddress) (*types.MsgProposeOperatorAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the signer is the current operator
	currentOperator := k.GetOperatorAddress(ctx)
	if msg.Signer != currentOperator {
		return nil, fmt.Errorf("only the current operator can propose a new operator address")
	}

	// Validate the new operator address
	newOperator, err := sdk.AccAddressFromBech32(msg.NewOperator)
	if err != nil {
		return nil, fmt.Errorf("invalid new operator address: %s", msg.NewOperator)
	}

	// Check if the new operator address is the same as the current one
	if newOperator.String() == currentOperator {
		return nil, fmt.Errorf("cannot propose the same operator address")
	}

	// Store the proposed operator address
	k.SetProposedOperatorAddress(ctx, newOperator.String())

	// Emit event for the operator address proposal
	types.EmitOperatorAddressProposedEvent(ctx, currentOperator, newOperator.String())

	return &types.MsgProposeOperatorAddressResponse{
		Signer:                  msg.Signer,
		ProposedOperatorAddress: newOperator.String(),
	}, nil
}
