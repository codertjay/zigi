package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"zigchain/testutil/sample"
	"zigchain/zutils/constants"
)

var msgUpdateParamsSample = MsgUpdateParams{
	// use sample.AccAddress() to get a valid address
	Authority: sample.AccAddress(),
	Params: Params{
		NewPoolFeePct:        uint32(600),
		CreationFee:          uint32(200000000),
		MinimalLiquidityLock: 1000,
	},
}

// Positive test cases

func TestMsgUpdateParams_ValidateBasic_Positive(t *testing.T) {
	// Test case: update parameters with valid input data

	// make a copy of sample message
	msg := msgUpdateParamsSample

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is nil
	require.NoError(t, err)
}

// Negative test cases

func TestMsgUpdateParams_ValidateBasic_InvalidAuthority(t *testing.T) {
	// Test case: invalid authority address

	// make a copy of sample message
	msg := MsgUpdateParams{
		Authority: "invalid_address",
		Params: Params{
			NewPoolFeePct: uint32(600),
			CreationFee:   uint32(200000000),
		},
	}

	// validate the message
	err := msg.ValidateBasic()

	// assert that the error is not nil
	require.Error(t, err)

	// assert that the error message matches the expected error message
	require.Contains(t, err.Error(), "invalid authority address")
}

func TestMsgUpdateParams_ValidateBasic_InvalidParams(t *testing.T) {
	// Test case: valid authority but invalid Params (e.g., NewPoolFeePct too large)

	msg := MsgUpdateParams{
		Authority: sample.AccAddress(),
		Params: Params{
			NewPoolFeePct:        constants.PoolFeeScalingFactor, // Invalid: too large
			CreationFee:          uint32(200000000),
			MinimalLiquidityLock: 1000,
		},
	}

	err := msg.ValidateBasic()
	require.Error(t, err)
	require.Contains(t, err.Error(), "Pool fee too large")
	require.Contains(t, err.Error(), "has to be less than scaling factor")
}
