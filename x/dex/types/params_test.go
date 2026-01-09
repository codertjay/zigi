package types_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"

	"zigchain/testutil/sample"
	"zigchain/x/dex/types"
	"zigchain/zutils/constants"

	"github.com/stretchr/testify/require"
)

// Positive test cases

func TestNewParams(t *testing.T) {
	// Test case: create new params with valid values
	newPoolFeePct := uint32(500)
	creationFee := uint32(100000000)
	beneficiary := sample.AccAddress()
	minimalLiquidityLock := uint32(1000)

	params := types.NewParams(newPoolFeePct, creationFee, beneficiary, minimalLiquidityLock)

	require.Equal(t, newPoolFeePct, params.NewPoolFeePct)
	require.Equal(t, creationFee, params.CreationFee)
	require.Equal(t, beneficiary, params.Beneficiary)
	require.Equal(t, minimalLiquidityLock, params.MinimalLiquidityLock)
}

func TestDefaultParams(t *testing.T) {
	// Test case: check default params values
	params := types.DefaultParams()

	require.Equal(t, types.DefaultNewPoolFeePct, params.NewPoolFeePct)
	require.Equal(t, types.DefaultCreationFee, params.CreationFee)
	require.Equal(t, types.DefaultBeneficiary, params.Beneficiary)
	require.Equal(t, types.DefaultMinimalLiquidityLock, params.MinimalLiquidityLock)
}

func TestParams_Validate_Valid(t *testing.T) {
	// Test case: validate params with valid values
	params := types.NewParams(500, 100000000, "", 1000)
	// Assuming Params has MaxSlippage field, set it to a valid value
	params.MaxSlippage = 100

	err := params.Validate()
	require.NoError(t, err, "Valid params should not return an error")
}

func TestParams_Validate_ValidWithBeneficiary(t *testing.T) {
	// Test case: validate params with valid beneficiary address
	beneficiary := sample.AccAddress()
	params := types.NewParams(500, 100000000, beneficiary, 1000)
	params.MaxSlippage = 100

	err := params.Validate()
	require.NoError(t, err, "Valid params with beneficiary should not return an error")
}

// Negative test cases

func TestParams_Validate_InvalidNewPoolFeePct(t *testing.T) {
	// Test case: validate params with invalid new pool fee pct
	params := types.NewParams(constants.PoolFeeScalingFactor, 100000000, "", 1000)
	params.MaxSlippage = 100

	err := params.Validate()
	require.Error(t, err, "Invalid new pool fee pct should return an error")
	require.Contains(t, err.Error(), "Pool fee too large")
}

func TestParams_Validate_InvalidBeneficiary(t *testing.T) {
	// Test case: validate params with invalid beneficiary address
	params := types.NewParams(500, 100000000, "invalid_address", 1000)
	params.MaxSlippage = 100

	err := params.Validate()
	require.Error(t, err, "Invalid beneficiary address should return an error")
	require.Contains(t, err.Error(), "Beneficiary address is invalid")
}

func TestParams_Validate_InvalidMinimalLiquidityLock(t *testing.T) {
	// Test case: validate params with invalid minimal liquidity lock
	params := types.NewParams(500, 100000000, "", 0)
	params.MaxSlippage = 100

	err := params.Validate()
	require.Error(t, err, "Invalid minimal liquidity lock should return an error")
	require.Contains(t, err.Error(), "MinimalLiquidityLock cannot be zero")
}

func TestParams_Validate_InvalidMaxSlippage(t *testing.T) {
	// Test case: validate params with invalid max slippage
	params := types.NewParams(500, 100000000, "", 1000)
	params.MaxSlippage = 10001

	err := params.Validate()
	require.Error(t, err, "Invalid max slippage should return an error")
	require.Contains(t, err.Error(), "MaxSlippage cannot be greater than 10000")
}

func TestParams_Validate_CreationFee_Zero(t *testing.T) {
	// Test case: validate params with zero creation fee
	params := types.NewParams(500, uint32(0), "", 1000)
	params.MaxSlippage = 100

	err := params.Validate()
	require.NoError(t, err, "Zero creation fee should be valid")
}

func TestParams_Validate_CreationFee_MaxUint32(t *testing.T) {
	// Test case: validate params with maximum uint32 value
	params := types.NewParams(500, uint32(4294967295), "", 1000)
	params.MaxSlippage = 100

	err := params.Validate()
	require.NoError(t, err, "Maximum uint32 creation fee should be valid")
}

func TestParams_Validate_CreationFee_EdgeCases(t *testing.T) {
	// Test case: validate params with various edge case values for creation fee
	testCases := []struct {
		name     string
		amount   uint32
		expected bool
	}{
		{"zero", 0, true},
		{"one", 1, true},
		{"small_value", 100, true},
		{"large_value", 1000000, true},
		{"max_uint32", 4294967295, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.NewParams(500, tc.amount, "", 1000)
			params.MaxSlippage = 100
			err := params.Validate()

			if tc.expected {
				require.NoError(t, err, "Creation fee %d should be valid", tc.amount)
			} else {
				require.Error(t, err, "Creation fee %d should be invalid", tc.amount)
			}
		})
	}
}

func TestParams_Validate_CreationFee_NegativeValue(t *testing.T) {
	// Test case: ensure that if somehow a negative value gets into CreationFee,
	// the validation will fail appropriately
	// Note: Since uint32 cannot hold negative values, this test ensures the validation
	// function handles edge cases correctly and doesn't panic

	// Test with zero value (closest to "negative" for uint32)
	params := types.NewParams(500, uint32(0), "", 1000)
	params.MaxSlippage = 100

	// This should not panic and should pass validation
	require.NotPanics(t, func() {
		err := params.Validate()
		require.NoError(t, err, "Zero creation fee should be valid")
	}, "Validation should not panic with zero creation fee")
}

func TestParams_Validate_CreationFee_ActualNegativeCast(t *testing.T) {
	// Test case: cast a negative value to uint32 and see what happens
	// This tests the actual behavior when a negative value is cast to uint32

	// Cast negative values to uint32
	negativeValues := []int{-1, -100, -1000, -1000000}

	for _, negVal := range negativeValues {
		t.Run(fmt.Sprintf("negative_%d", negVal), func(t *testing.T) {
			// Cast negative value to uint32
			castValue := uint32(negVal)

			// Create params with the cast value
			params := types.NewParams(500, castValue, "", 1000)
			params.MaxSlippage = 100

			// Test that validation doesn't panic
			require.NotPanics(t, func() {
				err := params.Validate()
				// Since uint32 wraps around, negative values become very large positive values
				// The validation should still pass as uint32 values are always valid
				require.NoError(t, err, "Cast negative value %d to uint32 should be valid", negVal)
			}, "Validation should not panic with cast negative value %d", negVal)

			// Log the actual cast value for verification
			t.Logf("Negative value %d cast to uint32 becomes: %d", negVal, castValue)
		})
	}
}

func TestParams_Validate_CreationFee_WhyNoNegativeCheck(t *testing.T) {
	// Test case: demonstrate why the validation function doesn't need to check for negative values
	// This shows the current behavior and explains the design decision

	t.Run("uint32_cannot_be_negative", func(t *testing.T) {
		// Demonstrate that uint32 cannot hold negative values
		// var u uint32 = 0  // This would be unused

		// This would cause a compilation error if we tried to assign a negative value directly
		// u = -1  // This line would not compile

		// But we can cast negative values, which wraps around
		negVal := -1
		castValue := uint32(negVal)
		require.Equal(t, uint32(4294967295), castValue, "Casting -1 to uint32 should wrap around to max value")

		// The validation function doesn't need to check for "negative" values because:
		// 1. uint32 cannot be negative by design
		// 2. Casting negative values results in very large positive values
		// 3. All uint32 values are technically valid for this field

		params := types.NewParams(500, castValue, "", 1000)
		params.MaxSlippage = 100
		err := params.Validate()
		require.NoError(t, err, "Even wrapped-around negative values should be valid")
	})

	t.Run("current_validation_logic", func(t *testing.T) {
		// Show what the current validation function actually does
		params := types.NewParams(500, 100000000, "", 1000)
		params.MaxSlippage = 100

		// The current validateCreationFee function just returns nil
		// because uint32 values can never be negative or exceed MaxUint32
		err := params.Validate()
		require.NoError(t, err, "Current validation should pass for any uint32 value")

		// This is the correct behavior because:
		// - uint32 is the correct type for this field
		// - All uint32 values are valid
		// - No additional validation is needed
	})
}

func TestParams_Validate_CreationFee_Robustness(t *testing.T) {
	// Test case: ensure the validation function is robust and handles edge cases
	// without panicking, even if somehow invalid data gets through

	testCases := []struct {
		name        string
		amount      uint32
		description string
	}{
		{"zero", 0, "Zero value should be handled gracefully"},
		{"one", 1, "Minimum positive value should be handled gracefully"},
		{"max_uint32", 4294967295, "Maximum uint32 value should be handled gracefully"},
		{"large_but_reasonable", 1000000, "Large but reasonable value should be handled gracefully"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.NewParams(500, tc.amount, "", 1000)
			params.MaxSlippage = 100

			// Ensure validation doesn't panic and handles the value appropriately
			require.NotPanics(t, func() {
				err := params.Validate()
				// All uint32 values should be considered valid since they can't be negative
				require.NoError(t, err, tc.description)
			}, "Validation should not panic with creation fee %d", tc.amount)
		})
	}
}

func TestParams_Validate_NewPoolFeePct_EdgeCases(t *testing.T) {
	// Test case: validate params with various edge case values for new pool fee pct
	testCases := []struct {
		name     string
		value    uint32
		expected bool
	}{
		{"zero", 0, true},
		{"small_value", 1, true},
		{"typical_value", 500, true},
		{"just_below_scaling", constants.PoolFeeScalingFactor - 1, true},
		{"equal_scaling", constants.PoolFeeScalingFactor, false},
		{"above_scaling", constants.PoolFeeScalingFactor + 1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.NewParams(tc.value, 100000000, "", 1000)
			params.MaxSlippage = 100
			err := params.Validate()

			if tc.expected {
				require.NoError(t, err, "New pool fee pct %d should be valid", tc.value)
			} else {
				require.Error(t, err, "New pool fee pct %d should be invalid", tc.value)
				require.Contains(t, err.Error(), "Pool fee too large")
			}
		})
	}
}

func TestParams_Validate_MinimalLiquidityLock_EdgeCases(t *testing.T) {
	// Test case: validate params with various edge case values for minimal liquidity lock
	testCases := []struct {
		name     string
		value    uint32
		expected bool
	}{
		{"zero", 0, false},
		{"one", 1, true},
		{"typical_value", 1000, true},
		{"large_value", 1000000, true},
		{"max_uint32", 4294967295, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.NewParams(500, 100000000, "", tc.value)
			params.MaxSlippage = 100
			err := params.Validate()

			if tc.expected {
				require.NoError(t, err, "Minimal liquidity lock %d should be valid", tc.value)
			} else {
				require.Error(t, err, "Minimal liquidity lock %d should be invalid", tc.value)
				require.Contains(t, err.Error(), "MinimalLiquidityLock cannot be zero")
			}
		})
	}
}

func TestParams_Validate_MaxSlippage_EdgeCases(t *testing.T) {
	// Test case: validate params with various edge case values for max slippage
	testCases := []struct {
		name     string
		value    uint32
		expected bool
	}{
		{"zero", 0, true},
		{"small_value", 1, true},
		{"typical_value", 100, true},
		{"max_allowed", 10000, true},
		{"above_max", 10001, false},
		{"large_value", 4294967295, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.NewParams(500, 100000000, "", 1000)
			params.MaxSlippage = tc.value
			err := params.Validate()

			if tc.expected {
				require.NoError(t, err, "Max slippage %d should be valid", tc.value)
			} else {
				require.Error(t, err, "Max slippage %d should be invalid", tc.value)
				require.Contains(t, err.Error(), "MaxSlippage cannot be greater than 10000")
			}
		})
	}
}

func TestParams_Validate_MultipleValidationErrors(t *testing.T) {
	// Test case: validate params with multiple validation errors
	params := types.NewParams(constants.PoolFeeScalingFactor, 100000000, "invalid_address", 0)
	params.MaxSlippage = 10001

	err := params.Validate()
	require.Error(t, err, "Params with multiple validation errors should return an error")
	// The first validation error (invalid new pool fee pct) should be returned
	require.Contains(t, err.Error(), "Pool fee too large")
}

func TestConvertFromBasisPointsToDecimal(t *testing.T) {
	// Test case: convert basis points to decimal
	testCases := []struct {
		name        string
		basisPoints uint32
		expected    sdkmath.LegacyDec
	}{
		{"zero", 0, sdkmath.LegacyZeroDec()},
		{"full", 10000, sdkmath.LegacyOneDec()},
		{"half", 5000, sdkmath.LegacyMustNewDecFromStr("0.5")},
		{"fraction", 1234, sdkmath.LegacyMustNewDecFromStr("0.1234")},
		{"max_uint32", 4294967295, sdkmath.LegacyMustNewDecFromStr("429496.7295")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.ConvertFromBasisPointsToDecimal(tc.basisPoints)
			require.True(t, tc.expected.Equal(result), "Expected %s, got %s", tc.expected, result)
		})
	}
}

func TestConvertFromDecimalToBasisPoints(t *testing.T) {
	// Test case: convert decimal to basis points
	testCases := []struct {
		name     string
		decimal  sdkmath.LegacyDec
		expected uint32
	}{
		{"zero", sdkmath.LegacyZeroDec(), 0},
		{"full", sdkmath.LegacyOneDec(), 10000},
		{"half", sdkmath.LegacyMustNewDecFromStr("0.5"), 5000},
		{"fraction", sdkmath.LegacyMustNewDecFromStr("0.1234"), 1234},
		{"truncate", sdkmath.LegacyMustNewDecFromStr("0.12345"), 1234},
		{"large", sdkmath.LegacyMustNewDecFromStr("429496.7295"), 4294967295},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.ConvertFromDecimalToBasisPoints(tc.decimal)
			require.Equal(t, tc.expected, result)
		})
	}
}
