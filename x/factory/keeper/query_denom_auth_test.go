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

func TestQueryListDenomAuth_Positive(t *testing.T) {
	// Test case: create two DenomAuth entries and query all

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create two sample denoms
	creator1 := sample.AccAddress()
	creator2 := sample.AccAddress()

	subDenom1 := "denom1"
	subDenom2 := "denom2"

	fullDenom1 := "coin" + types.FactoryDenomDelimiterChar + creator1 + types.FactoryDenomDelimiterChar + subDenom1
	fullDenom2 := "coin" + types.FactoryDenomDelimiterChar + creator2 + types.FactoryDenomDelimiterChar + subDenom2

	// create mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// create first denom authorization
	denomAuth1 := types.DenomAuth{
		Denom:         fullDenom1,
		BankAdmin:     creator1,
		MetadataAdmin: creator1,
	}
	k.SetDenomAuth(ctx, denomAuth1)

	// create second denom authorization
	denomAuth2 := types.DenomAuth{
		Denom:         fullDenom2,
		BankAdmin:     creator2,
		MetadataAdmin: creator2,
	}
	k.SetDenomAuth(ctx, denomAuth2)

	// query all denom auths
	req := &types.QueryAllDenomAuthRequest{
		Pagination: &query.PageRequest{},
	}
	resp, err := k.ListDenomAuth(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.DenomAuth, 2)

	// sort by subDenom for consistent ordering
	sort.Slice(resp.DenomAuth, func(i, j int) bool {
		partsI := strings.Split(resp.DenomAuth[i].Denom, ".")
		partsJ := strings.Split(resp.DenomAuth[j].Denom, ".")
		return partsI[2] < partsJ[2]
	})

	// validate first entry (denom1)
	require.Equal(t, fullDenom1, resp.DenomAuth[0].Denom)
	require.Equal(t, creator1, resp.DenomAuth[0].BankAdmin)
	require.Equal(t, creator1, resp.DenomAuth[0].MetadataAdmin)

	// validate second entry (denom2)
	require.Equal(t, fullDenom2, resp.DenomAuth[1].Denom)
	require.Equal(t, creator2, resp.DenomAuth[1].BankAdmin)
	require.Equal(t, creator2, resp.DenomAuth[1].MetadataAdmin)
}

func TestQueryDenomAuth_Positive(t *testing.T) {
	// Test case: query a single denom authorization

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a sample denom
	creator := sample.AccAddress()
	subDenom := "denom1"
	fullDenom := "coin" + types.FactoryDenomDelimiterChar + creator + types.FactoryDenomDelimiterChar + subDenom

	// create mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// create denom authorization entry
	denomAuth := types.DenomAuth{
		Denom:         fullDenom,
		BankAdmin:     creator,
		MetadataAdmin: creator,
	}
	k.SetDenomAuth(ctx, denomAuth)

	// query single denom auth
	req := &types.QueryGetDenomAuthRequest{Denom: fullDenom}
	resp, err := k.DenomAuth(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, fullDenom, resp.DenomAuth.Denom)
	require.Equal(t, creator, resp.DenomAuth.BankAdmin)
	require.Equal(t, creator, resp.DenomAuth.MetadataAdmin)
}

func TestQueryDenomAuth_CreateTwoDenomsAndCheckAuth(t *testing.T) {
	// Test case: create two denoms and verify their authorization data

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create two sample creators
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

	// query all denom auths
	authReq := &types.QueryAllDenomAuthRequest{
		Pagination: &query.PageRequest{},
	}

	authResp, err := k.ListDenomAuth(ctx, authReq)
	require.NoError(t, err)
	require.NotNil(t, authResp)
	require.Len(t, authResp.DenomAuth, 2)

	// Sort by subDenom to ensure order
	sort.Slice(authResp.DenomAuth, func(i, j int) bool {
		partsI := strings.Split(authResp.DenomAuth[i].Denom, ".")
		partsJ := strings.Split(authResp.DenomAuth[j].Denom, ".")
		return partsI[2] < partsJ[2]
	})

	// validate first denom auth (abc)
	require.Equal(t, fullDenom1, authResp.DenomAuth[0].Denom)
	require.Equal(t, creator1, authResp.DenomAuth[0].BankAdmin)
	require.Equal(t, creator1, authResp.DenomAuth[0].MetadataAdmin)

	// validate second denom auth (bcd)
	require.Equal(t, fullDenom2, authResp.DenomAuth[1].Denom)
	require.Equal(t, creator2, authResp.DenomAuth[1].BankAdmin)
	require.Equal(t, creator2, authResp.DenomAuth[1].MetadataAdmin)

	// query a single denom auth (fullDenom1)
	singleAuthReq := &types.QueryGetDenomAuthRequest{
		Denom: fullDenom1,
	}
	singleAuthResp, err := k.DenomAuth(ctx, singleAuthReq)
	require.NoError(t, err)
	require.NotNil(t, singleAuthResp)
	require.Equal(t, fullDenom1, singleAuthResp.DenomAuth.Denom)
	require.Equal(t, creator1, singleAuthResp.DenomAuth.BankAdmin)
	require.Equal(t, creator1, singleAuthResp.DenomAuth.MetadataAdmin)

	// query a single denom auth (fullDenom2)
	singleAuthReq2 := &types.QueryGetDenomAuthRequest{
		Denom: fullDenom2,
	}
	singleAuthResp2, err := k.DenomAuth(ctx, singleAuthReq2)
	require.NoError(t, err)
	require.NotNil(t, singleAuthResp2)
	require.Equal(t, fullDenom2, singleAuthResp2.DenomAuth.Denom)
	require.Equal(t, creator2, singleAuthResp2.DenomAuth.BankAdmin)
	require.Equal(t, creator2, singleAuthResp2.DenomAuth.MetadataAdmin)
}

func TestQueryListDenomAuth_EmptyStore(t *testing.T) {
	// Test case: query all denom auth if the store is empty

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// query all denom auths
	req := &types.QueryAllDenomAuthRequest{
		Pagination: &query.PageRequest{},
	}
	resp, err := k.ListDenomAuth(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.DenomAuth, 0)
}

// Negative test cases

func TestQueryListDenomAuth_InvalidRequest(t *testing.T) {
	// Test case: query all denoms auth with invalid request

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	_, err := k.ListDenomAuth(ctx, nil)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "invalid request")
}

func TestQueryDenomAuth_InvalidRequest(t *testing.T) {
	// Test case: query denom auth with invalid request

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	_, err := k.DenomAuth(ctx, nil)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "invalid request")
}

func TestQueryDenomAuth_EmptyDenom(t *testing.T) {
	// Test case: query denom auth with empty denom

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	req := &types.QueryGetDenomAuthRequest{
		Denom: "", // empty denom
	}

	_, err := k.DenomAuth(ctx, req)

	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "invalid denom")
}

func TestQueryDenomAuth_InvalidDenomFormat(t *testing.T) {
	// Test case: query denom auth with invalid denom format

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	req := &types.QueryGetDenomAuthRequest{
		Denom: "invalid!@#denom", // contains special characters
	}

	_, err := k.DenomAuth(ctx, req)

	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "invalid denom")
}

func TestQueryDenomAuth_NonExistingDenom(t *testing.T) {
	// Test case: query a non-existing denom auth

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	req := &types.QueryGetDenomAuthRequest{
		Denom: "coin" + types.FactoryDenomDelimiterChar + "zig1xyz123" + types.FactoryDenomDelimiterChar + "unknownDenom",
	}

	_, err := k.DenomAuth(ctx, req)

	require.Error(t, err)
	require.Equal(t, codes.NotFound.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "not found")
}

func TestQueryDenomAuth_ValidDenom(t *testing.T) {
	// Test case: query a valid denom
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	creator := sample.AccAddress()
	// Use a valid denom format that will pass validation
	subDenom := "subdenom"
	fullDenom := "coin." + creator + "." + subDenom
	transformedDenom := "coin." + creator + "." + subDenom

	// create mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// create denom authorization entry with transformed denom
	denomAuth := types.DenomAuth{
		Denom:         transformedDenom,
		BankAdmin:     creator,
		MetadataAdmin: creator,
	}
	k.SetDenomAuth(ctx, denomAuth)

	// query with original denom
	req := &types.QueryGetDenomAuthRequest{Denom: fullDenom}
	resp, err := k.DenomAuth(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, transformedDenom, resp.DenomAuth.Denom)
	require.Equal(t, creator, resp.DenomAuth.BankAdmin)
	require.Equal(t, creator, resp.DenomAuth.MetadataAdmin)
}

func TestQueryDenomAuth_InvalidDenom(t *testing.T) {
	// Test case: query with invalid denom
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Test with invalid denom format
	invalidDenom := "invalid'denom"
	req := &types.QueryGetDenomAuthRequest{Denom: invalidDenom}
	resp, err := k.DenomAuth(ctx, req)

	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), invalidDenom) // Error should contain original denom
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestQueryDenomAuth_PaginateError(t *testing.T) {
	// Test case: query denom auths with invalid pagination causing Paginate error

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	req := &types.QueryAllDenomAuthRequest{
		Pagination: &query.PageRequest{
			Key:    []byte("invalidkey"),
			Offset: 1, // Both Key and Offset set to trigger error
		},
	}

	resp, err := k.ListDenomAuth(ctx, req)

	require.Nil(t, resp)
	require.Error(t, err)
	require.Equal(t, codes.Internal.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "Internal desc = invalid request, either offset or key is expected, got both")
}
