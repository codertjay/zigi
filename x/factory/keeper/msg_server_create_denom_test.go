package keeper_test

import (
	"fmt"
	"testing"

	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

func TestMsgServer_CreateDenom_Positive(t *testing.T) {
	// Test case: create a new denom with valid input

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
		Times(1)

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
		Denom:               "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)
}

func TestMsgServer_CreateDenom_Positive_EmptyURI(t *testing.T) {
	// Test case: create a new denom with valid input
	// but with an empty URI

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
	uri := ""
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
		URIHash: "", // empty URIHash because URI is empty
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
		Times(1)

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
	// and we will set the expected response the URIHash should be empty
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               "coin." + creator + "." + subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             "", // empty URIHash because URI is empty
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// EVENT CHECK
	// ----------------------------------------------------------------

	// check that the event was emitted
	events := ctx.EventManager().Events()
	require.Greater(t, len(events), 0, "Expected at least one event to be emitted")

	// find the pauser address added event
	var foundEvent bool
	for _, event := range events {
		if event.Type == types.EventDenomCreated {
			foundEvent = true
			// check event attributes
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeKeyModule {
					require.Equal(t, "factory", string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyDenom {
					require.Equal(t, fullDenom, string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyMintingCap {
					require.Equal(t, expectedResp.MintingCap.String(), string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyTotalMinted {
					require.Equal(t, "0", string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyCanChangeMintingCap {
					require.Equal(t, "true", string(attr.Value))
				}
				if string(attr.Key) == types.AttributeKeyCreator {
					require.Equal(t, signer.String(), string(attr.Value))
				}
			}
			break
		}
	}
	require.True(t, foundEvent, "Expected EventDenomURIUpdated event to be emitted")
}

// Negative test cases

func TestMsgServer_CreateDenom_SameAsNativeDenom(t *testing.T) {
	// Test case: try to create a new denom with the same name as a native denom (e.g. "uzig")

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample subdenom name - full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "uzig"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(true).
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
	_, err := server.CreateDenom(ctx, msg)

	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDenom)

	// check if the error message is correct
	// "Can't create subdenoms (uzig) that are the same as a native denom"
	require.Equal(
		t,
		"Can't create subdenoms (uzig) that are the same as a native denom: Factory Denom name is not valid",
		err.Error(),
	)
}

func TestMsgServer_CreateDenom_DenomAlreadyExists(t *testing.T) {
	// Test case: try to create a new denom with the same name as an existing denom

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
		Times(2)

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
		Times(1)

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
		Denom:               "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	// make sure that response is not nil
	require.NotNil(t, resp)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, resp)

	// create a message to create a new denom
	msg = &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	_, err = server.CreateDenom(ctx, msg)

	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDenomExists)

	// check if the error message is correct
	// "Denom: coin.creator_address.abc already exists"
	require.Equal(
		t,
		fmt.Sprintf(
			"Denom: %s already exists: attempting to create a denom that already exists (has bank metadata)",
			fullDenom),
		err.Error(),
	)
}

func TestMsgServer_CreateDenom_InsufficientFunds(t *testing.T) {
	// Test case: try to create a new denom with insufficient funds

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
	// fee := sdk.NewCoin("uzig", cosmosmath.NewInt(100))

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// Code will check if the signer has the required balances to pay the fee
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, gomock.Any()).
		Return(false).
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
	_, err := server.CreateDenom(ctx, msg)

	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)

	// check if the error message is correct
	// "not enough balances to pay for denom creation fee"
	require.Equal(
		t,
		"not enough balance to pay for denom creation fee: 1000uzig: insufficient funds",
		err.Error(),
	)
}

func TestMsgServer_CreateDenom_FailedToSendCoins(t *testing.T) {
	// Test case: try to create a new denom but fail to send coins from sender's account to a module account

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

	// code will deduct the fee from the signer's account but return an error
	// SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(fmt.Errorf("failed to send coins")).
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
	_, err := server.CreateDenom(ctx, msg)

	require.Error(t, err)

	// check if the error message is correct
	// "failed to collect denom creation fee"
	require.Equal(
		t,
		"failed to collect denom creation fee: failed to send coins",
		err.Error(),
	)
}

func TestMsgServer_CreateDenom_MetaDataExists(t *testing.T) {
	// Test case: try to create a new denom, but metadata already exists

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

	// code will check if the denom metadata already exists
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, true).
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
	_, err := server.CreateDenom(ctx, msg)

	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDenomExists)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"denom: %s already exists: attempting to create a denom that already exists (has bank metadata)",
			fullDenom),
		err.Error(),
	)
}

func TestMsgServer_CreateDenom_InvalidCreatorAddress(t *testing.T) {
	// Test case: try to create a new denom with an invalid creator address

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample sub denom name - full name will be in format "factory/{creator}/{subDenom}"
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom with an invalid creator address
	msg := &types.MsgCreateDenom{
		Creator:             "invalid_address",
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	_, err := server.CreateDenom(ctx, msg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		"invalid denom: coin.invalid_address.abc",
		err.Error(),
	)
}

func TestMsgServer_CreateDenom_MissingCreatorAddress(t *testing.T) {
	// Test case: try to create a new denom with missing creator address

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample sub denom name - full name will be in format "factory/{creator}/{subDenom}"
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to create a new denom with an invalid creator address
	msg := &types.MsgCreateDenom{
		// create msg without creator address
		// Creator:              "invalid_address",
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// create a new denom
	_, err := server.CreateDenom(ctx, msg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		"empty address string is not allowed",
		err.Error(),
	)
}

func TestMsgServer_CreateDenom_InvalidSubDenom(t *testing.T) {
	// Test case: try to create a new denom with an invalid sub-denom name

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample sub denom name - full name will be in format "factory/{creator}/{subDenom}"
	subDenom := "invalid!@#subDenom"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
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
	_, err := server.CreateDenom(ctx, msg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf("invalid coin: '%s' only lowercase letters (a-z) and numbers (0-9) are allowed e.g. uzig123: invalid coins",
			subDenom),
		err.Error(),
	)
}
