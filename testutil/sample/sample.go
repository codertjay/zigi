package sample

import (
	// checked: used to generate sample account address
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"zigchain/zutils/constants"
)

func init() {
	// Make sure that addresses the same prefix in blockchain and tests/simulation
	sdk.GetConfig().SetBech32PrefixForAccount(constants.AddressPrefix, "pub")
}

// AccAddress returns a sample account address
func AccAddress() string {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()

	return sdk.AccAddress(addr).String()
}

// ZeroAccAddress returns a sample account with the zero address
func ZeroAccAddress() string {
	return sdk.AccAddress([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}).String()
}

func PoolModuleAccount(poolAddress sdk.AccAddress) sdk.AccountI {
	return authtypes.NewModuleAccount(
		// from the base account provided by the address (has only address)
		authtypes.NewBaseAccountWithAddress(poolAddress),
		poolAddress.String(),
	)
}

func Address() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(AccAddress())
}

func ZeroAddress() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(ZeroAccAddress())
}

func Coin(denom string, amount int64) sdk.Coin {
	return sdk.NewInt64Coin(denom, amount)
}
