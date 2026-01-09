package common

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/dex/keeper"
	"zigchain/x/dex/testutil"
	"zigchain/x/dex/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

func ServerDexKeeperWithRealBank(
	t *testing.T,
	signer sdk.AccAddress,
	base sdk.Coin,
	quote sdk.Coin,
	creationFee sdk.Coin,
	expectedLPCoin sdk.Coin,
) (server types.MsgServer, dexKeeper keeper.Keeper, ctx sdk.Context, pool types.Pool, poolAccount sdk.AccountI, bankKeeper bankkeeper.BaseKeeper) {
	dexKeeper, ctx, bankKeeper = keepertest.DexKeeperWithBank(t, nil)

	srv := keeper.NewMsgServerImpl(dexKeeper)

	// set the params
	params := dexKeeper.GetParams(ctx)
	params.MinimalLiquidityLock = 0
	// #nosec G104 -- test file, error handling not critical
	dexKeeper.SetParams(ctx, params)

	// mint signer balance x 2
	coins := sdk.NewCoins(base, quote, creationFee)
	// #nosec G104 -- test file, error handling not critical
	bankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	// #nosec G104 -- test file, error handling not critical
	bankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	// #nosec G104 -- test file, error handling not critical
	bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, coins)
	// #nosec G104 -- test file, error handling not critical
	bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, coins)

	// get id of the next pool
	poolId := dexKeeper.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount = sample.PoolModuleAccount(poolAddress)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: signer.String(),
		Base:    base,
		Quote:   quote,
	}

	// make rpc call to create a pool
	_, err := srv.CreatePool(ctx, txMsg)

	// make sure there is no error
	require.NoError(t, err)

	// get the pool from the keeper
	pool, found := dexKeeper.GetPool(ctx, poolId)

	// make sure the pool was found
	require.True(t, found)

	// expected (left) vs. actual (right)
	require.Equal(t, txMsg.Creator, pool.Creator)
	require.Equal(t, poolId, pool.PoolId)
	require.Equal(t, txMsg.Base, pool.Coins[0])
	require.Equal(t, txMsg.Quote, pool.Coins[1])
	require.Equal(t, expectedLPCoin, pool.LpToken)
	require.Equal(t, uint32(500), pool.Fee)
	require.Equal(t, "constant_product", pool.Formula)
	require.Equal(t, poolAddress.String(), pool.Address)

	return srv, dexKeeper, ctx, pool, poolAccount, bankKeeper
}

func ServerDexKeeperWithPoolMock(
	t *testing.T,
	ctrl *gomock.Controller,
	signer sdk.AccAddress,
	base sdk.Coin,
	quote sdk.Coin,
	creationFee sdk.Coin,
	expectedLPCoin sdk.Coin,
) (server types.MsgServer,
	dexKeeper keeper.Keeper,
	ctx sdk.Context,
	pool types.Pool,
	poolAccount sdk.AccountI,
	bankKeeper *testutil.MockBankKeeper,
	accountKeeper *testutil.MockAccountKeeper,
) {
	// create mock bank keeper
	bankKeeper = testutil.NewMockBankKeeper(ctrl)
	// create a mock account keeper
	accountKeeper = testutil.NewMockAccountKeeper(ctrl)

	mintKeeper := testutil.NewMockMintKeeper(ctrl)

	// create a mock keeper
	dexKeeper, ctx = keepertest.DexKeeper(t, bankKeeper, mintKeeper, accountKeeper)

	// set minimal lock to 0
	params := dexKeeper.GetParams(ctx)
	params.MinimalLiquidityLock = 0
	// #nosec G104 -- test file, error handling not critical
	dexKeeper.SetParams(ctx, params)

	// get access to the message server
	srv := keeper.NewMsgServerImpl(dexKeeper)

	// get id of the next pool
	poolId := dexKeeper.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount = sample.PoolModuleAccount(poolAddress)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(nil).
		Times(1)

	accountKeeper.
		EXPECT().
		NewAccount(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx sdk.Context, acc sdk.AccountI) sdk.AccountI {
			return acc
		})
	accountKeeper.
		EXPECT().
		SetAccount(gomock.Any(), gomock.Eq(poolAccount)).
		Times(1)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, base).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, quote).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of creationFee
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, creationFee).
		Return(true).
		Times(1)

	// code will transfer creationFee coins from the signer to the module
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(creationFee)).
		Return(nil).
		Times(1)

	// code will burn the creationFee coins from the signer
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(creationFee)).
		Return(nil).
		Times(1)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, base).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, quote).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(base, quote)).
		Return(nil).
		Times(1)

	// code will mint the LP token and drop it into the module dex module account
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(expectedLPCoin)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the minted LP token from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(expectedLPCoin)).
		Return(nil).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: signer.String(),
		Base:    base,
		Quote:   quote,
	}

	// make rpc call to create a pool
	_, err := srv.CreatePool(ctx, txMsg)

	// make sure there is no error
	require.NoError(t, err)

	// get the pool from the keeper
	pool, found := dexKeeper.GetPool(ctx,
		poolId,
	)

	// make sure the pool was found
	require.True(t, found)

	// expected (left) vs. actual (right)
	require.Equal(t, txMsg.Creator, pool.Creator)
	require.Equal(t, poolId, pool.PoolId)
	require.Equal(t, txMsg.Base, pool.Coins[0])
	require.Equal(t, txMsg.Quote, pool.Coins[1])
	require.Equal(t, expectedLPCoin, pool.LpToken)
	require.Equal(t, uint32(500), pool.Fee)
	require.Equal(t, "constant_product", pool.Formula)
	require.Equal(t, poolAddress.String(), pool.Address)

	// (server types.MsgServer, dexKeeper keeper.Keeper, ctx sdk.Context, poolID String, poolAccount sdk.AccountI)
	return srv, dexKeeper, ctx, pool, poolAccount, bankKeeper, accountKeeper
}

func PoolCheck(
	t *testing.T,
	pool types.Pool,
	poolId string,
	creator string,
	base sdk.Coin,
	quote sdk.Coin,
	expectedLPCoin sdk.Coin,
	poolAddress sdk.AccAddress,
) {

	// expected (left) vs. actual (right)
	require.Equal(t, creator, pool.Creator)
	require.Equal(t, poolId, pool.PoolId)
	require.Equal(t, base.Denom, pool.Coins[0].Denom)
	require.Equal(t, base.Amount.String(), pool.Coins[0].Amount.String())
	// require.Equal(t, base, pool.Coins[0])
	require.Equal(t, quote.Denom, pool.Coins[1].Denom)
	// fmt.Println(quote.Amount.String())
	// fmt.Println(pool.Coins[1].Amount.String())
	require.Equal(t, quote.Amount.String(), pool.Coins[1].Amount.String())
	require.Equal(t, quote.Denom, pool.Coins[1].Denom)
	// require.Equal(t, quote, pool.Coins[1])
	require.Equal(t, expectedLPCoin, pool.LpToken)
	require.Equal(t, uint32(500), pool.Fee)
	require.Equal(t, "constant_product", pool.Formula)
	require.Equal(t, poolAddress.String(), pool.Address)
}
