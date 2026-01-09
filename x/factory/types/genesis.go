package types

import (
	"fmt"
	"zigchain/zutils/validators"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		DenomList:     []Denom{},
		DenomAuthList: []DenomAuth{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in denom
	denomIndexMap := make(map[string]struct{})

	for _, denom := range gs.DenomList {
		index := string(DenomKey(denom.Denom))
		if _, ok := denomIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for denom")
		}
		denomIndexMap[index] = struct{}{}

		if err := denom.Validate(); err != nil {
			return err
		}
	}

	// Check for duplicated index in denomAuth
	denomAuthIndexMap := make(map[string]struct{})

	for _, denomAuth := range gs.DenomAuthList {
		index := string(DenomAuthKey(denomAuth.Denom))
		if _, ok := denomAuthIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for denomAuth")
		}
		denomAuthIndexMap[index] = struct{}{}

		if err := denomAuth.Validate(); err != nil {
			return err
		}
	}

	// Validate that every denom has a corresponding DenomAuth
	for _, denom := range gs.DenomList {
		if _, exists := denomAuthIndexMap[string(DenomAuthKey(denom.Denom))]; !exists {
			return fmt.Errorf("missing DenomAuth for denom '%s'", denom.Denom)
		}
	}

	// Validate that every DenomAuth corresponds to a Denom
	for _, denomAuth := range gs.DenomAuthList {
		if _, exists := denomIndexMap[string(DenomKey(denomAuth.Denom))]; !exists {
			return fmt.Errorf("missing Denom for DenomAuth '%s'", denomAuth.Denom)
		}
	}

	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}

// ValidateDenom validates a denom

func (d Denom) Validate() error {

	if err := validators.AddressCheck("creator", d.Creator); err != nil {
		return errorsmod.Wrapf(
			ErrInvalidDenom,
			"Denom creator invalid (%s) : %s",
			d.Denom,
			err,
		)
	}

	creator, _, err := DeconstructDenom(d.Denom)
	if err != nil {
		return errorsmod.Wrapf(
			ErrInvalidDenom,
			"(%s) : %s",
			d.Denom,
			err,
		)
	}

	if creator != d.Creator {
		return errorsmod.Wrapf(
			ErrInvalidDenom,
			"Denom creator (%s) does not match denom creator (%s)",
			creator,
			d.Creator,
		)

	}

	if d.MintingCap.IsNil() {
		return errorsmod.Wrapf(
			ErrInvalidDenom,
			"Minting Cap %s is nil",
			d.MintingCap,
		)
	}

	// MintingCap must be greater than zero
	if !d.MintingCap.GT(sdkmath.ZeroUint()) {
		return errorsmod.Wrapf(
			ErrInvalidDenom,
			"Minting Cap %s must be greater than 0",
			d.MintingCap,
		)
	}

	if d.Minted.GT(d.MintingCap) {
		return errorsmod.Wrapf(
			ErrInvalidDenom,
			"Denom minted (%s) exceeds minting cap (%s)",
			d.Minted,
			d.MintingCap,
		)
	}

	return nil

}

// ValidateDenomAuth validates a denomAuth
func (d DenomAuth) Validate() error {

	_, _, err := DeconstructDenom(d.Denom)
	if err != nil {
		return errorsmod.Wrapf(
			ErrInvalidDenom,
			"(%s) : %s",
			d.Denom,
			err,
		)
	}

	if d.BankAdmin != "" {
		if err = validators.AddressCheck("BankAdmin", d.BankAdmin); err != nil {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidAddress,
				"DenomAuth BankAdmin invalid (%s) : %s",
				d.Denom,
				err,
			)
		}
	}

	if d.MetadataAdmin != "" {
		if err = validators.AddressCheck("MetadataAdmin", d.MetadataAdmin); err != nil {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidAddress,
				"DenomAuth MetadataAdmin invalid (%s) : %s",
				d.Denom,
				err,
			)
		}
	}

	return nil
}
