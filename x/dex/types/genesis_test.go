package types_test

import (
	"fmt"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/testutil/sample"
	"zigchain/x/dex/types"
	factorytypes "zigchain/x/factory/types"
	"zigchain/zutils/constants"

	"github.com/stretchr/testify/require"
)

// Positive test cases

func TestGenesisState_Validate_Positive(t *testing.T) {
	// Test case: valid GenesisState

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	validPoolID := "zp1"
	validPoolList := []types.Pool{
		{
			PoolId:  validPoolID,
			LpToken: sdk.NewInt64Coin("zp1", 100),
			Creator: creator,
			Fee:     1,
			Formula: types.FormulaConstantProduct,
			Coins: []sdk.Coin{
				{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
				{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
			},
			Address: types.GetPoolAddress(validPoolID).String(),
		},
	}

	// poolUidString := pool.Coins[0].Denom + "/" + pool.Coins[1].Denom
	poolUidString := fullDenomAbc + types.PoolUidSeparator + fullDenomUsdt
	validPoolUidList := []types.PoolUids{
		{
			PoolUid: poolUidString,
			PoolId:  validPoolID,
		},
	}

	validGenesis := types.GenesisState{
		PoolList:     validPoolList,
		PoolUidsList: validPoolUidList,
		PoolsMeta:    &types.PoolsMeta{NextPoolId: 2},
		Params:       types.DefaultParams(),
	}

	err := validGenesis.Validate()
	require.NoError(t, err, "Valid GenesisState validation should not return an error")
}

func TestPool_Validate_Positive(t *testing.T) {
	// Test case: valid pool

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	validPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: types.GetPoolAddress("zp1").String(),
	}

	err := validPool.Validate()
	require.NoError(t, err, "Valid Pool validation should not return an error")
}

func TestDefaultGenesis(t *testing.T) {
	// Test case: DefaultGenesis should return a valid GenesisState

	genState := types.DefaultGenesis()

	require.NotNil(t, genState, "DefaultGenesis should return a non-nil GenesisState")
	require.Empty(t, genState.PoolList, "PoolList should be empty")
	require.Nil(t, genState.PoolsMeta, "PoolsMeta should be nil")
	require.Empty(t, genState.PoolUidsList, "PoolUidsList should be empty")
	require.Equal(t, types.DefaultParams(), genState.Params, "Params should match DefaultParams")
}

func TestDefaultGenesis_Validate(t *testing.T) {
	// Test case: DefaultGenesis should validate without errors

	genState := types.DefaultGenesis()

	err := genState.Validate()
	require.NoError(t, err, "DefaultGenesis should validate without errors")
}

// Negative test cases

func TestGenesisState_Validate_DuplicatedPool(t *testing.T) {
	// Test case: duplicate PoolID

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	duplicatedPoolID := "zp1"
	invalidGenesis := types.GenesisState{
		PoolList: []types.Pool{
			{
				PoolId:  duplicatedPoolID,
				LpToken: sdk.NewInt64Coin("zp1", 100),
				Creator: creator,
				Fee:     1,
				Formula: types.FormulaConstantProduct,
				Coins: []sdk.Coin{
					{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
					{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
				},
				Address: types.GetPoolAddress(duplicatedPoolID).String(),
			},
			{
				PoolId:  duplicatedPoolID,
				LpToken: sdk.NewInt64Coin("zp1", 100),
				Creator: creator,
				Fee:     1,
				Formula: types.FormulaConstantProduct,
				Coins: []sdk.Coin{
					{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
					{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
				},
				Address: types.GetPoolAddress(duplicatedPoolID).String(),
			},
		},
		Params: types.DefaultParams(),
	}

	err := invalidGenesis.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicated index for pool")
}

func TestGenesisState_Validate_DuplicatePoolUids(t *testing.T) {
	// Test case: duplicate PoolUids

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	poolUidString := fullDenomAbc + types.PoolUidSeparator + fullDenomUsdt
	invalidGenesis := types.GenesisState{
		PoolList: []types.Pool{
			{
				PoolId:  "zp1",
				LpToken: sdk.NewInt64Coin("zp1", 100),
				Creator: creator,
				Fee:     1,
				Formula: types.FormulaConstantProduct,
				Coins: []sdk.Coin{
					{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
					{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
				},
				Address: types.GetPoolAddress("zp1").String(),
			},
		},
		PoolsMeta: &types.PoolsMeta{NextPoolId: 2},
		PoolUidsList: []types.PoolUids{
			{
				PoolUid: poolUidString,
				PoolId:  "zp1",
			},
			{
				PoolUid: poolUidString,
				PoolId:  "zp2",
			},
		},
		Params: types.DefaultParams(),
	}

	err := invalidGenesis.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicated index for poolUids")
}

func TestGenesisState_Validate_MissingPoolUidString(t *testing.T) {
	// Test case: missing PoolUidString in PoolUidsList

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidGenesis := types.GenesisState{
		PoolList: []types.Pool{
			{
				PoolId:  "zp1",
				LpToken: sdk.NewInt64Coin("zp1", 100),
				Creator: creator,
				Fee:     1,
				Formula: types.FormulaConstantProduct,
				Coins: []sdk.Coin{
					{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
					{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
				},
				Address: types.GetPoolAddress("zp1").String(),
			},
		},
		PoolsMeta: &types.PoolsMeta{NextPoolId: 2},
		PoolUidsList: []types.PoolUids{
			{
				PoolUid: "invalid",
				PoolId:  "zp1",
			},
		},
		Params: types.DefaultParams(),
	}

	err := invalidGenesis.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing PoolUidString")
}

func TestGenesisState_Validate_MissingPoolsMeta(t *testing.T) {
	// Test case: missing PoolsMeta when PoolList is not empty

	creator := sample.AccAddress()
	fullDenomAbc := "coin." + creator + ".abc"
	fullDenomUsdt := "coin." + creator + ".usdt"

	invalidGenesis := types.GenesisState{
		PoolList: []types.Pool{
			{
				PoolId:  "zp1",
				LpToken: sdk.NewInt64Coin("zp1", 100),
				Creator: creator,
				Fee:     1,
				Formula: types.FormulaConstantProduct,
				Coins: []sdk.Coin{
					{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
					{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
				},
				Address: types.GetPoolAddress("zp1").String(),
			},
		},
		PoolsMeta:    nil,
		PoolUidsList: []types.PoolUids{},
		Params:       types.DefaultParams(),
	}

	err := invalidGenesis.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "PoolsMeta is nil but PoolList is not empty")
}

func TestGenesisState_Validate_InvalidPoolsMetaNextPoolId(t *testing.T) {
	// Test case: PoolsMeta.NextPoolId mismatch

	creator := sample.AccAddress()
	fullDenomAbc := "coin." + creator + ".abc"
	fullDenomUsdt := "coin." + creator + ".usdt"

	invalidGenesis := types.GenesisState{
		PoolList: []types.Pool{
			{
				PoolId:  "zp1",
				LpToken: sdk.NewInt64Coin("zp1", 100),
				Creator: creator,
				Fee:     1,
				Formula: types.FormulaConstantProduct,
				Coins: []sdk.Coin{
					{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
					{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
				},
				Address: types.GetPoolAddress("zp1").String(),
			},
		},
		PoolsMeta: &types.PoolsMeta{NextPoolId: 5},
		Params:    types.DefaultParams(),
	}

	err := invalidGenesis.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "NextPoolId: (5) must be equal to the number of pools in the PoolList (1 + 1 = 2 total)")
}

func TestGenesisState_Validate_InvalidLpTokenMissing(t *testing.T) {
	// Test case: missing LpToken in a pool

	creator := sample.AccAddress()
	fullDenomAbc := "coin." + creator + ".abc"
	fullDenomUsdt := "coin." + creator + ".usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: types.GetPoolAddress("zp1").String(),
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "lpToken: invalid coin amount: cannot be nil (<nil>): invalid coins: ")
}

func TestGenesisState_Validate_EmptyPoolListWithMeta(t *testing.T) {
	// Test case: empty PoolList but PoolsMeta exists

	validGenesis := types.GenesisState{
		PoolList:     []types.Pool{},
		PoolUidsList: []types.PoolUids{},
		PoolsMeta:    &types.PoolsMeta{NextPoolId: 2}, // Inconsistent because the PoolList is empty
		Params:       types.DefaultParams(),
	}

	err := validGenesis.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "NextPoolId: (2) must be equal to the number of pools in the PoolList (0 + 1 = 1 total)")
}

func TestPool_Validate_InvalidCreator(t *testing.T) {
	// Test case: invalid creator address

	creator := "invalid_creator"
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: types.GetPoolAddress("zp1").String(),
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "creator")
}

func TestPool_Validate_InvalidPoolIdEmpty(t *testing.T) {
	// Test case: invalid PoolID --> empty

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPool := types.Pool{
		PoolId:  "",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: sample.AccAddress(),
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.ErrorIs(t, sdkerrors.ErrInvalidCoins, err)
	require.EqualError(t, err, "Invalid pool id: pool id is empty: invalid coins")
}

func TestPool_Validate_InvalidPoolIdTooShort(t *testing.T) {
	// Test case: invalid PoolID --> too short

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPool := types.Pool{
		PoolId:  "z",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: sample.AccAddress(),
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.ErrorIs(t, sdkerrors.ErrInvalidCoins, err)
	require.EqualError(
		t,
		err,
		fmt.Sprintf(
			"Invalid pool id: '%s' pool id is too short, minimum %d characters: invalid coins",
			invalidPool.PoolId,
			constants.MinSubDenomLength,
		),
	)
}

func TestPool_Validate_InvalidPoolIdTooLong(t *testing.T) {
	// Test case: invalid PoolID --> too long

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPool := types.Pool{
		PoolId:  strings.Repeat("a", constants.MaxSubDenomLength+1),
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.ErrorIs(t, sdkerrors.ErrInvalidCoins, err)
	require.EqualError(
		t,
		err,
		fmt.Sprintf(
			"Invalid pool id: '%s' pool id is too long (%d), maximum %d characters: invalid coins",
			invalidPool.PoolId,
			len(invalidPool.PoolId),
			constants.MaxSubDenomLength,
		),
	)
}

func TestPool_Validate_InvalidPoolIdPrefix(t *testing.T) {
	// Test case: invalid PoolID --> invalid prefix

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPool := types.Pool{
		PoolId:  "invalid",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: sample.AccAddress(),
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.ErrorIs(t, sdkerrors.ErrInvalidCoins, err)
	require.EqualError(
		t,
		err,
		fmt.Sprintf(
			"Invalid pool id: '%s', pool id has to start with '%s' followed by numbers e.g. %s123: invalid coins",
			invalidPool.PoolId,
			constants.PoolPrefix,
			constants.PoolPrefix,
		),
	)
}

func TestPool_Validate_InvalidFee_TooLarge(t *testing.T) {
	// Test case: invalid fee --> too large

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     constants.PoolFeeScalingFactor,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(
		t,
		err,
		fmt.Sprintf(
			"Pool fee too large: %d, has to be less than scaling factor: %d: invalid amount",
			invalidPool.Fee,
			constants.PoolFeeScalingFactor,
		),
	)
}

func TestPool_Validate_InvalidFormulaEmpty(t *testing.T) {
	// Test case: invalid formula --> empty

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: "",
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: sample.AccAddress(),
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "invalid formula: ")
}

func TestPool_Validate_InvalidFormulaTooShort(t *testing.T) {
	// Test case: invalid formula --> too short

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: "ab",
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: sample.AccAddress(),
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "invalid formula: ab")
}

func TestPool_Validate_InvalidFormulaTooLong(t *testing.T) {
	// Test case: invalid formula --> too long

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: strings.Repeat("a", 256),
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: sample.AccAddress(),
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "invalid formula: "+strings.Repeat("a", 256))
}

func TestPool_Validate_InvalidCoinsEmpty(t *testing.T) {
	// Test case: invalid coins --> empty

	creator := sample.AccAddress()

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins:   nil,
		Address: sample.AccAddress(),
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "pool must have at least 2 coins")
}

func TestPool_Validate_InvalidCoinsSingle(t *testing.T) {
	// Test case: invalid coins --> single coin

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
		},
		Address: sample.AccAddress(),
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "pool must have at least 2 coins")
}

func TestPool_Validate_InvalidAddressEmpty(t *testing.T) {
	// Test case: invalid address --> empty

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: "",
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "address address: cannot be empty: invalid address")
}

func TestPool_Validate_InvalidAddressInvalidFormat(t *testing.T) {
	// Test case: invalid address --> Invalid Format

	creator := sample.AccAddress()
	fullDenomAbc := "coin." + creator + ".abc"
	fullDenomUsdt := "coin." + creator + ".usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: "invalid_address",
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "address address: 'invalid_address' (decoding bech32 failed: invalid separator index -1): invalid address")
}

func TestPool_Validate_InvalidAddressTooShort(t *testing.T) {
	// Test case: invalid address --> Too short

	creator := sample.AccAddress()
	fullDenomAbc := "coin." + creator + ".abc"
	fullDenomUsdt := "coin." + creator + ".usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: "zig123",
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "address address: 'zig123' (decoding bech32 failed: invalid bech32 string length 6): invalid address")
}

func TestPool_Validate_InvalidAddressTooLong(t *testing.T) {
	// Test case: invalid address --> Too Long

	creator := sample.AccAddress()
	fullDenomAbc := "coin." + creator + ".abc"
	fullDenomUsdt := "coin." + creator + ".usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: "zig1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "address address: 'zig1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890' (decoding bech32 failed: invalid checksum (expected yurny3 got 567890)): invalid address")
}

func TestPool_Validate_InvalidAddressLowerAndUpperMixed(t *testing.T) {
	// Test case: invalid address --> address is not all lowercase or all uppercase

	creator := sample.AccAddress()
	fullDenomAbc := "coin." + creator + ".abc"
	fullDenomUsdt := "coin." + creator + ".usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: "Zig12z7zhe2d03wn03e8lkw8At6zjxa7y65q5xnp5t",
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "address address: 'Zig12z7zhe2d03wn03e8lkw8At6zjxa7y65q5xnp5t' (decoding bech32 failed: string not all lowercase or all uppercase): invalid address")
}

func TestPool_Validate_InvalidAddressCharacters(t *testing.T) {
	// Test case: invalid address --> address contains invalid characters

	creator := sample.AccAddress()
	fullDenomAbc := "coin." + creator + ".abc"
	fullDenomUsdt := "coin." + creator + ".usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: "zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%",
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "address address: 'zi!12z7zhe2d03wn03e8lkw8at6zjxa7y65q5xnp5%' (decoding bech32 failed: invalid character not part of charset: 37): invalid address")
}

func TestPool_Validate_InvalidAddressOnlyNumbers(t *testing.T) {
	// Test case: invalid address --> address contains only numbers

	creator := sample.AccAddress()
	fullDenomAbc := "coin." + creator + ".abc"
	fullDenomUsdt := "coin." + creator + ".usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: "123121712321032203281238456456771651512351",
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "address address: '123121712321032203281238456456771651512351' (decoding bech32 failed: invalid separator index 41): invalid address")
}

func TestPool_Validate_InvalidAddressWithSpace(t *testing.T) {
	// Test case: invalid address --> address contains spaces

	creator := sample.AccAddress()
	fullDenomAbc := "coin." + creator + ".abc"
	fullDenomUsdt := "coin." + creator + ".usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: "abcdefghijabcdefghijabcde ghijabcdefghijabcdefghij",
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "address address: 'abcdefghijabcdefghijabcde ghijabcdefghijabcdefghij' (decoding bech32 failed: invalid character in string: ' '): invalid address")
}

func TestPool_Validate_InvalidAddressPrefix(t *testing.T) {
	// Test case: invalid address --> address has an invalid prefix

	creator := sample.AccAddress()
	fullDenomAbc := "coin." + creator + ".abc"
	fullDenomUsdt := "coin." + creator + ".usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: "yoy193fxruxcm8y32njzt23c86en7hd8tajma79tt3",
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "address address: 'yoy193fxruxcm8y32njzt23c86en7hd8tajma79tt3' (decoding bech32 failed: invalid checksum (expected z6kscw got a79tt3)): invalid address")
}

func TestPool_Validate_InvalidAddressSpecChar(t *testing.T) {
	// Test case: invalid address --> address has special characters

	creator := sample.AccAddress()
	fullDenomAbc := "coin." + creator + ".abc"
	fullDenomUsdt := "coin." + creator + ".usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: "zig1/\\.%&?32njzt23c86en7hd8tajma79tt3",
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "address address: 'zig1/\\.%&?32njzt23c86en7hd8tajma79tt3' (decoding bech32 failed: invalid character not part of charset: 47): invalid address")
}

func TestPool_Validate_InvalidAddressMismatch(t *testing.T) {
	// Test case: invalid address --> address doesn't match expected module address

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: sample.AccAddress(), // Wrong address - should be the module address for "zp1"
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidPoolAddress)
	require.Contains(t, err.Error(), "does not match expected module address")
	require.Contains(t, err.Error(), "zp1")
}

func TestPool_Validate_AllowedFormulas(t *testing.T) {
	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	validPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct, // valid enum
		Coins: []sdk.Coin{
			{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: types.GetPoolAddress("zp1").String(),
	}
	// Should pass
	err := validPool.Validate()
	require.NoError(t, err, "Valid formula should not return error")

	invalidPool := validPool
	invalidPool.Formula = "invalid_formula"
	// Should fail
	err = invalidPool.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid formula")
}

func TestGenesisState_Validate_InvalidPool(t *testing.T) {
	// Test case: GenesisState with an invalid pool (e.g., invalid fee)

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPoolID := "zp1"
	invalidPoolList := []types.Pool{
		{
			PoolId:  invalidPoolID,
			LpToken: sdk.NewInt64Coin("zp1", 100),
			Creator: creator,
			Fee:     constants.PoolFeeScalingFactor, // Invalid fee
			Formula: types.FormulaConstantProduct,
			Coins: []sdk.Coin{
				{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
				{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
			},
			Address: types.GetPoolAddress(invalidPoolID).String(),
		},
	}

	poolUidString := fullDenomAbc + types.PoolUidSeparator + fullDenomUsdt
	validPoolUidList := []types.PoolUids{
		{
			PoolUid: poolUidString,
			PoolId:  invalidPoolID,
		},
	}

	invalidGenesis := types.GenesisState{
		PoolList:     invalidPoolList,
		PoolUidsList: validPoolUidList,
		PoolsMeta:    &types.PoolsMeta{NextPoolId: 2},
		Params:       types.DefaultParams(),
	}

	err := invalidGenesis.Validate()
	require.Error(t, err)
	require.EqualError(
		t,
		err,
		fmt.Sprintf(
			"Pool fee too large: %d, has to be less than scaling factor: %d: invalid amount",
			constants.PoolFeeScalingFactor,
			constants.PoolFeeScalingFactor,
		),
	)
}

func TestGenesisState_Validate_InvalidPoolUidsPoolId(t *testing.T) {
	// Test case: GenesisState with an invalid PoolId in PoolUidsList (e.g., empty PoolId)

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	validPoolID := "zp1"
	validPoolList := []types.Pool{
		{
			PoolId:  validPoolID,
			LpToken: sdk.NewInt64Coin("zp1", 100),
			Creator: creator,
			Fee:     1,
			Formula: types.FormulaConstantProduct,
			Coins: []sdk.Coin{
				{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
				{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
			},
			Address: types.GetPoolAddress(validPoolID).String(),
		},
	}

	poolUidString := fullDenomAbc + types.PoolUidSeparator + fullDenomUsdt
	invalidPoolUidsList := []types.PoolUids{
		{
			PoolUid: poolUidString,
			PoolId:  "", // Invalid: empty PoolId
		},
	}

	invalidGenesis := types.GenesisState{
		PoolList:     validPoolList,
		PoolUidsList: invalidPoolUidsList,
		PoolsMeta:    &types.PoolsMeta{NextPoolId: 2},
		Params:       types.DefaultParams(),
	}

	err := invalidGenesis.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.EqualError(t, err, "Invalid pool id: pool id is empty: invalid coins")
}

func TestGenesisState_Validate_MissingPoolForPoolUids(t *testing.T) {
	// Test case: PoolUidsList references a PoolId that does not exist in PoolList

	creator := sample.AccAddress()
	fullDenomAbc := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "abc"
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	validPoolID := "zp1"
	missingPoolID := "zp2"
	validPoolList := []types.Pool{
		{
			PoolId:  validPoolID,
			LpToken: sdk.NewInt64Coin("zp1", 100),
			Creator: creator,
			Fee:     1,
			Formula: types.FormulaConstantProduct,
			Coins: []sdk.Coin{
				{Denom: fullDenomAbc, Amount: sdkmath.NewInt(1500)},
				{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
			},
			Address: types.GetPoolAddress(validPoolID).String(),
		},
	}

	poolUidString := fullDenomAbc + types.PoolUidSeparator + fullDenomUsdt
	invalidPoolUidsList := []types.PoolUids{
		{
			PoolUid: poolUidString,
			PoolId:  validPoolID,
		},
		{
			PoolUid: "different/denom1" + types.PoolUidSeparator + "denom2",
			PoolId:  missingPoolID,
		},
	}

	invalidGenesis := types.GenesisState{
		PoolList:     validPoolList,
		PoolUidsList: invalidPoolUidsList,
		PoolsMeta:    &types.PoolsMeta{NextPoolId: 2},
		Params:       types.DefaultParams(),
	}

	err := invalidGenesis.Validate()
	require.Error(t, err)
	require.EqualError(t, err, "missing PoolId 'zp2' from PoolUidsList in PoolList")
}

func TestPool_Validate_InvalidCoin(t *testing.T) {
	// Test case: Pool with an invalid coin in Coins slice (e.g., invalid denomination)

	creator := sample.AccAddress()
	fullDenomUsdt := "coin" + factorytypes.FactoryDenomDelimiterChar + creator + factorytypes.FactoryDenomDelimiterChar + "usdt"

	invalidPool := types.Pool{
		PoolId:  "zp1",
		LpToken: sdk.NewInt64Coin("zp1", 100),
		Creator: creator,
		Fee:     1,
		Formula: types.FormulaConstantProduct,
		Coins: []sdk.Coin{
			{Denom: "invalid_denom!", Amount: sdkmath.NewInt(1500)}, // Invalid denomination
			{Denom: fullDenomUsdt, Amount: sdkmath.NewInt(2500)},
		},
		Address: types.GetPoolAddress("zp1").String(),
	}

	err := invalidPool.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.Contains(t, err.Error(), "invalid_denom!")
	// The exact error message depends on the implementation of validators.CoinCheck.
	// Adjust the expected error message based on what validators.CoinCheck returns for an invalid denomination.
}
