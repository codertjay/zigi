package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"zigchain/x/factory/events"
	"zigchain/x/factory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdateDenomURI(goCtx context.Context, msg *types.MsgUpdateDenomURI) (*types.MsgUpdateDenomURIResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	currentDenom, isFound := k.GetDenom(
		ctx,
		msg.Denom,
	)
	if !isFound {
		return nil, errorsmod.Wrapf(
			types.ErrDenomDoesNotExist,
			"Denom: %s",
			msg.Denom,
		)
	}

	denomMetaDataCurrent, exists := k.bankKeeper.GetDenomMetaData(ctx, msg.Denom)
	if !exists {
		return nil,
			errorsmod.Wrapf(
				types.ErrDenomNotFound,
				"Denom: %s",
				msg.Denom,
			)
	}

	if currentDenom.Denom != denomMetaDataCurrent.Base {
		return nil,
			errorsmod.Wrapf(
				types.ErrInvalidDenom,
				"Denom: %s does not match base denom: %s",
				msg.Denom,
				denomMetaDataCurrent.Base,
			)
	}

	if err := k.Auth(
		ctx,
		msg.Denom,
		"metadata",
		msg.Signer,
	); err != nil {
		return nil, err
	}

	denomMetaData := banktypes.Metadata{
		// if DenomUnits are updated using bank module directly,
		// then setting it here would reset it,
		// so we will  pass it in here
		DenomUnits:  denomMetaDataCurrent.DenomUnits,
		Base:        denomMetaDataCurrent.Base,
		Name:        denomMetaDataCurrent.Name,
		Symbol:      denomMetaDataCurrent.Symbol,
		Display:     denomMetaDataCurrent.Display,
		URI:         msg.URI,
		URIHash:     msg.URIHash,
		Description: denomMetaDataCurrent.Description,
	}

	// Clear URIHash if URI is empty
	if denomMetaData.URI == "" {
		denomMetaData.URIHash = ""
	}

	k.bankKeeper.SetDenomMetaData(ctx, denomMetaData)

	events.EmitDenomURIUpdated(
		ctx,
		msg.Signer,
		msg.Denom,
		denomMetaData.URI,
		denomMetaData.URIHash,
	)

	return &types.MsgUpdateDenomURIResponse{
		Denom:   msg.Denom,
		URI:     denomMetaData.URI,
		URIHash: denomMetaData.URIHash,
	}, nil
}
