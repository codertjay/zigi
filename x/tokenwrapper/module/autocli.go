package tokenwrapper

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "zigchain/api/zigchain/tokenwrapper"
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
					Short:     "Shows the parameters of the module",
					Example:   "zigchaind q tokenwrapper params",
				},
				{
					RpcMethod:      "ModuleInfo",
					Use:            "module-info",
					Short:          "Show the information of the module",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
					Example:        "zigchaind q tokenwrapper module-info",
				},
				{
					RpcMethod:      "TotalTransfers",
					Use:            "total-transfers",
					Short:          "Show the total amount of ZIG tokens transferred in and out",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
					Example:        "zigchaind q tokenwrapper total-transfers",
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
					RpcMethod:      "FundModuleWallet",
					Use:            "fund-module-wallet [amount]",
					Short:          "Fund the module wallet with tokens",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "amount"}},
					Example:        "zigchaind tx tokenwrapper fund-module-wallet 1000uzig --from bob -y",
				},
				{
					RpcMethod:      "WithdrawFromModuleWallet",
					Use:            "withdraw-from-module-wallet [amount]",
					Short:          "Withdraw tokens from the module wallet",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "amount"}},
					Example:        "zigchaind tx tokenwrapper withdraw-from-module-wallet 1000uzig --from bob -y",
				},
				{
					RpcMethod:      "ProposeOperatorAddress",
					Use:            "propose-operator-address [new-operator]",
					Short:          "Propose a new operator address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "new_operator"}},
					Example:        "zigchaind tx tokenwrapper propose-operator-address zig15yk64u7zc9g9k2yr2wmzeva5qgwxps6y8c2amk --from bob -y",
				},
				{
					RpcMethod: "ClaimOperatorAddress",
					Use:       "claim-operator-address",
					Short:     "Claim the operator role",
					Example:   "zigchaind tx tokenwrapper claim-operator-address --from bob -y",
				},
				{
					RpcMethod: "EnableTokenWrapper",
					Use:       "enable-token-wrapper",
					Short:     "Enable the token wrapper functionality",
					Example:   "zigchaind tx tokenwrapper enable-token-wrapper --from bob -y",
				},
				{
					RpcMethod: "DisableTokenWrapper",
					Use:       "disable-token-wrapper",
					Short:     "Disable the token wrapper functionality",
					Example:   "zigchaind tx tokenwrapper disable-token-wrapper --from bob -y",
				},
				{
					RpcMethod: "UpdateIbcSettings",
					Use:       "update-ibc-settings [native-client-id] [counterparty-client-id] [native-port] [counterparty-port] [native-channel] [counterparty-channel] [denom] [decimal-difference]",
					Short:     "Update the IBC settings",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "native_client_id"},
						{ProtoField: "counterparty_client_id"},
						{ProtoField: "native_port"},
						{ProtoField: "counterparty_port"},
						{ProtoField: "native_channel"},
						{ProtoField: "counterparty_channel"},
						{ProtoField: "denom"},
						{ProtoField: "decimal_difference"},
					},
					Example: "zigchaind tx tokenwrapper update-ibc-settings 07-tendermint-0 07-tendermint-0 transfer transfer channel-0 channel-0 waxlzig 12 --from bob -y",
				},
				{
					RpcMethod:      "AddPauserAddress",
					Use:            "add-pauser-address [new-pauser]",
					Short:          "Add a new pauser address",
					Example:        "zigchaind tx tokenwrapper add-pauser-address zig15yk64u7zc9g9k2yr2wmzeva5qgwxps6y8c2amk --from bob -y",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "new_pauser"}},
				},
				{
					RpcMethod:      "RemovePauserAddress",
					Use:            "remove-pauser-address [pauser]",
					Short:          "Remove a pauser address",
					Example:        "zigchaind tx tokenwrapper remove-pauser-address zig15yk64u7zc9g9k2yr2wmzeva5qgwxps6y8c2amk --from bob -y",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "pauser"}},
				},
				{
					RpcMethod:      "RecoverZig",
					Use:            "recover-zig [address]",
					Short:          "Send a recover-zig tx",
					Example:        "zigchaind tx tokenwrapper recover-zig zig15yk64u7zc9g9k2yr2wmzeva5qgwxps6y8c2amk --from bob -y",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
