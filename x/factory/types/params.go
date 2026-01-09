package types

import (
	"zigchain/zutils/constants"
	"zigchain/zutils/validators"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// DefaultCreateFeeDenom is the denomination of the fee to create a new factory.
var DefaultCreateFeeDenom = constants.BondDenom

// DefaultCreateFeeAmount is the amount of the fee to create a new factory.
var DefaultCreateFeeAmount = uint32(1000)

// DefaultBeneficiary is an account that can pull funds from fees captured in the factory
var DefaultBeneficiary = ""

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	createFeeDenom string,
	createFeeAmount uint32,
	beneficiary string,
) Params {
	return Params{
		CreateFeeDenom:  createFeeDenom,
		CreateFeeAmount: createFeeAmount,
		Beneficiary:     beneficiary,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultCreateFeeDenom,
		DefaultCreateFeeAmount,
		DefaultBeneficiary,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validates the set of params
func (p Params) Validate() error {

	if err := validateCreateFeeDenom(p.CreateFeeDenom); err != nil {
		return err
	}

	if err := validateCreateFeeAmount(p.CreateFeeAmount); err != nil {
		return err
	}

	if p.Beneficiary != "" {
		if err := validators.AddressCheck("beneficiary", p.Beneficiary); err != nil {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidAddress,
				"Beneficiary address is invalid: %s : %s",
				p.Beneficiary,
				err,
			)
		}
	}

	return nil
}

func validateCreateFeeDenom(denom string) error {

	// check if existing param denom is the same as the new denom
	// this is probably not factory denom but uzig like coin
	// so we don't run deconstruction check on it
	if validators.CheckDenomString(denom) != nil {
		return errorsmod.Wrapf(
			ErrorInvalidParamValue,
			"invalid create fee denom parameter: %s",
			denom,
		)
	}

	return nil
}

// validateCreateFeeAmount validates the CreationFeeAmount parameter.
func validateCreateFeeAmount(createFeeAmount uint32) error {
	_ = createFeeAmount
	// No validation needed for uint32 - it can never be negative or exceed MaxUint32
	return nil
}
