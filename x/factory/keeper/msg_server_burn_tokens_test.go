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

func TestMsgServer_BurnTokens_Positive(t *testing.T) {
	// Test case: burn tokens successfully

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

	if err != nil {
		t.Fatal(err)
	}

	burnAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))

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

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, burnAmount).
		Return(true).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
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

	// create a message to mint and send tokens
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

	require.Equal(t, expectedRespMintAndSendTokens, respMintAndSendTokens)

	// create a message to burn tokens
	burnMsg := types.MsgBurnTokens{
		Signer: creator,
		Token:  burnAmount,
	}

	respBurnTokens, err := server.BurnTokens(ctx, &burnMsg)
	require.NoError(t, err)
	require.NotNil(t, respBurnTokens)

	expectedRespBurnTokens := &types.MsgBurnTokensResponse{
		AmountBurned: burnAmount,
	}

	require.Equal(t, expectedRespBurnTokens, respBurnTokens)
}

func TestMsgServer_BurnTokens_HalfMinted(t *testing.T) {
	// Test case: burn half minted tokens successfully

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

	if err != nil {
		t.Fatal(err)
	}

	burnAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(50))

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

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, burnAmount).
		Return(true).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
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

	// create a message to mint and send tokens
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

	require.Equal(t, expectedRespMintAndSendTokens, respMintAndSendTokens)

	// create a message to burn tokens
	burnMsg := types.MsgBurnTokens{
		Signer: creator,
		Token:  burnAmount,
	}

	respBurnTokens, err := server.BurnTokens(ctx, &burnMsg)
	require.NoError(t, err)
	require.NotNil(t, respBurnTokens)

	expectedRespBurnTokens := &types.MsgBurnTokensResponse{
		AmountBurned: burnAmount,
	}

	require.Equal(t, expectedRespBurnTokens, respBurnTokens)
}

func TestMsgServer_BurnTokens_Zero(t *testing.T) {
	// Test case: burn zero tokens successfully

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

	if err != nil {
		t.Fatal(err)
	}

	burnAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(0))

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

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, burnAmount).
		Return(true).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
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

	// create a message to mint and send tokens
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

	require.Equal(t, expectedRespMintAndSendTokens, respMintAndSendTokens)

	// create a message to burn tokens
	burnMsg := types.MsgBurnTokens{
		Signer: creator,
		Token:  burnAmount,
	}

	respBurnTokens, err := server.BurnTokens(ctx, &burnMsg)
	require.NoError(t, err)
	require.NotNil(t, respBurnTokens)

	expectedRespBurnTokens := &types.MsgBurnTokensResponse{
		AmountBurned: burnAmount,
	}

	require.Equal(t, expectedRespBurnTokens, respBurnTokens)
}

func TestMsgServer_BurnTokens_uzig(t *testing.T) {
	// Test case: burn non-factory tokens - uzig

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	nonFactoryDenom := "uzig"
	burnAmount := sdk.NewCoin(nonFactoryDenom, cosmosmath.NewInt(50))

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// expect `HasBalance` to be called and return true, meaning the signer has enough balances
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, burnAmount).
		Return(true).
		Times(1)

	// expect `SendCoinsFromAccountToModule` to be called, moving the tokens to the module account
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
		Times(1)

	// expect `BurnCoins` to be called, burning the tokens from the module account
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to burn tokens with a non-factory denom
	burnMsg := types.MsgBurnTokens{
		Signer: creator,
		Token:  burnAmount,
	}

	// call BurnTokens
	resp, err := server.BurnTokens(ctx, &burnMsg)

	// ensure no error occurs
	require.NoError(t, err)
	require.NotNil(t, resp)

	// check response
	expectedResp := &types.MsgBurnTokensResponse{
		AmountBurned: burnAmount,
	}
	require.Equal(t, expectedResp, resp)
}

func TestMsgServer_BurnTokens_NonFactoryDenom(t *testing.T) {
	// Test case: burn non-factory tokens - ibc

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	nonFactoryDenom := "ibc"
	burnAmount := sdk.NewCoin(nonFactoryDenom, cosmosmath.NewInt(50))

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// expect `HasBalance` to be called and return true, meaning the signer has enough balances
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, burnAmount).
		Return(true).
		Times(1)

	// expect `SendCoinsFromAccountToModule` to be called, moving the tokens to the module account
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
		Times(1)

	// expect `BurnCoins` to be called, burning the tokens from the module account
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to burn tokens with a non-factory denom
	burnMsg := types.MsgBurnTokens{
		Signer: creator,
		Token:  burnAmount,
	}

	// call BurnTokens
	resp, err := server.BurnTokens(ctx, &burnMsg)

	// ensure no error occurs
	require.NoError(t, err)
	require.NotNil(t, resp)

	// check response
	expectedResp := &types.MsgBurnTokensResponse{
		AmountBurned: burnAmount,
	}
	require.Equal(t, expectedResp, resp)
}

func TestMsgServer_BurnTokens_NonFactoryDenomZero(t *testing.T) {
	// Test case: burn non-factory token with zero amounts

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	nonFactoryDenom := "ibc"
	burnAmount := sdk.NewCoin(nonFactoryDenom, cosmosmath.NewInt(0))

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// expect `HasBalance` to be called and return true, meaning the signer has enough balances
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, burnAmount).
		Return(true).
		Times(1)

	// expect `SendCoinsFromAccountToModule` to be called, moving the tokens to the module account
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
		Times(1)

	// expect `BurnCoins` to be called, burning the tokens from the module account
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to burn tokens with a non-factory denom
	burnMsg := types.MsgBurnTokens{
		Signer: creator,
		Token:  burnAmount,
	}

	// call BurnTokens
	resp, err := server.BurnTokens(ctx, &burnMsg)

	// ensure no error occurs
	require.NoError(t, err)
	require.NotNil(t, resp)

	// check response
	expectedResp := &types.MsgBurnTokensResponse{
		AmountBurned: burnAmount,
	}
	require.Equal(t, expectedResp, resp)
}

// Negative test cases

func TestMsgServer_BurnTokens_InsufficientFunds(t *testing.T) {
	// Test case: trying to burn tokens when the signer does not have enough funds

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
	mintingCap := cosmosmath.NewUint(1_000)
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

	mintAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(1000))
	recipient := sample.AccAddress()
	recipientAddress, err := sdk.AccAddressFromBech32(recipient)

	if err != nil {
		t.Fatal(err)
	}

	burnAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))

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
		Return(sdk.NewCoin(fullDenom, cosmosmath.NewInt(1000))).
		Times(1)

	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, burnAmount).
		Return(false).
		Times(1)

	bankKeeper.
		EXPECT().
		GetBalance(gomock.Any(), signer, fullDenom).
		Return(sdk.NewCoin(fullDenom, cosmosmath.NewInt(0))).
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

	// create a message to mint and send tokens
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

	require.Equal(t, expectedRespMintAndSendTokens, respMintAndSendTokens)

	// create a message to burn tokens
	burnMsg := types.MsgBurnTokens{
		Signer: creator,
		Token:  burnAmount,
	}

	_, err = server.BurnTokens(ctx, &burnMsg)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInsufficientFunds)

	// check if the error message is correct
	require.Equal(t,
		fmt.Sprintf(
			"Signer: %s does not have enough funds, burn amount: %s, existing balance: 0%s: insufficient funds",
			signer,
			burnAmount,
			burnAmount.Denom,
		),
		err.Error(),
	)
}

func TestMsgServer_BurnTokens_InsufficientFunds_NonFactoryDenom(t *testing.T) {
	// Test case: trying to burn tokens when the signer does not have enough funds

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	nonFactoryDenom := "ibc"
	burnAmount := sdk.NewCoin(nonFactoryDenom, cosmosmath.NewInt(50))

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// expect `HasBalance` to be called and return true, meaning the signer has enough balances
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, burnAmount).
		Return(false).
		Times(1)

	bankKeeper.
		EXPECT().
		GetBalance(gomock.Any(), signer, nonFactoryDenom).
		Return(sdk.NewCoin(nonFactoryDenom, cosmosmath.NewInt(0))).
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to burn tokens with a non-factory denom
	burnMsg := types.MsgBurnTokens{
		Signer: creator,
		Token:  burnAmount,
	}

	_, err := server.BurnTokens(ctx, &burnMsg)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInsufficientFunds)

	// check if the error message is correct
	require.Equal(t,
		fmt.Sprintf(
			"Signer: %s does not have enough funds, burn amount: %s, existing balance: 0%s: insufficient funds",
			signer,
			burnAmount,
			burnAmount.Denom,
		),
		err.Error(),
	)
}

func TestMsgServer_BurnTokens_NonExisting(t *testing.T) {
	// Test case: try to burn token that does not exist

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

	// set the burn amount
	burnAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create a message to burn tokens
	burnMsg := types.MsgBurnTokens{
		Signer: creator,
		Token:  burnAmount,
	}

	// try to burn tokens
	_, err := server.BurnTokens(ctx, &burnMsg)
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrKeyNotFound)

	// check if the error message is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"Token: %s does not exist: key not found",
			fullDenom,
		),
		err.Error(),
	)
}

func TestMsgServer_BurnTokens_InvalidSignerAddress(t *testing.T) {
	// Test case: try to burn tokens with an invalid signer address

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

	burnAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))

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

	// create a message to mint and send tokens
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

	require.Equal(t, expectedRespMintAndSendTokens, respMintAndSendTokens)

	// create a message to burn tokens
	burnMsg := types.MsgBurnTokens{
		Signer: "invalid_address",
		Token:  burnAmount,
	}

	// try to burn tokens with a signer that is not the bank admin
	_, err = server.BurnTokens(ctx, &burnMsg)
	require.Error(t, err)

	// check if the error message is correct
	require.Equal(
		t,
		"decoding bech32 failed: invalid separator index -1",
		err.Error(),
	)
}

func TestMsgServer_BurnTokens_SendCoinsError(t *testing.T) {
	// Test case: burn tokens fails due to error in SendCoinsFromAccountToModule

	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // Assert that all expectations are met

	// Create a sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	// Create a sample sub-denom and full denom
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	// Create denom metadata
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

	// Define mint and burn amounts
	mintAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))
	burnAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(50))
	recipient := sample.AccAddress()
	recipientAddress, err := sdk.AccAddressFromBech32(recipient)
	require.NoError(t, err)

	// Create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// Mock expectations for creating a denom
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)
	bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// Mock expectations for minting tokens
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(mintAmount)).
		Return(nil).
		Times(1)
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, recipientAddress, sdk.NewCoins(mintAmount)).
		Return(nil).
		Times(1)
	bankKeeper.
		EXPECT().
		GetSupply(gomock.Any(), fullDenom).
		Return(sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))).
		Times(1)

	// Mock expectations for burning tokens
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, burnAmount).
		Return(true).
		Times(1)
	// Simulate an error in SendCoinsFromAccountToModule
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(fmt.Errorf("failed to send coins to module: invalid transfer")).
		Times(1)

	// Create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// Create a message server
	server := keeper.NewMsgServerImpl(k)

	// Create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// Create a new denom
	resp, err := server.CreateDenom(ctx, msg)
	require.NoError(t, err)
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               fullDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	require.NotNil(t, resp)
	require.Equal(t, expectedResp, resp)

	// Create a message to mint and send tokens
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
	require.Equal(t, expectedRespMintAndSendTokens, respMintAndSendTokens)

	// Create a message to burn tokens
	burnMsg := types.MsgBurnTokens{
		Signer: creator,
		Token:  burnAmount,
	}

	// Attempt to burn tokens
	_, err = server.BurnTokens(ctx, &burnMsg)
	require.Error(t, err)
	require.Equal(t, "failed to send coins to module: invalid transfer", err.Error(), "Expected error from SendCoinsFromAccountToModule")
}

func TestMsgServer_BurnTokens_BurnCoinsError(t *testing.T) {
	// Test case: burn tokens fails due to error in BurnCoins

	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // Assert that all expectations are met

	// Create a sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	// Create a sample sub-denom and full denom
	subDenom := "abc"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "sha256hash"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	// Create denom metadata
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

	// Define mint and burn amounts
	mintAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))
	burnAmount := sdk.NewCoin(fullDenom, cosmosmath.NewInt(50))
	recipient := sample.AccAddress()
	recipientAddress, err := sdk.AccAddressFromBech32(recipient)
	require.NoError(t, err)

	// Create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// Mock expectations for creating a denom
	bankKeeper.
		EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)
	bankKeeper.
		EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)
	bankKeeper.
		EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// Mock expectations for minting tokens
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(mintAmount)).
		Return(nil).
		Times(1)
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, recipientAddress, sdk.NewCoins(mintAmount)).
		Return(nil).
		Times(1)
	bankKeeper.
		EXPECT().
		GetSupply(gomock.Any(), fullDenom).
		Return(sdk.NewCoin(fullDenom, cosmosmath.NewInt(100))).
		Times(1)

	// Mock expectations for burning tokens
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, burnAmount).
		Return(true).
		Times(1)
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(nil).
		Times(1)
	// Simulate an error in BurnCoins
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(burnAmount)).
		Return(fmt.Errorf("failed to burn coins: invalid operation")).
		Times(1)

	// Create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// Create a message server
	server := keeper.NewMsgServerImpl(k)

	// Create a message to create a new denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	// Create a new denom
	resp, err := server.CreateDenom(ctx, msg)
	require.NoError(t, err)
	expectedResp := &types.MsgCreateDenomResponse{
		BankAdmin:           creator,
		MetadataAdmin:       creator,
		Denom:               fullDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}
	require.NotNil(t, resp)
	require.Equal(t, expectedResp, resp)

	// Create a message to mint and send tokens
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
	require.Equal(t, expectedRespMintAndSendTokens, respMintAndSendTokens)

	// Create a message to burn tokens
	burnMsg := types.MsgBurnTokens{
		Signer: creator,
		Token:  burnAmount,
	}

	// Attempt to burn tokens
	_, err = server.BurnTokens(ctx, &burnMsg)
	require.Error(t, err)
	require.Equal(t, "failed to burn coins: invalid operation", err.Error(), "Expected error from BurnCoins")
}
