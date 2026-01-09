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

log "Starting Axelar bridge transfer script..."

# Environment variables with defaults
log "Setting up environment variables..."
export ETH_RPC_URL=${ETH_RPC_URL:-"https://ethereum-sepolia-rpc.publicnode.com"}
export DEPOSIT_SERVICE_URL=${DEPOSIT_SERVICE_URL:-"https://deposit-service.testnet.axelar.dev"}
export ZIG_RPC_URL=${ZIG_RPC_URL:-"https://testnet-rpc.zigchain.com"}
export EVM_CHAIN=${EVM_CHAIN:-"ethereum-sepolia"}
export ZIG_CHAIN=${ZIG_CHAIN:-"zigchain-2"}
export ZIG_ADDRESS=${ZIG_ADDRESS:-""}
export EVM_ADDRESS=${EVM_ADDRESS:-""}
export ZIG_ERC_ADDRESS=${ZIG_ERC_ADDRESS:-"0x94999519B2FeD1bFcf597072A54b33f6b34f6Ca5"}
export EVM_PRIVATE_KEY=${EVM_PRIVATE_KEY:-""}
export ZIG_PRIVATE_KEY=${ZIG_PRIVATE_KEY:-""}
export ZIG_IBC_DENOM=${ZIG_IBC_DENOM:-"ibc/9BEE293E6559ED860CC702685996F394D4991D6DFFD60A19ABC3723E6F34788A"}

# Validate required variables
log "Validating required variables..."
if [ -z "$ZIG_ADDRESS" ]; then
    log "Error: ZIG_ADDRESS is not set"
    exit 1
fi

if [ -z "$EVM_ADDRESS" ]; then
    log "Error: EVM_ADDRESS is not set"
    exit 1
fi

if [ -z "$EVM_PRIVATE_KEY" ]; then
    log "Error: EVM_PRIVATE_KEY is not set"
    exit 1
fi

if [ -z "$ZIG_PRIVATE_KEY" ]; then
    log "Error: ZIG_PRIVATE_KEY is not set"
    exit 1
fi

log "All required variables are set successfully."

# Prompt user for transfer direction
log "Please choose the transfer direction:"
echo "1) Send from Sepolia to zig-test-2"
echo "2) Send from zig-test-2 to Sepolia"
read -p "Enter your choice (1 or 2): " choice

# 1) send from SEPOLIA to ZIGTEST1
if [ "$choice" = "1" ]; then
    log "Selected: Transfer from Sepolia to zig-test-2"
    
    # Ask user about deposit address generation method
    log "Please choose how to generate the deposit address:"
    echo "1) Generate automatically using the script"
    echo "2) Generate manually using satellite.money"
    echo "   URL: https://testnet.satellite.money/?source=ethereum+sepolia&destination=zigchain&asset_denom=unit-zig&destination_address=$ZIG_ADDRESS"
    read -p "Enter your choice (1 or 2): " deposit_choice

    if [ "$deposit_choice" = "1" ]; then
        # a) Generate the EVM deposit address automatically
        log "Generating EVM deposit address automatically..."
        export DEPOSIT_ADDRESS=$(curl -s -X POST -d '
        {
            "source_chain": "'$EVM_CHAIN'",
            "destination_chain": "'$ZIG_CHAIN'",
            "destination_address": "'$ZIG_ADDRESS'",
            "refund_address": "'$EVM_ADDRESS'"
        }' $DEPOSIT_SERVICE_URL/deposit/wrap | jq -r '.address')
        log "Generated deposit address: $DEPOSIT_ADDRESS"
    elif [ "$deposit_choice" = "2" ]; then
        log "Please generate the deposit address using the provided satellite.money URL"
        log "Once you have the deposit address, please enter it below:"
        read -p "Enter the deposit address: " DEPOSIT_ADDRESS
        log "Using manually entered deposit address: $DEPOSIT_ADDRESS"
    else
        log "Error: Invalid choice selected. Please enter 1 or 2."
        exit 1
    fi

    # b) Send the tokens to the EVM deposit address
    log "Initiating token transfer to deposit address..."
    cast send $ZIG_ERC_ADDRESS \
      "transfer(address,uint256)" \
      $DEPOSIT_ADDRESS \
      1000000000000000000 \
      --rpc-url $ETH_RPC_URL \
      --private-key $EVM_PRIVATE_KEY \
      --gas-price $(cast gas-price --rpc-url $ETH_RPC_URL)
    log "Token transfer transaction submitted successfully"

# 2) send from zig-test-2 to SEPOLIA
elif [ "$choice" = "2" ]; then
    log "Selected: Transfer from zig-test-2 to Sepolia"
    
    # Ask user about deposit address generation method
    log "Please choose how to generate the deposit address:"
    echo "1) Generate automatically using the script"
    echo "2) Generate manually using satellite.money"
    echo "   URL: https://testnet.satellite.money/?source=zigchain&destination=ethereum+sepolia&asset_denom=unit-zig&destination_address=$EVM_ADDRESS"
    read -p "Enter your choice (1 or 2): " deposit_choice

    if [ "$deposit_choice" = "1" ]; then
        # a) Generate the AXELAR deposit address automatically
        log "Generating AXELAR deposit address automatically..."
        export DEPOSIT_ADDRESS=$(curl -s -X POST -d '
        {
            "source_chain": "'$ZIG_CHAIN'",
            "destination_chain": "'$EVM_CHAIN'",
            "destination_address": "'$EVM_ADDRESS'",
            "refund_address": "'$EVM_ADDRESS'"
        }' $DEPOSIT_SERVICE_URL/deposit/unwrap | jq -r '.address')
        log "Generated deposit address: $DEPOSIT_ADDRESS"
    elif [ "$deposit_choice" = "2" ]; then
        log "Please generate the deposit address using the provided satellite.money URL"
        log "Once you have the deposit address, please enter it below:"
        read -p "Enter the deposit address: " DEPOSIT_ADDRESS
        log "Using manually entered deposit address: $DEPOSIT_ADDRESS"
    else
        log "Error: Invalid choice selected. Please enter 1 or 2."
        exit 1
    fi

    # b) Import the ZIG private key
    log "Importing ZIG private key..."
    zigchaind keys import-hex myaccount $ZIG_PRIVATE_KEY
    log "Private key imported successfully"

    # c) Send the tokens to the AXELAR deposit address
    log "Initiating IBC transfer..."
    zigchaind tx ibc-transfer transfer \
        transfer \
        channel-0 \
        $DEPOSIT_ADDRESS \
        1000000000000000000$ZIG_IBC_DENOM \
        --from myaccount \
        --node $ZIG_RPC_URL \
        --fees 50uzig \
        --chain-id zig-test-2 \
        -y
    log "IBC transfer transaction submitted successfully"
else
    log "Error: Invalid choice selected. Please enter 1 or 2."
    exit 1
fi

log "Script execution completed successfully."