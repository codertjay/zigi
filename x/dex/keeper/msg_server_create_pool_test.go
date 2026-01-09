package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	keepertest "zigchain/testutil/keeper"
	"zigchain/testutil/sample"
	"zigchain/x/dex/keeper"
	"zigchain/x/dex/testutil"
	"zigchain/x/dex/testutil/common"
	"zigchain/x/dex/types"
	"zigchain/zutils/constants"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

// Positive test cases

func TestPoolMsgServerCreate_Positive(t *testing.T) {
	// Test case: create a pool with valid inputs

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	expectedLPCoin := sample.Coin("zp1", 100)    // 100 (sqrt)
	expectedUserShares := sample.Coin("zp1", 90) // 100 - 10 (minimal lock)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 10
	k.SetParams(ctx, params)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(abc, usdt)).
		Return(nil).
		Times(1)

	// code will mint the LP token and drop it into the module dex module account
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(expectedUserShares)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the remaining LP tokens (excluding minimal lock) from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(expectedUserShares)).
		Return(nil).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	// make sure there is no error
	require.NoError(t, err)

	// get the pool from the keeper
	pool, found := k.GetPool(ctx,
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

	// quick pool check
	common.PoolCheck(
		t,
		pool,
		poolId,
		creator,
		abc,
		usdt,
		expectedLPCoin,
		poolAddress,
	)
}

func TestPoolMsgServerCreate_Valid(t *testing.T) {
	// Test case: create a pool with valid inputs

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 542)
	usdt := sample.Coin("usdt", 112)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	expectedLPCoin := sample.Coin("zp1", 246)     // 246 (sqrt)
	expectedUserShares := sample.Coin("zp1", 236) // 246 - 10 (minimal lock)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 10
	k.SetParams(ctx, params)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(abc, usdt)).
		Return(nil).
		Times(1)

	// code will mint the LP token and drop it into the module dex module account
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(expectedUserShares)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the remaining LP tokens (excluding minimal lock) from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(expectedUserShares)).
		Return(nil).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	// make sure there is no error
	require.NoError(t, err)

	// get the pool from the keeper
	pool, found := k.GetPool(ctx,
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

	// quick pool check
	common.PoolCheck(
		t,
		pool,
		poolId,
		creator,
		abc,
		usdt,
		expectedLPCoin,
		poolAddress,
	)
}

func TestPoolMsgServerCreate_WithReceiver(t *testing.T) {
	// Test case: create a pool with valid inputs --> with a receiver set

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create a sample receiver address
	receiver := sample.AccAddress()

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	expectedLPCoin := sample.Coin("zp1", 100)    // 100 (sqrt)
	expectedUserShares := sample.Coin("zp1", 90) // 100 - 10 (minimal lock)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 10
	k.SetParams(ctx, params)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(abc, usdt)).
		Return(nil).
		Times(1)

	// code will mint the LP token and drop it into the module dex module account
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(expectedUserShares)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the remaining LP tokens (excluding minimal lock) from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, sdk.MustAccAddressFromBech32(receiver), sdk.NewCoins(expectedUserShares)).
		Return(nil).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator:  creator,
		Base:     abc,
		Quote:    usdt,
		Receiver: receiver,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	// make sure there is no error
	require.NoError(t, err)

	// get the pool from the keeper
	pool, found := k.GetPool(ctx,
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
}

func TestPoolMsgServerCreate_WithBeneficiary(t *testing.T) {
	// Test case: create a pool with valid inputs --> with a beneficiary set

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	expectedLPCoin := sample.Coin("zp1", 100)    // 100 (sqrt)
	expectedUserShares := sample.Coin("zp1", 90) // 100 - 10 (minimal lock)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// update params to set the beneficiary
	params := types.DefaultParams()
	beneficiary := sample.AccAddress()
	params.Beneficiary = beneficiary
	params.MinimalLiquidityLock = 10
	require.NoError(t, k.SetParams(ctx, params))

	beneficiaryAddr := sdk.MustAccAddressFromBech32(beneficiary)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of creationFee
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, creationFee).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	// code will send the creation fee to the beneficiary
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, beneficiaryAddr, sdk.NewCoins(creationFee)).
		Return(nil).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(abc, usdt)).
		Return(nil).
		Times(1)

	// MintCoins(context.Context, string, sdk.Coins) error
	// code will mint the LP token and drop it into the module dex module account
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(expectedUserShares)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the remaining LP tokens (excluding minimal lock) from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(expectedUserShares)).
		Return(nil).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	// make sure there is no error
	require.NoError(t, err)

	// get the pool from the keeper
	pool, found := k.GetPool(ctx,
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

	// quick pool check
	common.PoolCheck(
		t,
		pool,
		poolId,
		creator,
		abc,
		usdt,
		expectedLPCoin,
		poolAddress,
	)
}

// Negative test cases

func TestPoolMsgServerCreate_InvalidSigner(t *testing.T) {
	// Test case: try to create a pool with an invalid signer address

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)

	// create mock bank keeper
	// we will still create it,
	// although it is never used to ensure no function is called on it
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, nil)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: "Bad signer",
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error msg is correct
	// "Invalid address: Bad signer: invalid address"
	require.Equal(t,
		"Invalid address: Bad signer: invalid address",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_EmptySigner(t *testing.T) {
	// Test case: try to create a pool with an empty signer address

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)

	// create mock bank keeper
	// we will still create it,
	// although it is never used to ensure no function is called on it
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, nil)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: "",
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error msg is correct
	require.Equal(t,
		"Invalid address: : invalid address",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_InsufficientBaseFunds(t *testing.T) {
	// Test case: try to create a pool if insufficient abc funds in signer wallet

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// will return false = No funds to check if error is properly thrown
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, abc).
		Return(false).
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, nil)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	// "Signer wallet does not have 100abc tokens: insufficient funds"
	require.Equal(t,
		"Signer wallet does not have 100abc tokens: insufficient funds",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_InsufficientQuoteFunds(t *testing.T) {
	// Test case: try to create a pool if insufficient usdt funds in signer wallet

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// will return false = No funds to check if error is properly thrown
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(false).
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, nil)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	// "Signer wallet does not have 100usdt tokens: insufficient funds"
	require.Equal(t,
		"Signer wallet does not have 100usdt tokens: insufficient funds",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_PollIdAlreadyExists(t *testing.T) {
	// Test case: try to create a pool if the pool id already exists

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	expectedLPCoin := sample.Coin("zp1", 100)    // 100 (sqrt)
	expectedUserShares := sample.Coin("zp1", 90) // 100 - 10 (minimal lock)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 10
	k.SetParams(ctx, params)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(abc, usdt)).
		Return(nil).
		Times(1)

	// code will mint the LP token and drop it into the module dex module account
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(expectedUserShares)).
		Return(nil).
		Times(1)

	// SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	// code will send the remaining LP tokens (excluding minimal lock) from dex module to the signer
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(expectedUserShares)).
		Return(nil).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	// make sure there is no error
	require.NoError(t, err)

	// get the pool from the keeper
	pool, found := k.GetPool(ctx,
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

	// tru to create the pool with the same id again
	_, err = srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	// "Pool with id: zp1 already exists with uid: abc-usdt: invalid request"
	require.Equal(t,
		"Pool with id: zp1 already exists with uid: abc-usdt: invalid request",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_InsufficientCreationFeeFunds(t *testing.T) {
	// Test case: try to create a pool if insufficient creation fee funds in signer wallet

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	// will return false = No funds to check if error is properly thrown
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, creationFee).
		Return(false).
		Times(1)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, nil)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)

	// check if the error msg is correct
	// "Signer wallet does not have 100uzig tokens: insufficient funds"
	require.Equal(t,
		fmt.Sprintf(
			"Signer wallet does not have %s tokens: insufficient funds",
			creationFee,
		),
		err.Error(),
	)
}

func TestPoolMsgServerCreate_SendCoinsFromAccountToModule_InsufficientCreationFeeFunds(t *testing.T) {
	// Test case: try to create a pool if insufficient creation fee funds in signer wallet

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, nil)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of creationFee
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, creationFee).
		Return(true).
		Times(1)

	// code will return ErrInsufficientFunds
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(creationFee)).
		Return(sdkerrors.ErrInsufficientFunds).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)

	// check if the error msg is correct
	require.Equal(t,
		fmt.Sprintf(
			"Error while sending coins %s from account: %s to module: dex: insufficient funds",
			creationFee,
			signer,
		),
		err.Error(),
	)
}

func TestPoolMsgServerCreate_WithBeneficiary_SendCoinsError(t *testing.T) {
	// Test case: failed in sending coins from module to account

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// update params to set the beneficiary
	params := types.DefaultParams()
	beneficiary := sample.AccAddress()
	params.Beneficiary = beneficiary
	require.NoError(t, k.SetParams(ctx, params))

	beneficiaryAddr := sdk.MustAccAddressFromBech32(beneficiary)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of creationFee
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, creationFee).
		Return(true).
		Times(1)

	// code will return ErrInsufficientFunds
	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, beneficiaryAddr, sdk.NewCoins(creationFee)).
		Return(sdkerrors.ErrInsufficientFunds).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	// check if the error is correct
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)
	require.Equal(
		t,
		fmt.Sprintf(
			"Error while sending coins %s from account: %s to beneficiary: %s: insufficient funds",
			creationFee,
			signer,
			beneficiary,
		),
		err.Error(),
	)
}

func TestPoolMsgServerCreate_BurnCoinsError(t *testing.T) {
	// Test case: error while burning coins

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// code will check if the signer has the required balance of abc
	// HasBalance(context.Context, sdk.AccAddress, sdk.Coin) bool
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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

	// code will return ErrInsufficientFunds
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(creationFee)).
		Return(sdkerrors.ErrInsufficientFunds).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	// check if there is an error
	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)
	require.Equal(
		t,
		fmt.Sprintf(
			"CreatePool: BurnCoins Failed in burning %s coins from module %s: insufficient funds",
			creationFee,
			types.ModuleName,
		),
		err.Error(),
	)
}

func TestPoolMsgServerCreate_SendCoinsInsufficientFunds(t *testing.T) {
	// Test case: error while sending coins

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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
		HasBalance(gomock.Any(), signer, abc).
		Return(false).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	// check if there is an error
	require.Error(t, err)

	// check if the error msg is correct
	require.Equal(t,
		fmt.Sprintf(
			"CreatePool: SendFromAddressToPool error while sending coins %s and %s: SendFromAddressToPool: Insufficient funds in sender %s to send %s coins to poolID %s (address: %s): insufficient funds",
			abc,
			usdt,
			signer,
			abc,
			poolId,
			poolAddress.String(),
		),
		err.Error(),
	)
}

func TestPoolMsgServerCreate_MintCoinsError(t *testing.T) {
	// Test case: try to create a pool if minting coins fails

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)
	// square root of a * b
	expectedUserShares := sample.Coin("zp1", 100-10) // 100 - 10 (minimal lock)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 10
	k.SetParams(ctx, params)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(abc, usdt)).
		Return(nil).
		Times(1)

	// code will return ErrInsufficientFunds
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(expectedUserShares)).
		Return(sdkerrors.ErrInsufficientFunds).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)

	// check if the error msg is correct
	require.Equal(t,
		"Error in minting coins: insufficient funds",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_SendCoinsInsufficient(t *testing.T) {
	// Test case: failed in sending coins from module to account

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)
	// // square root of a * b plus minimal lock
	expectedUserShares := sample.Coin("zp1", 100-10) // 100 - 10 (minimal lock)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 10
	k.SetParams(ctx, params)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(abc, usdt)).
		Return(nil).
		Times(1)

	// code will mint the LP token and drop it into the module dex module account
	bankKeeper.
		EXPECT().
		MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(expectedUserShares)).
		Return(nil).
		Times(1)

	// code will return ErrInsufficientFunds
	bankKeeper.
		EXPECT().
		SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, signer, sdk.NewCoins(expectedUserShares)).
		Return(sdkerrors.ErrInsufficientFunds).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)

	// check if the error msg is correct
	require.Equal(
		t,
		fmt.Sprintf(
			"CreatePool: SendCoinsFromModuleToAccount Failed in sending %s coins from module %s to account: %s: insufficient funds",
			expectedUserShares,
			types.ModuleName,
			creator,
		),
		err.Error(),
	)
}

func TestPoolMsgServerCreate_InvalidReceiver_BadAddress(t *testing.T) {
	// Test case: try to create a pool with invalid receiver address

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// String address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator:  creator,
		Base:     abc,
		Quote:    usdt,
		Receiver: "bad_receiver",
	}

	// make rpc call to add liquidity
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	require.Equal(t,
		"decoding bech32 failed: invalid separator index -1",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_InvalidReceiver_TooShort(t *testing.T) {
	// Test case: try to create a pool with invalid receiver address --> too short

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator:  creator,
		Base:     abc,
		Quote:    usdt,
		Receiver: "zig123",
	}

	// make rpc call to add liquidity
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	require.Equal(t,
		"decoding bech32 failed: invalid bech32 string length 6",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_InvalidReceiver_TooLong(t *testing.T) {
	// Test case: try to create a pool with invalid receiver address --> too long

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator:  creator,
		Base:     abc,
		Quote:    usdt,
		Receiver: "zig1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
	}

	// make rpc call to add liquidity
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	require.Equal(t,
		"decoding bech32 failed: invalid checksum (expected yurny3 got 567890)",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_InvalidReceiver_NotLowercaseOrUppercase(t *testing.T) {
	// Test case: try to create a pool with invalid receiver address --> not all lowercase or uppercase

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator:  creator,
		Base:     abc,
		Quote:    usdt,
		Receiver: "BAd Address",
	}

	// make rpc call to add liquidity
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	require.Equal(t,
		"decoding bech32 failed: string not all lowercase or all uppercase",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_InvalidReceiver_BadChars(t *testing.T) {
	// Test case: try to create a pool with invalid receiver address --> bad characters

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
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

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator:  creator,
		Base:     abc,
		Quote:    usdt,
		Receiver: "zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%",
	}

	// make rpc call to add liquidity
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	require.Equal(t,
		"decoding bech32 failed: invalid character not part of charset: 37",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_NoMinimalLiquidityLock(t *testing.T) {
	// Test case: verify that minimal liquidity is locked when creating a pool,
	// This test validates the security feature of minimal liquidity lock

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	// Create minimal amounts to test edge case
	abc := sample.Coin("abc", 1)
	usdt := sample.Coin("usdt", 1)
	creationFee := sample.Coin("uzig", 100000000)
	// With minimal amounts, sqrt(1*1) = 1, plus minimal lock
	expectedLPCoin := sample.Coin("zp1", 1)

	k, ctx, bankKeeper := keepertest.DexKeeperWithBank(t, nil)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 0
	k.SetParams(ctx, params)

	srv := keeper.NewMsgServerImpl(k)

	// mint signer balance
	coins := sdk.NewCoins(abc, usdt, creationFee)
	bankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, coins)

	poolId := k.GetNextPoolIDString(ctx)

	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// Create pool
	_, err := srv.CreatePool(ctx, txMsg)
	require.NoError(t, err)

	// Get the pool
	pool, found := k.GetPool(ctx, poolId)
	require.True(t, found)

	// Verify that LP tokens include minimal lock
	totalLPTokens := expectedLPCoin.Amount
	require.Equal(t, totalLPTokens.Uint64(), pool.LpToken.Amount.Uint64())
}

func TestPoolMsgServerCreate_InitialLiquidityCalculation(t *testing.T) {
	// Test case: verify the initial liquidity calculation includes minimal lock
	// This test validates that the sqrt calculation plus minimal lock is correct

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	// Use perfect square numbers to make calculation clear
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)
	// sqrt(100*100) = 100, plus minimal lock
	expectedLPCoin := sample.Coin("zp1", 100)

	k, ctx, bankKeeper := keepertest.DexKeeperWithBank(t, nil)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 10
	k.SetParams(ctx, params)

	srv := keeper.NewMsgServerImpl(k)

	// mint signer balance
	coins := sdk.NewCoins(abc, usdt, creationFee)
	bankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
	bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, coins)

	poolId := k.GetNextPoolIDString(ctx)

	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// Create pool
	_, err := srv.CreatePool(ctx, txMsg)
	require.NoError(t, err)

	// Get the pool
	pool, found := k.GetPool(ctx, poolId)
	require.True(t, found)

	// Verify that the LP token amount is sqrt(base * quote) + minimal lock
	totalLPTokens := expectedLPCoin.Amount
	require.Equal(t, totalLPTokens.Uint64(), pool.LpToken.Amount.Uint64())
	require.Equal(t, uint64(100), pool.LpToken.Amount.Uint64()) // 100
}

func TestPoolMsgServerCreate_InvalidBeneficiaryAddress(t *testing.T) {
	// Test case: try to create a pool with invalid beneficiary address in params

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, nil)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// update params to set an invalid beneficiary
	params := types.DefaultParams()
	params.Beneficiary = "bad_address"
	params.CreationFee = 100000000 // Ensure CreationFee > 0 to trigger the beneficiary check
	require.NoError(t, k.SetParams(ctx, params))

	// code will check if the signer has the required balance of abc
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	// code will check if the signer has the required balance of creationFee
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, creationFee).
		Return(true).
		Times(1)

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorIs(t, err, sdkerrors.ErrInvalidAddress)

	// check if the error message is correct
	require.Equal(t,
		"Invalid address: bad_address: invalid address",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_PoolIdAlreadyExists(t *testing.T) {
	// Test case: try to create a pool when the assigned pool ID already exists (just in case scenario)

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper (not used in this path, but pass it)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 10
	params.CreationFee = 100000000 // Ensure >0 to check fee
	k.SetParams(ctx, params)

	// Assuming the initial pool count is 0, the next ID is 1, poolIDString = "zp1"
	// We need to set a dummy pool with PoolId "zp1"
	dummyPool := types.Pool{
		PoolId: constants.PoolPrefix + "1",
		// Other fields minimal, different coins to avoid first check
		Coins: sdk.NewCoins(sample.Coin("dummy1", 1), sample.Coin("dummy2", 1)),
	}
	k.SetPool(ctx, dummyPool)

	// Now GetPool("zp1") will return found=true

	// Setup mocks for checks before ID assignment

	// HasBalance for base
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// HasBalance for quote
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	// HasBalance for creation fee
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, creationFee).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule for fee
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(creationFee)).
		Return(nil).
		Times(1)

	// BurnCoins for fee
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(creationFee)).
		Return(nil).
		Times(1)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// create a sample message
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	require.Equal(t,
		"Pool with id: zp1 already exists: invalid request",
		err.Error(),
	)
}

func TestPoolMsgServerCreate_CreatePoolAccountError_AccountAlreadyExists(t *testing.T) {
	// Test case: try to create a pool when the pool account already exists (error in CreatePoolAccount)

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 10
	params.CreationFee = 100000000 // Ensure >0 to check fee
	k.SetParams(ctx, params)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account (to simulate it already exists)
	poolAccount := sample.PoolModuleAccount(poolAddress)

	// Setup mocks

	// HasBalance for base
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(1)

	// HasBalance for quote
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(1)

	// HasBalance for creation fee
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, creationFee).
		Return(true).
		Times(1)

	// SendCoinsFromAccountToModule for fee (since CreationFee > 0 and no beneficiary)
	bankKeeper.
		EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), signer, types.ModuleName, sdk.NewCoins(creationFee)).
		Return(nil).
		Times(1)

	// BurnCoins for fee
	bankKeeper.
		EXPECT().
		BurnCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(creationFee)).
		Return(nil).
		Times(1)

	// Now, for CreatePoolAccount: GetAccount returns non-nil (existing account)
	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	// No further calls to NewAccount or SetAccount

	// create a sample message
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	require.Error(t, err)

	// check if the error is correct
	require.ErrorContains(t, err, "CreatePoolAccount: Failed to create module account at address")
}

func TestPoolMsgServerCreate_NonPositiveBaseAmount(t *testing.T) {
	// Test case: try to create a pool with non-positive base amount

	// create a mock keeper
	k, ctx, bankKeeper := keepertest.DexKeeperWithBank(t, nil)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 10
	params.CreationFee = 100000000
	k.SetParams(ctx, params)

	// create a sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 0)
	usdt := sample.Coin("usdt", 100)
	creationFee := sample.Coin("uzig", 100000000)

	// mint signer balance
	coins := sdk.NewCoins(abc, usdt, creationFee)
	err := bankKeeper.MintCoins(ctx, types.ModuleName, coins)
	require.NoError(t, err)
	err = bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, signer, coins)
	require.NoError(t, err)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// create a sample message
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err = srv.CreatePool(ctx, txMsg)

	// make sure there is an error
	require.ErrorIs(t, err, types.ErrNonPositiveAmounts)
}

func TestPoolMsgServerCreate_NonPositiveQuoteAmount(t *testing.T) {
	// Test case: try to create a pool with non-positive quote amount

	// create a mock keeper
	k, ctx, bankKeeper := keepertest.DexKeeperWithBank(t, nil)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 10
	params.CreationFee = 100000000
	k.SetParams(ctx, params)

	// create a sample signer address
	creator := sample.AccAddress()
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	abc := sample.Coin("abc", 100)
	usdt := sample.Coin("usdt", 0)
	creationFee := sample.Coin("uzig", 100000000)

	// mint signer balance
	coins := sdk.NewCoins(abc, usdt, creationFee)
	err := bankKeeper.MintCoins(ctx, types.ModuleName, coins)
	require.NoError(t, err)
	err = bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, signer, coins)
	require.NoError(t, err)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// create a sample message
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err = srv.CreatePool(ctx, txMsg)

	// make sure there is an error
	require.ErrorIs(t, err, types.ErrNonPositiveAmounts)
}

func TestPoolMsgServerCreate_InsufficientLiquidityLock(t *testing.T) {
	// Test case: try to create a pool with insufficient initial liquidity (lpAmount < minimalLock)

	// create a mock controller
	ctrl := gomock.NewController(t)

	// call this when the test is done, or exits prematurely
	defer ctrl.Finish() // assert that all expectations are met

	// create a sample signer address
	// string address for message creator
	creator := sample.AccAddress()

	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// create all coins required for the pool creation
	// 9*9=81, sqrt=9 <10
	abc := sample.Coin("abc", 9)
	usdt := sample.Coin("usdt", 9)
	creationFee := sample.Coin("uzig", 100000000)

	// create mock bank keeper
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	// create mock account keeper
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)

	// create a mock keeper
	k, ctx := keepertest.DexKeeper(t, bankKeeper, nil, accountKeeper)

	// set the params
	params := types.DefaultParams()
	params.MinimalLiquidityLock = 10
	params.CreationFee = 100000000
	k.SetParams(ctx, params)

	// get access to message server
	srv := keeper.NewMsgServerImpl(k)

	// get id of next pool
	poolId := k.GetNextPoolIDString(ctx)

	// get the pool address
	poolAddress := types.GetPoolAddress(poolId)

	// create a pool account
	poolAccount := sample.PoolModuleAccount(poolAddress)

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
		}).
		Times(1)

	accountKeeper.
		EXPECT().
		SetAccount(gomock.Any(), gomock.Eq(poolAccount)).
		Times(1)

	// code will check if the signer has the required balance of abc
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, abc).
		Return(true).
		Times(2)

	// code will check if the signer has the required balance of usdt
	bankKeeper.
		EXPECT().
		HasBalance(gomock.Any(), signer, usdt).
		Return(true).
		Times(2)

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

	accountKeeper.
		EXPECT().
		GetAccount(ctx, poolAddress).
		Return(poolAccount).
		Times(1)

	bankKeeper.
		EXPECT().
		SendCoins(gomock.Any(), signer, poolAddress, sdk.NewCoins(abc, usdt)).
		Return(nil).
		Times(1)

	// No mint or send LP

	// create a sample message to compare
	txMsg := &types.MsgCreatePool{
		Creator: creator,
		Base:    abc,
		Quote:   usdt,
	}

	// make rpc call to create pool
	_, err := srv.CreatePool(ctx, txMsg)

	// make sure there is an error
	require.ErrorIs(t, err, types.ErrInsufficientLiquidityLock)
}
