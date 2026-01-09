package constants

// BondDenom coin denomination
const (
	BondDenom = "uzig"

	// BondDenomDecimals number of decimal places in the native coin
	BondDenomDecimals = 6

	// MinSubDenomLength minimum length of the coin denomination
	MinSubDenomLength = 3

	// MaxSubDenomLength maximum length of the coin denomination
	MaxSubDenomLength = 44

	// MaxDenomLength len("factory") = 7 + 1 + macCreatorLength + 1 + MaxSubdenomLength = 8 + 1 + 59 + 1 + 44 = 112
	// We leave some buffer - 128 is cosmos limitation anyway
	MaxDenomLength = 127

	MaxURILength = 256

	// MaxURIHashLength The length matches base64 of 32 bytes without padding.
	// SHA256 produces a string of 32 bytes.
	// This string is usually represented in the hexadecimal format (as 64 characters [0-9a-f]),
	// but it's not a requirement. One may choose to use a different encoding to make the produced string shorter.
	MaxURIHashLength = 64
)
