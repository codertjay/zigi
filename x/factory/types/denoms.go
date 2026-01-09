package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zigchain/zutils/validators"
)

const (
	ModuleDenomPrefix = "coin"
	// See the TokenFactory readme for a derivation of these.
	// TL;DR, MaxSubdenomLength + MaxHrpLength = 60 comes from SDK max denom length = 128
	// and the structure of tokenfactory denoms.
	MaxSubdenomLength = 44
	MaxHrpLength      = 16
	MaxCreatorLength  = 59 + MaxHrpLength
	// FactoryDenomDelimiterChar is the character used to separate parts of a factory denom
	FactoryDenomDelimiterChar = "."
)

// GetTokenDenom constructs a denom string for tokens created by tokenfactory
// based on an input creator address and a subdenom
// The denom constructed is coin.{creator}.{subdenom}
func GetTokenDenom(creator, subdenom string) (string, error) {

	if err := validators.CheckSubDenomString(subdenom); err != nil {
		return "", err
	}

	if len(creator) > MaxCreatorLength {
		return "", ErrCreatorTooLong
	}

	// Validate creator content
	if strings.ContainsAny(creator, "/\\.%&?") {
		return "", ErrInvalidSignerAccount // Extend validation for special characters
	}
	// above might be overkill, since creator is address and ValidateDenom will catch this anyway
	// but better to be safe than sorry
	//if strings.Contains(creator, "/") {
	//	return "", ErrInvalidSignerAccount
	//}
	denom := strings.Join([]string{ModuleDenomPrefix, creator, subdenom}, FactoryDenomDelimiterChar)
	return denom, sdk.ValidateDenom(denom)
}

// DeconstructDenom takes a token denom string and verifies that it is a valid
// denom of the tokenfactory module, and is of the form `coin.{creator}.{subdenom}`
// If valid, it returns the creator address and subdenom
func DeconstructDenom(denom string) (creator string, subdenom string, err error) {
	err = sdk.ValidateDenom(denom)
	if err != nil {
		return "", "", err
	}

	strParts := strings.Split(denom, FactoryDenomDelimiterChar)
	if len(strParts) < 3 {

		return "", "", errorsmod.Wrapf(ErrInvalidDenom, "not enough parts of denom %s", denom)
	}

	if len(strParts) > 3 {
		return "", "", errorsmod.Wrapf(ErrInvalidDenom, "too many parts of denom %s", denom)
	}

	if strParts[0] != ModuleDenomPrefix {
		return "", "", errorsmod.Wrapf(ErrInvalidDenom, "denom prefix is incorrect. Is: %s.  Should be: %s", strParts[0], ModuleDenomPrefix)
	}

	creator = strParts[1]
	err = validators.AddressCheck("Creator", creator)
	if err != nil {
		return "", "", errorsmod.Wrapf(ErrInvalidDenom, "Invalid creator address (%s)", err)
	}

	// Handle the case where a denom has a slash in its subdenom. For example,
	// when we did the split, we'd turn coin.accaddr.der.test into ["coin", "accaddr", "der", "test"]
	// So we have to join [2:] with a "/" as the delimiter to get back the correct subdenom which should be "der/test"
	// This is not used as it is limited to 3 parts above, but if the limit is removed, it will fit in
	subdenom = strings.Join(strParts[2:], FactoryDenomDelimiterChar)

	if subdenom == "" {
		return "", "", errorsmod.Wrapf(ErrInvalidDenom, "subdenom is empty")
	}

	if err = validators.CheckSubDenomString(subdenom); err != nil {
		return "", "", err
	}

	return creator, subdenom, nil
}
