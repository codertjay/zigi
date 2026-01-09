package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zigchain/x/factory/types"
)

func (k Keeper) Auth(ctx context.Context, denom string, adminType string, signer string) error {
	// Check if the value exists
	denomAuth, isFound := k.GetDenomAuth(
		ctx,
		denom,
	)
	if !isFound {
		return errorsmod.Wrapf(
			types.ErrDenomAuthNotFound,
			"Auth: denom (%s)",
			denom,
		)
	}

	switch adminType {
	case "bank":

		// Check if bank admin is disabled
		if denomAuth.BankAdmin == "" {
			return errorsmod.Wrapf(
				types.ErrDenomLocked,
				"Auth: bank admin disabled (%s)",
				denom,
			)
		}
		// Checks if the msg admin is the same as the current admin
		if signer != denomAuth.BankAdmin {
			return errorsmod.Wrapf(
				sdkerrors.ErrUnauthorized,
				"Auth: incorrect admin for denom (%s) only the current bank admin (%s) can perform this action "+
					"(attempted with: %s)",
				denom,
				denomAuth.BankAdmin,
				signer,
			)
		}
		return nil
	case "metadata":
		// First check if the signer is the bank admin
		if signer == denomAuth.BankAdmin {
			return nil
		}

		// If not bank admin, then check metadata admin permissions
		if denomAuth.MetadataAdmin == "" {
			return errorsmod.Wrapf(
				types.ErrDenomLocked,
				"Auth: metadata admin disabled (%s)",
				denom,
			)
		}

		// Checks if the msg admin is the same as the current metadata admin
		if signer != denomAuth.MetadataAdmin {
			return errorsmod.Wrapf(
				sdkerrors.ErrUnauthorized,
				"Auth: incorrect admin for denom (%s) only the current admins bank (%s) or metadata (%s) can update the denom metadata",
				denom,
				denomAuth.BankAdmin,
				denomAuth.MetadataAdmin,
			)
		}
		return nil
	default:
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"Auth: invalid admin type (%s), for denom (%s), only 'bank' or 'metadata' are allowed",
			adminType,
			denom,
		)
	}

}
