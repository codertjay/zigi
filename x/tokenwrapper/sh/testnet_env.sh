#!/bin/bash
# Environment variables for tokenwrapper tests

export ZIGCHAIN_NODE=https://testnet-rpc.zigchain.com:443
export ZIGCHAIN_GRPC=https://testnet-grpc.zigchain.com:443
export ZIGCHAIN_WS=wss://testnet-rpc.zigchain.com/websocket
export ZIGCHAIN_CHAIN_ID=zig-test-2
export ZIGCHAIN_FEES=1000uzig

export AXELAR_NODE=https://tm.axelar-testnet.lava.build:443
export AXELAR_GRPC=https://grpc.axelar-testnet.lava.build:443
export AXELAR_WS=wss://tm.axelar-testnet.lava.build/websocket
export AXELAR_CHAIN_ID=axelar-testnet-lisbon-3
export AXELAR_FEES=1400uaxl

export NOBLE_NODE=https://noble-testnet-rpc.polkachu.com:443
export NOBLE_GRPC=http://noble-testnet-grpc.polkachu.com:21590
export NOBLE_WS=wss://noble-testnet-rpc.polkachu.com/websocket
export NOBLE_CHAIN_ID=grand-1

export COSMOS_NODE=https://rpc.provider-sentry-01.ics-testnet.polypore.xyz:443
export COSMOS_GRPC=https://grpc.provider-sentry-01.ics-testnet.polypore.xyz:443
export COSMOS_WS=wss://rpc.provider-sentry-01.ics-testnet.polypore.xyz/websocket
export COSMOS_CHAIN_ID=provider

export ZIGCHAIN_AXELAR_CHANNEL_ID=channel-0
export ZIGCHAIN_AXELAR_CONNECTION_ID=connection-0
export ZIGCHAIN_AXELAR_CLIENT_ID=07-tendermint-0

export AXELAR_ZIGCHAIN_CHANNEL_ID=channel-612
export AXELAR_ZIGCHAIN_CONNECTION_ID=connection-916
export AXELAR_ZIGCHAIN_CLIENT_ID=07-tendermint-1163

export ZIGCHAIN_NOBLE_CHANNEL_ID=channel-44
export ZIGCHAIN_NOBLE_CONNECTION_ID=connection-62
export ZIGCHAIN_NOBLE_CLIENT_ID=07-tendermint-84

export NOBLE_ZIGCHAIN_CONNECTION_ID=connection-520
export NOBLE_ZIGCHAIN_CHANNEL_ID=channel-704
export NOBLE_ZIGCHAIN_CLIENT_ID=07-tendermint-572

export ZIGCHAIN_COSMOS_CHANNEL_ID=channel-43
export ZIGCHAIN_COSMOS_CONNECTION_ID=connection-61
export ZIGCHAIN_COSMOS_CLIENT_ID=07-tendermint-83

export COSMOS_ZIGCHAIN_CONNECTION_ID=connection-279
export COSMOS_ZIGCHAIN_CHANNEL_ID=channel-566
export COSMOS_ZIGCHAIN_CLIENT_ID=07-tendermint-388

export WORK_DIR="."
export EXPECTED_DECIMAL_DIFFERENCE="12"
export EXPECTED_DENOM="unit-zig"
export EXPECTED_NATIVE_CHANNEL="channel-0"
export EXPECTED_NATIVE_PORT="transfer"
export EXPECTED_NATIVE_CLIENT_ID="07-tendermint-0"
export EXPECTED_COUNTERPARTY_CHANNEL="channel-612"
export EXPECTED_COUNTERPARTY_PORT="transfer"
export EXPECTED_COUNTERPARTY_CLIENT_ID="07-tendermint-1163"
export EXPECTED_ENABLE="true"
export EXPECTED_OPERATOR_ADDRESS="zig1fw7yrucw09guffffdac64c3x5yc6rtlpt0nm5a"
export EXPECTED_PROPOSED_OPERATOR_ADDRESS="zig1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqk9whvl"
export EXPECTED_MODULE_ADDRESS="zig1hdq87rzf327fwz8rw9rnmchj7qa3uxrpxds2fw"

export OP_ACCOUNT_NAME="op"
export TEST_ACCOUNT_NAME="testaccount"

export AXELAR_ADDRESS_PREFIX="axelar"
export ZIGCHAIN_ADDRESS_PREFIX="zig"
export DUMMY_ADDRESS_PREFIX="dummy"
export AXELAR_DENOM="unit-zig"
export ZIGCHAIN_DENOM="uzig"
export DUMMY_DENOM="udummy"
export ZIG_ALICE_INITIAL_BALANCE="100000000uzig"
export ZIG_BOB_INITIAL_BALANCE="500000000uzig"
export ZIG_OPERATOR_INITIAL_BALANCE="100000000uzig"
export ZIG_FAUCET_INITIAL_BALANCE="500000000uzig"
export ZIG_FAUCET_AMOUNT="500uzig"
export ZIG_BOB_MNEMONIC="diagram saddle click pipe medal text bounce spread elbow rebel couple grocery exotic piano okay kiwi tornado summer tube cool pipe tower scare spin"
export ZIG_OPERATOR_MNEMONIC="helmet tornado sure split just dream orbit give explain check pact hat similar silk jelly cinnamon aim escape myself moment drive orange blue knife"
export AXELAR_ALICE_INITIAL_BALANCE="100000000000000000000unit-zig"
export AXELAR_BOB_INITIAL_BALANCE="500000000000000000000000unit-zig"
export AXELAR_FAUCET_INITIAL_BALANCE="500000000000000000000000unit-zig"
export AXELAR_FAUCET_AMOUNT="500000000000000000000unit-zig"
export AXELAR_BOB_MNEMONIC="canyon depart chief choose winner bone blouse zone mandate feature note sport seat increase sell history patrol renew ozone travel scare easy drift fog"
export DUMMY_ALICE_INITIAL_BALANCE="100000000000000000000unit-zig"
export DUMMY_BOB_INITIAL_BALANCE="5500000000000000000000unit-zig"
export DUMMY_FAUCET_INITIAL_BALANCE="5500000000000000000000unit-zig"
export DUMMY_FAUCET_AMOUNT="500000000000000000000unit-zig"
export DUMMY_BOB_MNEMONIC="spike work diary decrease ribbon already real recycle run sad ball patch economy help tooth file embark recycle acquire belt series slogan goat twin"
export COSMOS_ALICE_INITIAL_BALANCE="100000000000000000000uatom"
export COSMOS_BOB_INITIAL_BALANCE="5000000000000000000000uatom"
export COSMOS_FAUCET_INITIAL_BALANCE="5000000000000000000000uatom"
export COSMOS_FAUCET_AMOUNT="500000000000000000000uatom"
export COSMOS_BOB_MNEMONIC="flight portion muscle angle way between pumpkin kit dry age visual stool axis snake bracket slogan kite rug huge typical this argue secret broken"
export TOKEN_WRAPPER_OPERATOR_ADDRESS="zig199d3ngzyz8up8wm4605wtadnnj44chn9wz4la6"
export TOKEN_WRAPPER_ENABLED="true"
export ZIG_TRANSFER_AMOUNT="10uzig"
export AXELAR_TRANSFER_AMOUNT="10000000000000unit-zig"
export MODULE_FUND_AMOUNT="1000uzig"
export ZIGCHAIN_REPO_URL="http://18.183.194.220:3001/nat/zigchain_private.git"
export ZIGCHAIN_BRANCH=""
export ZIGCHAIN_LOCAL_PATH="/Users/caner/Dev/zignaly/zigchain"
export CHAIN_START_TIMEOUT="90"
export RELAYER_START_TIMEOUT="10"
export IBC_TRANSFER_TRANSACTION_TIMEOUT="20"
export TRANSACTION_TIMEOUT="10"
export TIMEOUT_TEST_ENABLED="true"
export TIMEOUT_TEST_AMOUNT="1000000000000unit-zig"
export TIMEOUT_TEST_NATIVE_AMOUNT="1uzig"
export TIMEOUT_TEST_DURATION="1000000000"

# Test configuration for TW-002: Set to true if module wallet is expected to have balances on testnet
export EXPECTED_BALANCES="true"