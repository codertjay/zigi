package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// PoolUidsKeyPrefix is the prefix to retrieve all PoolUids
	PoolUidsKeyPrefix = "PoolUids/value/"

	// PoolUidSeparator is the separator between the two denoms in the poolUid
	PoolUidSeparator = "-"
)

// PoolUidsKey returns the store key to retrieve a PoolUids from the index fields
func PoolUidsKey(
	poolUid string,
) []byte {
	var key []byte

	poolUidBytes := []byte(poolUid)
	key = append(key, poolUidBytes...)
	key = append(key, []byte("/")...)

	return key
}
