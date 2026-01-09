package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

type MsgToTestType interface {
	ValidateBasic() error
}

func CheckMessageStringFieldErrors(
	t *testing.T,
	msgToTest MsgToTestType,
	fieldPtr *string,
	testsPtr *[]struct {
		TestName   string
		FieldValue string
		ErrorText  string
	},
	expectedErr error,
) {

	for _, test := range *testsPtr {

		t.Run(test.TestName, func(t *testing.T) {

			// change Creator address to the one from the test
			*fieldPtr = test.FieldValue

			// Run validations
			errorFromValidator := msgToTest.ValidateBasic()
			// Make sure we got error
			require.Error(t, errorFromValidator)
			// assert that the error is of type ErrInvalidDenom
			require.ErrorIs(t, errorFromValidator, expectedErr)
			// assert that the error message matches the expected error message
			require.EqualError(t, errorFromValidator, test.ErrorText+": "+expectedErr.Error())

		})
	}
}

func CheckMessageCoinFieldErrors(
	t *testing.T,
	msgToTest MsgToTestType,
	fieldPtr *sdk.Coin,
	testsPtr *[]struct {
		TestName   string
		FieldValue sdk.Coin
		ErrorText  string
	},
	expectedErr error,
) {

	for _, test := range *testsPtr {

		t.Run(test.TestName, func(t *testing.T) {

			// change Creator address to the one from the test
			*fieldPtr = test.FieldValue

			// Run validations
			errorFromValidator := msgToTest.ValidateBasic()
			// Make sure we got error
			require.Error(t, errorFromValidator)
			// assert that the error is of type ErrInvalidDenom
			require.ErrorIs(t, errorFromValidator, expectedErr)
			// assert that the error message matches the expected error message
			require.EqualError(t, errorFromValidator, test.ErrorText+": "+expectedErr.Error())

		})
	}
}

func CheckMessageInt64FieldErrors(
	t *testing.T,
	msgToTest MsgToTestType,
	fieldPtr *int64,
	testsPtr *[]struct {
		TestName   string
		FieldValue int64
		ErrorText  string
	},
	expectedErr error,
) {

	for _, test := range *testsPtr {

		t.Run(test.TestName, func(t *testing.T) {

			// change Creator address to the one from the test
			*fieldPtr = test.FieldValue

			// Run validations
			errorFromValidator := msgToTest.ValidateBasic()
			// Make sure we got error
			require.Error(t, errorFromValidator)
			// assert that the error is of type ErrInvalidDenom
			require.ErrorIs(t, errorFromValidator, expectedErr)
			// assert that the error message matches the expected error message
			require.EqualError(t, errorFromValidator, test.ErrorText+": "+expectedErr.Error())

		})
	}
}

func CheckMessageInt32FieldErrors(
	t *testing.T,
	msgToTest MsgToTestType,
	fieldPtr *int32,
	testsPtr *[]struct {
		TestName   string
		FieldValue int32
		ErrorText  string
	},
	expectedErr error,
) {

	for _, test := range *testsPtr {

		t.Run(test.TestName, func(t *testing.T) {

			// change Creator address to the one from the test
			*fieldPtr = test.FieldValue

			// Run validations
			errorFromValidator := msgToTest.ValidateBasic()
			// Make sure we got error
			require.Error(t, errorFromValidator)
			// assert that the error is of type ErrInvalidDenom
			require.ErrorIs(t, errorFromValidator, expectedErr)
			// assert that the error message matches the expected error message
			require.EqualError(t, errorFromValidator, test.ErrorText+": "+expectedErr.Error())

		})
	}
}
