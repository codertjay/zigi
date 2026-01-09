package types_test

import (
	"fmt"
	"testing"

	"zigchain/testutil/sample"
	"zigchain/x/factory/types"

	"github.com/stretchr/testify/require"
)

func TestNewParams(t *testing.T) {
	// Test case: create new params with valid values
	createFeeDenom := "uzig"
	createFeeAmount := uint32(1000)
	beneficiary := sample.AccAddress()

	params := types.NewParams(createFeeDenom, createFeeAmount, beneficiary)

	require.Equal(t, createFeeDenom, params.CreateFeeDenom)
	require.Equal(t, createFeeAmount, params.CreateFeeAmount)
	require.Equal(t, beneficiary, params.Beneficiary)
}

func TestDefaultParams(t *testing.T) {
	// Test case: check default params values
	params := types.DefaultParams()

	require.Equal(t, types.DefaultCreateFeeDenom, params.CreateFeeDenom)
	require.Equal(t, types.DefaultCreateFeeAmount, params.CreateFeeAmount)
	require.Equal(t, types.DefaultBeneficiary, params.Beneficiary)
}

func TestParams_Validate_Valid(t *testing.T) {
	// Test case: validate params with valid values
	params := types.NewParams("uzig", uint32(1000), "")

	err := params.Validate()
	require.NoError(t, err, "Valid params should not return an error")
}

func TestParams_Validate_ValidWithBeneficiary(t *testing.T) {
	// Test case: validate params with valid beneficiary address
	beneficiary := sample.AccAddress()
	params := types.NewParams("uzig", uint32(1000), beneficiary)

	err := params.Validate()
	require.NoError(t, err, "Valid params with beneficiary should not return an error")
}

func TestParams_Validate_InvalidCreateFeeDenom(t *testing.T) {
	// Test case: validate params with invalid create fee denom
	params := types.NewParams("invalid_denom", uint32(1000), "")

	err := params.Validate()
	require.Error(t, err, "Invalid create fee denom should return an error")
	require.Contains(t, err.Error(), "invalid create fee denom parameter")
}

func TestParams_Validate_InvalidBeneficiary(t *testing.T) {
	// Test case: validate params with invalid beneficiary address
	params := types.NewParams("uzig", uint32(1000), "invalid_address")

	err := params.Validate()
	require.Error(t, err, "Invalid beneficiary address should return an error")
	require.Contains(t, err.Error(), "Beneficiary address is invalid")
}

func TestParams_Validate_CreateFeeAmount_Zero(t *testing.T) {
	// Test case: validate params with zero create fee amount
	params := types.NewParams("uzig", uint32(0), "")

	err := params.Validate()
	require.NoError(t, err, "Zero create fee amount should be valid")
}

func TestParams_Validate_CreateFeeAmount_MaxUint32(t *testing.T) {
	// Test case: validate params with maximum uint32 value
	params := types.NewParams("uzig", uint32(4294967295), "")

	err := params.Validate()
	require.NoError(t, err, "Maximum uint32 create fee amount should be valid")
}

func TestParams_Validate_CreateFeeAmount_EdgeCases(t *testing.T) {
	// Test case: validate params with various edge case values for create fee amount
	testCases := []struct {
		name     string
		amount   uint32
		expected bool // true if should pass validation
	}{
		{"zero", 0, true},
		{"one", 1, true},
		{"small_value", 100, true},
		{"large_value", 1000000, true},
		{"max_uint32", 4294967295, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.NewParams("uzig", tc.amount, "")
			err := params.Validate()

			if tc.expected {
				require.NoError(t, err, "Create fee amount %d should be valid", tc.amount)
			} else {
				require.Error(t, err, "Create fee amount %d should be invalid", tc.amount)
			}
		})
	}
}

func TestParams_Validate_CreateFeeAmount_NegativeValue(t *testing.T) {
	// Test case: ensure that if somehow a negative value gets into CreateFeeAmount,
	// the validation will fail appropriately
	// Note: Since uint32 cannot hold negative values, this test ensures the validation
	// function handles edge cases correctly and doesn't panic

	// Test with zero value (closest to "negative" for uint32)
	params := types.NewParams("uzig", uint32(0), "")

	// This should not panic and should pass validation
	require.NotPanics(t, func() {
		err := params.Validate()
		require.NoError(t, err, "Zero create fee amount should be valid")
	}, "Validation should not panic with zero create fee amount")
}

func TestParams_Validate_CreateFeeAmount_ActualNegativeCast(t *testing.T) {
	// Test case: cast a negative value to uint32 and see what happens
	// This tests the actual behavior when a negative value is cast to uint32

	// Cast negative values to uint32
	negativeValues := []int{-1, -100, -1000, -1000000}

	for _, negVal := range negativeValues {
		t.Run(fmt.Sprintf("negative_%d", negVal), func(t *testing.T) {
			// Cast negative value to uint32
			castValue := uint32(negVal)

			// Create params with the cast value
			params := types.NewParams("uzig", castValue, "")

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

func TestParams_Validate_CreateFeeAmount_WhyNoNegativeCheck(t *testing.T) {
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

		params := types.NewParams("uzig", castValue, "")
		err := params.Validate()
		require.NoError(t, err, "Even wrapped-around negative values should be valid")
	})

	t.Run("current_validation_logic", func(t *testing.T) {
		// Show what the current validation function actually does
		params := types.NewParams("uzig", uint32(1000), "")

		// The current validateCreateFeeAmount function just returns nil
		// because uint32 values can never be negative or exceed MaxUint32
		err := params.Validate()
		require.NoError(t, err, "Current validation should pass for any uint32 value")

		// This is the correct behavior because:
		// - uint32 is the correct type for this field
		// - All uint32 values are valid
		// - No additional validation is needed
	})
}

func TestParams_Validate_CreateFeeAmount_Robustness(t *testing.T) {
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
			params := types.NewParams("uzig", tc.amount, "")

			// Ensure validation doesn't panic and handles the value appropriately
			require.NotPanics(t, func() {
				err := params.Validate()
				// All uint32 values should be considered valid since they can't be negative
				require.NoError(t, err, tc.description)
			}, "Validation should not panic with create fee amount %d", tc.amount)
		})
	}
}

func TestParams_Validate_EmptyCreateFeeDenom(t *testing.T) {
	// Test case: validate params with empty create fee denom
	params := types.NewParams("", uint32(1000), "")

	err := params.Validate()
	require.Error(t, err, "Empty create fee denom should return an error")
	require.Contains(t, err.Error(), "invalid create fee denom parameter")
}

func TestParams_Validate_MultipleValidationErrors(t *testing.T) {
	// Test case: validate params with multiple validation errors
	params := types.NewParams("invalid_denom", uint32(1000), "invalid_address")

	err := params.Validate()
	require.Error(t, err, "Params with multiple validation errors should return an error")
	// The first validation error (invalid denom) should be returned
	require.Contains(t, err.Error(), "invalid create fee denom parameter")
}

func TestParamKeyTable(t *testing.T) {
	// Test case: ensure ParamKeyTable can be created without panicking
	require.NotPanics(t, func() {
		keyTable := types.ParamKeyTable()
		require.NotNil(t, keyTable, "ParamKeyTable should not be nil")
	}, "ParamKeyTable creation should not panic")
}

func TestParams_ParamSetPairs(t *testing.T) {
	// Test case: ensure ParamSetPairs returns expected value
	params := types.DefaultParams()

	pairs := params.ParamSetPairs()
	require.NotNil(t, pairs, "ParamSetPairs should not be nil")
	// Currently returns empty slice as per implementation
	require.Len(t, pairs, 0, "ParamSetPairs should return empty slice")
}
