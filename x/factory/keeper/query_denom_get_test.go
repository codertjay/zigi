package keeper_test

import (
	"fmt"
	"testing"

	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/factory/keeper"
	"zigchain/x/factory/testutil"
	"zigchain/x/factory/types"
)

// Positive test cases

func TestQueryDenomGet_Positive_Create(t *testing.T) {
	// Test case: create two denoms and query them

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // ensure all expectations are met

	// create two sample signers
	creator := sample.AccAddress()

	signer := sdk.MustAccAddressFromBech32(creator)

	// create sub-denoms
	subDenom := "abc"

	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "b598d682c86bc9dafc92b33789efc2f94929deabf1cc3fbbb45b8375cba78124"
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

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// expectations for creating denom
	bankKeeper.EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)
	bankKeeper.EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)
	bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)
	bankKeeper.EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)
	bankKeeper.EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	bankKeeper.EXPECT().
		GetSupply(gomock.Any(), fullDenom).
		Return(sdk.NewCoin(fullDenom, cosmosmath.NewInt(0))).
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create denom
	createMsg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	respCreate1, err := server.CreateDenom(ctx, createMsg)
	require.NoError(t, err)
	require.NotNil(t, respCreate1)

	// query denom
	queryReq := &types.QueryGetDenomRequest{
		Denom: fullDenom,
	}

	queryResp, err := k.Denom(ctx, queryReq)
	require.NoError(t, err)
	require.NotNil(t, queryResp)

	// validate denom
	require.Equal(t, fullDenom, queryResp.Denom)
	require.Equal(t, cosmosmath.NewUint(0), queryResp.TotalMinted)
	require.Equal(t, cosmosmath.NewUint(0), queryResp.TotalSupply)
	require.Equal(t, cosmosmath.NewUint(0), queryResp.TotalBurned)
	require.Equal(t, mintingCap, queryResp.MintingCap)
	require.Equal(t, mintingCap, queryResp.MaxSupply)
	require.True(t, queryResp.CanChangeMintingCap)
	require.Equal(t, creator, queryResp.Creator)
	require.Equal(t, creator, queryResp.BankAdmin)
	require.Equal(t, creator, queryResp.MetadataAdmin)
}

func TestQueryDenomGet_Positive_Mint(t *testing.T) {
	// Test case: create two denoms, mint tokens and query them

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // ensure all expectations are met

	// create two sample signers
	creator1 := sample.AccAddress()
	creator2 := sample.AccAddress()

	signer1 := sdk.MustAccAddressFromBech32(creator1)
	signer2 := sdk.MustAccAddressFromBech32(creator2)

	// create two different sub-denoms
	subDenom1 := "abc"
	subDenom2 := "bcd"

	mintingCap := cosmosmath.NewUint(1000000)
	uri1 := "ipfs://uri1"
	uri2 := "ipfs://uri2"
	uriHash1 := "b598d682c86bc9dafc92b33789efc2f94929deabf1cc3fbbb45b8375cba78124"
	uriHash2 := "ab4ff2837b90af0e129b7a3e3af949022e1e3215ae1b2b2f502ce15281d35b32"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom1 := "coin" + types.FactoryDenomDelimiterChar + creator1 + types.FactoryDenomDelimiterChar + subDenom1
	fullDenom2 := "coin" + types.FactoryDenomDelimiterChar + creator2 + types.FactoryDenomDelimiterChar + subDenom2

	denomMetaData1 := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom1,
			Exponent: 0,
		}},
		Base:    fullDenom1,
		Name:    fullDenom1,
		Symbol:  subDenom1,
		Display: fullDenom1,
		URI:     uri1,
		URIHash: uriHash1,
	}

	denomMetaData2 := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom2,
			Exponent: 0,
		}},
		Base:    fullDenom2,
		Name:    fullDenom2,
		Symbol:  subDenom2,
		Display: fullDenom2,
		URI:     uri2,
		URIHash: uriHash2,
	}

	mintAmount1 := sdk.NewCoin(fullDenom1, cosmosmath.NewInt(100))
	mintAmount2 := sdk.NewCoin(fullDenom2, cosmosmath.NewInt(200))

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// expectations for creating denom1
	bankKeeper.EXPECT().
		HasSupply(gomock.Any(), subDenom1).
		Return(false).
		Times(1)
	bankKeeper.EXPECT().
		HasBalance(gomock.Any(), signer1, fee).
		Return(true).
		Times(1)
	bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer1, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)
	bankKeeper.EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom1).
		Return(banktypes.Metadata{}, false).
		Times(1)
	bankKeeper.EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData1).
		Return().
		Times(1)

	// expectations for creating denom2
	bankKeeper.EXPECT().
		HasSupply(gomock.Any(), subDenom2).
		Return(false).
		Times(1)
	bankKeeper.EXPECT().
		HasBalance(gomock.Any(), signer2, fee).
		Return(true).
		Times(1)
	bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer2, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)
	bankKeeper.EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom2).
		Return(banktypes.Metadata{}, false).
		Times(1)
	bankKeeper.EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData2).
		Return().
		Times(1)

	// mint coins for denom1
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(mintAmount1)).
		Return(nil).
		Times(1)
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(
			gomock.Any(),
			types.ModuleName,
			signer1,
			sdk.NewCoins(mintAmount1)).
		Return(nil).
		Times(1)

	// mint coins for denom2
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(mintAmount2)).
		Return(nil).
		Times(1)
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(
			gomock.Any(),
			types.ModuleName,
			signer2,
			sdk.NewCoins(mintAmount2)).
		Return(nil).
		Times(1)

	// expect GetSupply to return a total supply value for each denom
	bankKeeper.EXPECT().
		GetSupply(gomock.Any(), fullDenom1).
		Return(sdk.NewCoin(fullDenom1, cosmosmath.NewInt(100))).
		Times(2)

	bankKeeper.EXPECT().
		GetSupply(gomock.Any(), fullDenom2).
		Return(sdk.NewCoin(fullDenom2, cosmosmath.NewInt(200))).
		Times(2)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create first denom
	createMsg1 := &types.MsgCreateDenom{
		Creator:             creator1,
		SubDenom:            subDenom1,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri1,
		URIHash:             uriHash1,
	}

	respCreate1, err := server.CreateDenom(ctx, createMsg1)
	require.NoError(t, err)
	require.NotNil(t, respCreate1)

	// create second denom
	createMsg2 := &types.MsgCreateDenom{
		Creator:             creator2,
		SubDenom:            subDenom2,
		MintingCap:          mintingCap,
		CanChangeMintingCap: false,
		URI:                 uri2,
		URIHash:             uriHash2,
	}

	respCreate2, err := server.CreateDenom(ctx, createMsg2)
	require.NoError(t, err)
	require.NotNil(t, respCreate2)

	// mint coins for the first denom
	mintMsg1 := types.MsgMintAndSendTokens{
		Signer:    creator1,
		Token:     mintAmount1,
		Recipient: creator1,
	}

	respMintAndSendTokens1, err := server.MintAndSendTokens(ctx, &mintMsg1)
	require.NoError(t, err)
	require.NotNil(t, respMintAndSendTokens1)

	// mint coins for the second denom
	mintMsg2 := types.MsgMintAndSendTokens{
		Signer:    creator2,
		Token:     mintAmount2,
		Recipient: creator2,
	}

	respMintAndSendTokens2, err := server.MintAndSendTokens(ctx, &mintMsg2)
	require.NoError(t, err)
	require.NotNil(t, respMintAndSendTokens2)

	// query denom1
	queryReq1 := &types.QueryGetDenomRequest{
		Denom: fullDenom1,
	}

	queryResp1, err := k.Denom(ctx, queryReq1)
	require.NoError(t, err)
	require.NotNil(t, queryResp1)

	// validate denom1
	require.Equal(t, fullDenom1, queryResp1.Denom)
	require.Equal(t, cosmosmath.NewUint(mintAmount1.Amount.Uint64()), queryResp1.TotalMinted)
	require.Equal(t, cosmosmath.NewUint(100), queryResp1.TotalSupply)
	require.Equal(t, cosmosmath.NewUint(0), queryResp1.TotalBurned)
	require.Equal(t, mintingCap, queryResp1.MintingCap)
	require.Equal(t, mintingCap, queryResp1.MaxSupply)
	require.True(t, queryResp1.CanChangeMintingCap)
	require.Equal(t, creator1, queryResp1.Creator)
	require.Equal(t, creator1, queryResp1.BankAdmin)
	require.Equal(t, creator1, queryResp1.MetadataAdmin)

	// query denom2
	queryReq2 := &types.QueryGetDenomRequest{
		Denom: fullDenom2,
	}

	queryResp2, err := k.Denom(ctx, queryReq2)
	require.NoError(t, err)
	require.NotNil(t, queryResp2)

	// validate denom1
	require.Equal(t, fullDenom2, queryResp2.Denom)
	require.Equal(t, cosmosmath.NewUint(mintAmount2.Amount.Uint64()), queryResp2.TotalMinted)
	require.Equal(t, cosmosmath.NewUint(200), queryResp2.TotalSupply)
	require.Equal(t, cosmosmath.NewUint(0), queryResp2.TotalBurned)
	require.Equal(t, mintingCap, queryResp2.MintingCap)
	require.Equal(t, mintingCap, queryResp2.MaxSupply)
	require.False(t, queryResp2.CanChangeMintingCap)
	require.Equal(t, creator2, queryResp2.Creator)
	require.Equal(t, creator2, queryResp2.BankAdmin)
	require.Equal(t, creator2, queryResp2.MetadataAdmin)
}

func TestQueryDenomGet_Positive_Burn(t *testing.T) {
	// Test case: create two denoms, mint tokens, burn and query them

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // ensure all expectations are met

	// create two sample signers
	creator1 := sample.AccAddress()
	creator2 := sample.AccAddress()

	signer1 := sdk.MustAccAddressFromBech32(creator1)
	signer2 := sdk.MustAccAddressFromBech32(creator2)

	// create two different sub-denoms
	subDenom1 := "abc"
	subDenom2 := "bcd"

	mintingCap := cosmosmath.NewUint(1000000)
	uri1 := "ipfs://uri1"
	uri2 := "ipfs://uri2"
	uriHash1 := "b598d682c86bc9dafc92b33789efc2f94929deabf1cc3fbbb45b8375cba78124"
	uriHash2 := "ab4ff2837b90af0e129b7a3e3af949022e1e3215ae1b2b2f502ce15281d35b32"
	fee := sdk.NewCoin("uzig", cosmosmath.NewInt(1000))

	fullDenom1 := "coin." + creator1 + "." + subDenom1
	fullDenom2 := "coin." + creator2 + "." + subDenom2

	denomMetaData1 := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom1,
			Exponent: 0,
		}},
		Base:    fullDenom1,
		Name:    fullDenom1,
		Symbol:  subDenom1,
		Display: fullDenom1,
		URI:     uri1,
		URIHash: uriHash1,
	}

	denomMetaData2 := banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    fullDenom2,
			Exponent: 0,
		}},
		Base:    fullDenom2,
		Name:    fullDenom2,
		Symbol:  subDenom2,
		Display: fullDenom2,
		URI:     uri2,
		URIHash: uriHash2,
	}

	mintAmount1 := sdk.NewCoin(fullDenom1, cosmosmath.NewInt(100))
	mintAmount2 := sdk.NewCoin(fullDenom2, cosmosmath.NewInt(200))

	burnAmount1 := sdk.NewCoin(fullDenom1, cosmosmath.NewInt(50))
	burnAmount2 := sdk.NewCoin(fullDenom2, cosmosmath.NewInt(200))

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// expectations for creating denom1
	bankKeeper.EXPECT().
		HasSupply(gomock.Any(), subDenom1).
		Return(false).
		Times(1)
	bankKeeper.EXPECT().
		HasBalance(gomock.Any(), signer1, fee).
		Return(true).
		Times(1)
	bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer1, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)
	bankKeeper.EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom1).
		Return(banktypes.Metadata{}, false).
		Times(1)
	bankKeeper.EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData1).
		Return().
		Times(1)

	// expectations for creating denom2
	bankKeeper.EXPECT().
		HasSupply(gomock.Any(), subDenom2).
		Return(false).
		Times(1)
	bankKeeper.EXPECT().
		HasBalance(gomock.Any(), signer2, fee).
		Return(true).
		Times(1)
	bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer2, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)
	bankKeeper.EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom2).
		Return(banktypes.Metadata{}, false).
		Times(1)
	bankKeeper.EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData2).
		Return().
		Times(1)

	// mint coins for denom1
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(mintAmount1)).
		Return(nil).
		Times(1)
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(
			gomock.Any(),
			types.ModuleName,
			signer1,
			sdk.NewCoins(mintAmount1)).
		Return(nil).
		Times(1)

	// mint coins for denom2
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(mintAmount2)).
		Return(nil).
		Times(1)
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(
			gomock.Any(),
			types.ModuleName,
			signer2,
			sdk.NewCoins(mintAmount2)).
		Return(nil).
		Times(1)

	// expect GetSupply to return a total supply value for each denom
	bankKeeper.EXPECT().
		GetSupply(gomock.Any(), fullDenom1).
		Return(sdk.NewCoin(fullDenom1, cosmosmath.NewInt(100))).
		Times(1)

	bankKeeper.EXPECT().
		GetSupply(gomock.Any(), fullDenom2).
		Return(sdk.NewCoin(fullDenom2, cosmosmath.NewInt(200))).
		Times(1)

	// burn coins for denom1
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer1, burnAmount1).
		Return(true).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer1, types.ModuleName, sdk.NewCoins(burnAmount1)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(burnAmount1)).
		Return(nil).
		Times(1)

	// burn coins for denom1
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer2, burnAmount2).
		Return(true).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer2, types.ModuleName, sdk.NewCoins(burnAmount2)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(burnAmount2)).
		Return(nil).
		Times(1)

	// expect GetSupply to return a total supply value for each denom
	bankKeeper.EXPECT().
		GetSupply(gomock.Any(), fullDenom1).
		Return(sdk.NewCoin(fullDenom1, cosmosmath.NewInt(50))).
		Times(1)

	bankKeeper.EXPECT().
		GetSupply(gomock.Any(), fullDenom2).
		Return(sdk.NewCoin(fullDenom2, cosmosmath.NewInt(0))).
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create first denom
	createMsg1 := &types.MsgCreateDenom{
		Creator:             creator1,
		SubDenom:            subDenom1,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri1,
		URIHash:             uriHash1,
	}

	respCreate1, err := server.CreateDenom(ctx, createMsg1)
	require.NoError(t, err)
	require.NotNil(t, respCreate1)

	// create second denom
	createMsg2 := &types.MsgCreateDenom{
		Creator:             creator2,
		SubDenom:            subDenom2,
		MintingCap:          mintingCap,
		CanChangeMintingCap: false,
		URI:                 uri2,
		URIHash:             uriHash2,
	}

	respCreate2, err := server.CreateDenom(ctx, createMsg2)
	require.NoError(t, err)
	require.NotNil(t, respCreate2)

	// mint coins for the first denom
	mintMsg1 := types.MsgMintAndSendTokens{
		Signer:    creator1,
		Token:     mintAmount1,
		Recipient: creator1,
	}

	respMintAndSendTokens1, err := server.MintAndSendTokens(ctx, &mintMsg1)
	require.NoError(t, err)
	require.NotNil(t, respMintAndSendTokens1)

	// mint coins for the second denom
	mintMsg2 := types.MsgMintAndSendTokens{
		Signer:    creator2,
		Token:     mintAmount2,
		Recipient: creator2,
	}

	respMintAndSendTokens2, err := server.MintAndSendTokens(ctx, &mintMsg2)
	require.NoError(t, err)
	require.NotNil(t, respMintAndSendTokens2)

	// create a message to burn tokens for denom1
	burnMsg1 := types.MsgBurnTokens{
		Signer: creator1,
		Token:  burnAmount1,
	}

	respBurnTokens1, err := server.BurnTokens(ctx, &burnMsg1)
	require.NoError(t, err)
	require.NotNil(t, respBurnTokens1)

	expectedRespBurnTokens1 := &types.MsgBurnTokensResponse{
		AmountBurned: burnAmount1,
	}

	require.Equal(t, expectedRespBurnTokens1, respBurnTokens1)

	// create a message to burn tokens for denom2
	burnMsg2 := types.MsgBurnTokens{
		Signer: creator2,
		Token:  burnAmount2,
	}

	respBurnTokens2, err := server.BurnTokens(ctx, &burnMsg2)
	require.NoError(t, err)
	require.NotNil(t, respBurnTokens2)

	expectedRespBurnTokens2 := &types.MsgBurnTokensResponse{
		AmountBurned: burnAmount2,
	}

	require.Equal(t, expectedRespBurnTokens2, respBurnTokens2)

	// query denom1
	queryReq1 := &types.QueryGetDenomRequest{
		Denom: fullDenom1,
	}

	queryResp1, err := k.Denom(ctx, queryReq1)
	require.NoError(t, err)
	require.NotNil(t, queryResp1)

	// validate denom1
	require.Equal(t, fullDenom1, queryResp1.Denom)
	require.Equal(t, cosmosmath.NewUint(mintAmount1.Amount.Uint64()), queryResp1.TotalMinted)
	require.Equal(t, cosmosmath.NewUint(50), queryResp1.TotalSupply)
	require.Equal(t, cosmosmath.NewUint(burnAmount1.Amount.Uint64()), queryResp1.TotalBurned)
	require.Equal(t, mintingCap, queryResp1.MintingCap)
	// new max supply after burn should be mintingCap - totalBurned
	maxSupply1 := mintingCap.Sub(cosmosmath.NewUint(50))
	require.Equal(t, maxSupply1, queryResp1.MaxSupply)
	require.Equal(t, cosmosmath.NewUint(999950), queryResp1.MaxSupply)
	require.True(t, queryResp1.CanChangeMintingCap)
	require.Equal(t, creator1, queryResp1.Creator)
	require.Equal(t, creator1, queryResp1.BankAdmin)
	require.Equal(t, creator1, queryResp1.MetadataAdmin)

	// query denom2
	queryReq2 := &types.QueryGetDenomRequest{
		Denom: fullDenom2,
	}

	queryResp2, err := k.Denom(ctx, queryReq2)
	require.NoError(t, err)
	require.NotNil(t, queryResp2)

	// validate denom1
	require.Equal(t, fullDenom2, queryResp2.Denom)
	require.Equal(t, cosmosmath.NewUint(mintAmount2.Amount.Uint64()), queryResp2.TotalMinted)
	require.Equal(t, cosmosmath.NewUint(0), queryResp2.TotalSupply)
	require.Equal(t, cosmosmath.NewUint(burnAmount2.Amount.Uint64()), queryResp2.TotalBurned)
	require.Equal(t, mintingCap, queryResp2.MintingCap)
	// new max supply after burn should be mintingCap - totalBurned
	maxSupply2 := mintingCap.Sub(cosmosmath.NewUint(200))
	require.Equal(t, maxSupply2, queryResp2.MaxSupply)
	require.Equal(t, cosmosmath.NewUint(999800), queryResp2.MaxSupply)
	require.False(t, queryResp2.CanChangeMintingCap)
	require.Equal(t, creator2, queryResp2.Creator)
	require.Equal(t, creator2, queryResp2.BankAdmin)
	require.Equal(t, creator2, queryResp2.MetadataAdmin)
}

// Negative test cases

func TestQueryDenom_InvalidRequest(t *testing.T) {
	// Test case: query denom with invalid request

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	_, err := k.Denom(ctx, nil)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "invalid request")
}

func TestQueryDenom_InvalidDenomFormat(t *testing.T) {
	// Test case: query a denom with an invalid format

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	req := &types.QueryGetDenomRequest{
		Denom: "invalid!denom",
	}

	_, err := k.Denom(ctx, req)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "invalid denom")
}

func TestQueryDenom_DenomNotFound(t *testing.T) {
	// Test case: query a denom that does not exist in the store

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	creator := sample.AccAddress()
	subDenom := "abc"
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	req := &types.QueryGetDenomRequest{
		Denom: fullDenom,
	}

	_, err := k.Denom(ctx, req)
	require.Error(t, err)
	require.Equal(t, codes.NotFound.String(), status.Code(err).String())
	require.Contains(t, err.Error(), fmt.Sprintf("denom (%s) not found", fullDenom))
}

func TestQueryDenom_DenomAuthNotFound(t *testing.T) {
	// Test case: query a denom where DenomAuth does not exist

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	creator := sample.AccAddress()
	subDenom := "abc"
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	req := &types.QueryGetDenomRequest{
		Denom: fullDenom,
	}

	// simulate that the denom exists but no DenomAuth is found
	k.SetDenom(ctx, types.Denom{
		Denom: fullDenom,
	})

	_, err := k.Denom(ctx, req)
	require.Error(t, err)
	require.Equal(t, codes.NotFound.String(), status.Code(err).String())
	require.Contains(t, err.Error(), fmt.Sprintf("denomAuth (%s) not found", fullDenom))
}

func TestQueryDenom_ValidDenom(t *testing.T) {
	// Test case: query a valid denom

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	creator := sample.AccAddress()

	// Use a valid denom format that will pass validation
	subDenom := "subdenom"
	fullDenom := "coin." + creator + "." + subDenom
	transformedDenom := "coin." + creator + "." + subDenom

	// bank keeper mock
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	bankKeeper.EXPECT().
		GetSupply(gomock.Any(), fullDenom).
		Return(sdk.NewCoin(fullDenom, cosmosmath.NewInt(1000))).
		Times(1)

	// create mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create denom entry with transformed denom
	denom := types.Denom{
		Denom:               transformedDenom,
		Creator:             creator,
		Minted:              cosmosmath.NewUint(1000),
		MintingCap:          cosmosmath.NewUint(10000),
		CanChangeMintingCap: true,
	}
	k.SetDenom(ctx, denom)

	// create denom auth entry with transformed denom
	denomAuth := types.DenomAuth{
		Denom:         transformedDenom,
		BankAdmin:     creator,
		MetadataAdmin: creator,
	}
	k.SetDenomAuth(ctx, denomAuth)

	// query with original denom
	req := &types.QueryGetDenomRequest{Denom: fullDenom}
	resp, err := k.Denom(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, transformedDenom, resp.Denom)
	require.Equal(t, creator, resp.Creator)
	require.Equal(t, creator, resp.BankAdmin)
	require.Equal(t, creator, resp.MetadataAdmin)
}

func TestQueryDenom_InvalidDenom(t *testing.T) {
	// Test case: query with invalid denom
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Test with invalid denom format
	invalidDenom := "invalid'denom"
	req := &types.QueryGetDenomRequest{Denom: invalidDenom}
	resp, err := k.Denom(ctx, req)

	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), invalidDenom) // Error should contain original denom
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestQueryDenom_UnderflowProtection(t *testing.T) {
	// Test case: verify underflow protection when total supply is greater than minted amount
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)
	subDenom := "underflow"
	mintingCap := cosmosmath.NewUint(1000000)
	uri := "ipfs://uri"
	uriHash := "b598d682c86bc9dafc92b33789efc2f94929deabf1cc3fbbb45b8375cba78124"
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

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// expectations for creating denom
	bankKeeper.EXPECT().
		HasSupply(gomock.Any(), subDenom).
		Return(false).
		Times(1)
	bankKeeper.EXPECT().
		HasBalance(gomock.Any(), signer, fee).
		Return(true).
		Times(1)
	bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(fee)).
		Return(nil).
		Times(1)
	bankKeeper.EXPECT().
		GetDenomMetaData(gomock.Any(), fullDenom).
		Return(banktypes.Metadata{}, false).
		Times(1)
	bankKeeper.EXPECT().
		SetDenomMetaData(gomock.Any(), denomMetaData).
		Return().
		Times(1)

	// Set up a scenario where total supply (200) is greater than mint amount (100)
	bankKeeper.EXPECT().
		GetSupply(gomock.Any(), fullDenom).
		Return(sdk.NewCoin(fullDenom, cosmosmath.NewInt(200))).
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, bankKeeper, nil)

	// create a message server
	server := keeper.NewMsgServerImpl(k)

	// create denom
	msg := &types.MsgCreateDenom{
		Creator:             creator,
		SubDenom:            subDenom,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri,
		URIHash:             uriHash,
	}

	resp, err := server.CreateDenom(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Manually set the minted amount to 100 (less than the total supply of 200)
	denom, found := k.GetDenom(ctx, fullDenom)
	require.True(t, found)
	denom.Minted = cosmosmath.NewUint(100)
	k.SetDenom(ctx, denom)

	// Create denom auth
	denomAuth := types.DenomAuth{
		Denom:         fullDenom,
		BankAdmin:     creator,
		MetadataAdmin: creator,
	}
	k.SetDenomAuth(ctx, denomAuth)

	// query denom
	queryReq := &types.QueryGetDenomRequest{
		Denom: fullDenom,
	}

	// Verify that the query doesn't panic
	require.NotPanics(t, func() {
		queryResp, err := k.Denom(ctx, queryReq)
		require.NoError(t, err)
		require.NotNil(t, queryResp)

		// Verify the burned amount is 0 (due to underflow protection)
		require.Equal(t, cosmosmath.ZeroUint(), queryResp.TotalBurned)
		require.Equal(t, cosmosmath.NewUint(100), queryResp.TotalMinted)
		require.Equal(t, cosmosmath.NewUint(200), queryResp.TotalSupply)
	})
}
