package ztests

import (
	"fmt"

	// we use math/rand to generate random numbers predictably
	// so we can reproduce the same results in tests
	// nosem: math-random-used
	"math/rand" // checked: used for simulation

	"zigchain/zutils/constants"
)

const (
	subDenomCharset     = "abcdefghijklmnopqrstuvwxyz" // Lowercase only
	denomAllowedChars   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/:._-"
	alphanumericCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// RandomDenom generates a random string of the specified length using the provided rand.Rand instance.
func RandomDenom(r *rand.Rand, length int) string {
	if length < 1 {
		return ""
	}

	result := make([]byte, length)

	// The First character must be from the subDenomCharset (no numbers)
	result[0] = subDenomCharset[r.Intn(len(subDenomCharset))]

	// Remaining characters can be from the full charset (including numbers)
	for i := 1; i < length; i++ {
		result[i] = denomAllowedChars[r.Intn(len(denomAllowedChars))]
	}

	return string(result)
}

// RandomAlphanumeric generates a random string of the specified length using the provided rand.Rand instance.
func RandomAlphanumeric(r *rand.Rand, length int) string {
	if length < 1 {
		return ""
	}

	result := make([]byte, length)

	// The First character must be from the subDenomCharset (no numbers)
	result[0] = subDenomCharset[r.Intn(len(subDenomCharset))]

	for i := 0; i < length; i++ {
		result[i] = alphanumericCharset[r.Intn(len(alphanumericCharset))]
	}

	return string(result)
}

// RandomSubDenomRandomLength generates a random string of random length using the provided rand.Rand instance.
func RandomSubDenomRandomLength(r *rand.Rand) string {

	// get number between 3 and 44
	subDenomLength := r.Intn(
		constants.MaxSubDenomLength-constants.MinSubDenomLength-1) + constants.MinSubDenomLength

	return RandomSubDenom(r, subDenomLength)
}

// RandomSubDenom generates a random string using subDenomCharset characters.
func RandomSubDenom(r *rand.Rand, n int) string {
	h := make([]byte, n)
	for i := range h {
		h[i] = subDenomCharset[r.Intn(len(subDenomCharset))]
	}
	return string(h)
}

// Generate a valid SHA256 hash string (64 hex characters)
func RandomSHA256Hash(r *rand.Rand) string {
	return fmt.Sprintf("%064x", r.Uint64())
}
