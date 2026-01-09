package types

import (
	"zigchain/zutils/validators"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/zutils/constants"
)

// DefaultCreationFee represents the CreationFee default value.
//var DefaultCreationFee = uint32(100 * math.Pow(10, constants.BondDenomDecimals))

const (
	// DefaultNewPoolFeePct represents the NewPoolFeePct default value.
	DefaultNewPoolFeePct uint32 = 500

	// DefaultCreationFee Deterministic computation using integer arithmetic
	// ZIG: 100 * 1_000_000 = 100_000_000 uzig (100 ZIG, ZIG to an uzig conversion factor is 1_000_000)
	DefaultCreationFee uint32 = 100 * 1_000_000

	// DefaultBeneficiary represents the Beneficiary default value.
	DefaultBeneficiary = ""

	// DefaultMaxSlippage represents the MaxSlippage default value.
	DefaultMaxSlippage uint32 = 100

	// MaximumMaxSlippage represents the MaximumMaxSlippage default value.
	MaximumMaxSlippage uint32 = 10000

	// DefaultMinimalLiquidityLock represents the default value for minimal liquidity lock
	// This is set to 1000 units to prevent share inflation attacks
	DefaultMinimalLiquidityLock uint32 = 1000

	// MinSwapFee is the minimum fee in token's smallest unit for swaps
	MinSwapFee = 1

	// BasisPoints represents 10000, which is equivalent to 100%
	BasisPoints = 10000
)

// NewParams creates a new Params instance.
func NewParams(
	newPoolFeePct uint32,
	creationFee uint32,
	beneficiary string,
	minimalLiquidityLock uint32,
) Params {
	return Params{
		NewPoolFeePct:        newPoolFeePct,
		CreationFee:          creationFee,
		Beneficiary:          beneficiary,
		MinimalLiquidityLock: minimalLiquidityLock,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultNewPoolFeePct,
		DefaultCreationFee,
		DefaultBeneficiary,
		DefaultMinimalLiquidityLock,
	)
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if err := validateNewPoolFeePct(p.NewPoolFeePct); err != nil {
		return err
	}

	if err := validateCreationFee(p.CreationFee); err != nil {
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

	if err := validateMinimalLiquidityLock(p.MinimalLiquidityLock); err != nil {
		return err
	}

	if err := validateMaxSlippage(p.MaxSlippage); err != nil {
		return err
	}

	return nil
}

// validateNewPoolFeePct validates the NewPoolFeePct parameter.
func validateNewPoolFeePct(poolFee uint32) error {

	if poolFee >= constants.PoolFeeScalingFactor {
		// Handle the overflow case
		return errorsmod.Wrapf(
			ErrorInvalidAmount,
			"Pool fee too large: %d, has to be less than scaling factor: %d",
			poolFee,
			constants.PoolFeeScalingFactor,
		)
	}

	return nil
}

// validateMaxSlippage validates the MaxSlippage parameter.
func validateMaxSlippage(maxSlippage uint32) error {
	// max slippage is 100% = 10000 basis points
	if maxSlippage > MaximumMaxSlippage {
		return errorsmod.Wrapf(ErrorInvalidAmount, "MaxSlippage cannot be greater than 10000: %d", maxSlippage)
	}
	return nil
}

// validateCreationFee validates the CreationFee parameter.
func validateCreationFee(creationFee uint32) error {
	_ = creationFee
	// No validation needed for uint32 - it can never be negative or exceed MaxUint32
	return nil
}

// validateMinimalLiquidityLock validates the MinimalLiquidityLock parameter.
func validateMinimalLiquidityLock(minimalLiquidityLock uint32) error {
	if minimalLiquidityLock == 0 {
		return errorsmod.Wrapf(
			ErrorInvalidAmount,
			"MinimalLiquidityLock cannot be zero",
		)
	}

	return nil
}

func ConvertFromBasisPointsToDecimal(basisPoints uint32) math.LegacyDec {
	return math.NewInt(int64(basisPoints)).ToLegacyDec().QuoInt64(BasisPoints)
}

func ConvertFromDecimalToBasisPoints(decimal math.LegacyDec) uint32 {
	// #nosec G115 -- integer overflow conversion is safe here as the input is validated
	return uint32(decimal.MulInt64(BasisPoints).TruncateInt64())
}
