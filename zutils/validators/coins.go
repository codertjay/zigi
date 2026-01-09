package validators

import (
	"fmt"
	"regexp"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/constants"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	poolIDRegexString = "zp[0-9]+"
	regexPoolID       *regexp.Regexp
	DenomRegexString  = `[a-zA-Z][a-zA-Z0-9./-]+`
)

func init() {
	sdk.SetCoinDenomRegex(func() string {
		return DenomRegexString
	})
	// compile poolIDRegex
	regexPoolID = regexp.MustCompile(fmt.Sprintf(`^%s$`, poolIDRegexString))
}

func CoinCheck(coin sdk.Coin, zeroOK bool) error {

	err := CheckCoinAmount(coin, zeroOK)
	if err != nil {
		return err
	}

	err = CheckCoinDenom(coin)
	if err != nil {
		return err
	}

	return nil
}

// CheckCoinAmount validates the amount of a given sdk.Coin based on specified conditions.
//
// Parameters:
// - coin: The `sdk.Coin` instance to validate. Represents the token denomination and its amount.
// - zeroOK: A boolean indicating if a zero amount is acceptable. If false, the amount must be strictly positive.
//
// Returns:
// - error: An error if the coin amount is invalid; nil otherwise.
//
// Validation Criteria:
// - The amount in the coin must not be nil.
// - The amount must not be negative.
// - If zeroOK is false, the amount must be strictly positive.
func CheckCoinAmount(coin sdk.Coin, zeroOK bool) error {

	// Ensure the coin amount is not nil.
	if coin.Amount.IsNil() {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin amount: cannot be nil (%s)",
			coin.String(),
		)
	}

	// Ensure the coin amount is not negative.
	if coin.IsNegative() {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin amount: %s cannot be negative (%s)",
			coin.Amount.String(),
			coin.String(),
		)
	}

	// If zero amounts are not allowed, ensure the coin amount is positive.
	if !zeroOK && !coin.IsPositive() {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin amount: %s has to be positive (%s)",
			coin.Amount.String(),
			coin.String(),
		)
	}

	return nil
}

// CheckDenomString validates the denomination string of a coin based on specific criteria.
//
// Parameters:
// - denom: The denomination string to validate.
//
// Returns:
// - error: An error if the denomination is invalid; nil otherwise.
//
// Validation Criteria:
// - The denomination must not be empty.
// - The denomination length must be within the allowed range (MinSubDenomLength to MaxDenomLength).
// - The denomination must conform to the required regex pattern.
//
// Notes:
// - `constants.MinSubDenomLength` specifies the minimum length for a denomination.
// - `constants.MaxDenomLength` specifies the maximum length for a denomination.
// - The regex check (`sdk.ValidateDenom`) ensures that the denomination adheres to the allowed format.
func CheckDenomString(denom string) error {

	if denom == "" {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin: denomination '%s' cannot be empty (e.g., 10uzig)",
			denom,
		)
	}

	if len(denom) < constants.MinSubDenomLength {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin: '%s' denom name is too short, minimum %d characters e.g. 10uzig",
			denom,
			constants.MinSubDenomLength,
		)

	}

	if len(denom) > constants.MaxDenomLength {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin: '%s' denom name is too long (%d), maximum %d characters e.g. uzig",
			denom,
			len(denom),
			constants.MaxDenomLength,
		)

	}

	// regex check (do it last as it is most expensive)
	if err := sdk.ValidateDenom(denom); err != nil {

		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin: '%s' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
			denom,
		)
	}

	return nil

}

func CheckCoinDenom(coin sdk.Coin) error {

	// Check if the denomination is empty.
	if coin.Denom == "" {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin: denomination %s cannot be empty (e.g., 10uzig)",
			coin.String(),
		)
	}

	// Check if the denomination is too short.
	if len(coin.Denom) < constants.MinSubDenomLength {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin: '%s' denom name is too short for %s, minimum %d characters e.g. 10uzig",
			coin.Denom,
			coin.String(),
			constants.MinSubDenomLength,
		)

	}

	// Check if the denomination is too long.
	if len(coin.Denom) > constants.MaxDenomLength {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin: '%s' denom name is too long (%d) for %s, maximum %d characters e.g. uzig",
			coin.Denom,
			len(coin.Denom),
			coin.String(),
			constants.MaxDenomLength,
		)

	}

	// regex check (do it last as it is most expensive)
	if err := sdk.ValidateDenom(coin.Denom); err != nil {

		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin: '%s' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. 10uzig",
			coin.String(),
		)
	}

	return nil
}

func CheckSubDenomString(denom string) error {
	if denom == "" {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidCoins,
			"Invalid subdenom name: denom name is empty e.g. uzig",
		)
	}

	// first size checks for performance
	if len(denom) < constants.MinSubDenomLength {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin: '%s' denom name is too short, minimum %d characters e.g. uzig",
			denom,
			constants.MinSubDenomLength,
		)

	}

	if len(denom) > constants.MaxSubDenomLength {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin: '%s' denom name is too long (%d), maximum %d characters e.g. uzig",
			denom,
			len(denom),
			constants.MaxSubDenomLength,
		)

	}

	// Make sure the first character is an ascii uppercase or lowercase letter
	firstChar := denom[0]
	if firstChar < 'a' || firstChar > 'z' {
		// if !(firstChar >= 'a' && firstChar <= 'z') {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin: '%s' denom name has to start with a lowercase letter e.g. uzig",
			denom,
		)
	}

	// make sure that the rest of the name is alphanumeric
	for i := 1; i < len(denom); i++ {
		if (denom[i] < 'a' || denom[i] > 'z') && (denom[i] < '0' || denom[i] > '9') {
			// if !((denom[i] >= 'a' && denom[i] <= 'z') || (denom[i] >= '0' && denom[i] <= '9')) {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidCoins,
				"invalid coin: '%s' only lowercase letters (a-z) and numbers (0-9) are allowed e.g. uzig123",
				denom,
			)
		}
	}

	// regex check (do it last as it is most expensive)
	if err := sdk.ValidateDenom(denom); err != nil {

		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"invalid coin: '%s' only letters (a-z, A-Z), numbers (0-9), dots (.) and forward slashes (/) are allowed e.g. uzig",
			denom,
		)
	}

	return nil
}

func CheckPoolId(poolId string) error {
	if poolId == "" {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidCoins,
			"Invalid pool id: pool id is empty",
		)
	}

	if len(poolId) < constants.MinSubDenomLength {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"Invalid pool id: '%s' pool id is too short, minimum %d characters",
			poolId,
			constants.MinSubDenomLength,
		)

	}

	if len(poolId) > constants.MaxSubDenomLength {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"Invalid pool id: '%s' pool id is too long (%d), maximum %d characters",
			poolId,
			len(poolId),
			constants.MaxSubDenomLength,
		)

	}

	// regex check that pool id start with constants.PoolPrefix and followed by numbers
	if poolId[:2] != constants.PoolPrefix {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"Invalid pool id: '%s', pool id has to start with '%s' followed by numbers e.g. %s123",
			poolId,
			constants.PoolPrefix,
			constants.PoolPrefix,
		)
	}

	if !regexPoolID.MatchString(poolId) {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"Invalid pool id: '%s', pool id has to start with '%s' followed by numbers e.g. %s123",
			poolId,
			constants.PoolPrefix,
			constants.PoolPrefix,
		)
	}
	return nil
}
