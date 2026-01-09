package keeper_test

import (
	"context"
	"math"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/nullify"
	"zigchain/testutil/sample"
	"zigchain/x/dex/keeper"
	"zigchain/x/dex/types"
	"zigchain/zutils/constants"

	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNPool(keeper keeper.Keeper, ctx context.Context, n int) []types.Pool {
	items := make([]types.Pool, n)
	for i := range items {
		items[i].PoolId = constants.PoolPrefix + strconv.Itoa(i)

		keeper.SetPool(ctx, items[i])
	}
	return items
}

// Positive test cases

func TestPoolGet(t *testing.T) {
	// Test case: get a pool by its ID

	// get the keeper and context
	dexKeeper, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	// create a number of pools
	items := createNPool(dexKeeper, ctx, 10)
	for _, item := range items {
		rst, found := dexKeeper.GetPool(ctx,
			item.PoolId,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}

func TestPoolGetAll(t *testing.T) {
	// Test case: get all pools

	dexKeeper, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	items := createNPool(dexKeeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(dexKeeper.GetAllPool(ctx)),
	)
}

func TestGetNextPoolID_Default(t *testing.T) {
	// Test case: get the next pool ID when no meta

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	// no meta-set yet, should return 1
	nextID := k.GetNextPoolID(ctx)
	require.Equal(t, uint64(1), nextID)
}

func TestSetNextPoolID_Increments(t *testing.T) {
	// Test case: set the next pool ID

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	k.SetNextPoolID(ctx)
	require.Equal(t, uint64(2), k.GetNextPoolID(ctx))

	k.SetNextPoolID(ctx)
	require.Equal(t, uint64(3), k.GetNextPoolID(ctx))
}

func TestGetAndSetNextPoolID(t *testing.T) {
	// Test case: get and set the next pool ID

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	id1 := k.GetAndSetNextPoolID(ctx)
	id2 := k.GetAndSetNextPoolID(ctx)

	require.Equal(t, uint64(1), id1)
	require.Equal(t, uint64(2), id2)
}

func TestGetNextPoolIDString(t *testing.T) {
	// Test case: get the next pool ID as a string

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	id := k.GetNextPoolIDString(ctx)
	require.Equal(t, constants.PoolPrefix+"1", id)
}

func TestGetAndSetNextPoolIDString(t *testing.T) {
	// Test case: get and set the next pool ID as a string

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	id1 := k.GetAndSetNextPoolIDString(ctx)
	require.Equal(t, constants.PoolPrefix+"1", id1)

	id2 := k.GetAndSetNextPoolIDString(ctx)
	require.Equal(t, constants.PoolPrefix+"2", id2)
}

// Negative test cases

func TestSetNextPoolID_OverflowPanics(t *testing.T) {
	// Test case: set the next pool ID to overflow

	k, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	// force meta to max
	k.SetPoolsMeta(ctx, types.PoolsMeta{NextPoolId: math.MaxUint64})

	require.PanicsWithValue(t, "overflow on pool id", func() {
		k.SetNextPoolID(ctx)
	})
}

func TestSendFromAddressToPool_EmptyCoins(t *testing.T) {
	// Test case: sending empty coins from address to pool should return an error

	dexKeeper, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	sender := sdk.AccAddress([]byte("sender"))
	poolID := "zp1"
	coins := sdk.NewCoins()

	err := dexKeeper.SendFromAddressToPool(ctx, sender, poolID, coins)
	require.Error(t, err)
	require.Contains(t, err.Error(), "SendFromAddressToPool: coins cannot be empty")
}

func TestSendFromAddressToPool_ZeroCoins(t *testing.T) {
	// Test case: sending zero coins from address to pool should return an error

	dexKeeper, ctx := keepertest.DexKeeper(t, nil, nil, nil)
	sender := sdk.AccAddress([]byte("sender"))
	poolID := "zp1"
	coins := sdk.NewCoins(sample.Coin("abc", 0))

	err := dexKeeper.SendFromAddressToPool(ctx, sender, poolID, coins)
	require.Error(t, err)
	require.Contains(t, err.Error(), "SendFromAddressToPool: coins cannot be empty: invalid request")
}

func TestSendFromAddressToPool_Empty(t *testing.T) {
	// Test case: sending empty coins should return an error
	dexKeeper, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	sender := sdk.AccAddress([]byte("sender"))
	poolID := "zp1"

	// Test with empty coins
	coins := sdk.NewCoins()

	err := dexKeeper.SendFromAddressToPool(ctx, sender, poolID, coins)
	require.Error(t, err)
	require.Contains(t, err.Error(), "SendFromAddressToPool: coins cannot be empty")
}

func TestSendFromAddressToPool_ZeroCoin(t *testing.T) {
	// Test case: sending zero coins should return an error
	dexKeeper, ctx := keepertest.DexKeeper(t, nil, nil, nil)

	sender := sdk.AccAddress([]byte("sender"))
	poolID := "zp1"

	// Use sample coin utility with zero amount
	coins := sdk.Coins{sample.Coin("abc", 0)}

	err := dexKeeper.SendFromAddressToPool(ctx, sender, poolID, coins)
	require.Error(t, err)
	require.Contains(t, err.Error(), "SendFromAddressToPool: coins cannot be zero")
	require.Contains(t, err.Error(), "invalid request")
}

func TestSendFromAddressToPool_PoolAccountNotFound(t *testing.T) {
	// Test case: sending coins to a pool that doesn't exist should return an error
	dexKeeper, ctx, bankKeeper := keepertest.DexKeeperWithBank(t, nil)

	// Create and fund a sender account
	sender := sdk.AccAddress([]byte("sender"))
	coins := sdk.Coins{sample.Coin("abc", 100)}

	// Fund the sender account
	err := bankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	require.NoError(t, err)
	err = bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, sender, coins)
	require.NoError(t, err)

	// Now try to send to a non-existent pool
	poolID := "zp1"
	err = dexKeeper.SendFromAddressToPool(ctx, sender, poolID, sdk.Coins{sample.Coin("abc", 1)})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrPoolAccountNotFound)
	require.Contains(t, err.Error(), "Pool ID: zp1 account does not exist")
}

func TestSendFromPoolToAddress_EmptyCoins(t *testing.T) {
	// Test case: sending empty coins from pool to address should return an error
	dexKeeper, ctx, _ := keepertest.DexKeeperWithBank(t, nil)

	poolID := "zp1"
	receiver := sdk.AccAddress([]byte("receiver"))
	coins := sdk.NewCoins() // Empty coins

	err := dexKeeper.SendFromPoolToAddress(ctx, poolID, receiver, coins)
	require.Error(t, err)
	require.Contains(t, err.Error(), "SendFromPoolToAddress: No coins to send")
	require.Contains(t, err.Error(), "invalid request")
	require.Contains(t, err.Error(), poolID)
	require.Contains(t, err.Error(), receiver.String())
}

func TestSendFromPoolToAddress_PoolAccountNotFound(t *testing.T) {
	// Test case: sending coins from a non-existent pool should return an error
	dexKeeper, ctx, _ := keepertest.DexKeeperWithBank(t, nil)

	poolID := "zp999" // Non-existent pool
	receiver := sdk.AccAddress([]byte("receiver"))
	coins := sdk.Coins{sample.Coin("abc", 100)}

	err := dexKeeper.SendFromPoolToAddress(ctx, poolID, receiver, coins)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrPoolAccountNotFound)
	require.Contains(t, err.Error(), "Pool ID: zp999 account already does not exists")
}
