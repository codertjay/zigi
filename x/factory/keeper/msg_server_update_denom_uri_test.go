package keeper_test

import (
	"fmt"
	"testing"

	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/factory/keeper"
	"zigchain/x/factory/testutil"
	"zigchain/x/factory/types"
)

// Positive test cases

func TestMsgServer_UpdateDenomURI_Positive(t *testing.T) {
	// Test case: update the URI of a denom

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name - the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	denomMetaData := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  subDenom,
		Display: fullDenom,
		URI:     uri,
		URIHash: uriHash,
	}

	// create a message to update uri of the denom
	newUri := "ipfs://newuri"
	newHash := "newsha256hash"

	newDenomMetaData := banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  subDenom,
		Display: fullDenom,
		URI:     newUri,
		URIHash: newHash,
	}

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// ensure this order of calls
	gomock.InOrder(
		// code will first check if subDenom is matching one of the code tokens (e.g. "uzig")
		bankKeeper.
			EXPECT().
			HasSupply(gomock.Any(), subDenom).
			Return(false).
			Times(1),

		// code will check if the signer has the required balances to pay the fee
		// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
		bankKeeper.
			EXPECT().
			HasBalance(gomock.Any(), signer, fee).
			Return(true).
			Times(1),

		// code will deduct the fee from the signer's account
		// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
		// it will return nil if the operation is successful - no error
		bankKeeper.
			EXPECT().
			SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
			Return(nil).
			Times(1),

		// check if denom already has metadata
		// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
		bankKeeper.
			EXPECT().
			GetDenomMetaData(gomock.Any(), fullDenom).
			Return(banktypes.Metadata{}, false).
			Times(1),

		// SetDenomMetaData(context.Context, banktypes.Metadata)
		bankKeeper.
			EXPECT().
			SetDenomMetaData(gomock.Any(), denomMetaData).
			Return().
			Times(1),

		bankKeeper.
			EXPECT().
			GetDenomMetaData(gomock.Any(), fullDenom).
			Return(banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    fullDenom,
					Exponent: 0,
				}},
				Base:    fullDenom,
				Name:    fullDenom,
				Symbol:  subDenom,
				Display: fullDenom,
				URI:     uri,
				URIHash: uriHash,
			}, true).
			Times(1),

		// new data is being set
		bankKeeper.
			EXPECT().
			SetDenomMetaData(gomock.Any(), newDenomMetaData).
			Return().
			Times(1),
	)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               fullDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	updateMsg := &types.MsgUpdateDenomURI{
		Signer:  creator,
		Denom:   fullDenom,
		URI:     newUri,
		URIHash: newHash,
	}

	// update the uri of the denom
	updateResp, err := server.UpdateDenomURI(ctx, updateMsg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedUpdateResp := &types.MsgUpdateDenomURIResponse{
		Denom:   fullDenom,
		URI:     newUri,
		URIHash: newHash,
	}

	// make sure that response is not nil
	require.NotNil(t, updateResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedUpdateResp, updateResp)

	// EVENT CHECK
	// ----------------------------------------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomURIUpdated {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeValueCategory {
					require.Equal(t, "factory", string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, signer.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenom {
					require.Equal(t, fullDenom, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenomURI {
					require.Equal(t, newUri, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenomURIHash {
					require.Equal(t, newHash, string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventDenomURIUpdated event to be emitted")
}

func TestMsgServer_UpdateDenomURI_Positive_EmptyURI(t *testing.T) {
	// Test case: update the URI of a denom with an empty URI but not empty URIHash

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name - the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin." + creator + "." + subDenom

	denomMetaData := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  subDenom,
		Display: fullDenom,
		URI:     uri,
		URIHash: uriHash,
	}

	// create a message to update uri of the denom
	newUri := ""
	newHash := ""

	newDenomMetaData := banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  subDenom,
		Display: fullDenom,
		URI:     newUri,
		URIHash: newHash,
	}

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// ensure this order of calls
	gomock.InOrder(

		// Part of the Denom creation

		// code will first check if subDenom is matching one of the code tokens (e.g. "uzig")
		bankKeeper.
			EXPECT().
			HasSupply(gomock.Any(), subDenom).
			Return(false).
			Times(1),

		// code will check if the signer has the required balances to pay the fee
		// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
		bankKeeper.
			EXPECT().
			HasBalance(gomock.Any(), signer, fee).
			Return(true).
			Times(1),

		// code will deduct the fee from the signer's account
		// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
		// it will return nil if the operation is successful - no error
		bankKeeper.
			EXPECT().
			SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
			Return(nil).
			Times(1),

		// check if denom already has metadata
		// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
		bankKeeper.
			EXPECT().
			GetDenomMetaData(gomock.Any(), fullDenom).
			Return(banktypes.Metadata{}, false).
			Times(1),

		// SetDenomMetaData(context.Context, banktypes.Metadata)
		bankKeeper.
			EXPECT().
			SetDenomMetaData(gomock.Any(), denomMetaData).
			Return().
			Times(1),

		// Part of the Denom URI update
		bankKeeper.
			EXPECT().
			GetDenomMetaData(gomock.Any(), fullDenom).
			Return(banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    fullDenom,
					Exponent: 0,
				}},
				Base:    fullDenom,
				Name:    fullDenom,
				Symbol:  subDenom,
				Display: fullDenom,
				URI:     uri,
				URIHash: uriHash,
			}, true).
			Times(1),

		// new data is being set
		// for this call we need to replace the URIHash with empty string
		bankKeeper.
			EXPECT().
			SetDenomMetaData(gomock.Any(), newDenomMetaData).
			Return().
			Times(1),
	)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               fullDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// UPDATE URI
	// ---------------------------------------------------------------

	// In the request the URI is empty, but URIHash is not empty
	updateMsg := &types.MsgUpdateDenomURI{
		Signer:  creator,
		Denom:   fullDenom,
		URI:     newUri,
		URIHash: "newsha256hash",
	}

	// update the uri of the denom
	updateResp, err := server.UpdateDenomURI(ctx, updateMsg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedUpdateResp := &types.MsgUpdateDenomURIResponse{
		Denom:   fullDenom,
		URI:     newUri,
		URIHash: newHash,
	}

	// make sure that response is not nil
	require.NotNil(t, updateResp)
	// compare the expected response with the actual response
	require.Equal(t, expectedUpdateResp, updateResp)

	// EVENT CHECK
	// ----------------------------------------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomURIUpdated {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeValueCategory {
					require.Equal(t, "factory", string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, signer.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenom {
					require.Equal(t, fullDenom, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenomURI {
					require.Equal(t, "", string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenomURIHash {
					require.Equal(t, "", string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventDenomURIUpdated event to be emitted")
}

// Negative test cases

func TestMsgServer_UpdateDenomURI_DenomDoesNotExist(t *testing.T) {
	// Test case: try to update the URI of a denom that does not exist

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample sub denom name - the full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"

	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to update the URI of a denom
	updateMsg := &types.MsgUpdateDenomURI{
		Signer:  creator,
		Denom:   fullDenom,
		URI:     "ipfs://uri",
		URIHash: "sha256hash",
	}

	// update the max supply of the denom
	_, err := server.UpdateDenomURI(ctx, updateMsg)

	// check if the error is correct
	require.ErrorIs(t, err, types.ErrDenomDoesNotExist)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf("Denom: %s: denom does not exist", fullDenom),
		err.Error(),
	)
}

func TestMsgServer_UpdateDenomURI_DenomNotFound(t *testing.T) {
	// Test case: try to update the URI of a denom that does not have metadata

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name - full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	denomMetaData := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  subDenom,
		Display: fullDenom,
		URI:     uri,
		URIHash: uriHash,
	}

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balances to pay the fee
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)

	// code will deduct the fee from the signer's account
	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	// it will return nil if the operation is successful - no error
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)

	// check if denom already has metadata
	// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(2)

	// SetDenomMetaData(context.Context, banktypes.Metadata)
	bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               fullDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// create a message to update the URI of a denom
	updateMsg := &types.MsgUpdateDenomURI{
		Signer:  creator,
		Denom:   fullDenom,
		URI:     "ipfs://uri",
		URIHash: "sha256hash",
	}

	// update the uri of the denom
	_, err = server.UpdateDenomURI(ctx, updateMsg)

	// check if the error is correct
	require.ErrorIs(t, err, types.ErrDenomNotFound)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf("Denom: %s: denom not found", fullDenom),
		err.Error(),
	)
}

func TestMsgServer_UpdateDenomURI_InvalidDenom(t *testing.T) {
	// Test case: try to update the URI of a denom with an invalid denom base

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name - full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	denomMetaData := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  subDenom,
		Display: fullDenom,
		URI:     uri,
		URIHash: uriHash,
	}

	// create a message to update uri of the denom
	newUri := "ipfs://newuri"
	newHash := "newsha256hash"
	newBase := "btc"

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// ensure this order of calls
	gomock.InOrder(
		// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
		bankKeeper.
			EXPECT().
			HasSupply(gomock.Any(), subDenom).
			Return(false).
			Times(1),

		// code will check if the signer has the required balances to pay the fee
		// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
		bankKeeper.
			EXPECT().
			HasBalance(gomock.Any(), signer, fee).
			Return(true).
			Times(1),

		// code will deduct the fee from the signer's account
		// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
		// it will return nil if the operation is successful - no error
		bankKeeper.
			EXPECT().
			SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
			Return(nil).
			Times(1),

		// check if denom already has metadata
		// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
		bankKeeper.
			EXPECT().
			GetDenomMetaData(gomock.Any(), fullDenom).
			Return(banktypes.Metadata{}, false).
			Times(1),

		// SetDenomMetaData(context.Context, banktypes.Metadata)
		bankKeeper.
			EXPECT().
			SetDenomMetaData(gomock.Any(), denomMetaData).
			Return().
			Times(1),

		bankKeeper.
			EXPECT().
			GetDenomMetaData(gomock.Any(), fullDenom).
			Return(banktypes.Metadata{
				Description: "",
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    fullDenom,
					Exponent: 0,
				}},
				Base:    newBase,
				Name:    fullDenom,
				Symbol:  subDenom,
				Display: fullDenom,
				URI:     uri,
				URIHash: uriHash,
			}, true).
			Times(1),
	)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               fullDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	updateMsg := &types.MsgUpdateDenomURI{
		Signer:  creator,
		Denom:   fullDenom,
		URI:     newUri,
		URIHash: newHash,
	}

	// update the uri of the denom
	_, err = server.UpdateDenomURI(ctx, updateMsg)

	// check if the error is correct
	require.ErrorIs(t, err, types.ErrInvalidDenom)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"Denom: %s does not match base denom: %s: Factory Denom name is not valid",
			fullDenom,
			newBase,
		),
		err.Error(),
	)
}

func TestMsgServer_UpdateDenomURI_InvalidSignerAddress(t *testing.T) {
	// Test case: try to update the URI of a denom with an invalid signer address

	// create denom first
	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample signer address
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample sub denom name - full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	denomMetaData := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  subDenom,
		Display: fullDenom,
		URI:     uri,
		URIHash: uriHash,
	}

	// create a message to update uri of the denom
	newUri := "ipfs://newuri"
	newHash := "newsha256hash"

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// ensure this order of calls
	gomock.InOrder(
		// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
		bankKeeper.
			EXPECT().
			HasSupply(gomock.Any(), subDenom).
			Return(false).
			Times(1),

		// code will check if the signer has the required balances to pay the fee
		// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
		bankKeeper.
			EXPECT().
			HasBalance(gomock.Any(), signer, fee).
			Return(true).
			Times(1),

		// code will deduct the fee from the signer's account
		// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
		// it will return nil if the operation is successful - no error
		bankKeeper.
			EXPECT().
			SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
			Return(nil).
			Times(1),

		// check if denom already has metadata
		// GetDenomMetaData(context.Context, string) (banktypes.Metadata, bool)
		bankKeeper.
			EXPECT().
			GetDenomMetaData(gomock.Any(), fullDenom).
			Return(banktypes.Metadata{}, false).
			Times(1),

		// SetDenomMetaData(context.Context, banktypes.Metadata)
		bankKeeper.
			EXPECT().
			SetDenomMetaData(gomock.Any(), denomMetaData).
			Return().
			Times(1),

		bankKeeper.
			EXPECT().
			GetDenomMetaData(gomock.Any(), fullDenom).
			Return(banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    fullDenom,
					Exponent: 0,
				}},
				Base:    fullDenom,
				Name:    fullDenom,
				Symbol:  subDenom,
				Display: fullDenom,
				URI:     uri,
				URIHash: uriHash,
			}, true).
			Times(1),
	)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	resp, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	// and we will set the expected response,
	// so we can compare it
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               fullDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	updateMsg := &types.MsgUpdateDenomURI{
		Signer:  "creator",
		Denom:   fullDenom,
		URI:     newUri,
		URIHash: newHash,
	}

	// update the uri of the denom
	_, err = server.UpdateDenomURI(ctx, updateMsg)
	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"Auth: incorrect admin for denom (%s) only the current admins bank (%s) or metadata (%s) can update the denom metadata: unauthorized",
			fullDenom,
			creator,
			creator,
		),
		err.Error(),
	)
}
