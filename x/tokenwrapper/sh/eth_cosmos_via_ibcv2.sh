#!/bin/bash

# Exit on error
set -e

# Logging function
log() {
    local prefix="✅"
    if [[ "$1" == *"Error"* ]] || [[ "$1" == *"error"* ]]; then
        prefix="❌"
    fi
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $prefix $1"
}

# Function to check if a command exists
check_command() {
    if ! command -v $1 &> /dev/null; then
        log "Error: $1 is not installed"
        exit 1
    fi
}

# Check required commands
check_command cast
check_command zigchaind
check_command jq

log "Starting IBCv2 bridge transfer script..."

# Environment variables with defaults
log "Setting up environment variables..."
export RPC_URL=${RPC_URL:-"https://ethereum-rpc.publicnode.com"}
export PUBLIC_KEY=${PUBLIC_KEY:-""}
export PRIVATE_KEY=${PRIVATE_KEY:-""}
export CONTRACT_ADDRESS=${CONTRACT_ADDRESS:-"0x47a4b9F949E98a49Be500753c19a8f9c9d6b7689"}
export AMOUNT=${AMOUNT:-"526"}
export TOKEN_ADDRESS=${TOKEN_ADDRESS:-"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"}
export RECEIVER_ADDRESS=${RECEIVER_ADDRESS:-""}
export SOURCE_CLIENT=${SOURCE_CLIENT:-"cosmoshub-0"}
export DEST_PORT=${DEST_PORT:-"transfer"}
export TIMEOUT=$(($(date +%s) + 7200))
export MEMO=${MEMO:-"Test transaction"}
export RELAYER_FEE=${RELAYER_FEE:-"200000"}
export RELAYER_ADDRESS=${RELAYER_ADDRESS:-"0x33C4DaD158F1E2cCF97bF17d1574d5b7b9f43002"}

# Validate required variables
log "Validating required variables..."
if [ -z "$PUBLIC_KEY" ]; then
    log "Error: PUBLIC_KEY is not set"
    exit 1
fi

if [ -z "$PRIVATE_KEY" ]; then
    log "Error: PRIVATE_KEY is not set"
    exit 1
fi

if [ -z "$RECEIVER_ADDRESS" ]; then
    log "Error: RECEIVER_ADDRESS is not set"
    exit 1
fi

log "Configuration:"
for var in PUBLIC_KEY PRIVATE_KEY RECEIVER_ADDRESS TOKEN_ADDRESS CONTRACT_ADDRESS AMOUNT SOURCE_CLIENT DEST_PORT TIMEOUT MEMO RELAYER_FEE RELAYER_ADDRESS RPC_URL; do
    log "$var: ${!var}"
done

log "All required variables are set successfully."

log "Approving ERC20 tokens..."

cast send \
    $TOKEN_ADDRESS \
    "approve(address,uint256)(bool)" \
    $CONTRACT_ADDRESS \
    1500000 \
    --rpc-url $RPC_URL \
    --private-key $PRIVATE_KEY \
    --gas-price $(echo "$(cast gas-price --rpc-url $RPC_URL) + 10000000" | bc)

log "Checking allowance..."

cast call \
    $TOKEN_ADDRESS \
    "allowance(address,address)(uint256)" \
    $PUBLIC_KEY \
    $CONTRACT_ADDRESS \
    --rpc-url $RPC_URL

log "Sending ERC20 tokens to Cosmos..."

cast send \
    $CONTRACT_ADDRESS \
    "transfer(uint256,(address,string,string,string,uint64,string),(uint256,address,uint64))" \
    $AMOUNT \
    "($TOKEN_ADDRESS,$RECEIVER_ADDRESS,$SOURCE_CLIENT,$DEST_PORT,$TIMEOUT,$MEMO)" \
    "($RELAYER_FEE,$RELAYER_ADDRESS,$TIMEOUT)" \
    --rpc-url $RPC_URL \
    --private-key $PRIVATE_KEY \
    --gas-price $(echo "$(cast gas-price --rpc-url $RPC_URL) + 10000000" | bc)

log "Transaction submitted successfully"