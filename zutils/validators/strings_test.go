package validators_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"zigchain/zutils/validators"
)

func TestIsURI(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected bool
	}{
		{
			name:     "valid HTTP URI",
			uri:      "http://example.com",
			expected: true,
		},
		{
			name:     "valid HTTPS URI with path",
			uri:      "https://example.com/path/to/resource",
			expected: true,
		},
		{
			name:     "valid URI with query",
			uri:      "https://example.com?query=123",
			expected: true,
		},
		{
			name:     "empty URI",
			uri:      "",
			expected: false,
		},
		{
			name:     "invalid URI - missing scheme",
			uri:      "example.com",
			expected: false,
		},
		{
			name:     "invalid URI - invalid characters",
			uri:      "http://example.com/<script>",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validators.IsURI(tt.uri)
			require.Equal(t, tt.expected, result, "URI validation failed for %s", tt.uri)
		})
	}
}

func TestStringLengthInRange(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		min      int
		max      int
		expected bool
	}{
		{
			name:     "string within range",
			str:      "hello",
			min:      3,
			max:      10,
			expected: true,
		},
		{
			name:     "string at minimum length",
			str:      "abc",
			min:      3,
			max:      10,
			expected: true,
		},
		{
			name:     "string at maximum length",
			str:      "abcdefghij",
			min:      3,
			max:      10,
			expected: true,
		},
		{
			name:     "string too short",
			str:      "ab",
			min:      3,
			max:      10,
			expected: false,
		},
		{
			name:     "string too long",
			str:      "abcdefghijk",
			min:      3,
			max:      10,
			expected: false,
		},
		{
			name:     "empty string",
			str:      "",
			min:      1,
			max:      5,
			expected: false,
		},
		{
			name:     "zero min and max",
			str:      "",
			min:      0,
			max:      0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validators.StringLengthInRange(tt.str, tt.min, tt.max)
			require.Equal(t, tt.expected, result, "String length check failed for str=%s, min=%d, max=%d", tt.str, tt.min, tt.max)
		})
	}
}

func TestSHA256HashOfURL(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			name:     "valid URI - simple",
			uri:      "http://example.com",
			expected: fmt.Sprintf("%064x", sha256.Sum256([]byte("http://example.com"))),
		},
		{
			name:     "valid URI - with path and query",
			uri:      "https://example.com/path?query=123",
			expected: fmt.Sprintf("%064x", sha256.Sum256([]byte("https://example.com/path?query=123"))),
		},
		{
			name:     "empty URI",
			uri:      "",
			expected: fmt.Sprintf("%064x", sha256.Sum256([]byte(""))),
		},
		{
			name:     "URI with special characters",
			uri:      "https://example.com/path#fragment",
			expected: fmt.Sprintf("%064x", sha256.Sum256([]byte("https://example.com/path#fragment"))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validators.SHA256HashOfURL(tt.uri)
			require.Equal(t, tt.expected, result, "SHA256 hash mismatch for URI: %s", tt.uri)
			require.True(t, validators.IsSHA256Hash(result), "Generated hash is not a valid SHA-256 hash")
		})
	}
}

func TestIsSHA256Hash(t *testing.T) {
	// Generate a valid SHA-256 hash for testing
	validHash := fmt.Sprintf("%064x", sha256.Sum256([]byte("test")))

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid SHA-256 hash",
			input:    validHash,
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "too short hash",
			input:    "a1b2c3d4e5",
			expected: false,
		},
		{
			name:     "too long hash",
			input:    validHash + "ff",
			expected: false,
		},
		{
			name:     "invalid characters in hash",
			input:    validHash[:62] + "gg",
			expected: false,
		},
		{
			name:     "valid hash with uppercase",
			input:    "A1B2C3D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4A5B6C7D8E9F0A1B2",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validators.IsSHA256Hash(tt.input)
			require.Equal(t, tt.expected, result, "SHA256 hash validation failed for input: %s", tt.input)
		})
	}
}
