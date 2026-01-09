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

func TestMsgServer_MintAndSendTokens_Positive(t *testing.T) {
	// Test case: mint and send tokens successfully

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

	mintAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))
	recipient := sample.AccAddress()
	recipientAddress, err := sdk.AccAddressFromBech32(recipient)

	require.NoError(t, err)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balance of base
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

	// check if denom has metadata already
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

	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(mintAmount)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(
			gomock.Any(),
			types.ModuleName,
			recipientAddress,
			sdk.NewCoins(mintAmount)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		GetSupply(gomock.Any(), fullDenom).
		Return(sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))).
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
	respCreateDenom, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct
	// so we will set the expected response,
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
	require.NotNil(t, respCreateDenom)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, respCreateDenom)

	mintMsg := types.MsgMintAndSendTokens{
		Signer:    creator,
		Token:     mintAmount,
		Recipient: recipient,
	}

	respMintAndSendTokens, err := server.MintAndSendTokens(ctx, &mintMsg)

	require.NoError(t, err)

	require.NotNil(t, respMintAndSendTokens)

	expectedRespMintAndSendTokens := &types.MsgMintAndSendTokensResponse{
		TokenMinted: &mintAmount,
		Recipient:   recipient,
		TotalMinted: &mintAmount,
		TotalSupply: &mintAmount,
	}

	// compare the expected response with the actual response
	require.Equal(t, expectedRespMintAndSendTokens, respMintAndSendTokens)
}

// Negative test cases

func TestMsgServer_MintAndSendTokens_TokenDoesNotExist(t *testing.T) {
	// Test case: tyring to mint and send tokens for a token that does not exist

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// create a sample sub denom name - full name will be in format "coin.{creator}.{subDenom}"
	subDenom := "abc"
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	mintAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))
	recipient := sample.AccAddress()

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	mintMsg := types.MsgMintAndSendTokens{
		Signer:    creator,
		Token:     mintAmount,
		Recipient: recipient,
	}

	// mint and send tokens
	_, err := server.MintAndSendTokens(ctx, &mintMsg)

	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrKeyNotFound)

	// check if the error message is correct
	require.Equal(t,
		fmt.Sprintf(
			"Token: %s does not exist: key not found",
			fullDenom),
		err.Error(),
	)
}

func TestMsgServer_MintAndSendTokens_MintingExceedMintingCap(t *testing.T) {
	// Test case: trying to mint and send tokens when minting would exceed minting cap

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

	mintAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(10000000))
	recipient := sample.AccAddress()
	// recipientAddress, err := sdk.AccAddressFromBech32(recipient)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balance of base
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

	// check if denom has metadata already
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
	respCreateDenom, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct
	// so we will set the expected response,
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
	require.NotNil(t, respCreateDenom)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, respCreateDenom)

	mintMsg := types.MsgMintAndSendTokens{
		Signer:    creator,
		Token:     mintAmount,
		Recipient: recipient,
	}

	// mint and send tokens
	_, err = server.MintAndSendTokens(ctx, &mintMsg)

	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)

	// check if the error message is correct
	require.Equal(t,
		fmt.Sprintf(
			"Minting %s would exceed Minting Cap of %s: invalid request",
			mintAmount,
			mintingCap),
		err.Error(),
	)
}

func TestMsgServer_MintAndSendTokens_BadSigner(t *testing.T) {
	// Test case: try to mint and send tokens with a bad signer

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

	mintAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))
	recipient := sample.AccAddress()

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balance of base
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

	// check if denom has metadata already
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
	respCreateDenom, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct
	// so we will set the expected response,
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
	require.NotNil(t, respCreateDenom)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, respCreateDenom)

	// create a message to mint and send tokens
	mintMsg := types.MsgMintAndSendTokens{
		Signer:    recipient,
		Token:     mintAmount,
		Recipient: recipient,
	}

	// try to mint and send tokens with a bad signer
	_, err = server.MintAndSendTokens(ctx, &mintMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"Auth: incorrect admin for denom (%s) only the current bank admin (%s) can perform this action (attempted with: %s): unauthorized",
			fullDenom,
			creator,
			recipient,
		),
		err.Error(),
	)
}

func TestMsgServer_MintAndSendTokens_InvalidRecipientAddress(t *testing.T) {
	// Test case: try to mint and send tokens with invalid recipient address

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

	mintAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balance of base
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

	// check if denom has metadata already
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
	respCreateDenom, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct
	// so we will set the expected response,
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
	require.NotNil(t, respCreateDenom)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, respCreateDenom)

	// create a message to mint and send tokens
	mintMsg := types.MsgMintAndSendTokens{
		Signer:    creator,
		Token:     mintAmount,
		Recipient: "invalid_address",
	}

	// try to mint and send tokens with a bad signer
	_, err = server.MintAndSendTokens(ctx, &mintMsg)

	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		"decoding bech32 failed: invalid separator index -1",
		err.Error(),
	)
}

func TestMsgServer_MintAndSendTokens_MintCoinsError(t *testing.T) {
	// Test case: handle error from bankKeeper.MintCoins

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

	mintAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))
	recipient := sample.AccAddress()

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balance of base
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

	// check if denom has metadata already
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

	expectedErr := fmt.Errorf("mint coins error")

	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(mintAmount)).
		Return(expectedErr).
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
	respCreateDenom, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct
	// so we will set the expected response,
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
	require.NotNil(t, respCreateDenom)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, respCreateDenom)

	mintMsg := types.MsgMintAndSendTokens{
		Signer:    creator,
		Token:     mintAmount,
		Recipient: recipient,
	}

	respMintAndSendTokens, err := server.MintAndSendTokens(ctx, &mintMsg)

	require.Nil(t, respMintAndSendTokens)
	require.ErrorIs(t, err, expectedErr)

	// Check that the denom's minted amount was not updated
	currentDenom, found := k.GetDenom(ctx, fullDenom)
	require.True(t, found)
	require.Equal(t, cosmosmath.ZeroUint(), currentDenom.Minted)
}

func TestMsgServer_MintAndSendTokens_SendCoinsFromModuleToAccountError(t *testing.T) {
	// Test case: handle error from bankKeeper.SendCoinsFromModuleToAccount

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

	mintAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))
	recipient := sample.AccAddress()
	recipientAddress, err := sdk.AccAddressFromBech32(recipient)
	require.NoError(t, err)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will fist check if subDenom is matching one of the code tokens (e.g. "uzig")
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)

	// code will check if the signer has the required balance of base
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

	// check if denom has metadata already
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

	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(mintAmount)).
		Return(nil).
		Times(1)

	expectedErr := fmt.Errorf("send coins error")

	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(
			gomock.Any(),
			types.ModuleName,
			recipientAddress,
			sdk.NewCoins(mintAmount)).
		Return(expectedErr).
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
	respCreateDenom, err := server.CreateDenom(ctx, msg)

	// check if the response is correct
	require.NoError(t, err)

	// we will also need to check if the response is correct
	// so we will set the expected response,
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
	require.NotNil(t, respCreateDenom)
	// compare the expected response with the actual response
	require.Equal(t, expectedResp, respCreateDenom)

	mintMsg := types.MsgMintAndSendTokens{
		Signer:    creator,
		Token:     mintAmount,
		Recipient: recipient,
	}

	respMintAndSendTokens, err := server.MintAndSendTokens(ctx, &mintMsg)

	require.Nil(t, respMintAndSendTokens)
	require.ErrorIs(t, err, expectedErr)

	// Check that the denom's minted amount was not updated
	currentDenom, found := k.GetDenom(ctx, fullDenom)
	require.True(t, found)
	require.Equal(t, cosmosmath.ZeroUint(), currentDenom.Minted)
}
