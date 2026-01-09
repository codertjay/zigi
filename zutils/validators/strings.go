package validators

import (
	"crypto/sha256"
	"fmt"
	"regexp"

	"github.com/asaskevich/govalidator"
)

// IsURI validates a URI format
func IsURI(uri string) bool {
	return govalidator.IsRequestURI(uri)
}

// StringLengthInRange function checks if the given string is within a specified range.
func StringLengthInRange(str string, min int, max int) bool {
	return len(str) >= min && len(str) <= max
}

// SHA256HashOfURL SHA-256 hash of a given URL
func SHA256HashOfURL(uri string) string {
	// will assume the URL is valid, so error messages are specific to the hash check
	// if !IsURI(uri) {
	// 	return false
	// }
	hashBytes := sha256.Sum256([]byte(uri))
	hashString := fmt.Sprintf("%064x", hashBytes[:])
	return hashString
}

// IsSHA256Hash checks if a given string is a valid SHA-256 hash
func IsSHA256Hash(input string) bool {
	// A valid SHA-256 hash is 64 hexadecimal characters (256 bits = 64 hex chars)
	match, _ := regexp.MatchString(`^[a-fA-F0-9]{64}$`, input)
	return match
}
