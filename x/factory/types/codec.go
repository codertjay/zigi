package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"

	// checked: included in an early version, not needed anymore, remove for next release
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	// this line is used by starport scaffolding # 1
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgCreateDenom{}, "zigchain/factory/create-denom")
	legacy.RegisterAminoMsg(cdc, &MsgMintAndSendTokens{}, "zigchain/factory/mint-and-send")
	legacy.RegisterAminoMsg(cdc, &MsgSetDenomMetadata{}, "zigchain/factory/set-denom-metadata")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateDenomURI{}, "zigchain/factory/set-denom-uri")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateDenomMintingCap{}, "zigchain/factory/set-denom-minting-cap")
	legacy.RegisterAminoMsg(cdc, &MsgProposeDenomAdmin{}, "zigchain/factory/propose-denom-admin")
	legacy.RegisterAminoMsg(cdc, &MsgClaimDenomAdmin{}, "zigchain/factory/claim-denom-admin")
	legacy.RegisterAminoMsg(cdc, &MsgDisableDenomAdmin{}, "zigchain/factory/disable-denom-admin")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgCreateDenom{},
		&MsgMintAndSendTokens{},
		&MsgSetDenomMetadata{},
		&MsgUpdateDenomURI{},
		&MsgUpdateDenomMintingCap{},
		&MsgUpdateDenomMetadataAuth{},
		&MsgBurnTokens{},
		&MsgWithdrawModuleFees{},
		&MsgProposeDenomAdmin{},
		&MsgClaimDenomAdmin{},
		&MsgDisableDenomAdmin{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// generated with ignite, but is it needed?
var (
	amino = codec.NewLegacyAmino()
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)
	amino.Seal()

}
