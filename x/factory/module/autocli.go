package factory

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "zigchain/api/zigchain/factory"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: modulev1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module.",
					Example:   "  zigchaind query factory params --chain-id zigchain",
				},
				{
					RpcMethod: "DenomAll",
					Use:       "list-denom",
					Short:     "List all Denoms.",
					Alias:     []string{"denoms", "all-denom"},
					Example:   "  zigchaind query factory list-denom --chain-id zigchain",
				},
				{
					RpcMethod: "Denom",
					Use:       "show-denom [denom]",
					Short:     "Shows a Denom.",
					Alias:     []string{"denom", "denom-info"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
					},
					Example: "  zigchaind query factory show-denom coin.zig1vhn9xt0tnmve9nr9dtqt4emcsae40a3cs4v049.abcd --chain-id zigchain",
				},
				{
					RpcMethod: "DenomsByAdmin",
					Use:       "denoms-by-admin [admin]",
					Short:     "Get all denoms by admin.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "admin"},
					},
					Example: "  zigchaind query factory denoms-by-admin zig1w8vt9ec0s0cg8umgv3ln5furqhzaa0nrcakgks --chain-id zigchain\n" +
						"  zigchaind query factory denoms-by-admin $(zigchaind keys show -a z)",
				},

				{
					RpcMethod: "DenomAuth",
					Use:       "denom-auth [denom]",
					Short:     "Get a denomAuth.",
					Alias:     []string{"show-denom-auth"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
					},
					Example: "  zigchaind query factory denom-auth coin.zig1vhn9xt0tnmve9nr9dtqt4emcsae40a3cs4v049.abcd --chain-id zigchain",
				},
				// this line is used by ignite scaffolding # autocli/query
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              modulev1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod: "CreateDenom",
					Use:       "create-denom [sub-denom] [minting-cap] [can-change-minting-cap] [uri] [uri-hash]",
					Short:     "Create a new factory denom.",
					Alias:     []string{"create"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sub_denom"},
						{ProtoField: "minting_cap"},
						{ProtoField: "can_change_minting_cap"},
						{ProtoField: "URI"},
						{ProtoField: "URI_hash"},
					},
					Example: "  zigchaind tx factory create-denom abcd 1000000000 true 'ipfs://ipfs.io/XXX' 'hrekwjhrewhjkrhew321312klj' --from z --chain-id zigchain --gas-prices 0.0025uzig  --gas auto --gas-adjustment 1.3",
				},
				{
					RpcMethod: "MintAndSendTokens",
					Use:       "mint-and-send-tokens [amount-token] [recipient-address]",
					Short:     "Mint and send tokens.",
					Long:      "Mint and send tokens to a recipient address.",
					Alias:     []string{"mint"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "token"},
						{ProtoField: "recipient"},
					},
					Example: "  zigchaind tx factory mint-and-send-tokens 100coin.zig1vhn9xt0tnmve9nr9dtqt4emcsae40a3cs4v049.abcd zig1q2w3e4r5t6y7u8i9o0p1a2s3d4f5g6h7j8k9l0 --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3\n" +
						"  zigchaind tx factory mint-and-send-tokens 100coin.zig1vhn9xt0tnmve9nr9dtqt4emcsae40a3cs4v049.abcd $(zigchaind keys show -a account1) --from z --chain-id zigchain --dry-run ",
				},
				{
					RpcMethod: "BurnTokens",
					Use:       "burn-tokens [amount-denom]",
					Short:     "Burn specific amount of token from bank admin account.",
					Long: "Burn specific amount of token from bank admin account. " +
						"Only bank admin can perform this action. ",
					Alias: []string{"burn"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "token"},
					},
					Example: "  zigchaind tx factory burn-tokens 50coin.zig1vhn9xt0tnmve9nr9dtqt4emcsae40a3cs4v049.abcd --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3\n" +
						"  zigchaind tx factory burn-tokens 50coin.zig1vhn9xt0tnmve9nr9dtqt4emcsae40a3cs4v049.abcd --from z --chain-id zigchain --dry-run",
				},
				{
					RpcMethod: "SetDenomMetadata",
					Use:       "set-denom-metadata [metadata]",
					Short:     "Set denom metadata.",
					Long:      "Set the metadata of a token.",
					Alias:     []string{"set-metadata", "set-meta", "denom-meta"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "metadata"},
					},
					Example: "zigchaind tx factory set-denom-metadata '{metadata}' --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3",
				},
				{
					RpcMethod: "ProposeDenomAdmin",
					Use:       "propose-denom-admin [denom] [bank-admin-address] [metadata-admin-address]",
					Short:     "Propose new admin addresses for a denom",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
						{ProtoField: "bank_admin"},
						{ProtoField: "metadata_admin"},
					},
					Example: "  zigchaind tx factory propose-denom-admin coin.zig1vhn9xt0tnmve9nr9dtqt4emcsae40a3cs4v049.abcd zig1q2w3e4r5t6y7u8i9o0p1a2s3d4f5g6h7j8k9l0 zig1ulwc5xa4pp4sa0x04aff2tluvq2epqp5xt2kfj --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3",
				},
				{
					RpcMethod: "ClaimDenomAdmin",
					Use:       "claim-denom-admin [denom]",
					Short:     "Claim the denom admin role",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
					},
					Example: "  zigchaind tx factory claim-denom-admin coin.zig1vhn9xt0tnmve9nr9dtqt4emcsae40a3cs4v049.abcd --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3",
				},
				{
					RpcMethod: "DisableDenomAdmin",
					Use:       "disable-denom-admin [denom]",
					Short:     "Disable the denom admin role",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
					},
					Example: "  zigchaind tx factory disable-denom-admin coin.zig1vhn9xt0tnmve9nr9dtqt4emcsae40a3cs4v049.abcd --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3",
				},
				{
					RpcMethod: "UpdateDenomURI",
					Use:       "update-denom-uri [denom] [uri] [uri-hash]",
					Short:     "Change denom meta data URI and URI hash",
					Alias:     []string{"update-uri", "uri"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
						{ProtoField: "URI"},
						{ProtoField: "URI_hash"},
					},
					Example: "  zigchaind tx factory update-denom-uri coin.zig1vhn9xt0tnmve9nr9dtqt4emcsae40a3cs4v049.abcd 'ipfs://ipfs.io/XXX' 'hrekwjhrewhjkrhew321312klj' --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3",
				},
				{
					RpcMethod: "UpdateDenomMintingCap",
					Use:       "update-denom-minting-cap [denom] [minting-cap] [can-change-minting-cap]",
					Short:     "Change denom minting cap and ability to change minting cap",
					Alias:     []string{"update-minting-cap", "minting-cap"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
						{ProtoField: "minting_cap"},
						{ProtoField: "can_change_minting_cap"},
					},
					Example: "  zigchaind tx factory update-denom-minting-cap coin.zig1vhn9xt0tnmve9nr9dtqt4emcsae40a3cs4v049.abcd 2000000000 true --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3",
				},
				{
					RpcMethod: "UpdateDenomMetadataAuth",
					Use:       "update-denom-metadata-auth [denom] [metadata-admin]",
					Short:     "Send a updateDenomMetadataAuth tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
						{ProtoField: "metadata_admin"},
					},
					Example: "  zigchaind tx factory update-denom-metadata-auth factory/zig1vhn9xt0tnmve9nr9dtqt4emcsae40a3cs4v049/abcd zig1q2w3e4r5t6y7u8i9o0p1a2s3d4f5g6h7j8k9l0 --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3",
				},
				{
					RpcMethod: "WithdrawModuleFees",
					Use:       "withdraw-module-fees [receiver]",
					Short:     "Send a withdrawModuleFees tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "receiver"},
					},
					Example: "  zigchaind tx factory withdraw-module-fees zig1q2w3e4r5t6y7u8i9o0p1a2s3d4f5g6h7j8k9l0 --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3",
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
