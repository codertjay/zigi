package events

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"zigchain/x/factory/types"
)

func EmitDenomCreated(ctx sdk.Context, denom *types.Denom, feeAmount sdk.Coin) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newDenomCreateEvent(denom, feeAmount),
	})
}

func newDenomCreateEvent(denom *types.Denom, feeAmount sdk.Coin) sdk.Event {

	return sdk.NewEvent(
		types.EventDenomCreated,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(types.AttributeKeyDenom, denom.Denom),
		sdk.NewAttribute(types.AttributeKeyMintingCap, denom.MintingCap.String()),
		sdk.NewAttribute(types.AttributeKeyTotalMinted, denom.Minted.String()),
		sdk.NewAttribute(types.AttributeKeyCanChangeMintingCap, strconv.FormatBool(denom.CanChangeMintingCap)),
		sdk.NewAttribute(types.AttributeKeyCreator, denom.Creator),
		sdk.NewAttribute(types.AttributeKeyFee, feeAmount.String()),
	)
}

func EmitDenomMintedAndSent(
	ctx sdk.Context,
	denom *types.Denom,
	mintedToken sdk.Coin,
	recipient string,
	totalMinted sdk.Coin,
	totalSupply sdk.Coin,

) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newDenomMintedAndSentEvent(
			denom,
			mintedToken,
			recipient,
			totalMinted,
			totalSupply,
		),
	})

}

func newDenomMintedAndSentEvent(
	denom *types.Denom,
	mintedToken sdk.Coin,
	recipient string,
	totalMinted sdk.Coin,
	totalSupply sdk.Coin,
) sdk.Event {
	return sdk.NewEvent(
		types.EventDenomMintedAndSent,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		// Redundant but friendly for somebody filtering the events by denom name
		sdk.NewAttribute(types.AttributeKeyDenom, denom.Denom),
		sdk.NewAttribute(types.AttributeKeyDenomMintedAndSent, mintedToken.String()),
		sdk.NewAttribute(types.AttributeKeyRecipient, recipient),
		sdk.NewAttribute(types.AttributeKeyTotalMinted, totalMinted.String()),
		sdk.NewAttribute(types.AttributeKeyTotalSupply, totalSupply.String()),
	)
}

func EmitMintingCapChanged(ctx sdk.Context, signer string, denom *types.Denom) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newMintingCapChangedEvent(signer, denom),
	})
}

func newMintingCapChangedEvent(signer string, denom *types.Denom) sdk.Event {
	return sdk.NewEvent(
		types.EventDenomMintingCapChanged,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(types.AttributeKeySigner, signer),
		sdk.NewAttribute(types.AttributeKeyDenom, denom.Denom),
		sdk.NewAttribute(types.AttributeKeyMintingCap, denom.MintingCap.String()),
		sdk.NewAttribute(types.AttributeKeyCanChangeMintingCap, strconv.FormatBool(denom.CanChangeMintingCap)),
	)
}

func EmitDenomURIUpdated(ctx sdk.Context, signer string, denom string, uri string, uriHash string) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newDenomURIUpdatedEvent(signer, denom, uri, uriHash),
	})
}

func newDenomURIUpdatedEvent(signer string, denom string, uri string, uriHash string) sdk.Event {
	return sdk.NewEvent(
		types.EventDenomURIUpdated,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(types.AttributeKeySigner, signer),
		sdk.NewAttribute(types.AttributeKeyDenom, denom),
		sdk.NewAttribute(types.AttributeKeyDenomURI, uri),
		sdk.NewAttribute(types.AttributeKeyDenomURIHash, uriHash),
	)
}

func EmitDenomMetadataUpdated(ctx sdk.Context, signer string, metadata *banktypes.Metadata) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newDenomMetadataUpdatedEvent(signer, metadata),
	})
}

func newDenomMetadataUpdatedEvent(signer string, metadata *banktypes.Metadata) sdk.Event {
	return sdk.NewEvent(
		types.EventDenomMetadataUpdated,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(types.AttributeKeySigner, signer),
		sdk.NewAttribute(types.AttributeKeyDenomMetadata, metadata.String()),
	)
}

func EmitDenomBurned(ctx sdk.Context, signer string, burned sdk.Coin) {

	ctx.EventManager().EmitEvents(sdk.Events{
		newDenomBurnedEvent(signer, burned.Denom, burned.Amount.String()),
	})
}

func newDenomBurnedEvent(signer string, denom string, amount string) sdk.Event {
	return sdk.NewEvent(
		types.EventDenomBurned,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(types.AttributeKeySigner, signer),
		sdk.NewAttribute(types.AttributeKeyDenom, denom),
		sdk.NewAttribute(types.AttributeKeyAmount, amount),
	)

}

func EmitDenomAuthUpdated(ctx sdk.Context, signer string, denomAuth *types.DenomAuth) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newDenomAuthUpdatedEvent(signer, denomAuth),
	})
}

func newDenomAuthUpdatedEvent(signer string, denomAuth *types.DenomAuth) sdk.Event {
	return sdk.NewEvent(
		types.EventDenomAuthUpdated,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(types.AttributeKeySigner, signer),
		sdk.NewAttribute(types.AttributeKeyDenom, denomAuth.Denom),
		sdk.NewAttribute(types.AttributeKeyBankAdmin, denomAuth.BankAdmin),
		sdk.NewAttribute(types.AttributeKeyMetadataAdmin, denomAuth.MetadataAdmin),
	)
}

func EmitDenomAuthProposed(ctx sdk.Context, signer string, denomAuth *types.DenomAuth) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newDenomAuthProposedEvent(signer, denomAuth),
	})
}

func newDenomAuthProposedEvent(signer string, denomAuth *types.DenomAuth) sdk.Event {
	return sdk.NewEvent(
		types.EventDenomAuthProposed,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(types.AttributeKeySigner, signer),
		sdk.NewAttribute(types.AttributeKeyDenom, denomAuth.Denom),
		sdk.NewAttribute(types.AttributeKeyBankAdmin, denomAuth.BankAdmin),
		sdk.NewAttribute(types.AttributeKeyMetadataAdmin, denomAuth.MetadataAdmin),
	)
}

func EmitDenomAuthClaimed(ctx sdk.Context, signer string, denomAuth *types.DenomAuth) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newDenomAuthClaimedEvent(signer, denomAuth),
	})
}

func newDenomAuthClaimedEvent(signer string, denomAuth *types.DenomAuth) sdk.Event {
	return sdk.NewEvent(
		types.EventDenomAuthClaimed,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(types.AttributeKeySigner, signer),
		sdk.NewAttribute(types.AttributeKeyDenom, denomAuth.Denom),
		sdk.NewAttribute(types.AttributeKeyBankAdmin, denomAuth.BankAdmin),
		sdk.NewAttribute(types.AttributeKeyMetadataAdmin, denomAuth.MetadataAdmin),
	)
}

func EmitDenomAuthDisabled(ctx sdk.Context, signer string, denomAuth *types.DenomAuth) {
	ctx.EventManager().EmitEvents(sdk.Events{
		newDenomAuthDisabledEvent(signer, denomAuth),
	})
}

func newDenomAuthDisabledEvent(signer string, denomAuth *types.DenomAuth) sdk.Event {
	return sdk.NewEvent(
		types.EventDenomAuthDisabled,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		sdk.NewAttribute(types.AttributeKeySigner, signer),
		sdk.NewAttribute(types.AttributeKeyDenom, denomAuth.Denom),
		sdk.NewAttribute(types.AttributeKeyBankAdmin, denomAuth.BankAdmin),
		sdk.NewAttribute(types.AttributeKeyMetadataAdmin, denomAuth.MetadataAdmin),
	)
}
