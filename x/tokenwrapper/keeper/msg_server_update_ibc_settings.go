package keeper

import (
	"context"
	"fmt"

	"zigchain/x/tokenwrapper/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UpdateIbcSettings updates the IBC settings
func (k msgServer) UpdateIbcSettings(goCtx context.Context, msg *types.MsgUpdateIbcSettings) (*types.MsgUpdateIbcSettingsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the signer is the current operator
	currentOperator := k.GetOperatorAddress(ctx)
	if msg.Signer != currentOperator {
		return nil, fmt.Errorf("only the current operator can update the IBC settings")
	}

	// Update the IBC settings
	k.SetNativeClientId(ctx, msg.NativeClientId)
	k.SetCounterpartyClientId(ctx, msg.CounterpartyClientId)
	k.SetNativePort(ctx, msg.NativePort)
	k.SetCounterpartyPort(ctx, msg.CounterpartyPort)
	k.SetNativeChannel(ctx, msg.NativeChannel)
	k.SetCounterpartyChannel(ctx, msg.CounterpartyChannel)
	k.SetDenom(ctx, msg.Denom)

	if err := k.SetDecimalDifference(ctx, msg.DecimalDifference); err != nil {
		return nil, err
	}

	// Emit event for IBC settings update
	settings := map[string]string{
		types.AttributeKeyNativeClientId:       msg.NativeClientId,
		types.AttributeKeyCounterpartyClientId: msg.CounterpartyClientId,
		types.AttributeKeyNativePort:           msg.NativePort,
		types.AttributeKeyCounterpartyPort:     msg.CounterpartyPort,
		types.AttributeKeyNativeChannel:        msg.NativeChannel,
		types.AttributeKeyCounterpartyChannel:  msg.CounterpartyChannel,
		types.AttributeKeyDenom:                msg.Denom,
		types.AttributeKeyDecimalDifference:    fmt.Sprintf("%d", msg.DecimalDifference),
	}
	types.EmitIbcSettingsUpdatedEvent(ctx, msg.Signer, settings)

	return &types.MsgUpdateIbcSettingsResponse{
		Signer:               msg.Signer,
		NativeClientId:       msg.NativeClientId,
		CounterpartyClientId: msg.CounterpartyClientId,
		NativePort:           msg.NativePort,
		CounterpartyPort:     msg.CounterpartyPort,
		NativeChannel:        msg.NativeChannel,
		CounterpartyChannel:  msg.CounterpartyChannel,
		Denom:                msg.Denom,
		DecimalDifference:    msg.DecimalDifference,
	}, nil
}
