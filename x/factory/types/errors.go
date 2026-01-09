package types

// DONTCOVER

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
)

// x/factory module sentinel errors
var (
	ErrInvalidSignerAccount        = sdkerrors.Register(ModuleName, 1100, "Signer account is not valid")
	ErrInvalidDenom                = sdkerrors.Register(ModuleName, 1101, "Factory Denom name is not valid")
	ErrInvalidMintingCap           = sdkerrors.Register(ModuleName, 1102, "Minting Cap is not valid")
	ErrInvalidMetadata             = sdkerrors.Register(ModuleName, 1103, "Metadata is not valid")
	ErrCreatorTooLong              = sdkerrors.Register(ModuleName, 1104, fmt.Sprintf("creator too long, max length is %d bytes", MaxCreatorLength))
	ErrDenomDoesNotExist           = sdkerrors.Register(ModuleName, 1105, "denom does not exist")
	ErrDenomExists                 = sdkerrors.Register(ModuleName, 1106, "attempting to create a denom that already exists (has bank metadata)")
	ErrDenomNotFound               = sdkerrors.Register(ModuleName, 1107, "denom not found")
	ErrDenomAuthNotFound           = sdkerrors.Register(ModuleName, 1108, "denom auth not found")
	ErrDenomLocked                 = sdkerrors.Register(ModuleName, 1109, "denom changes are permanently disabled")
	ErrInvalidBankAdminAddress     = sdkerrors.Register(ModuleName, 1110, "Bank admin address is not valid")
	ErrInvalidMetadataAdminAddress = sdkerrors.Register(ModuleName, 1111, "Metadata admin address is not valid")
	ErrCannotChangeMintingCap      = sdkerrors.Register(ModuleName, 1112, "cannot change minting cap")
	ErrInsufficientFunds           = sdkerrors.Register(ModuleName, 1113, "insufficient funds")
	ErrorInvalidParamValue         = sdkerrors.Register(ModuleName, 1114, "invalid factory parameter")
	ErrDuplicateBankAdminProposal  = sdkerrors.Register(ModuleName, 1115, "cannot propose the same bank admin address")
	ErrNoAdminProposal             = sdkerrors.Register(ModuleName, 1116, "no admin has been proposed for denom")
	ErrUnauthorizedAdminClaim      = sdkerrors.Register(ModuleName, 1117, "only the proposed admin can claim the role")
)
