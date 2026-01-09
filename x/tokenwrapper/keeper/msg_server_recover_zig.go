package keeper

import (
	"context"

	"zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RecoverZig(goCtx context.Context, msg *types.MsgRecoverZig) (*types.MsgRecoverZigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate the signer
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Validate the address
	address, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	// Attempt recovery on address
	receivingAddress, lockedIbcAmount, unlockedNativeAmount, err := k.Keeper.RecoverZig(ctx, address)
	if err != nil {
		return nil, err
	}

	// Emit event
	types.EmitAddressZigRecovered(ctx, signer, receivingAddress, lockedIbcAmount, unlockedNativeAmount)

	// Response
	return &types.MsgRecoverZigResponse{
		Signer:               signer.String(),
		ReceivingAddress:     receivingAddress.String(),
		LockedIbcAmount:      lockedIbcAmount,
		UnlockedNativeAmount: unlockedNativeAmount,
	}, nil
}
