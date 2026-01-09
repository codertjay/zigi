package validators_test

import (
	"fmt"
	"strings"
	"testing"

	errorPacks "zigchain/testutil/data"
	"zigchain/testutil/sample"
	"zigchain/zutils/constants"
	"zigchain/zutils/validators"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Positive test cases

func TestCoinCheckValid(t *testing.T) {
	// Test case: valid coin

	validCoin := sample.Coin("abc", 100)

	err := validators.CoinCheck(validCoin, true)
	require.NoError(t, err, "expected no error for valid coin")
}

func TestCoinCheck_ZeroOK(t *testing.T) {
	// Test case: valid coin

	validCoin := sample.Coin("abc", 0)

	err := validators.CoinCheck(validCoin, true)
	require.NoError(t, err, "expected no error for valid coin")
}

func TestCoinCheckAmount_ValidAmount(t *testing.T) {
	// Test case: valid coins when checking amount for zero ok and not zero ok

	// Loop through the test cases
	for _, tc := range []struct {
		desc   string
		amount int64
		zeroOK bool
	}{
		{
			desc:   "Valid Amount - Zero Ok: Zero",
			amount: 0,
			zeroOK: true,
		},
		{
			desc:   "Valid Amount - Zero Ok: Smallest positive amount",
			amount: 1,
			zeroOK: true,
		},
		{
			desc:   "Valid Amount - Zero Ok: Regular number",
			amount: 123456,
			zeroOK: true,
		},
		{
			desc:   "Valid Amount - Zero Ok: max int64 value",
			amount: 9223372036854775807,
			zeroOK: true,
		},
		{
			desc:   "Valid Amount - Zero Not Ok: Smallest positive amount",
			amount: 1,
			zeroOK: false,
		},
		{
			desc:   "Valid Amount - Zero Not Ok: Regular number",
			amount: 123456,
			zeroOK: false,
		},
		{
			desc:   "Valid Amount - Zero not Ok: max int64 value",
			amount: 9223372036854775807,
			zeroOK: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {

			coin := sdk.Coin{Denom: "lpk", Amount: math.NewInt(tc.amount)}
			err := validators.CheckCoinAmount(coin, tc.zeroOK)

			require.NoError(t, err, "expected no error for valid coin denom")

			// Assert that the error is nil
			require.NoError(t, err)
		})
	}

}

func TestCheckDenomString_ValidDenom(t *testing.T) {
	// Test cases for CheckDenomString function with valid inputs for denom
	// The function should return no error

	// Loop through the denom test cases
	// Denom default regex is [a-zA-Z][a-zA-Z0-9./]+
	for _, tc := range []struct {
		desc  string
		denom string
	}{
		{
			desc:  "normal",
			denom: "bitcoin",
		},
		{
			desc:  "denom is exactly min length 3 characters",
			denom: "bit",
		},
		{
			desc:  "denom is exactly max length 127 characters",
			denom: "bitcoinfsadfghjklqwertyuiopzxcvbnmasdfghjklqwertyuiopzxcvbnmasdfghjklqwertyuiopzxcvbnm",
		},
		{
			desc:  "denom includes all the allowed numbers",
			denom: "bit0123456789",
		},
		{
			desc:  "denom includes all the allowed letters lowercase",
			denom: "abcdefghijklmnopqrstuvwxyz",
		},
		{
			desc:  "denom includes all the allowed letters uppercase",
			denom: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
		{
			desc:  "denom includes all the allowed special characters",
			denom: "bit/.",
		},
		{
			desc:  "denom includes letters, numbers and special characters",
			denom: "bit/.0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
		{
			desc:  "normal",
			denom: "coin.zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5.bitcoin",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {

			err := validators.CheckDenomString(tc.denom)
			require.NoError(t, err, "expected no error for valid coin denom")

			// Assert that the error is nil
			require.NoError(t, err)
		})
	}
}

func TestCheckCoinDenom_ValidDenom(t *testing.T) {
	// Test cases for CheckCoinDenom function with valid inputs for denom
	// The function should return no error

	// Loop through the denom test cases
	// Denom default regex is [a-zA-Z][a-zA-Z0-9./]+
	for _, tc := range []struct {
		desc  string
		denom string
	}{
		{
			desc:  "normal",
			denom: "bitcoin",
		},
		{
			desc:  "denom is exactly min length 3 characters",
			denom: "bit",
		},
		{
			desc:  "denom is exactly max length 127 characters",
			denom: "bitcoinfsadfghjklqwertyuiopzxcvbnmasdfghjklqwertyuiopzxcvbnmasdfghjklqwertyuiopzxcvbnm",
		},
		{
			desc:  "denom includes all the allowed numbers",
			denom: "bit0123456789",
		},
		{
			desc:  "denom includes all the allowed letters lowercase",
			denom: "abcdefghijklmnopqrstuvwxyz",
		},
		{
			desc:  "denom includes all the allowed letters uppercase",
			denom: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
		{
			desc:  "denom includes all the allowed special characters",
			denom: "bit/.",
		},
		{
			desc:  "denom includes letters, numbers and special characters",
			denom: "bit/.0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
		{
			desc:  "normal",
			denom: "coin.zig1vm3v4yrd3rrwkf3fe8qxutaz27098t76270qc5.bitcoin",
		},
		{
			desc:  "mix upper and lowercase letters",
			denom: "BitCoin",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			coin := sdk.Coin{Denom: tc.denom, Amount: math.NewInt(100)}
			err := validators.CheckCoinDenom(coin)
			require.NoError(t, err, "expected no error for valid coin denom")

			// Assert that the error is nil
			require.NoError(t, err)
		})
	}
}

func TestCheckPoolId_Valid(t *testing.T) {
	// Test cases for valid pool IDs
	// Pool ID must start with constants.PoolPrefix ("zp") followed by numbers
	testCases := []struct {
		desc   string
		poolId string
	}{
		{
			desc:   "valid pool ID: minimum length",
			poolId: "zp1",
		},
		{
			desc:   "valid pool ID: typical case",
			poolId: "zp123",
		},
		{
			desc:   "valid pool ID: maximum length",
			poolId: "zp" + strings.Repeat("1", constants.MaxSubDenomLength-2),
		},
		{
			desc:   "valid pool ID: multiple digits",
			poolId: "zp1234567890",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := validators.CheckPoolId(tc.poolId)
			require.NoError(t, err, "expected no error for valid pool ID: %s", tc.poolId)
		})
	}
}

// Negative test cases

func TestCoinCheck_InvalidDenom(t *testing.T) {
	// Test case: invalid coin denom

	invalidCoin := sample.Coin("ab", 10)

	err := validators.CoinCheck(invalidCoin, false)
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.Equal(
		t,
		fmt.Sprintf(
			"invalid coin: '%s' denom name is too short for 10ab,"+
				" minimum %d characters e.g. 10uzig: invalid coins",
			invalidCoin.Denom,
			constants.MinSubDenomLength,
		),
		err.Error(),
	)
}

func TestCoinCheckAmount_InvalidAmountZeroOk(t *testing.T) {
	// Test case: invalid coins when checking amount

	// Get the invalid Coins from the error pack
	var InvalidDenomAmountZeroOK = &errorPacks.InvalidDenomAmountZeroOK

	// Loop through the invalid Coins
	for _, tc := range *InvalidDenomAmountZeroOK {
		t.Run(tc.TestName, func(t *testing.T) {

			// Extract the denom string from the coin
			err := validators.CheckCoinAmount(tc.FieldValue, true)

			require.Error(t, err)

			// Assert that the error message is equal to the expected error message
			// append to the Error Text ": invalid coins
			var errorText = tc.ErrorText + ": invalid coins"

			assert.Equal(t, errorText, err.Error())

		})
	}
}

func TestCoinCheckAmount_InvalidAmountZeroNotOk(t *testing.T) {
	// Test case: invalid coins when checking amount and zero is not allowed

	// Get the invalid Coins from the error pack
	var InvalidDenomAmountZeroNotOK = &errorPacks.InvalidDenomZeroAmountNotOKOnlyAmounts

	// Loop through the invalid
	for _, tc := range *InvalidDenomAmountZeroNotOK {
		t.Run(tc.TestName, func(t *testing.T) {

			// Extract the denom string from the coin
			err := validators.CheckCoinAmount(tc.FieldValue, false)

			require.Error(t, err)

			// Assert that the error message is equal to the expected error message
			// append to the Error Text ": invalid coins
			var errorText = tc.ErrorText + ": invalid coins"

			assert.Equal(t, errorText, err.Error())

		})
	}
}

func TestCheckDenomString_InvalidDenom(t *testing.T) {
	// Test case: invalid denom
	// The function should return an error

	// Get the invalid Subdenom from the error pack
	var InvalidDenomNameString = &errorPacks.InvalidDenomNameString

	// Loop through the invalid SubDenoms
	for _, tc := range *InvalidDenomNameString {
		t.Run(tc.TestName, func(t *testing.T) {

			// Extract the denom string from the coin
			err := validators.CheckDenomString(tc.FieldValue)

			require.Error(t, err)

			// Assert that the error message is equal to the expected error message
			// append to the Error Text ": invalid coins
			var errorText = tc.ErrorText + ": invalid coins"

			assert.Equal(t, errorText, err.Error())

		})
	}
}

func TestCheckCoinDenom_InvalidDenom(t *testing.T) {
	// Test case: invalid denom
	// The function should return an error

	// Get the invalid Subdenom from the error pack
	var InvalidDenomNameString = &errorPacks.InvalidDenomName

	// Loop through the invalid SubDenoms
	for _, tc := range *InvalidDenomNameString {
		t.Run(tc.TestName, func(t *testing.T) {

			// Extract the denom string from the coin
			err := validators.CheckCoinDenom(tc.FieldValue)

			require.Error(t, err)

			// Assert that the error message is equal to the expected error message
			// append to the Error Text ": invalid coins
			var errorText = tc.ErrorText + ": invalid coins"

			assert.Equal(t, errorText, err.Error())

		})
	}
}

func TestCoinCheck_InvalidAmount(t *testing.T) {
	// Test case: invalid coin amounts in CoinCheck
	// Covers the error path where CheckCoinAmount returns an error

	validDenom := "uzig"

	// Test cases for invalid coin amounts
	testCases := []struct {
		desc      string
		coin      sdk.Coin
		zeroOK    bool
		errorText string
	}{
		{
			desc: "Nil Amount",
			coin: sdk.Coin{
				Denom:  validDenom,
				Amount: math.Int{}, // Nil amount
			},
			zeroOK:    true,
			errorText: fmt.Sprintf("invalid coin amount: cannot be nil (%s)", (sdk.Coin{Denom: validDenom, Amount: math.Int{}}).String()),
		},
		{
			desc: "Negative Amount",
			coin: sdk.Coin{
				Denom:  validDenom,
				Amount: math.NewInt(-10),
			},
			zeroOK:    true,
			errorText: fmt.Sprintf("invalid coin amount: -10 cannot be negative (%s)", (sdk.Coin{Denom: validDenom, Amount: math.NewInt(-10)}).String()),
		},
		{
			desc: "Zero Amount with ZeroOK false",
			coin: sdk.Coin{
				Denom:  validDenom,
				Amount: math.NewInt(0),
			},
			zeroOK:    false,
			errorText: fmt.Sprintf("invalid coin amount: 0 has to be positive (%s)", (sdk.Coin{Denom: validDenom, Amount: math.NewInt(0)}).String()),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := validators.CoinCheck(tc.coin, tc.zeroOK)
			require.Error(t, err, "expected an error for invalid coin amount")
			require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
			require.Equal(t, tc.errorText+": invalid coins", err.Error(), "error message mismatch")
		})
	}
}

func TestCheckSubDenomString_Empty(t *testing.T) {
	// Test case: empty sub-denom string

	emptyDenom := ""

	err := validators.CheckSubDenomString(emptyDenom)
	require.Error(t, err, "expected an error for empty sub-denom string")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.Equal(
		t,
		"Invalid subdenom name: denom name is empty e.g. uzig: invalid coins",
		err.Error(),
	)
}

func TestCheckSubDenomString_TooShort(t *testing.T) {
	// Test case: sub-denom string shorter than minimum length

	shortDenom := "ab"

	err := validators.CheckSubDenomString(shortDenom)
	require.Error(t, err, "expected an error for sub-denom string too short")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.Equal(
		t,
		fmt.Sprintf(
			"invalid coin: '%s' denom name is too short, minimum %d characters e.g. uzig: invalid coins",
			shortDenom,
			constants.MinSubDenomLength,
		),
		err.Error(),
	)
}

func TestCheckSubDenomString_TooLong(t *testing.T) {
	// Test case: sub-denom string longer than maximum length

	// Create a string longer than MaxSubDenomLength (44 characters)
	longDenom := "a" + strings.Repeat("b", constants.MaxSubDenomLength)

	err := validators.CheckSubDenomString(longDenom)
	require.Error(t, err, "expected an error for sub-denom string too long")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.Equal(
		t,
		fmt.Sprintf(
			"invalid coin: '%s' denom name is too long (%d), maximum %d characters e.g. uzig: invalid coins",
			longDenom,
			len(longDenom),
			constants.MaxSubDenomLength,
		),
		err.Error(),
	)
}

func TestCheckSubDenomString_InvalidFirstChar(t *testing.T) {
	// Test case: sub-denom string with invalid first character (not a lowercase letter)

	// Use an uppercase letter as the first character to trigger the error
	invalidDenom := "Abcdef"

	err := validators.CheckSubDenomString(invalidDenom)
	require.Error(t, err, "expected an error for sub-denom string with invalid first character")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.Equal(
		t,
		fmt.Sprintf(
			"invalid coin: '%s' denom name has to start with a lowercase letter e.g. uzig: invalid coins",
			invalidDenom,
		),
		err.Error(),
	)
}

func TestCheckSubDenomString_NonAlphanumeric(t *testing.T) {
	// Test case: sub-denom string with non-alphanumeric characters after the first character

	// Use a string with a special character (_) to trigger the non-alphanumeric error
	invalidDenom := "abc_def" // Starts with 'a', contains '_' which is neither a-z nor 0-9

	err := validators.CheckSubDenomString(invalidDenom)
	require.Error(t, err, "expected an error for sub-denom string with non-alphanumeric characters")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.Equal(
		t,
		fmt.Sprintf(
			"invalid coin: '%s' only lowercase letters (a-z) and numbers (0-9) are allowed e.g. uzig123: invalid coins",
			invalidDenom,
		),
		err.Error(),
	)
}

func TestCheckPoolId_Empty(t *testing.T) {
	// Test case: empty pool ID

	err := validators.CheckPoolId("")
	require.Error(t, err, "expected an error for empty pool ID")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.Equal(t, "Invalid pool id: pool id is empty: invalid coins", err.Error())
}

func TestCheckPoolId_TooShort(t *testing.T) {
	// Test case: pool ID shorter than minimum length

	shortPoolId := "zp"
	err := validators.CheckPoolId(shortPoolId)
	require.Error(t, err, "expected an error for pool ID too short")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.Equal(t,
		fmt.Sprintf(
			"Invalid pool id: '%s' pool id is too short, minimum %d characters: invalid coins",
			shortPoolId,
			constants.MinSubDenomLength,
		),
		err.Error(),
	)
}

func TestCheckPoolId_TooLong(t *testing.T) {
	// Test case: pool ID longer than maximum length

	longPoolId := "zp" + strings.Repeat("1", constants.MaxSubDenomLength)
	err := validators.CheckPoolId(longPoolId)
	require.Error(t, err, "expected an error for pool ID too long")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.Equal(t,
		fmt.Sprintf(
			"Invalid pool id: '%s' pool id is too long (%d), maximum %d characters: invalid coins",
			longPoolId,
			len(longPoolId),
			constants.MaxSubDenomLength,
		),
		err.Error(),
	)
}

func TestCheckPoolId_InvalidPrefix(t *testing.T) {
	// Test case: pool ID with invalid prefix

	invalidPoolId := "xp123"
	err := validators.CheckPoolId(invalidPoolId)
	require.Error(t, err, "expected an error for pool ID with invalid prefix")
	require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
	require.Equal(t,
		fmt.Sprintf(
			"Invalid pool id: '%s', pool id has to start with '%s' followed by numbers e.g. %s123: invalid coins",
			invalidPoolId,
			constants.PoolPrefix,
			constants.PoolPrefix,
		),
		err.Error(),
	)
}

func TestCheckPoolId_InvalidFormat(t *testing.T) {
	// Test case: pool ID with invalid format

	testCases := []struct {
		desc   string
		poolId string
	}{
		{
			desc:   "pool ID with letters after prefix",
			poolId: "zpabc",
		},
		{
			desc:   "pool ID with special characters",
			poolId: "zp12#",
		},
		{
			desc:   "pool ID with mixed characters",
			poolId: "zp12ab34",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := validators.CheckPoolId(tc.poolId)
			require.Error(t, err, "expected an error for invalid pool ID format: %s", tc.poolId)
			require.ErrorIs(t, err, sdkerrors.ErrInvalidCoins)
			require.Equal(t,
				fmt.Sprintf(
					"Invalid pool id: '%s', pool id has to start with '%s' followed by numbers e.g. %s123: invalid coins",
					tc.poolId,
					constants.PoolPrefix,
					constants.PoolPrefix,
				),
				err.Error(),
			)
		})
	}
}
