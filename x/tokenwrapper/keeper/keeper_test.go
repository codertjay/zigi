package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	"zigchain/app"
	"zigchain/testutil/sample"
	"zigchain/x/tokenwrapper/types"
)

const (
	initChain = true
)

// Positive test cases

func TestLockTokens_Valid(t *testing.T) {
	// Test case: lock tokens successfully when sender has sufficient balance

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a test account with some coins
	creator := sample.AccAddress()
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)
	initialCoins := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000)))

	// mint coins to the mint module
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initialCoins))
	// send coins from mint module to the test account
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, initialCoins))

	// call the LockTokens method
	amount := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(200)))
	err := k.LockTokens(ctx, signer, amount)
	// check if the method executed without error
	require.NoError(t, err)

	// verify the coins were moved to the module account
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "uzig")
	require.Equal(t, sdkmath.NewInt(200), moduleBalance.Amount)

	// verify the sender's balance was reduced
	senderBalance := testApp.BankKeeper.GetBalance(ctx, signer, "uzig")
	require.Equal(t, sdkmath.NewInt(800), senderBalance.Amount)
}

func TestUnlockTokens_Valid(t *testing.T) {
	// Test case: unlock tokens from a module account successfully

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a test account with some coins
	creator := sample.AccAddress()
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)
	moduleCoins := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000)))

	// mint coins to the mint module
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, moduleCoins))
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, moduleCoins))

	// call the UnlockTokens method
	amount := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(100)))
	err := k.UnlockTokens(ctx, signer, amount)
	require.NoError(t, err)

	// verify the coins were moved from the module account
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "uzig")
	require.Equal(t, sdkmath.NewInt(900), moduleBalance.Amount)

	// verify the signer's balance was updated
	recipientBalance := testApp.BankKeeper.GetBalance(ctx, signer, "uzig")
	require.Equal(t, sdkmath.NewInt(100), recipientBalance.Amount)
}

func TestBurnIbcTokens_Valid(t *testing.T) {
	// Test case: burn IBC tokens from a module account

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	coins := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000)))

	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, coins))

	err := k.BurnIbcTokens(ctx, coins)
	require.NoError(t, err)

	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	balance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "uzig")
	require.True(t, balance.IsZero())
}

func TestGetModuleWalletBalances(t *testing.T) {
	// Test case: get the module wallet balances

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	// mint to a module account
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	coins := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000)))

	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins))
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, coins))

	// get balances
	addrStr, balances := k.GetModuleWalletBalances(ctx)
	require.Equal(t, moduleAddr.String(), addrStr)
	require.Equal(t, coins, balances)
}

func TestTotalTransferredIn_GetSet(t *testing.T) {
	// Test case: set and get the total transferred in value

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	value := sdkmath.NewInt(500)
	k.SetTotalTransferredIn(ctx, value)

	result := k.GetTotalTransferredIn(ctx)
	require.Equal(t, value, result)
}

func TestTotalTransferredIn_Add(t *testing.T) {
	// Test case: add to the total transferred in value

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	k.SetTotalTransferredIn(ctx, sdkmath.NewInt(1000))
	k.AddToTotalTransferredIn(ctx, sdkmath.NewInt(250))

	result := k.GetTotalTransferredIn(ctx)
	require.Equal(t, sdkmath.NewInt(1250), result)
}

func TestTotalTransferredOut_GetSet(t *testing.T) {
	// Test case: set and get the total transferred out value

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	value := sdkmath.NewInt(300)
	k.SetTotalTransferredOut(ctx, value)

	result := k.GetTotalTransferredOut(ctx)
	require.Equal(t, value, result)
}

func TestTotalTransferredOut_Add(t *testing.T) {
	// Test case: add to the total transferred out value

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	k.SetTotalTransferredOut(ctx, sdkmath.NewInt(400))
	k.AddToTotalTransferredOut(ctx, sdkmath.NewInt(100))

	result := k.GetTotalTransferredOut(ctx)
	require.Equal(t, sdkmath.NewInt(500), result)
}

func TestSetGetOperatorAddress(t *testing.T) {
	// Test case: set and get the operator address

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	addr := sample.AccAddress()
	k.SetOperatorAddress(ctx, addr)
	require.Equal(t, addr, k.GetOperatorAddress(ctx))
}

func TestSetGetEnabled(t *testing.T) {
	// Test case: set and get the enabled status

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	k.SetEnabled(ctx, true)
	require.True(t, k.IsEnabled(ctx))

	k.SetEnabled(ctx, false)
	require.False(t, k.IsEnabled(ctx))
}

func TestSetGetNativeClientId(t *testing.T) {
	// Test case: set and get the native client ID

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	clientId := "ibc-native-client-1"
	k.SetNativeClientId(ctx, clientId)
	require.Equal(t, clientId, k.GetNativeClientId(ctx))
}

func TestSetGetCounterpartyClientId(t *testing.T) {
	// Test case: set and get the counterparty client ID

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	clientId := "ibc-counterparty-client-1"
	k.SetCounterpartyClientId(ctx, clientId)
	require.Equal(t, clientId, k.GetCounterpartyClientId(ctx))
}

func TestSetGetNativePort(t *testing.T) {
	// Test case: set and get the native port

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	port := "transfer"
	k.SetNativePort(ctx, port)
	require.Equal(t, port, k.GetNativePort(ctx))
}

func TestSetGetCounterpartyPort(t *testing.T) {
	// Test case: set and get the counterparty port

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	port := "transfer"
	k.SetCounterpartyPort(ctx, port)
	require.Equal(t, port, k.GetCounterpartyPort(ctx))
}

func TestSetGetNativeChannel(t *testing.T) {
	// Test case: set and get the native channel

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	channel := "channel-0"
	k.SetNativeChannel(ctx, channel)
	require.Equal(t, channel, k.GetNativeChannel(ctx))
}

func TestSetGetCounterpartyChannel(t *testing.T) {
	// Test case: set and get the counterparty channel

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	channel := "channel-0"
	k.SetCounterpartyChannel(ctx, channel)
	require.Equal(t, channel, k.GetCounterpartyChannel(ctx))
}

func TestSetGetDecimalDifference(t *testing.T) {
	// Test case: set and get the decimal difference

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	_ = k.SetDecimalDifference(ctx, 18)
	require.Equal(t, uint32(18), k.GetDecimalDifference(ctx))
}

func TestSetGetDenom(t *testing.T) {
	// Test case: set and get the denom

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	denom := "uzig"
	k.SetDenom(ctx, denom)
	require.Equal(t, denom, k.GetDenom(ctx))
}

func TestGetDecimalConversionFactor(t *testing.T) {
	// Test case: returns correct 10^decimalDifference value

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	_ = k.SetDecimalDifference(ctx, 6)
	conversionFactor := k.GetDecimalConversionFactor(ctx)

	expected := sdkmath.NewIntFromUint64(1_000_000)
	require.Equal(t, expected, conversionFactor)
}

func TestDecimalConversionRounding_ScaleDown(t *testing.T) {
	// Test case: verify that decimal conversion using Quo rounds down

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	// Set up test cases with amounts that would have decimal places
	testCases := []struct {
		name           string
		amount         sdkmath.Int
		decimalDiff    uint32
		expectedResult sdkmath.Int
	}{
		{
			name:           "exact division",
			amount:         sdkmath.NewInt(1000000000000000000), // 1.0 with 18 decimals
			decimalDiff:    12,                                  // 18-6 decimals
			expectedResult: sdkmath.NewInt(1000000),             // 1.0 with 6 decimals
		},
		{
			name:           "rounding down with remainder",
			amount:         sdkmath.NewInt(1234567890000000000), // 1.23456789 with 18 decimals
			decimalDiff:    12,                                  // 18-6 decimals
			expectedResult: sdkmath.NewInt(1234567),             // 1.234567 with 6 decimals
		},
		{
			name:           "small amount that rounds to zero",
			amount:         sdkmath.NewInt(999999999999), // 0.000999999999999 with 18 decimals
			decimalDiff:    12,                           // 18-6 decimals
			expectedResult: sdkmath.NewInt(0),            // rounds down to 0
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the decimal difference in the keeper
			_ = k.SetDecimalDifference(ctx, tc.decimalDiff)

			// Get the conversion factor
			conversionFactor := k.GetDecimalConversionFactor(ctx)

			// Perform the division
			result := tc.amount.Quo(conversionFactor)

			// Verify the result
			require.Equal(t, tc.expectedResult.String(), result.String(),
				"Quo operation did not round down correctly. Expected %s, got %s",
				tc.expectedResult.String(), result.String())
		})
	}
}

func TestScaleUpTokenPrecision_Valid(t *testing.T) {
	// Test case: scale up token precision successfully

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	// Set decimal difference to 12 (converting from 6 to 18 decimals)
	_ = k.SetDecimalDifference(ctx, 12)

	// Test scaling up 1.0 with 6 decimals to 1.0 with 18 decimals
	amount := sdkmath.NewInt(1_000_000)                   // 1.0 with 6 decimals
	expected := sdkmath.NewInt(1_000_000_000_000_000_000) // 1.0 with 18 decimals

	result, err := k.ScaleUpTokenPrecision(ctx, amount)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestScaleUpTokenPrecision_WithRemainder(t *testing.T) {
	// Test case: scale up token precision with fractional amount

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	// Set decimal difference to 12 (converting from 6 to 18 decimals)
	_ = k.SetDecimalDifference(ctx, 12)

	// Test scaling up 1.234567 with 6 decimals
	amount := sdkmath.NewInt(1_234_567)                   // 1.234567 with 6 decimals
	expected := sdkmath.NewInt(1_234_567_000_000_000_000) // 1.234567 with 18 decimals

	result, err := k.ScaleUpTokenPrecision(ctx, amount)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestScaleUpTokenPrecision_ZeroDecimalDifference(t *testing.T) {
	// Test case: scale up with zero decimal difference (no conversion needed)

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	// Set decimal difference to 0 (same decimals)
	_ = k.SetDecimalDifference(ctx, 0)

	amount := sdkmath.NewInt(1_000_000)
	expected := sdkmath.NewInt(1_000_000) // Should remain the same

	result, err := k.ScaleUpTokenPrecision(ctx, amount)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestScaleUpTokenPrecision_MaxDecimalDifference(t *testing.T) {
	// Test case: scale up with maximum decimal difference (18)

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	// Set decimal difference to 18 (converting from 0 to 18 decimals)
	_ = k.SetDecimalDifference(ctx, 18)

	amount := sdkmath.NewInt(1)                           // 1 with 0 decimals
	expected := sdkmath.NewInt(1_000_000_000_000_000_000) // 1 with 18 decimals

	result, err := k.ScaleUpTokenPrecision(ctx, amount)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestGetPauserAddressesEmpty(t *testing.T) {
	// Test case: get pauser addresses when none are set

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	// Should return empty slice when no addresses are set
	addresses := k.GetPauserAddresses(ctx)
	require.Empty(t, addresses)
	require.Equal(t, []string{}, addresses)
}

func TestSetPauserAddresses_OneAddressValid(t *testing.T) {
	// Test case: set and get one pauser address

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	// Test with single address
	addresses := []string{"zig1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq9hc4ar"}
	k.SetPauserAddresses(ctx, addresses)
	require.Equal(t, addresses, k.GetPauserAddresses(ctx))
}

func TestSetPauserAddresses_MultipleAddressValid(t *testing.T) {
	// Test case: set and get multiple pauser addresses

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	// Test with multiple addresses
	addresses := []string{
		"zig1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq9hc4ar",
		"zig1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"zig1bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}
	k.SetPauserAddresses(ctx, addresses)
	require.Equal(t, addresses, k.GetPauserAddresses(ctx))
}

func TestAddPauserAddress(t *testing.T) {
	// Test case: add pauser addresses

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	address1 := "zig1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq9hc4ar"
	address2 := "zig1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	// Add first address
	k.AddPauserAddress(ctx, address1)
	addresses := k.GetPauserAddresses(ctx)
	require.Len(t, addresses, 1)
	require.Contains(t, addresses, address1)

	// Add second address
	k.AddPauserAddress(ctx, address2)
	addresses = k.GetPauserAddresses(ctx)
	require.Len(t, addresses, 2)
	require.Contains(t, addresses, address1)
	require.Contains(t, addresses, address2)

	// Try to add duplicate address - should not increase length
	k.AddPauserAddress(ctx, address1)
	addresses = k.GetPauserAddresses(ctx)
	require.Len(t, addresses, 2)
	require.Contains(t, addresses, address1)
	require.Contains(t, addresses, address2)
}

func TestAddPauserAddress_SameAddress(t *testing.T) {
	// Test case: adding the same address multiple times should not create duplicates

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	address := "zig1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq9hc4ar"

	// Add address multiple times
	k.AddPauserAddress(ctx, address)
	k.AddPauserAddress(ctx, address)
	k.AddPauserAddress(ctx, address)

	// Should only appear once
	addresses := k.GetPauserAddresses(ctx)
	require.Len(t, addresses, 1)
	require.Equal(t, address, addresses[0])
	require.True(t, k.IsPauserAddress(ctx, address))
}

func TestRemovePauserAddress(t *testing.T) {
	// Test case: remove pauser addresses

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	address1 := "zig1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq9hc4ar"
	address2 := "zig1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	address3 := "zig1bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	// Set up initial addresses
	initialAddresses := []string{address1, address2, address3}
	k.SetPauserAddresses(ctx, initialAddresses)

	// Remove middle address
	k.RemovePauserAddress(ctx, address2)
	addresses := k.GetPauserAddresses(ctx)
	require.Len(t, addresses, 2)
	require.Contains(t, addresses, address1)
	require.Contains(t, addresses, address3)
	require.NotContains(t, addresses, address2)

	// Remove first address
	k.RemovePauserAddress(ctx, address1)
	addresses = k.GetPauserAddresses(ctx)
	require.Len(t, addresses, 1)
	require.Contains(t, addresses, address3)
	require.NotContains(t, addresses, address1)

	// Remove last address
	k.RemovePauserAddress(ctx, address3)
	addresses = k.GetPauserAddresses(ctx)
	require.Empty(t, addresses)
}

func TestRemovePauserAddress_NotExistingValid(t *testing.T) {
	// Test case: remove pauser addresses

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	address1 := "zig1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq9hc4ar"
	address2 := "zig1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	address3 := "zig1bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	// Set up initial addresses
	initialAddresses := []string{address1, address2, address3}
	k.SetPauserAddresses(ctx, initialAddresses)

	// Try to remove non-existent address - should not panic
	k.RemovePauserAddress(ctx, "zig1nonexistent")
	addresses := k.GetPauserAddresses(ctx)
	require.Contains(t, addresses, address1)
	require.Contains(t, addresses, address2)
	require.Contains(t, addresses, address3)
}

func TestRemovePauserAddressFromEmpty(t *testing.T) {
	// Test case: remove pauser address when the list is empty

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	// Try to remove from empty list - should not panic
	k.RemovePauserAddress(ctx, "zig1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq9hc4ar")
	addresses := k.GetPauserAddresses(ctx)
	require.Empty(t, addresses)
}

func TestIsPauserAddress(t *testing.T) {
	// Test case: check if address is a pauser

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	address1 := "zig1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq9hc4ar"
	address2 := "zig1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	nonPauserAddress := "zig1bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	// Initially, no addresses should be pausers
	require.False(t, k.IsPauserAddress(ctx, address1))
	require.False(t, k.IsPauserAddress(ctx, address2))
	require.False(t, k.IsPauserAddress(ctx, nonPauserAddress))

	// Set pauser addresses addresses 1 and 2
	addresses := []string{address1, address2}
	k.SetPauserAddresses(ctx, addresses)

	// Check pauser addresses
	require.True(t, k.IsPauserAddress(ctx, address1))
	require.True(t, k.IsPauserAddress(ctx, address2))
	require.False(t, k.IsPauserAddress(ctx, nonPauserAddress))
}

// Negative test cases

func TestLockTokens_InsufficientBalance(t *testing.T) {
	// Test case: try to lock tokens if sender has insufficient balance

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a test account with some coins
	creator := sample.AccAddress()
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)
	initialCoins := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000)))
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initialCoins))
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, signer, initialCoins))

	// call the LockTokens method with an amount greater than the balance
	amount := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(2000)))
	err := k.LockTokens(ctx, signer, amount)
	// check if the method returned an error
	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		fmt.Sprintf(
			"account %s does not have enough balance of 2000uzig",
			signer,
		),
		err.Error(),
	)
}

func TestLockTokens_InvalidDenom(t *testing.T) {
	// Test case: try to lock tokens if sender has insufficient balance

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a test account with some coins
	creator := sample.AccAddress()
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)

	// call the LockTokens method with an amount greater than the balance
	amount := sdk.NewCoins(sdk.NewCoin("abc", sdkmath.NewInt(2000)))
	err := k.LockTokens(ctx, signer, amount)
	// check if the method returned an error
	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		fmt.Sprintf(
			"account %s does not have enough balance of 2000abc",
			signer,
		),
		err.Error(),
	)
}

func TestUnlockTokens_InsufficientBalance(t *testing.T) {
	// Test case: try to unlock tokens if a module account has insufficient balance

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a test account with some coins
	creator := sample.AccAddress()
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)
	moduleCoins := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000)))
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, moduleCoins))
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, moduleCoins))

	// call the UnlockTokens method with an amount greater than the balance
	amount := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(2000)))
	err := k.UnlockTokens(ctx, signer, amount)
	// check if the method returned an error
	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		fmt.Sprintf(
			"module does not have enough balance of %s",
			amount,
		),
		err.Error(),
	)
}

func TestUnlockTokens_InvalidDenom(t *testing.T) {
	// Test case: try to unlock tokens if a module account has insufficient balance

	// initialize test blockchain app
	// simulating Cosmos SDK module environment
	testApp := app.InitTestApp(initChain, t)
	// retrieve the tokenwrapper keeper
	k := testApp.TokenwrapperKeeper
	// create a fresh context
	ctx := testApp.BaseApp.NewContext(initChain)

	// generate a test account with some coins
	creator := sample.AccAddress()
	// Bech32 address for verifying method calls
	signer := sdk.MustAccAddressFromBech32(creator)
	moduleCoins := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000)))
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, moduleCoins))
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, moduleCoins))

	// call the UnlockTokens method with an amount greater than the balance
	amount := sdk.NewCoins(sdk.NewCoin("abc", sdkmath.NewInt(2000)))
	err := k.UnlockTokens(ctx, signer, amount)
	// check if the method returned an error
	require.Error(t, err)

	// check if the error message is correct
	require.Equal(t,
		fmt.Sprintf(
			"module does not have enough balance of %s",
			amount,
		),
		err.Error(),
	)
}

func TestBurnIbcTokens_InvalidBalance(t *testing.T) {
	// Test case: burn tokens not present in a module account

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	coins := sdk.NewCoins(sdk.NewCoin("uzig", sdkmath.NewInt(1000)))

	err := k.BurnIbcTokens(ctx, coins)
	require.Error(t, err)
	require.Contains(t, err.Error(), "module does not have enough balance of 1000uzig")
}

func TestSetDecimalDifference_Validation(t *testing.T) {
	// Test case: validate decimal difference range (0-18)

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	// Test valid cases
	validCases := []uint32{0, 9, 18}
	for _, validCase := range validCases {
		err := k.SetDecimalDifference(ctx, validCase)
		require.NoError(t, err)
		require.Equal(t, validCase, k.GetDecimalDifference(ctx))
	}

	// Test invalid case
	err := k.SetDecimalDifference(ctx, 19)
	require.Error(t, err)
	require.Contains(t, err.Error(), "decimal difference must be between 0 and 18")
}

func TestScaleUpTokenPrecision_ZeroAmount(t *testing.T) {
	// Test case: scale up zero amount (should return error)

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	_ = k.SetDecimalDifference(ctx, 12)

	amount := sdkmath.NewInt(0)

	result, err := k.ScaleUpTokenPrecision(ctx, amount)
	require.Error(t, err)
	require.True(t, result.IsNil() || result.IsZero())
	require.Contains(t, err.Error(), "converted amount is zero or negative")
}

func TestScaleUpTokenPrecision_NegativeAmount(t *testing.T) {
	// Test case: scale up negative amount (should return error)

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	_ = k.SetDecimalDifference(ctx, 12)

	amount := sdkmath.NewInt(-1000)

	result, err := k.ScaleUpTokenPrecision(ctx, amount)
	require.Error(t, err)
	require.True(t, result.IsNil() || result.IsNegative())
	require.Contains(t, err.Error(), "converted amount is zero or negative")
}

func TestScaleDownTokenPrecision_ZeroAmount(t *testing.T) {
	// Test case: scale down zero amount (should return error)

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	_ = k.SetDecimalDifference(ctx, 12)

	amount := sdkmath.NewInt(0)

	result, err := k.ScaleDownTokenPrecision(ctx, amount)
	require.Error(t, err)
	require.True(t, result.IsNil() || result.IsZero())
	require.Contains(t, err.Error(), "converted amount is zero or negative")
}

func TestScaleDownTokenPrecision_NegativeAmount(t *testing.T) {
	// Test case: scale down negative amount (should return error)

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	_ = k.SetDecimalDifference(ctx, 12)

	amount := sdkmath.NewInt(-1000)

	result, err := k.ScaleDownTokenPrecision(ctx, amount)
	require.Error(t, err)
	require.True(t, result.IsNil() || result.IsNegative())
	require.Contains(t, err.Error(), "converted amount is zero or negative")
}

func TestUnlockNativeTokens_UnlockFailsWithRecovery(t *testing.T) {
	// Test case: UnlockTokens fails but successfully recovers locked IBC tokens

	testApp := app.InitTestApp(initChain, t)
	k := testApp.TokenwrapperKeeper
	ctx := testApp.BaseApp.NewContext(initChain)

	// Setup: Create a receiver account with some IBC tokens locked in module
	receiver := sample.AccAddress()
	receiverAddr := sdk.MustAccAddressFromBech32(receiver)

	// Create IBC coins that were previously locked
	ibcCoins := sdk.NewCoins(sdk.NewCoin("ibc/ABC123", sdkmath.NewInt(1000)))

	// Mint IBC tokens to module (simulating they were previously locked)
	require.NoError(t, testApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, ibcCoins))
	require.NoError(t, testApp.BankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, types.ModuleName, ibcCoins))

	// Create a mock amount for native tokens that will fail to unlock
	amount := sdkmath.NewInt(1000)

	// The UnlockTokens should fail because the module doesn't have enough native tokens
	result, err := k.UnlockNativeTokens(ctx, receiverAddr, amount, ibcCoins)

	// Should return the original error
	require.Error(t, err)
	require.True(t, result.IsZero())
	require.Contains(t, err.Error(), "failed to unlock tokens")

	// Verify that the IBC tokens were successfully recovered (unlocked back to receiver)
	receiverBalance := testApp.BankKeeper.GetBalance(ctx, receiverAddr, "ibc/ABC123")
	require.Equal(t, sdkmath.NewInt(1000), receiverBalance.Amount)

	// Verify module no longer has the IBC tokens
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	moduleBalance := testApp.BankKeeper.GetBalance(ctx, moduleAddr, "ibc/ABC123")
	require.True(t, moduleBalance.IsZero())
}
