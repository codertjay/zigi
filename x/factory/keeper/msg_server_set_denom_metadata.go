package keeper

import (
	"context"

	"zigchain/x/factory/events"
	"zigchain/x/factory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SetDenomMetadata(goCtx context.Context, msg *types.MsgSetDenomMetadata) (*types.MsgSetDenomMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if a signer has metadata admin rights
	if err := k.Auth(ctx, msg.Metadata.Base, "metadata", msg.Signer); err != nil {
		return nil, err
	}

	// Clear URIHash if URI is empty
	if msg.Metadata.URI == "" {
		msg.Metadata.URIHash = ""
	}

	// set metadata on banker module
	k.Keeper.bankKeeper.SetDenomMetaData(ctx, msg.Metadata)

	// emit event about metadata change
	events.EmitDenomMetadataUpdated(ctx, msg.Signer, &msg.Metadata)

	// return response
	return &types.MsgSetDenomMetadataResponse{
		Metadata: &msg.Metadata,
	}, nil
}
