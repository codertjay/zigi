package dex

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "zigchain/api/zigchain/dex"
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
					Example:   "  zigchaind query dex params --chain-id zigchain",
				},
				{
					RpcMethod: "ListPool",
					Use:       "list-pool",
					Short:     "List all pools",
					Example:   "  zigchaind query dex list-pool --chain-id zigchain",
				},
				{
					RpcMethod: "GetPool",
					Use:       "get-pool [pool-id]",
					Short:     "Gets a pool meta data",
					Alias:     []string{"show-pool", "pool"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "pool_id"},
					},
					Example: "  zigchaind query dex get-pool zp1 --chain-id zigchain\n" +
						"  zigchaind query dex show-pool zp1 --chain-id zigchain",
				},
				{
					RpcMethod: "GetPoolBalances",
					Use:       "get-pool-balances [pool-id]",
					Short:     "Gets a pool meta plus balances to compare accounting with account balances",
					Alias:     []string{"show-pool-balances", "pool-balances"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "pool_id"},
					},
					Example: "  zigchaind query dex get-pool-balances zp1 --chain-id zigchain\n" +
						"  zigchaind query dex pool-balances zp1 --chain-id zigchain",
				},
				// {
				// 	RpcMethod: "GetPoolsMeta",
				// 	Use:       "get-pools-meta",
				// 	Short:     "Gets a pools-meta",
				// 	Alias:     []string{"show-pools-meta"},
				// },
				{
					RpcMethod: "ListPoolUids",
					Use:       "list-pool-uids",
					Short:     "List all pool-uids",
					Example:   "  zigchaind query dex list-pool-uids --chain-id zigchain",
				},
				{
					RpcMethod: "GetPoolUid",
					Use:       "get-pool-uid [base] [quote]",
					Short:     "Gets a pool-id from tokens",
					Alias:     []string{"show-pool-uid", "pool-id-by-tokens"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "base"},
						{ProtoField: "quote"},
					},
					Example: "  zigchaind query dex get-pool-uid coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.abc coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.usdt --chain-id zigchain",
				},
				{
					RpcMethod: "SwapIn",
					Use:       "swap-in [pool-id] [incoming-token]",
					Short:     "Calculates the amount of token going out given token amount coming in",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "pool_id"},
						{ProtoField: "coin_in"},
					},
					Example: "  zigchaind query dex swap-in zp1 100coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.abc --chain-id zigchain",
				},
				{
					RpcMethod: "SwapOut",
					Use:       "swap-out [pool-id] [outgoing-token]",
					Short:     "Calculates the amount of token going in given token amount coming out",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "pool_id"},
						{ProtoField: "coin_out"},
					},
					Example: "  zigchaind query dex swap-out zp1 100coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.abc --chain-id zigchain",
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
					RpcMethod: "CreatePool",
					Use:       "create-pool [base] [quote] --receiver (optional)",
					Short:     "Create a new pool",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						// {ProtoField: "poolId"},
						{ProtoField: "base"},
						{ProtoField: "quote"},
						// {ProtoField: "receiver", Optional: true},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"receiver": {
							Name:         "receiver",
							Shorthand:    "r",
							Usage:        "Address of the receiver of liquidity coins (zig1...)",
							DefaultValue: "", // no default
						},
					},
					Example: "  zigchaind tx dex create-pool 10coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.abc 30coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.usdt --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3",
				},
				{
					RpcMethod: "SwapExactIn",
					Use:       "swap-exact-in [pool-id] [token] --receiver (optional) --outgoing-min (optional)]",
					Short:     "Send a swap tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "pool_id"},
						{ProtoField: "incoming"},
						// the optional fields - below
						// {ProtoField: "receiver", Optional: true},
						// {ProtoField: "outgoingMin", Optional: true},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"receiver": {
							Name:         "receiver",
							Shorthand:    "r",
							Usage:        "Address of the receiver (zig1...)",
							DefaultValue: "", // no default
						},
						"outgoing_min": {
							Name:      "outgoing-min",
							Shorthand: "m",
							Usage:     "Minimum outgoing amount (123abc), how much you are willing to receive, if not set, it will be 0",
						},
					},
					Example: "  zigchaind tx dex swap zp1 100coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.abc --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3\n" +
						"  zigchaind tx dex swap zp1 100coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.abc --receiver zig1abc --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3\n" +
						"  zigchaind tx dex swap zp1 100coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.abc --receiver zig1abc --outgoing-min 123abc --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3\n",
				},
				{
					RpcMethod: "SwapExactOut",
					Use:       "swap-exact-out [pool-id] [outgoing_token_amount] --receiver (optional) --incoming-max (optional)]",
					Short:     "Send a swap exact out tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "pool_id"},
						{ProtoField: "outgoing"},
						// the optional fields - below
						// {ProtoField: "receiver", Optional: true},
						// {ProtoField: "incomingMax", Optional: true},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"receiver": {
							Name:         "receiver",
							Shorthand:    "r",
							Usage:        "Address of the receiver (zig1...)",
							DefaultValue: "", // no default
						},
						"incoming_max": {
							Name:      "incoming-max",
							Shorthand: "m",
							Usage:     "Maximum incoming amount (123abc), how much you are willing to pay, if not set, it can slip",
						},
					},
					Example: "  zigchaind tx dex swap zp1 100coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.abc --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3\n" +
						"  zigchaind tx dex swap zp1 100coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.abc --receiver zig1abc --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3\n" +
						"  zigchaind tx dex swap zp1 100coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.abc --receiver zig1abc --outgoing-min 123abc --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3\n",
				},
				{
					RpcMethod: "AddLiquidity",
					Use:       "add-liquidity [pool-id] [base] [quote] --receiver (optional)",
					Short:     "Send a addLiquidity tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "pool_id"},
						{ProtoField: "base"},
						{ProtoField: "quote"},
						// {ProtoField: "receiver", Optional: true},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"receiver": {
							Name:         "receiver",
							Shorthand:    "r",
							Usage:        "Address of the receiver of liquidity coins (zig1...)",
							DefaultValue: "", // no default
						},
					},
					Example: "  zigchaind tx dex add-liquidity zp1 10coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.abc 30coin.zig1ajg7jku4crf46lcskykwvkjrwfj7zan98az4k2.usdt --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3"},
				{
					RpcMethod: "RemoveLiquidity",
					Use:       "remove-liquidity [lptoken] --receiver (optional)",
					Short:     "Send a removeLiquidity tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "lptoken"},
						// {ProtoField: "receiver", Optional: true},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"receiver": {
							Name:         "receiver",
							Shorthand:    "r",
							Usage:        "Address of the receiver of swap coins (zig1...)",
							DefaultValue: "", // no default
						},
					},
					Example: "  zigchaind tx dex remove-liquidity 10zp1 --from z --chain-id zigchain --gas-prices 0.25uzig --gas auto --gas-adjustment 1.3",
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
