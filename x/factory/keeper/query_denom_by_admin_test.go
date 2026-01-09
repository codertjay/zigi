package keeper_test

import (
	"fmt"
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

func TestQueryDenomByAdmin_Positive(t *testing.T) {
	// Test case: create multiple denoms with different admins and query by admin

	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create two different admin addresses
	admin1 := sample.AccAddress()
	admin2 := sample.AccAddress()

	// Create different sub-denoms
	subDenom1 := "denom1"
	subDenom2 := "denom2"
	subDenom3 := "denom3"

	fullDenom1 := "coin" + types.FactoryDenomDelimiterChar + admin1 + types.FactoryDenomDelimiterChar + subDenom1
	fullDenom2 := "coin" + types.FactoryDenomDelimiterChar + admin1 + types.FactoryDenomDelimiterChar + subDenom2
	fullDenom3 := "coin" + types.FactoryDenomDelimiterChar + admin2 + types.FactoryDenomDelimiterChar + subDenom3 // Different admin

	// Create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Create first denom (admin1)
	denomAuth1 := types.DenomAuth{
		Denom:         fullDenom1,
		BankAdmin:     admin1,
		MetadataAdmin: admin1,
	}
	k.SetDenomAuth(ctx, denomAuth1)
	k.AddDenomToAdminDenomAuthList(ctx, admin1, fullDenom1)

	// Create second denom (admin1)
	denomAuth2 := types.DenomAuth{
		Denom:         fullDenom2,
		BankAdmin:     admin1,
		MetadataAdmin: admin1,
	}
	k.SetDenomAuth(ctx, denomAuth2)
	k.AddDenomToAdminDenomAuthList(ctx, admin1, fullDenom2)

	// Create third denom (admin2)
	denomAuth3 := types.DenomAuth{
		Denom:         fullDenom3,
		BankAdmin:     admin2,
		MetadataAdmin: admin2,
	}
	k.SetDenomAuth(ctx, denomAuth3)
	k.AddDenomToAdminDenomAuthList(ctx, admin2, fullDenom3)

	// Query denoms owned by admin1
	req := &types.QueryDenomByAdminRequest{
		Admin:      admin1,
		Pagination: &query.PageRequest{},
	}

	resp, err := k.DenomsByAdmin(ctx, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Denoms, 2)

	// Ensure the returned denoms match the expected ones
	expectedDenoms := map[string]bool{
		fullDenom1: true,
		fullDenom2: true,
	}

	// Check if both expected denoms exist in response
	for _, denom := range resp.Denoms {
		require.True(t, expectedDenoms[denom], "denom in response")
	}

	// Query denoms owned by admin2
	req2 := &types.QueryDenomByAdminRequest{
		Admin:      admin2,
		Pagination: &query.PageRequest{},
	}

	resp2, err := k.DenomsByAdmin(ctx, req2)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, resp2)
	require.Len(t, resp2.Denoms, 1)

	// Ensure only the correct denom is returned
	require.Equal(t, fullDenom3, resp2.Denoms[0])
}

func TestQueryDenomByAdmin_CreateTwoDenoms(t *testing.T) {
	// Test case: Create two denoms and query denoms by admin

	// create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create two sample creators
	creator1 := sample.AccAddress()

	signer1 := sdk.MustAccAddressFromBech32(creator1)

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
	fullDenom2 := "coin" + types.FactoryDenomDelimiterChar + creator1 + types.FactoryDenomDelimiterChar + subDenom2

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
		HasBalance(gomock.Any(), signer1, fee).
		Return(true).
		Times(1)
	bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer1, types.ModuleName, sdk.NewCoins(fee)).
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
		Creator:             creator1,
		SubDenom:            subDenom2,
		MintingCap:          mintingCap,
		CanChangeMintingCap: false,
		URI:                 uri2,
		URIHash:             uriHash2,
	}

	respCreate2, err := server.CreateDenom(ctx, createMsg2)
	require.NoError(t, err)
	require.NotNil(t, respCreate2)

	// query denom by admin
	req := &types.QueryDenomByAdminRequest{
		Admin:      creator1,
		Pagination: &query.PageRequest{},
	}

	resp, err := k.DenomsByAdmin(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Denoms, 2)

	// sort by subDenom to ensure order
	sort.Slice(resp.Denoms, func(i, j int) bool {
		partsI := strings.Split(resp.Denoms[i], ".")
		partsJ := strings.Split(resp.Denoms[j], ".")
		return partsI[2] < partsJ[2]
	})

	// ensure the returned denoms match the expected ones
	require.Equal(t, fullDenom1, resp.Denoms[0])
	require.Equal(t, fullDenom2, resp.Denoms[1])
}

func TestQueryDenomByAdmin_EmptyStore(t *testing.T) {
	// Test case: query all denom by admin the store is empty

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// query all denom auths
	req := &types.QueryDenomByAdminRequest{
		Admin:      sample.AccAddress(),
		Pagination: &query.PageRequest{},
	}
	resp, err := k.DenomsByAdmin(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Denoms, 0)
}

// Pagination edge case tests

func TestQueryDenomByAdmin_PaginationLimit(t *testing.T) {
	// Test case: Create many denoms and test pagination limit
	// This ensures that when pagination limits to N items, only N items are returned

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	admin := sample.AccAddress()
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Create 25 denoms for the same admin
	totalDenoms := 25
	createdDenoms := make([]string, totalDenoms)

	for i := 0; i < totalDenoms; i++ {
		subDenom := fmt.Sprintf("denom%d", i)
		fullDenom := "coin" + types.FactoryDenomDelimiterChar + admin + types.FactoryDenomDelimiterChar + subDenom

		denomAuth := types.DenomAuth{
			Denom:         fullDenom,
			BankAdmin:     admin,
			MetadataAdmin: admin,
		}
		k.SetDenomAuth(ctx, denomAuth)
		k.AddDenomToAdminDenomAuthList(ctx, admin, fullDenom)
		createdDenoms[i] = fullDenom
	}

	// Test with limit of 10
	limit := uint64(10)
	req := &types.QueryDenomByAdminRequest{
		Admin: admin,
		Pagination: &query.PageRequest{
			Limit:      limit,
			CountTotal: true,
		},
	}

	resp, err := k.DenomsByAdmin(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Should return exactly 10 items
	require.Len(t, resp.Denoms, int(limit))
	require.NotNil(t, resp.Pagination)
	require.True(t, resp.Pagination.NextKey != nil, "should have next key for pagination")
	require.Equal(t, uint64(totalDenoms), resp.Pagination.Total)
}

func TestQueryDenomByAdmin_PaginationOffset(t *testing.T) {
	// Test case: Test pagination with offset to skip items

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	admin := sample.AccAddress()
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Create 15 denoms for the same admin
	totalDenoms := 15
	createdDenoms := make([]string, totalDenoms)

	for i := 0; i < totalDenoms; i++ {
		subDenom := fmt.Sprintf("denom%d", i)
		fullDenom := "coin" + types.FactoryDenomDelimiterChar + admin + types.FactoryDenomDelimiterChar + subDenom

		denomAuth := types.DenomAuth{
			Denom:         fullDenom,
			BankAdmin:     admin,
			MetadataAdmin: admin,
		}
		k.SetDenomAuth(ctx, denomAuth)
		k.AddDenomToAdminDenomAuthList(ctx, admin, fullDenom)
		createdDenoms[i] = fullDenom
	}

	// Sort the created denoms to ensure consistent ordering
	sort.Strings(createdDenoms)

	// Test with offset of 5 and limit of 7
	offset := uint64(5)
	limit := uint64(7)
	req := &types.QueryDenomByAdminRequest{
		Admin: admin,
		Pagination: &query.PageRequest{
			Offset:     offset,
			Limit:      limit,
			CountTotal: true,
		},
	}

	resp, err := k.DenomsByAdmin(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Should return 7 items (limit)
	require.Len(t, resp.Denoms, int(limit))
	require.NotNil(t, resp.Pagination)
	require.Equal(t, uint64(totalDenoms), resp.Pagination.Total)

	// The returned denoms should be the ones starting from index 5
	expectedDenoms := createdDenoms[offset : offset+limit]
	require.Equal(t, expectedDenoms, resp.Denoms)
}

func TestQueryDenomByAdmin_PaginationWithNextKey(t *testing.T) {
	// Test case: Test pagination using NextKey from previous response

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	admin := sample.AccAddress()
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Create 20 denoms for the same admin
	totalDenoms := 20
	createdDenoms := make([]string, totalDenoms)

	for i := 0; i < totalDenoms; i++ {
		subDenom := fmt.Sprintf("denom%d", i)
		fullDenom := "coin" + types.FactoryDenomDelimiterChar + admin + types.FactoryDenomDelimiterChar + subDenom

		denomAuth := types.DenomAuth{
			Denom:         fullDenom,
			BankAdmin:     admin,
			MetadataAdmin: admin,
		}
		k.SetDenomAuth(ctx, denomAuth)
		k.AddDenomToAdminDenomAuthList(ctx, admin, fullDenom)
		createdDenoms[i] = fullDenom
	}

	// Sort the created denoms to ensure consistent ordering
	sort.Strings(createdDenoms)

	// First page: get first 8 items
	limit := uint64(8)
	req1 := &types.QueryDenomByAdminRequest{
		Admin: admin,
		Pagination: &query.PageRequest{
			Limit:      limit,
			CountTotal: true,
		},
	}

	resp1, err := k.DenomsByAdmin(ctx, req1)
	require.NoError(t, err)
	require.NotNil(t, resp1)
	require.Len(t, resp1.Denoms, int(limit))
	require.NotNil(t, resp1.Pagination.NextKey, "should have next key")

	// Second page: use NextKey from first response
	req2 := &types.QueryDenomByAdminRequest{
		Admin: admin,
		Pagination: &query.PageRequest{
			Key:   resp1.Pagination.NextKey,
			Limit: limit,
		},
	}

	resp2, err := k.DenomsByAdmin(ctx, req2)
	require.NoError(t, err)
	require.NotNil(t, resp2)
	require.Len(t, resp2.Denoms, int(limit))

	// Third page: get remaining items
	req3 := &types.QueryDenomByAdminRequest{
		Admin: admin,
		Pagination: &query.PageRequest{
			Key:   resp2.Pagination.NextKey,
			Limit: limit,
		},
	}

	resp3, err := k.DenomsByAdmin(ctx, req3)
	require.NoError(t, err)
	require.NotNil(t, resp3)
	require.Len(t, resp3.Denoms, totalDenoms-2*int(limit)) // Remaining items

	// Verify all items are returned across all pages
	allReturnedDenoms := append(resp1.Denoms, resp2.Denoms...)
	allReturnedDenoms = append(allReturnedDenoms, resp3.Denoms...)
	require.Len(t, allReturnedDenoms, totalDenoms)

	// Verify no duplicates
	denomSet := make(map[string]bool)
	for _, denom := range allReturnedDenoms {
		require.False(t, denomSet[denom], "duplicate denom found: %s", denom)
		denomSet[denom] = true
	}
}

func TestQueryDenomByAdmin_PaginationLimitExceedsTotal(t *testing.T) {
	// Test case: Test when pagination limit exceeds the total number of items

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	admin := sample.AccAddress()
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Create only 5 denoms
	totalDenoms := 5
	createdDenoms := make([]string, totalDenoms)

	for i := 0; i < totalDenoms; i++ {
		subDenom := fmt.Sprintf("denom%d", i)
		fullDenom := "coin" + types.FactoryDenomDelimiterChar + admin + types.FactoryDenomDelimiterChar + subDenom

		denomAuth := types.DenomAuth{
			Denom:         fullDenom,
			BankAdmin:     admin,
			MetadataAdmin: admin,
		}
		k.SetDenomAuth(ctx, denomAuth)
		k.AddDenomToAdminDenomAuthList(ctx, admin, fullDenom)
		createdDenoms[i] = fullDenom
	}

	// Request with limit of 20 (more than total items)
	limit := uint64(20)
	req := &types.QueryDenomByAdminRequest{
		Admin: admin,
		Pagination: &query.PageRequest{
			Limit:      limit,
			CountTotal: true,
		},
	}

	resp, err := k.DenomsByAdmin(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Should return only the available items (5)
	require.Len(t, resp.Denoms, totalDenoms)
	require.NotNil(t, resp.Pagination)
	require.Nil(t, resp.Pagination.NextKey, "should not have next key when all items returned")
	require.Equal(t, uint64(totalDenoms), resp.Pagination.Total)
}

func TestQueryDenomByAdmin_PaginationOffsetExceedsTotal(t *testing.T) {
	// Test case: Test when pagination offset exceeds the total number of items

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	admin := sample.AccAddress()
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Create only 3 denoms
	totalDenoms := 3
	createdDenoms := make([]string, totalDenoms)

	for i := 0; i < totalDenoms; i++ {
		subDenom := fmt.Sprintf("denom%d", i)
		fullDenom := "coin" + types.FactoryDenomDelimiterChar + admin + types.FactoryDenomDelimiterChar + subDenom

		denomAuth := types.DenomAuth{
			Denom:         fullDenom,
			BankAdmin:     admin,
			MetadataAdmin: admin,
		}
		k.SetDenomAuth(ctx, denomAuth)
		k.AddDenomToAdminDenomAuthList(ctx, admin, fullDenom)
		createdDenoms[i] = fullDenom
	}

	// Request with offset of 10 (more than total items)
	offset := uint64(10)
	limit := uint64(5)
	req := &types.QueryDenomByAdminRequest{
		Admin: admin,
		Pagination: &query.PageRequest{
			Offset:     offset,
			Limit:      limit,
			CountTotal: true,
		},
	}

	resp, err := k.DenomsByAdmin(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Should return empty list
	require.Len(t, resp.Denoms, 0)
	require.NotNil(t, resp.Pagination)
	require.Equal(t, uint64(totalDenoms), resp.Pagination.Total)
}

func TestQueryDenomByAdmin_PaginationCountTotal(t *testing.T) {
	// Test case: Test pagination with count_total flag

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	admin := sample.AccAddress()
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Create 12 denoms
	totalDenoms := 12
	createdDenoms := make([]string, totalDenoms)

	for i := 0; i < totalDenoms; i++ {
		subDenom := fmt.Sprintf("denom%d", i)
		fullDenom := "coin" + types.FactoryDenomDelimiterChar + admin + types.FactoryDenomDelimiterChar + subDenom

		denomAuth := types.DenomAuth{
			Denom:         fullDenom,
			BankAdmin:     admin,
			MetadataAdmin: admin,
		}
		k.SetDenomAuth(ctx, denomAuth)
		k.AddDenomToAdminDenomAuthList(ctx, admin, fullDenom)
		createdDenoms[i] = fullDenom
	}

	// Test with count_total = true
	limit := uint64(5)
	req := &types.QueryDenomByAdminRequest{
		Admin: admin,
		Pagination: &query.PageRequest{
			Limit:      limit,
			CountTotal: true,
		},
	}

	resp, err := k.DenomsByAdmin(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Should return 5 items and total count
	require.Len(t, resp.Denoms, int(limit))
	require.NotNil(t, resp.Pagination)
	require.Equal(t, uint64(totalDenoms), resp.Pagination.Total)
}

func TestQueryDenomByAdmin_PaginationReverse(t *testing.T) {
	// Test case: Test pagination with reverse flag

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	admin := sample.AccAddress()
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	// Create 8 denoms
	totalDenoms := 8
	createdDenoms := make([]string, totalDenoms)

	for i := 0; i < totalDenoms; i++ {
		subDenom := fmt.Sprintf("denom%d", i)
		fullDenom := "coin" + types.FactoryDenomDelimiterChar + admin + types.FactoryDenomDelimiterChar + subDenom

		denomAuth := types.DenomAuth{
			Denom:         fullDenom,
			BankAdmin:     admin,
			MetadataAdmin: admin,
		}
		k.SetDenomAuth(ctx, denomAuth)
		k.AddDenomToAdminDenomAuthList(ctx, admin, fullDenom)
		createdDenoms[i] = fullDenom
	}

	// Sort the created denoms to ensure consistent ordering
	sort.Strings(createdDenoms)

	// Test with reverse = true
	limit := uint64(4)
	req := &types.QueryDenomByAdminRequest{
		Admin: admin,
		Pagination: &query.PageRequest{
			Limit:   limit,
			Reverse: true,
		},
	}

	resp, err := k.DenomsByAdmin(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Should return 4 items in reverse order
	require.Len(t, resp.Denoms, int(limit))
	require.NotNil(t, resp.Pagination)

	// Verify reverse order (last 4 items in reverse)
	expectedReverse := make([]string, limit)
	for i := 0; i < int(limit); i++ {
		expectedReverse[i] = createdDenoms[totalDenoms-1-i]
	}
	require.Equal(t, expectedReverse, resp.Denoms)
}

// Negative test cases

func TestQueryDenomByAdmin_InvalidRequest(t *testing.T) {
	// Test case: query denoms by admin with invalid request

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	_, err := k.DenomsByAdmin(ctx, nil)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "invalid request")
}

func TestQueryDenomsByAdmin_PaginateError(t *testing.T) {
	// Test case: query denoms by admin with invalid pagination causing Paginate error

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a mock keeper
	k, ctx := keepertest.FactoryKeeper(t, nil, nil)

	req := &types.QueryDenomByAdminRequest{
		Admin: sample.AccAddress(),
		Pagination: &query.PageRequest{
			Key:    []byte("invalidkey"),
			Offset: 1, // Both Key and Offset set to trigger error
		},
	}

	resp, err := k.DenomsByAdmin(ctx, req)

	require.Nil(t, resp)
	require.Error(t, err)
	require.Equal(t, codes.Internal.String(), status.Code(err).String())
	require.Contains(t, err.Error(), "Internal desc = invalid request, either offset or key is expected, got both")
}
