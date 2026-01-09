package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// DenomAuthKeyPrefix is the prefix to retrieve all DenomAuth
	DenomAuthKeyPrefix = "DenomAuth/value/"

	// ProposedDenomAuthKeyPrefix is the prefix to retrieve all proposed DenomAuth
	ProposedDenomAuthKeyPrefix = "ProposedDenomAuth/value/"

	// AdminDenomAuthKeyPrefix is the prefix to retrieve all admin denom auth indexes
	AdminDenomAuthKeyPrefix = "AdminDenomAuth/value/"
)

// DenomAuthKey returns the store key to retrieve a DenomAuth from the index fields
func DenomAuthKey(
	denom string,
) []byte {
	var key []byte

	denomBytes := []byte(denom)
	key = append(key, denomBytes...)
	key = append(key, []byte("/")...)

	return key
}

func AdminDenomAuthListKey(admin string) []byte {
	var key []byte
	adminBytes := []byte(admin)
	key = append(key, []byte(AdminDenomAuthKeyPrefix)...)
	key = append(key, adminBytes...)
	key = append(key, []byte("/")...)
	return key
}

// DenomAuthNameKey returns the store key to retrieve a DenomAuth from the index fields as a name
func DenomAuthNameKey(
	denom string,
) []byte {
	var key []byte

	denomBytes := []byte(denom)
	key = append(key, denomBytes...)

	return key
}
