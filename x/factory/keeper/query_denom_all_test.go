package keeper_test

import (
	"sort"
	"strings"
	"testing"

	cosmosmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
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

func TestQueryDenomAll_CreateTwoDenomsAndQueryAll(t *testing.T) {
	// Test case: create two denoms and query all

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // ensure all expectations are met

	// create two sample signers
	creator1 := sample.AccAddress()
	creator2 := sample.AccAddress()

	signer1 := sdk.MustAccAddressFromBech32(creator1)
	signer2 := sdk.MustAccAddressFromBech32(creator2)

	// create two different sub-denoms
	subDenom1 := "denom1"
	subDenom2 := "denom2"

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

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// expectations for denom1
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

	// expectations for denom2
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

	// expect GetSupply to return a total supply value for each denom
	// not minted any coins yet so total supply should be 0
	bankKeeper.EXPECT().
		GetSupply(gomock.Any(), fullDenom1).
		Return(sdk.NewCoin(fullDenom1, cosmosmath.NewInt(0))).
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
	msg1 := &types.MsgCreateDenom{
		Creator:             creator1,
		SubDenom:            subDenom1,
		MintingCap:          mintingCap,
		CanChangeMintingCap: true,
		URI:                 uri1,
		URIHash:             uriHash1,
	}

	resp1, err := server.CreateDenom(ctx, msg1)
	require.NoError(t, err)
	require.NotNil(t, resp1)

	// create second denom
	msg2 := &types.MsgCreateDenom{
		Creator:             creator2,
		SubDenom:            subDenom2,
		MintingCap:          mintingCap,
		CanChangeMintingCap: false,
		URI:                 uri2,
		URIHash:             uriHash2,
	}

	resp2, err := server.CreateDenom(ctx, msg2)
	require.NoError(t, err)
	require.NotNil(t, resp2)

	// query all denoms
	queryReq := &types.QueryAllDenomRequest{
		Pagination: &query.PageRequest{},
	}

	queryResp, err := k.DenomAll(ctx, queryReq)
	require.NoError(t, err)
	require.NotNil(t, queryResp)

	// validate response contains the two denoms
	require.Len(t, queryResp.Denom, 2)

	// ensure both denoms exist in response
	foundDenoms := map[string]bool{
		queryResp.Denom[0].Denom: true,
		queryResp.Denom[1].Denom: true,
	}

	require.True(t, foundDenoms[fullDenom1], "Denom1 found")
	require.True(t, foundDenoms[fullDenom2], "Denom2 found")
}

func TestQueryDenomAll_Positive(t *testing.T) {
	// Test case: create two denoms, mint tokens and query all denoms

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

	// query all denoms
	queryReq := &types.QueryAllDenomRequest{
		Pagination: &query.PageRequest{},
	}

	queryResp, err := k.DenomAll(ctx, queryReq)
	require.NoError(t, err)
	require.NotNil(t, queryResp)

	// validate response contains the two denoms
	require.Len(t, queryResp.Denom, 2)

	// ensure both denoms exist in response
	foundDenoms := map[string]bool{
		queryResp.Denom[0].Denom: true,
		queryResp.Denom[1].Denom: true,
	}

	require.True(t, foundDenoms[fullDenom1], "Denom1 found")
	require.True(t, foundDenoms[fullDenom2], "Denom2 found")

	// ensure the denoms are sorted by subdenom name to guarantee order
	sort.Slice(queryResp.Denom, func(i, j int) bool {
		// extract subDenom by splitting on "/"
		partsI := strings.Split(queryResp.Denom[i].Denom, ".")
		partsJ := strings.Split(queryResp.Denom[j].Denom, ".")

		// ensure a correct format before sorting
		if len(partsI) < 3 || len(partsJ) < 3 {
			return queryResp.Denom[i].Denom < queryResp.Denom[j].Denom
		}

		// compare subDenoms
		return partsI[2] < partsJ[2]
	})

	// check first denom (should be fullDenom1)
	require.Equal(t, creator1, queryResp.Denom[0].Creator)
	require.Equal(t, fullDenom1, queryResp.Denom[0].Denom)
	require.Equal(t, mintingCap, queryResp.Denom[0].MintingCap)
	require.Equal(t, mintingCap, queryResp.Denom[0].MaxSupply)
	require.True(t, queryResp.Denom[0].CanChangeMintingCap)
	require.Equal(t, cosmosmath.NewUint(mintAmount1.Amount.Uint64()), queryResp.Denom[0].TotalMinted)
	require.Equal(t, cosmosmath.NewUint(100), queryResp.Denom[0].TotalSupply)
	require.Equal(t, cosmosmath.NewUint(0), queryResp.Denom[0].TotalBurned)

	// check second denom (should be fullDenom2)
	require.Equal(t, creator2, queryResp.Denom[1].Creator)
	require.Equal(t, fullDenom2, queryResp.Denom[1].Denom)
	require.Equal(t, mintingCap, queryResp.Denom[1].MintingCap)
	require.Equal(t, mintingCap, queryResp.Denom[1].MaxSupply)
	require.False(t, queryResp.Denom[1].CanChangeMintingCap)
	require.Equal(t, cosmosmath.NewUint(mintAmount2.Amount.Uint64()), queryResp.Denom[1].TotalMinted)
	require.Equal(t, cosmosmath.NewUint(200), queryResp.Denom[1].TotalSupply)
	require.Equal(t, cosmosmath.NewUint(0), queryResp.Denom[1].TotalBurned)
}

func TestQueryDenomAll_EmptyStore(t *testing.T) {
	// Test case: query all denoms if the store is empty

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	req := &types.QueryAllDenomRequest{
		Pagination: &query.PageRequest{},
	}

	resp, err := k.DenomAll(ctx, req)

	require.NoError(t, err)
	require.Len(t, resp.Denom, 0) // expect an empty list
}

// Negative test cases

func TestQueryDenomAll_InvalidRequest(t *testing.T) {
	// Test case: query all denoms with invalid request

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	_, err := k.DenomAll(ctx, nil)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "invalid request")
}

func TestQueryDenomAll_UnderflowProtection(t *testing.T) {
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

	// Set up a scenario where total supply (200) is greater than the minted amount (100)
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

	// query all denoms
	queryReq := &types.QueryAllDenomRequest{
		Pagination: &query.PageRequest{},
	}

	// Verify that the query doesn't panic
	require.NotPanics(t, func() {
		queryResp, err := k.DenomAll(ctx, queryReq)
		require.NoError(t, err)
		require.NotNil(t, queryResp)
		require.Len(t, queryResp.Denom, 1)

		// Verify that the burned amount is 0 (due to underflow protection)
		require.Equal(t, cosmosmath.ZeroUint(), queryResp.Denom[0].TotalBurned)
		require.Equal(t, cosmosmath.NewUint(100), queryResp.Denom[0].TotalMinted)
		require.Equal(t, cosmosmath.NewUint(200), queryResp.Denom[0].TotalSupply)
	})
}

func TestQueryDenomAll_MaxSupplyZero(t *testing.T) {
	// Test case: verify max supply is zero when the minting cap is less than or equal to total burned

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)
	subDenom := "underflow"
	mintingCap := cosmosmath.NewUint(100)
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
		Return(sdk.NewCoin(fullDenom, cosmosmath.NewInt(50))).
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

	// Manually set the minted amount to 200 (greater than the total supply + minting cap)
	denom, found := k.GetDenom(ctx, fullDenom)
	require.True(t, found)
	denom.Minted = cosmosmath.NewUint(200)
	k.SetDenom(ctx, denom)

	// query all denoms
	queryReq := &types.QueryAllDenomRequest{
		Pagination: &query.PageRequest{},
	}

	// Verify that the query doesn't panic
	require.NotPanics(t, func() {
		queryResp, err := k.DenomAll(ctx, queryReq)
		require.NoError(t, err)
		require.NotNil(t, queryResp)
		require.Len(t, queryResp.Denom, 1)

		// Verify that max supply is 0 (due to underflow protection)
		require.Equal(t, cosmosmath.ZeroUint(), queryResp.Denom[0].MaxSupply)
		require.Equal(t, cosmosmath.NewUint(150), queryResp.Denom[0].TotalBurned)
		require.Equal(t, cosmosmath.NewUint(200), queryResp.Denom[0].TotalMinted)
		require.Equal(t, cosmosmath.NewUint(50), queryResp.Denom[0].TotalSupply)
	})
}

func TestQueryDenomAll_PaginateError(t *testing.T) {
	// Test case: query all denoms with invalid pagination causing Paginate error

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	req := &types.QueryAllDenomRequest{
		Pagination: &query.PageRequest{
			Key:    []byte("invalidkey"),
			Offset: 1, // Both Key and Offset set to trigger error
		},
	}

	resp, err := k.DenomAll(ctx, req)

	require.Nil(t, resp)
	require.Error(t, err)
	require.Equal(t, codes.Internal.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "Internal desc = invalid request, either offset or key is expected, got both")
}
