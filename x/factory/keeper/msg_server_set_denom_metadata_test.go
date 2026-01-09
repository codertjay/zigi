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

func TestMsgServer_SetDenomMetadata_Positive(t *testing.T) {
	// Test case: set metadata of a denom with valid input data

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

	newDenomMetaData := banktypes.Metadata{
		Description: "new description",
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

	// set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   creator,
		Metadata: newDenomMetaData,
	}

	setMetadataResp, err := server.SetDenomMetadata(ctx, setMetadataMsg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	expectedSetMetadataResp := &types.MsgSetDenomMetadataResponse{
		Metadata: &newDenomMetaData,
	}

	// make sure that response is not nil
	require.NotNil(t, setMetadataResp)

	// compare the expected response with the actual response
	require.Equal(t, expectedSetMetadataResp, setMetadataResp)

	// EVENT CHECK
	// ----------------------------------------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomMetadataUpdated {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyModule {
					require.Equal(t, "factory", string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, signer.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenomMetadata {
					require.Equal(t, expectedSetMetadataResp.Metadata.String(), string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventDenomURIUpdated event to be emitted")
}

func TestMsgServer_SetDenomMetadata_Positive_EmptyURI(t *testing.T) {
	// Test case: set metadata of a denom with valid input data but empty URI

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

	newDenomMetaData := banktypes.Metadata{
		Description: "new description",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  subDenom,
		Display: fullDenom,
		URI:     newUri,
		URIHash: "", // empty URIHash because URI is empty
	}

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
		URIHash:             uriHash, // empty URIHash because URI is empty
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   creator,
		Metadata: newDenomMetaData,
	}

	setMetadataResp, err := server.SetDenomMetadata(ctx, setMetadataMsg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct,
	expectedSetMetadataResp := &types.MsgSetDenomMetadataResponse{
		Metadata: &newDenomMetaData,
	}

	// make sure that response is not nil
	require.NotNil(t, setMetadataResp)

	// compare the expected response with the actual response
	require.Equal(t, expectedSetMetadataResp, setMetadataResp)

	// Check that the URIHash is empty on expectedSetMetadataResp
	require.Equal(t, "", expectedSetMetadataResp.Metadata.URIHash) // URIHash is empty because URI is empty

	// EVENT CHECK
	// ----------------------------------------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomMetadataUpdated {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyModule {
					require.Equal(t, "factory", string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeySigner {
					require.Equal(t, signer.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenomMetadata {
					require.Equal(t, expectedSetMetadataResp.Metadata.String(), string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventDenomURIUpdated event to be emitted")

}

// Negative test cases

func TestMsgServer_SetDenomMetadata_NameBlank(t *testing.T) {
	// Test case: try to set denom metadata with blank name

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

	newDenomMetaData := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    "",
		Symbol:  subDenom,
		Display: fullDenom,
		URI:     newUri,
		URIHash: newHash,
	}

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

	// create a message to set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   creator,
		Metadata: newDenomMetaData,
	}

	err = setMetadataMsg.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, "URI: ipfs://newuri length must be between 15 and 256 characters: Metadata is not valid", err.Error())
}

func TestMsgServer_SetDenomMetadata_SymbolBlank(t *testing.T) {
	// Test case: try to set denom metadata with blank symbol

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

	newDenomMetaData := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  "",
		Display: fullDenom,
		URI:     newUri,
		URIHash: newHash,
	}

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

	// create a message to set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   creator,
		Metadata: newDenomMetaData,
	}

	err = setMetadataMsg.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, "URI: ipfs://newuri length must be between 15 and 256 characters: Metadata is not valid", err.Error())
}

func TestMsgServer_SetDenomMetadata_DisplayBlank(t *testing.T) {
	// Test case: try to set denom metadata with blank display

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

	newDenomMetaData := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  subDenom,
		Display: "",
		URI:     newUri,
		URIHash: newHash,
	}

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

	// create a message to set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   creator,
		Metadata: newDenomMetaData,
	}

	err = setMetadataMsg.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, "URI: ipfs://newuri length must be between 15 and 256 characters: Metadata is not valid", err.Error())

}

func TestMsgServer_SetDenomMetadata_DisplayBadChars(t *testing.T) {
	// Test case: try to set denom metadata with bad characters in display

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

	// create a sample sub denom name - full name will be in format "factory/{creator}/{subDenom}"
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
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  subDenom,
		Display: "#$%§",
		URI:     newUri,
		URIHash: newHash,
	}

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

	// create a message to set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   creator,
		Metadata: newDenomMetaData,
	}

	err = setMetadataMsg.ValidateBasic()
	require.Error(t, err)

	require.Equal(t, "URI: ipfs://newuri length must be between 15 and 256 characters: Metadata is not valid", err.Error())
}

func TestMsgServer_SetDenomMetadata_DisplayInvalid(t *testing.T) {
	// Test case: try to set denom metadata with invalid display

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

	// create a sample sub denom name - full name will be in format "factory/{creator}/{subDenom}"
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
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    fullDenom,
		Name:    fullDenom,
		Symbol:  subDenom,
		Display: "btc",
		URI:     newUri,
		URIHash: newHash,
	}

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

	// create a message to set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   creator,
		Metadata: newDenomMetaData,
	}

	err = setMetadataMsg.ValidateBasic()

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		"URI: ipfs://newuri length must be between 15 and 256 characters: Metadata is not valid",
		err.Error(),
	)
}

func TestMsgServer_SetDenomMetadata_BaseEmpty(t *testing.T) {
	// Test case: try to set denom metadata with empty base

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

	// create a sample sub denom name - full name will be in format "factory/{creator}/{subDenom}"
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
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    "",
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

	// create a message to set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   creator,
		Metadata: newDenomMetaData,
	}

	_, err = server.SetDenomMetadata(ctx, setMetadataMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		"Auth: denom (): denom auth not found",
		err.Error(),
	)
}

func TestMsgServer_SetDenomMetadata_BaseInvalid(t *testing.T) {
	// Test case: try to set denom metadata with invalid base

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

	// create a sample sub denom name - full name will be in format "factory/{creator}/{subDenom}"
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

	base := "coin." + creator + "." + "abc%#!"
	newDenomMetaData := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    base,
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

	// create a message to set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   creator,
		Metadata: newDenomMetaData,
	}

	_, err = server.SetDenomMetadata(ctx, setMetadataMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"Auth: denom (%s): denom auth not found",
			base,
		),
		err.Error(),
	)
}

func TestMsgServer_SetDenomMetadata_BaseInvalidFirstDenomination(t *testing.T) {
	// Test case: try to set denom metadata with invalid base

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

	// create a sample sub denom name - full name will be in format "factory/{creator}/{subDenom}"
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

	base := "coin." + creator + "." + "abcd"
	newDenomMetaData := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 0,
		}},
		Base:    base,
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

	// create a message to set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   creator,
		Metadata: newDenomMetaData,
	}

	_, err = server.SetDenomMetadata(ctx, setMetadataMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"Auth: denom (%s): denom auth not found",
			base,
		),
		err.Error(),
	)
}

func TestMsgServer_SetDenomMetadata_DenomUnitBaseExponent(t *testing.T) {
	// Test case: try to set denom metadata with base exponent different from 0

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

	// create a sample sub denom name - full name will be in format "factory/{creator}/{subDenom}"
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
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom,
			Exponent: 10,
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

	// create a message to set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   creator,
		Metadata: newDenomMetaData,
	}

	err = setMetadataMsg.ValidateBasic()

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		"URI: ipfs://newuri length must be between 15 and 256 characters: Metadata is not valid",
		err.Error(),
	)
}

func TestMsgServer_SetDenomMetadata_DenomUnitInvalidDenom(t *testing.T) {
	// Test case: try to set denom metadata with invalid denom in denomination unit

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

	// create a sample sub denom name - full name will be in format "factory/{creator}/{subDenom}"
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
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    "factory/" + creator + "/",
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

	// create a message to set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   creator,
		Metadata: newDenomMetaData,
	}

	err = setMetadataMsg.ValidateBasic()

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		"URI: ipfs://newuri length must be between 15 and 256 characters: Metadata is not valid",
		err.Error(),
	)
}

func TestMsgServer_SetDenomMetadata_Unauthorized(t *testing.T) {
	// Test case: try to set denom metadata with unauthorized signer

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

	// create a sample sub denom name - full name will be in format "factory/{creator}/{subDenom}"
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
		Description: "",
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

	// create a message to set metadata of the denom
	setMetadataMsg := &types.MsgSetDenomMetadata{
		Signer:   sample.AccAddress(),
		Metadata: newDenomMetaData,
	}

	_, err = server.SetDenomMetadata(ctx, setMetadataMsg)

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
