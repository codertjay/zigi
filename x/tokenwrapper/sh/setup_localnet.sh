#!/bin/bash

# Localnet Environment Setup Script
#
# This script sets up the localnet environment with 4 chains (Axelar, ZIGChain, Dummy, Cosmos)
# and 3 IBC relayers for testing the tokenwrapper module's IBC functionality.

# Exit on error
set -e

# Environment variables with defaults
# Chain configurations
export AXELAR_CHAIN_ID=${AXELAR_CHAIN_ID:-"axelar"}
export ZIGCHAIN_CHAIN_ID=${ZIGCHAIN_CHAIN_ID:-"zigchain"}
export DUMMY_CHAIN_ID=${DUMMY_CHAIN_ID:-"dummy"}
export COSMOS_CHAIN_ID=${COSMOS_CHAIN_ID:-"cosmos"}

export AXELAR_ADDRESS_PREFIX=${AXELAR_ADDRESS_PREFIX:-"axelar"}
export ZIGCHAIN_ADDRESS_PREFIX=${ZIGCHAIN_ADDRESS_PREFIX:-"zig"}
export DUMMY_ADDRESS_PREFIX=${DUMMY_ADDRESS_PREFIX:-"dummy"}
export COSMOS_ADDRESS_PREFIX=${COSMOS_ADDRESS_PREFIX:-"cosmos"}

export AXELAR_DENOM=${AXELAR_DENOM:-"unit-zig"}
export ZIGCHAIN_DENOM=${ZIGCHAIN_DENOM:-"uzig"}
export DUMMY_DENOM=${DUMMY_DENOM:-"unit-zig"}
export COSMOS_DENOM=${COSMOS_DENOM:-"uatom"}

# Ports
export AXELAR_RPC_PORT=${AXELAR_RPC_PORT:-"26657"}
export ZIGCHAIN_RPC_PORT=${ZIGCHAIN_RPC_PORT:-"26659"}
export DUMMY_RPC_PORT=${DUMMY_RPC_PORT:-"26661"}
export COSMOS_RPC_PORT=${COSMOS_RPC_PORT:-"26663"}

export AXELAR_P2P_PORT=${AXELAR_P2P_PORT:-"26656"}
export ZIGCHAIN_P2P_PORT=${ZIGCHAIN_P2P_PORT:-"26658"}
export DUMMY_P2P_PORT=${DUMMY_P2P_PORT:-"26660"}
export COSMOS_P2P_PORT=${COSMOS_P2P_PORT:-"26662"}

export AXELAR_PPROF_PORT=${AXELAR_PPROF_PORT:-"6061"}
export ZIGCHAIN_PPROF_PORT=${ZIGCHAIN_PPROF_PORT:-"6062"}
export DUMMY_PPROF_PORT=${DUMMY_PPROF_PORT:-"6063"}
export COSMOS_PPROF_PORT=${COSMOS_PPROF_PORT:-"6064"}

export AXELAR_GRPC_PORT=${AXELAR_GRPC_PORT:-"9090"}
export ZIGCHAIN_GRPC_PORT=${ZIGCHAIN_GRPC_PORT:-"9092"}
export DUMMY_GRPC_PORT=${DUMMY_GRPC_PORT:-"9094"}
export COSMOS_GRPC_PORT=${COSMOS_GRPC_PORT:-"9096"}

export AXELAR_GRPCWEB_PORT=${AXELAR_GRPCWEB_PORT:-"9091"}
export ZIGCHAIN_GRPCWEB_PORT=${ZIGCHAIN_GRPCWEB_PORT:-"9093"}
export DUMMY_GRPCWEB_PORT=${DUMMY_GRPCWEB_PORT:-"9095"}
export COSMOS_GRPCWEB_PORT=${COSMOS_GRPCWEB_PORT:-"9097"}

export AXELAR_API_PORT=${AXELAR_API_PORT:-"1317"}
export ZIGCHAIN_API_PORT=${ZIGCHAIN_API_PORT:-"1318"}
export DUMMY_API_PORT=${DUMMY_API_PORT:-"1319"}
export COSMOS_API_PORT=${COSMOS_API_PORT:-"1320"}

export AXELAR_FAUCET_PORT=${AXELAR_FAUCET_PORT:-"4500"}
export ZIGCHAIN_FAUCET_PORT=${ZIGCHAIN_FAUCET_PORT:-"4501"}
export DUMMY_FAUCET_PORT=${DUMMY_FAUCET_PORT:-"4502"}
export COSMOS_FAUCET_PORT=${COSMOS_FAUCET_PORT:-"4503"}

# Zig account configurations
export ZIG_ALICE_INITIAL_BALANCE=${ZIG_ALICE_INITIAL_BALANCE:-"100000000uzig"}
export ZIG_BOB_INITIAL_BALANCE=${ZIG_BOB_INITIAL_BALANCE:-"500000000uzig"}
export ZIG_OPERATOR_INITIAL_BALANCE=${ZIG_OPERATOR_INITIAL_BALANCE:-"100000000uzig"}
export ZIG_FAUCET_INITIAL_BALANCE=${ZIG_FAUCET_INITIAL_BALANCE:-"500000000uzig"}
export ZIG_FAUCET_AMOUNT=${ZIG_FAUCET_AMOUNT:-"500uzig"}
export ZIG_BOB_MNEMONIC=${ZIG_BOB_MNEMONIC:-"diagram saddle click pipe medal text bounce spread elbow rebel couple grocery exotic piano okay kiwi tornado summer tube cool pipe tower scare spin"}
export ZIG_OPERATOR_MNEMONIC=${ZIG_OPERATOR_MNEMONIC:-"helmet tornado sure split just dream orbit give explain check pact hat similar silk jelly cinnamon aim escape myself moment drive orange blue knife"}

# Axelar account configurations
export AXELAR_ALICE_INITIAL_BALANCE=${AXELAR_ALICE_INITIAL_BALANCE:-"100000000000000000000unit-zig"}
export AXELAR_BOB_INITIAL_BALANCE=${AXELAR_BOB_INITIAL_BALANCE:-"500000000000000000000000unit-zig"}
export AXELAR_FAUCET_INITIAL_BALANCE=${AXELAR_FAUCET_INITIAL_BALANCE:-"500000000000000000000000unit-zig"}
export AXELAR_FAUCET_AMOUNT=${AXELAR_FAUCET_AMOUNT:-"500000000000000000000unit-zig"}
export AXELAR_BOB_MNEMONIC=${AXELAR_BOB_MNEMONIC:-"canyon depart chief choose winner bone blouse zone mandate feature note sport seat increase sell history patrol renew ozone travel scare easy drift fog"}

# Dummy account configurations
export DUMMY_ALICE_INITIAL_BALANCE=${DUMMY_ALICE_INITIAL_BALANCE:-"100000000000000000000unit-zig"}
export DUMMY_BOB_INITIAL_BALANCE=${DUMMY_BOB_INITIAL_BALANCE:-"5500000000000000000000unit-zig"}
export DUMMY_FAUCET_INITIAL_BALANCE=${DUMMY_FAUCET_INITIAL_BALANCE:-"5500000000000000000000unit-zig"}
export DUMMY_FAUCET_AMOUNT=${DUMMY_FAUCET_AMOUNT:-"500000000000000000000unit-zig"}
export DUMMY_BOB_MNEMONIC=${DUMMY_BOB_MNEMONIC:-"spike work diary decrease ribbon already real recycle run sad ball patch economy help tooth file embark recycle acquire belt series slogan goat twin"}

# Cosmos account configurations
export COSMOS_ALICE_INITIAL_BALANCE=${COSMOS_ALICE_INITIAL_BALANCE:-"100000000000000000000uatom"}
export COSMOS_BOB_INITIAL_BALANCE=${COSMOS_BOB_INITIAL_BALANCE:-"5000000000000000000000uatom"}
export COSMOS_FAUCET_INITIAL_BALANCE=${COSMOS_FAUCET_INITIAL_BALANCE:-"5000000000000000000000uatom"}
export COSMOS_FAUCET_AMOUNT=${COSMOS_FAUCET_AMOUNT:-"500000000000000000000uatom"}
export COSMOS_BOB_MNEMONIC=${COSMOS_BOB_MNEMONIC:-"flight portion muscle angle way between pumpkin kit dry age visual stool axis snake bracket slogan kite rug huge typical this argue secret broken"}

# Token wrapper configuration
export TOKEN_WRAPPER_OPERATOR_ADDRESS=${TOKEN_WRAPPER_OPERATOR_ADDRESS:-"zig199d3ngzyz8up8wm4605wtadnnj44chn9wz4la6"}
export TOKEN_WRAPPER_ENABLED=${TOKEN_WRAPPER_ENABLED:-"true"}

# Transfer amounts
export ZIG_TRANSFER_AMOUNT=${ZIG_TRANSFER_AMOUNT:-"10uzig"}
export AXELAR_TRANSFER_AMOUNT=${AXELAR_TRANSFER_AMOUNT:-"10000000000000unit-zig"}
export COSMOS_TRANSFER_AMOUNT=${COSMOS_TRANSFER_AMOUNT:-"80uatom"}
export DUMMY_TRANSFER_AMOUNT=${DUMMY_TRANSFER_AMOUNT:-"10000000000000unit-zig"}
export MODULE_FUND_AMOUNT=${MODULE_FUND_AMOUNT:-"1000uzig"}

# Git repository
export ZIGCHAIN_REPO_URL=${ZIGCHAIN_REPO_URL:-""}
export ZIGCHAIN_BRANCH=${ZIGCHAIN_BRANCH:-""}
export ZIGCHAIN_LOCAL_PATH=${ZIGCHAIN_LOCAL_PATH:-"/Users/nat/src/zigchain_public/zigchain_private"}

# Timeouts
export CHAIN_START_TIMEOUT=${CHAIN_START_TIMEOUT:-"90"}
export RELAYER_START_TIMEOUT=${RELAYER_START_TIMEOUT:-"10"}
export IBC_TRANSFER_TRANSACTION_TIMEOUT=${IBC_TRANSFER_TRANSACTION_TIMEOUT:-"20"}
export TRANSACTION_TIMEOUT=${TRANSACTION_TIMEOUT:-"10"}

# Timeout test configuration
export TIMEOUT_TEST_ENABLED=${TIMEOUT_TEST_ENABLED:-"true"}
export TIMEOUT_TEST_AMOUNT=${TIMEOUT_TEST_AMOUNT:-"1000000000000unit-zig"}
export TIMEOUT_TEST_NATIVE_AMOUNT=${TIMEOUT_TEST_NATIVE_AMOUNT:-"1uzig"}
export TIMEOUT_TEST_DURATION=${TIMEOUT_TEST_DURATION:-"1000000000"}  # 1 second in nanoseconds

# Ignite
export IGNITE_CMD="ignite"

echo "🚀 Starting localnet environment setup..."
echo "Using configuration:"
echo "  Axelar Chain ID: $AXELAR_CHAIN_ID"
echo "  Zigchain Chain ID: $ZIGCHAIN_CHAIN_ID"
echo "  Dummy Chain ID: $DUMMY_CHAIN_ID"
echo "  Cosmos Chain ID: $COSMOS_CHAIN_ID"
echo "  Axelar RPC Port: $AXELAR_RPC_PORT"
echo "  Zigchain RPC Port: $ZIGCHAIN_RPC_PORT"
echo "  Dummy RPC Port: $DUMMY_RPC_PORT"
echo "  Cosmos RPC Port: $COSMOS_RPC_PORT"
echo "  Zigchain Transfer Amount: $ZIG_TRANSFER_AMOUNT"
echo "  Axelar Transfer Amount: $AXELAR_TRANSFER_AMOUNT"
echo "  Cosmos Transfer Amount: $COSMOS_TRANSFER_AMOUNT"
echo "  Module Fund Amount: $MODULE_FUND_AMOUNT"
echo "  Zigchain Repo URL: $ZIGCHAIN_REPO_URL"
echo "  Zigchain Branch: $ZIGCHAIN_BRANCH"
echo "  Zigchain Local Path: $ZIGCHAIN_LOCAL_PATH"

# Function to check if a command exists
check_command() {
  if ! command -v $1 &>/dev/null; then
    echo "Error: $1 is not installed"
    exit 1
  fi
}

# Check required commands
check_command $IGNITE_CMD
check_command git
check_command jq
check_command bc
check_command osascript

# Set up exit trap to beep on completion
beep_on_exit() {
  local exit_code=$?
  if [[ $exit_code -eq 0 ]]; then
    # Success - beep 5 times
    osascript -e 'beep 5' 2>/dev/null || true
  else
    # Failure - beep 10 times
    osascript -e 'beep 10' 2>/dev/null || true
  fi
  exit $exit_code
}
trap beep_on_exit EXIT

# Clean up previous runs
pkill hermes zigchaind axelard dummyd cosmosd || true

# Create temporary directory for all chains and logs
WORK_DIR=$(mktemp -d)
cd $WORK_DIR

echo "Created working directory: WORK_DIR: $WORK_DIR"

# Create logs directory
mkdir -p logs

# 1. Generate axelar chain
echo "Generating axelar chain..."
$IGNITE_CMD s chain axelar --address-prefix $AXELAR_ADDRESS_PREFIX --no-module

# 2. Create axelar config file
echo "Creating axelar config file..."
cat >$WORK_DIR/axelar/config.yml <<EOL
version: 1
validation: sovereign
accounts:
  - name: alice
    coins:
      - $AXELAR_ALICE_INITIAL_BALANCE
  - name: bob
    mnemonic: "$AXELAR_BOB_MNEMONIC"
    coins:
      - $AXELAR_BOB_INITIAL_BALANCE
  - name: faucet
    coins:
      - $AXELAR_FAUCET_INITIAL_BALANCE
faucet:
  name: faucet
  coins:
    - $AXELAR_FAUCET_AMOUNT
  host: :$AXELAR_FAUCET_PORT
genesis:
  chain_id: $AXELAR_CHAIN_ID
  app_state:
    staking:
      params:
        bond_denom: $AXELAR_DENOM
    mint:
      params:
        mint_denom: $AXELAR_DENOM
    gov:
      params:
        expedited_min_deposit:
          - denom: $AXELAR_DENOM
            amount: "100000000000000"
        min_deposit:
          - denom: $AXELAR_DENOM
            amount: "5000000000000"
validators:
  - name: alice
    bonded: $AXELAR_ALICE_INITIAL_BALANCE
    home: \$HOME/.axelar
    app:
      api:
        address: :$AXELAR_API_PORT
      grpc:
        address: :$AXELAR_GRPC_PORT
      grpc-web:
        address: :$AXELAR_GRPCWEB_PORT
    config:
      p2p:
        laddr: :$AXELAR_P2P_PORT
      rpc:
        laddr: :$AXELAR_RPC_PORT
        pprof_laddr: :$AXELAR_PPROF_PORT
      consensus:
        timeout_commit: 5s
        timeout_propose: 5s
      minimum-gas-prices: 0$AXELAR_DENOM
    client:
      node: http://localhost:$AXELAR_RPC_PORT
EOL

# 3. Clone zigchain repo
echo "Cloning zigchain repository..."
if [ -n "$ZIGCHAIN_LOCAL_PATH" ]; then
  echo "Using local path: $ZIGCHAIN_LOCAL_PATH"
  cp -r "$ZIGCHAIN_LOCAL_PATH" "$WORK_DIR/zigchain_private"
else
  echo "Cloning from repository: $ZIGCHAIN_REPO_URL"
  git clone --branch $ZIGCHAIN_BRANCH $ZIGCHAIN_REPO_URL $WORK_DIR/zigchain_private
fi
cd $WORK_DIR/zigchain_private

# 4. Create zigchain config file
echo "Creating zigchain config file..."
cat >$WORK_DIR/zigchain_private/config.yml <<EOL
version: 1
validation: sovereign
accounts:
  - name: alice
    coins:
      - $ZIG_ALICE_INITIAL_BALANCE
  - name: bob
    mnemonic: "$ZIG_BOB_MNEMONIC"
    coins:
      - $ZIG_BOB_INITIAL_BALANCE
  - name: operator
    mnemonic: "$ZIG_OPERATOR_MNEMONIC"
    coins:
      - $ZIG_OPERATOR_INITIAL_BALANCE
  - name: faucet
    coins:
      - $ZIG_FAUCET_INITIAL_BALANCE
faucet:
  name: faucet
  coins:
    - $ZIG_FAUCET_AMOUNT
  host: :$ZIGCHAIN_FAUCET_PORT
genesis:
  chain_id: $ZIGCHAIN_CHAIN_ID
  app_state:
    staking:
      params:
        bond_denom: $ZIGCHAIN_DENOM
    mint:
      params:
        mint_denom: $ZIGCHAIN_DENOM
    gov:
      params:
        expedited_min_deposit:
          - denom: $ZIGCHAIN_DENOM
            amount: "1000000"
        expedited_voting_period: 1m
        min_deposit:
          - denom: $ZIGCHAIN_DENOM
            amount: "500000"
    tokenwrapper:
      operator_address: "$TOKEN_WRAPPER_OPERATOR_ADDRESS"
      enabled: $TOKEN_WRAPPER_ENABLED
      native_client_id: "07-tendermint-1"
      counterparty_client_id: "07-tendermint-0"
      native_port: "transfer"
      counterparty_port: "transfer"
      native_channel: "channel-1"
      counterparty_channel: "channel-0"
      denom: $AXELAR_DENOM
      decimal_difference: 12

validators:
  - name: alice
    bonded: $ZIG_ALICE_INITIAL_BALANCE
    home: \$HOME/.zigchain
    app:
      api:
        address: :$ZIGCHAIN_API_PORT
      grpc:
        address: :$ZIGCHAIN_GRPC_PORT
      grpc-web:
        address: :$ZIGCHAIN_GRPCWEB_PORT
    config:
      p2p:
        laddr: :$ZIGCHAIN_P2P_PORT
      rpc:
        laddr: :$ZIGCHAIN_RPC_PORT
        pprof_laddr: :$ZIGCHAIN_PPROF_PORT
      consensus:
        timeout_commit: 5s
        timeout_propose: 5s
      minimum-gas-prices: 0$ZIGCHAIN_DENOM
    client:
      node: http://localhost:$ZIGCHAIN_RPC_PORT
EOL

# 5. Generate dummy chain
cd $WORK_DIR
echo "Generating dummy chain..."
$IGNITE_CMD s chain dummy --address-prefix $DUMMY_ADDRESS_PREFIX --no-module

# 6. Create dummy config file
cat >$WORK_DIR/dummy/config.yml <<EOL
version: 1
validation: sovereign
accounts:
  - name: alice
    coins:
      - $DUMMY_ALICE_INITIAL_BALANCE
  - name: bob
    mnemonic: "$DUMMY_BOB_MNEMONIC"
    coins:
      - $DUMMY_BOB_INITIAL_BALANCE
  - name: faucet
    coins:
      - $DUMMY_FAUCET_INITIAL_BALANCE
faucet:
  name: faucet
  coins:
    - $DUMMY_FAUCET_AMOUNT
  host: :$DUMMY_FAUCET_PORT
genesis:
  chain_id: $DUMMY_CHAIN_ID
  app_state:
    staking:
      params:
        bond_denom: $DUMMY_DENOM
    mint:
      params:
        mint_denom: $DUMMY_DENOM
    gov:
      params:
        expedited_min_deposit:
          - denom: $DUMMY_DENOM
            amount: "100000000000000"
        min_deposit:
          - denom: $DUMMY_DENOM
            amount: "5000000000000"
validators:
  - name: alice
    bonded: $DUMMY_ALICE_INITIAL_BALANCE
    home: \$HOME/.dummy
    app:
      api:
        address: :$DUMMY_API_PORT
      grpc:
        address: :$DUMMY_GRPC_PORT
      grpc-web:
        address: :$DUMMY_GRPCWEB_PORT
    config:
      p2p:
        laddr: :$DUMMY_P2P_PORT
      rpc:
        laddr: :$DUMMY_RPC_PORT
        pprof_laddr: :$DUMMY_PPROF_PORT
      consensus:
        timeout_commit: 5s
        timeout_propose: 5s
      minimum-gas-prices: 0$DUMMY_DENOM
    client:
      node: http://localhost:$DUMMY_RPC_PORT
EOL

# 7. Generate cosmos chain
echo "Generating cosmos chain..."
$IGNITE_CMD s chain cosmos --address-prefix $COSMOS_ADDRESS_PREFIX --no-module

# 8. Create cosmos config file
cat >$WORK_DIR/cosmos/config.yml <<EOL
version: 1
validation: sovereign
accounts:
  - name: alice
    coins:
      - $COSMOS_ALICE_INITIAL_BALANCE
  - name: bob
    mnemonic: "$COSMOS_BOB_MNEMONIC"
    coins:
      - $COSMOS_BOB_INITIAL_BALANCE
  - name: faucet
    coins:
      - $COSMOS_FAUCET_INITIAL_BALANCE
faucet:
  name: faucet
  coins:
    - $COSMOS_FAUCET_AMOUNT
  host: :$COSMOS_FAUCET_PORT
genesis:
  chain_id: $COSMOS_CHAIN_ID
  app_state:
    staking:
      params:
        bond_denom: $COSMOS_DENOM
    mint:
      params:
        mint_denom: $COSMOS_DENOM
    gov:
      params:
        expedited_min_deposit:
          - denom: $COSMOS_DENOM
            amount: "100000000000000"
        min_deposit:
          - denom: $COSMOS_DENOM
            amount: "5000000000000"
validators:
  - name: alice
    bonded: $COSMOS_ALICE_INITIAL_BALANCE
    home: \$HOME/.cosmos
    app:
      api:
        address: :$COSMOS_API_PORT
      grpc:
        address: :$COSMOS_GRPC_PORT
      grpc-web:
        address: :$COSMOS_GRPCWEB_PORT
    config:
      p2p:
        laddr: :$COSMOS_P2P_PORT
      rpc:
        laddr: :$COSMOS_RPC_PORT
        pprof_laddr: :$COSMOS_PPROF_PORT
      consensus:
        timeout_commit: 5s
        timeout_propose: 5s
      minimum-gas-prices: 0$COSMOS_DENOM
    client:
      node: http://localhost:$COSMOS_RPC_PORT
EOL

echo "Config files created successfully!"

# 9. Start all chains in background with logging
echo "Starting chains..."
cd $WORK_DIR/axelar && $IGNITE_CMD c serve -r -v > $WORK_DIR/logs/axelar.log 2>&1 &
AXELAR_PID=$!
cd $WORK_DIR/zigchain_private && $IGNITE_CMD c serve -r -v > $WORK_DIR/logs/zigchain.log 2>&1 &
ZIGCHAIN_PID=$!
cd $WORK_DIR/dummy && $IGNITE_CMD c serve -r -v > $WORK_DIR/logs/dummy.log 2>&1 &
DUMMY_PID=$!
cd $WORK_DIR/cosmos && $IGNITE_CMD c serve -r -v > $WORK_DIR/logs/cosmos.log 2>&1 &
COSMOS_PID=$!
cd $WORK_DIR

# Wait for chains to start
echo "Waiting $CHAIN_START_TIMEOUT seconds for chains to start..."
sleep $CHAIN_START_TIMEOUT

# 10. Configure relayer (add dummy chain to relayer config if needed)
echo "Configuring relayer..."
rm -rf ~/.hermes ~/.ignite/relayer

# dummy IBC connections to increment to following IBC ids
$IGNITE_CMD relayer hermes configure \
    "$DUMMY_CHAIN_ID" "http://localhost:$DUMMY_RPC_PORT" "http://localhost:$DUMMY_GRPC_PORT" \
    "$ZIGCHAIN_CHAIN_ID" "http://localhost:$ZIGCHAIN_RPC_PORT" "http://localhost:$ZIGCHAIN_GRPC_PORT" \
    --chain-a-faucet "http://0.0.0.0:$DUMMY_FAUCET_PORT" \
    --chain-b-faucet "http://0.0.0.0:$ZIGCHAIN_FAUCET_PORT" \
    --chain-a-gas-price 0$DUMMY_DENOM \
    --chain-b-gas-price 0$ZIGCHAIN_DENOM \
    --chain-a-account-prefix $DUMMY_ADDRESS_PREFIX \
    --chain-b-account-prefix $ZIGCHAIN_ADDRESS_PREFIX \
    --channel-version "ics20-1" \
    --generate-wallets

$IGNITE_CMD relayer hermes configure \
    "$AXELAR_CHAIN_ID" "http://localhost:$AXELAR_RPC_PORT" "http://localhost:$AXELAR_GRPC_PORT" \
    "$ZIGCHAIN_CHAIN_ID" "http://localhost:$ZIGCHAIN_RPC_PORT" "http://localhost:$ZIGCHAIN_GRPC_PORT" \
    --chain-a-faucet "http://0.0.0.0:$AXELAR_FAUCET_PORT" \
    --chain-b-faucet "http://0.0.0.0:$ZIGCHAIN_FAUCET_PORT" \
    --chain-a-gas-price 0$AXELAR_DENOM \
    --chain-b-gas-price 0$ZIGCHAIN_DENOM \
    --chain-a-account-prefix $AXELAR_ADDRESS_PREFIX \
    --chain-b-account-prefix $ZIGCHAIN_ADDRESS_PREFIX \
    --channel-version "ics20-1" \
    --generate-wallets

$IGNITE_CMD relayer hermes configure \
    "$COSMOS_CHAIN_ID" "http://localhost:$COSMOS_RPC_PORT" "http://localhost:$COSMOS_GRPC_PORT" \
    "$ZIGCHAIN_CHAIN_ID" "http://localhost:$ZIGCHAIN_RPC_PORT" "http://localhost:$ZIGCHAIN_GRPC_PORT" \
    --chain-a-faucet "http://0.0.0.0:$COSMOS_FAUCET_PORT" \
    --chain-b-faucet "http://0.0.0.0:$ZIGCHAIN_FAUCET_PORT" \
    --chain-a-gas-price 0$COSMOS_DENOM \
    --chain-b-gas-price 0$ZIGCHAIN_DENOM \
    --chain-a-account-prefix $COSMOS_ADDRESS_PREFIX \
    --chain-b-account-prefix $ZIGCHAIN_ADDRESS_PREFIX \
    --channel-version "ics20-1" \
    --generate-wallets

# 11. Start relayers in background with logging
echo "Starting relayers..."
$IGNITE_CMD relayer hermes start $DUMMY_CHAIN_ID $ZIGCHAIN_CHAIN_ID > $WORK_DIR/logs/relayer1.log 2>&1 &
RELAYER_PID1=$!
$IGNITE_CMD relayer hermes start $AXELAR_CHAIN_ID $ZIGCHAIN_CHAIN_ID > $WORK_DIR/logs/relayer2.log 2>&1 &
RELAYER_PID2=$!
$IGNITE_CMD relayer hermes start $COSMOS_CHAIN_ID $ZIGCHAIN_CHAIN_ID > $WORK_DIR/logs/relayer3.log 2>&1 &
RELAYER_PID3=$!

# Wait for relayer to start
echo "Waiting $RELAYER_START_TIMEOUT seconds for relayer to start..."
sleep $RELAYER_START_TIMEOUT

echo "✅ Localnet environment setup completed successfully!"
echo "Working directory: $WORK_DIR"
echo "Logs are available in: $WORK_DIR/logs/"
echo "Chains are available in: $WORK_DIR/axelar/ and $WORK_DIR/zigchain_private/"
echo ""

# Create environment file for test script
ENV_FILE="$WORK_DIR/test_env.sh"
cat > "$ENV_FILE" << EOF
#!/bin/bash
# Environment variables for tokenwrapper tests
export WORK_DIR="$WORK_DIR"
export AXELAR_CHAIN_ID="$AXELAR_CHAIN_ID"
export ZIGCHAIN_CHAIN_ID="$ZIGCHAIN_CHAIN_ID"
export DUMMY_CHAIN_ID="$DUMMY_CHAIN_ID"
export COSMOS_CHAIN_ID="$COSMOS_CHAIN_ID"
export AXELAR_ADDRESS_PREFIX="$AXELAR_ADDRESS_PREFIX"
export ZIGCHAIN_ADDRESS_PREFIX="$ZIGCHAIN_ADDRESS_PREFIX"
export DUMMY_ADDRESS_PREFIX="$DUMMY_ADDRESS_PREFIX"
export COSMOS_ADDRESS_PREFIX="$COSMOS_ADDRESS_PREFIX"
export AXELAR_DENOM="$AXELAR_DENOM"
export ZIGCHAIN_DENOM="$ZIGCHAIN_DENOM"
export DUMMY_DENOM="$DUMMY_DENOM"
export COSMOS_DENOM="$COSMOS_DENOM"
export AXELAR_RPC_PORT="$AXELAR_RPC_PORT"
export ZIGCHAIN_RPC_PORT="$ZIGCHAIN_RPC_PORT"
export DUMMY_RPC_PORT="$DUMMY_RPC_PORT"
export COSMOS_RPC_PORT="$COSMOS_RPC_PORT"
export AXELAR_P2P_PORT="$AXELAR_P2P_PORT"
export ZIGCHAIN_P2P_PORT="$ZIGCHAIN_P2P_PORT"
export DUMMY_P2P_PORT="$DUMMY_P2P_PORT"
export COSMOS_P2P_PORT="$COSMOS_P2P_PORT"
export AXELAR_PPROF_PORT="$AXELAR_PPROF_PORT"
export ZIGCHAIN_PPROF_PORT="$ZIGCHAIN_PPROF_PORT"
export DUMMY_PPROF_PORT="$DUMMY_PPROF_PORT"
export COSMOS_PPROF_PORT="$COSMOS_PPROF_PORT"
export AXELAR_GRPC_PORT="$AXELAR_GRPC_PORT"
export ZIGCHAIN_GRPC_PORT="$ZIGCHAIN_GRPC_PORT"
export DUMMY_GRPC_PORT="$DUMMY_GRPC_PORT"
export COSMOS_GRPC_PORT="$COSMOS_GRPC_PORT"
export AXELAR_GRPCWEB_PORT="$AXELAR_GRPCWEB_PORT"
export ZIGCHAIN_GRPCWEB_PORT="$ZIGCHAIN_GRPCWEB_PORT"
export DUMMY_GRPCWEB_PORT="$DUMMY_GRPCWEB_PORT"
export COSMOS_GRPCWEB_PORT="$COSMOS_GRPCWEB_PORT"
export AXELAR_API_PORT="$AXELAR_API_PORT"
export ZIGCHAIN_API_PORT="$ZIGCHAIN_API_PORT"
export DUMMY_API_PORT="$DUMMY_API_PORT"
export COSMOS_API_PORT="$COSMOS_API_PORT"
export AXELAR_FAUCET_PORT="$AXELAR_FAUCET_PORT"
export ZIGCHAIN_FAUCET_PORT="$ZIGCHAIN_FAUCET_PORT"
export DUMMY_FAUCET_PORT="$DUMMY_FAUCET_PORT"
export COSMOS_FAUCET_PORT="$COSMOS_FAUCET_PORT"
export ZIG_ALICE_INITIAL_BALANCE="$ZIG_ALICE_INITIAL_BALANCE"
export ZIG_BOB_INITIAL_BALANCE="$ZIG_BOB_INITIAL_BALANCE"
export ZIG_FAUCET_INITIAL_BALANCE="$ZIG_FAUCET_INITIAL_BALANCE"
export ZIG_FAUCET_AMOUNT="$ZIG_FAUCET_AMOUNT"
export ZIG_BOB_MNEMONIC="$ZIG_BOB_MNEMONIC"
export AXELAR_ALICE_INITIAL_BALANCE="$AXELAR_ALICE_INITIAL_BALANCE"
export AXELAR_BOB_INITIAL_BALANCE="$AXELAR_BOB_INITIAL_BALANCE"
export AXELAR_FAUCET_INITIAL_BALANCE="$AXELAR_FAUCET_INITIAL_BALANCE"
export AXELAR_FAUCET_AMOUNT="$AXELAR_FAUCET_AMOUNT"
export AXELAR_BOB_MNEMONIC="$AXELAR_BOB_MNEMONIC"
export DUMMY_ALICE_INITIAL_BALANCE="$DUMMY_ALICE_INITIAL_BALANCE"
export DUMMY_BOB_INITIAL_BALANCE="$DUMMY_BOB_INITIAL_BALANCE"
export DUMMY_FAUCET_INITIAL_BALANCE="$DUMMY_FAUCET_INITIAL_BALANCE"
export DUMMY_FAUCET_AMOUNT="$DUMMY_FAUCET_AMOUNT"
export DUMMY_BOB_MNEMONIC="$DUMMY_BOB_MNEMONIC"
export COSMOS_ALICE_INITIAL_BALANCE="$COSMOS_ALICE_INITIAL_BALANCE"
export COSMOS_BOB_INITIAL_BALANCE="$COSMOS_BOB_INITIAL_BALANCE"
export COSMOS_FAUCET_INITIAL_BALANCE="$COSMOS_FAUCET_INITIAL_BALANCE"
export COSMOS_FAUCET_AMOUNT="$COSMOS_FAUCET_AMOUNT"
export COSMOS_BOB_MNEMONIC="$COSMOS_BOB_MNEMONIC"
export TOKEN_WRAPPER_OPERATOR_ADDRESS="$TOKEN_WRAPPER_OPERATOR_ADDRESS"
export TOKEN_WRAPPER_ENABLED="$TOKEN_WRAPPER_ENABLED"
export ZIG_TRANSFER_AMOUNT="$ZIG_TRANSFER_AMOUNT"
export AXELAR_TRANSFER_AMOUNT="$AXELAR_TRANSFER_AMOUNT"
export COSMOS_TRANSFER_AMOUNT="$COSMOS_TRANSFER_AMOUNT"
export DUMMY_TRANSFER_AMOUNT="$DUMMY_TRANSFER_AMOUNT"
export MODULE_FUND_AMOUNT="$MODULE_FUND_AMOUNT"
export ZIGCHAIN_REPO_URL="$ZIGCHAIN_REPO_URL"
export ZIGCHAIN_BRANCH="$ZIGCHAIN_BRANCH"
export ZIGCHAIN_LOCAL_PATH="$ZIGCHAIN_LOCAL_PATH"
export CHAIN_START_TIMEOUT="$CHAIN_START_TIMEOUT"
export RELAYER_START_TIMEOUT="$RELAYER_START_TIMEOUT"
export IBC_TRANSFER_TRANSACTION_TIMEOUT="$IBC_TRANSFER_TRANSACTION_TIMEOUT"
export TRANSACTION_TIMEOUT="$TRANSACTION_TIMEOUT"
export TIMEOUT_TEST_ENABLED="$TIMEOUT_TEST_ENABLED"
export TIMEOUT_TEST_AMOUNT="$TIMEOUT_TEST_AMOUNT"
export TIMEOUT_TEST_NATIVE_AMOUNT="$TIMEOUT_TEST_NATIVE_AMOUNT"
export TIMEOUT_TEST_DURATION="$TIMEOUT_TEST_DURATION"
export IGNITE_CMD="$IGNITE_CMD"
export AXELAR_PID="$AXELAR_PID"
export ZIGCHAIN_PID="$ZIGCHAIN_PID"
export DUMMY_PID="$DUMMY_PID"
export COSMOS_PID="$COSMOS_PID"
export RELAYER_PID1="$RELAYER_PID1"
export RELAYER_PID2="$RELAYER_PID2"
export RELAYER_PID3="$RELAYER_PID3"
EOF

echo "Environment file created: $ENV_FILE"
echo ""
echo "🚀 To run tests, execute:"
echo "  source $ENV_FILE && ./x/tokenwrapper/sh/run_tests.sh"
echo ""
echo "To manually check chains:"
echo "  - Axelar: axelard q bank balances bob"
echo "  - ZIGChain: zigchaind q bank balances bob"
echo "  - Dummy: dummyd q bank balances bob"
echo "  - Cosmos: cosmosd q bank balances bob"
echo ""
echo "Service management commands:"
echo "  Note: All chains and relayers are currently running in the background."
echo "  Start all services (chains + relayers):"
echo "    source $ENV_FILE && \\"
echo "    (cd \$WORK_DIR/axelar && axelard start > \$WORK_DIR/logs/axelar.log 2>&1 &) && \\"
echo "    (cd \$WORK_DIR/zigchain_private && zigchaind start > \$WORK_DIR/logs/zigchain.log 2>&1 &) && \\"
echo "    (cd \$WORK_DIR/dummy && dummyd start > \$WORK_DIR/logs/dummy.log 2>&1 &) && \\"
echo "    (cd \$WORK_DIR/cosmos && cosmosd start > \$WORK_DIR/logs/cosmos.log 2>&1 &) && \\"
echo "    sleep 5 && \\"
echo "    ($IGNITE_CMD relayer hermes start $DUMMY_CHAIN_ID $ZIGCHAIN_CHAIN_ID > \$WORK_DIR/logs/relayer1.log 2>&1 &) && \\"
echo "    ($IGNITE_CMD relayer hermes start $AXELAR_CHAIN_ID $ZIGCHAIN_CHAIN_ID > \$WORK_DIR/logs/relayer2.log 2>&1 &) && \\"
echo "    ($IGNITE_CMD relayer hermes start $COSMOS_CHAIN_ID $ZIGCHAIN_CHAIN_ID > \$WORK_DIR/logs/relayer3.log 2>&1 &)"
echo ""
echo "  Stop all services:"
echo "    source $ENV_FILE && \\"
echo "    pkill hermes zigchaind axelard dummyd cosmosd || true"
echo "    echo 'All services stopped'"
