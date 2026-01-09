package keeper

import (
	"context"

	cosmosmath "cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"zigchain/x/factory/events"
	"zigchain/x/factory/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateDenom(
	goCtx context.Context,
	msg *types.MsgCreateDenom,
) (
	*types.MsgCreateDenomResponse,
	error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get keeper params
	params := k.GetParams(ctx)
	// Check if the sender has enough funds to pay the fee
	feeAmount := sdk.NewCoin(
		params.CreateFeeDenom,
		cosmosmath.NewInt(int64(params.CreateFeeAmount)),
	) // Adjust the "token" and amount as needed

	// Native denom names are protected uzig, or bridged stables, etc.
	if k.bankKeeper.HasSupply(ctx, msg.SubDenom) {
		return nil,
			errorsmod.Wrapf(
				types.ErrInvalidDenom,
				"Can't create subdenoms (%s) that are the same as a native denom",
				msg.SubDenom,
			)
	}

	fullDenom, err := types.GetTokenDenom(msg.Creator, msg.SubDenom)
	if err != nil {
		return nil, err
	}
	// Check if the denom already exists
	_, isFound := k.GetDenom(ctx, fullDenom)
	if isFound {
		return nil, errorsmod.Wrapf(
			types.ErrDenomExists,
			"Denom: %s already exists",
			fullDenom,
		)
	}

	// Deduct the fee from the sender's account
	sender, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	// Ensure the sender has enough balances to pay the fee
	if !k.bankKeeper.HasBalance(ctx, sender, feeAmount) {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"not enough balance to pay for denom creation fee: %s",
			feeAmount.String(),
		)
	}

	// Deduct fee from sender's account to a module account
	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(feeAmount)); err != nil {
		return nil, errorsmod.Wrap(err, "failed to collect denom creation fee")
	}

	_, exists := k.bankKeeper.GetDenomMetaData(ctx, fullDenom)
	if exists {
		return nil,
			errorsmod.Wrapf(
				types.ErrDenomExists,
				"denom: %s already exists",
				fullDenom,
			)
	}

	// Description is purposely left empty as it stored out of the chain
	// in json with icon, image and other metadata
	// advance users can bypass this use UpdateMetadata
	//is they are really keen on setting it up
	denomMetaData := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  msg.SubDenom,
		Display: fullDenom,
		URI:     msg.URI,
		URIHash: msg.URIHash,
	}

	// Clear URIHash if URI is empty
	if denomMetaData.URI == "" {
		denomMetaData.URIHash = ""
	}

	k.bankKeeper.SetDenomMetaData(ctx, denomMetaData)

	denomAuth := types.DenomAuth{
		Denom:         fullDenom,
		BankAdmin:     msg.Creator,
		MetadataAdmin: msg.Creator,
	}
	k.SetDenomAuth(ctx, denomAuth)

	// Add the denom to the admin denom auth list
	k.AddDenomToAdminDenomAuthList(ctx, denomAuth.BankAdmin, fullDenom)

	// Proceed to create the denom
	var denom = types.Denom{
		Creator:             msg.Creator,
		Denom:               fullDenom,
		Minted:              cosmosmath.ZeroUint(),
		MintingCap:          msg.MintingCap,
		CanChangeMintingCap: msg.CanChangeMintingCap,
	}

	k.SetDenom(ctx, denom)

	events.EmitDenomCreated(ctx, &denom, feeAmount)

	return &types.MsgCreateDenomResponse{
		BankAdmin:           denomAuth.BankAdmin,
		MetadataAdmin:       denomAuth.MetadataAdmin,
		Denom:               denom.Denom,
		MintingCap:          denom.MintingCap,
		CanChangeMintingCap: denom.CanChangeMintingCap,
		URI:                 denomMetaData.URI,
		URIHash:             denomMetaData.URIHash,
	}, nil
}
