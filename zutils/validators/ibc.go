package validators

import (
	"fmt"

	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
)

func ValidateClientId(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == "" {
		return fmt.Errorf("%w: client ID cannot be empty", clienttypes.ErrInvalidClient)
	}
	if !clienttypes.IsValidClientID(v) {
		return fmt.Errorf("%w: invalid client ID format", clienttypes.ErrInvalidClient)
	}
	return nil
}

func ValidatePort(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == "" {
		return fmt.Errorf("%w: port cannot be empty", porttypes.ErrInvalidPort)
	}
	if len(v) < 2 || len(v) > 128 {
		return fmt.Errorf("%w: port length must be between 2 and 128 characters", porttypes.ErrInvalidPort)
	}
	if !IsValidIdentifier(v) {
		return fmt.Errorf("%w: port contains invalid characters. Only alphanumeric, ., _, +, -, #, [, ], <, > are allowed", porttypes.ErrInvalidPort)
	}
	return nil
}

func ValidateChannel(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == "" {
		return fmt.Errorf("%w: channel cannot be empty", channeltypes.ErrInvalidChannelIdentifier)
	}
	if !channeltypes.IsValidChannelID(v) {
		return fmt.Errorf("%w: invalid channel ID format", channeltypes.ErrInvalidChannelIdentifier)
	}
	return nil
}

func ValidateDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == "" {
		return fmt.Errorf("denom cannot be empty")
	}
	if !IsValidIdentifier(v) {
		return fmt.Errorf("denom contains invalid characters. Only alphanumeric, ., _, +, -, #, [, ], <, > are allowed")
	}
	if err := CheckDenomString(v); err != nil {
		return err
	}
	return nil
}

func ValidateDecimalDifference(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v > 18 {
		return fmt.Errorf("decimal difference cannot be greater than 18")
	}
	return nil
}

// IsValidIdentifier checks if a string contains only valid IBC identifier characters
func IsValidIdentifier(s string) bool {
	for _, c := range s {
		if !IsValidIdentifierChar(c) {
			return false
		}
	}
	return true
}

// IsValidIdentifierChar checks if a rune is a valid IBC identifier character
func IsValidIdentifierChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '.' || c == '_' || c == '+' || c == '-' || c == '#' ||
		c == '[' || c == ']' || c == '<' || c == '>'
}
