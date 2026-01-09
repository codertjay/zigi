#!/bin/bash

# IBC TokenWrapper Test Runner Script
#
# This script runs tests against the localnet environment to verify
# the tokenwrapper module's IBC functionality.

# Exit on error
set -e

# Source common utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common_test_utils.sh"

# Check required commands
check_command jq
check_command bc
check_command axelard
check_command zigchaind
check_command dummyd
check_command ignite
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

# Parse arguments
parse_test_args "$@"

# Check if WORK_DIR is set (indicating setup has been run)
if [ -z "$WORK_DIR" ]; then
    echo "❌ Error: WORK_DIR environment variable not set."
    echo "Please run setup_localnet.sh first, then source the generated environment file."
    echo "Example:"
    echo "  ./x/tokenwrapper/sh/setup_localnet.sh"
    echo "  source /tmp/tmp.XXX/test_env.sh  # (path shown in setup output)"
    echo "  ./x/tokenwrapper/sh/run_tests.sh"
    echo ""
    echo "Usage: ./run_tests.sh [TEST_ID...]"
    echo "  TEST_ID: Test ID(s) to run (e.g., TW-002, TW-003, TW-002-TW-005)"
    echo "  Note: Some tests (${SKIP_BY_DEFAULT_TESTS[*]}) only run when explicitly specified"
    echo "  Examples:"
    echo "    ./run_tests.sh                    # Run all tests except skip-by-default ones"
    echo "    ./run_tests.sh TW-002             # Run only test TW-002"
    echo "    ./run_tests.sh TW-002 TW-003      # Run tests TW-002 and TW-003"
    echo "    ./run_tests.sh TW-002-TW-005      # Run tests TW-002 through TW-005"
    echo "    ./run_tests.sh TW-999              # Run only skip-by-default test TW-999"
    exit 1
fi

echo "🚀 Starting tokenwrapper IBC tests..."
echo "Using working directory: $WORK_DIR"

if [ "$RUN_SPECIFIC_TESTS" = true ]; then
    echo "Running specific tests: ${SPECIFIED_TESTS[*]}"
else
    echo "Running all tests (${SKIP_BY_DEFAULT_TESTS[*]} tests are skipped by default)"
fi

# Navigate to working directory
if [ -n "$WORK_DIR" ]; then
    cd "$WORK_DIR"
fi

# Test function definitions
test_001_module_info() {
    echo '🔍 Checking token wrapper module info...'

    # Get module info using common utility
    MODULE_INFO_OUTPUT=$(get_module_info)

    # Extract expected values (with defaults)
    EXPECTED_DECIMAL_DIFFERENCE=${EXPECTED_DECIMAL_DIFFERENCE:-'12'}
    EXPECTED_DENOM=${EXPECTED_DENOM:-$AXELAR_DENOM}
    EXPECTED_NATIVE_CHANNEL=${EXPECTED_NATIVE_CHANNEL:-'channel-1'}
    EXPECTED_NATIVE_PORT=${EXPECTED_NATIVE_PORT:-'transfer'}
    EXPECTED_NATIVE_CLIENT_ID=${EXPECTED_NATIVE_CLIENT_ID:-'07-tendermint-1'}
    EXPECTED_COUNTERPARTY_CHANNEL=${EXPECTED_COUNTERPARTY_CHANNEL:-'channel-0'}
    EXPECTED_COUNTERPARTY_PORT=${EXPECTED_COUNTERPARTY_PORT:-'transfer'}
    EXPECTED_COUNTERPARTY_CLIENT_ID=${EXPECTED_COUNTERPARTY_CLIENT_ID:-'07-tendermint-0'}
    EXPECTED_ENABLE=${EXPECTED_ENABLE:-'true'}
    EXPECTED_OPERATOR_ADDRESS=${EXPECTED_OPERATOR_ADDRESS:-'zig199d3ngzyz8up8wm4605wtadnnj44chn9wz4la6'}
    EXPECTED_PROPOSED_OPERATOR_ADDRESS=${EXPECTED_PROPOSED_OPERATOR_ADDRESS:-'zig1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqk9whvl'}
    EXPECTED_MODULE_ADDRESS=${EXPECTED_MODULE_ADDRESS:-'zig1hdq87rzf327fwz8rw9rnmchj7qa3uxrpxds2fw'}

    # Extract actual values using common utility functions
    ACTUAL_DECIMAL_DIFFERENCE=$(echo $MODULE_INFO_OUTPUT | jq -r '.decimal_difference')
    ACTUAL_DENOM=$(echo $MODULE_INFO_OUTPUT | jq -r '.denom')
    ACTUAL_NATIVE_CHANNEL=$(echo $MODULE_INFO_OUTPUT | jq -r '.native_channel')
    ACTUAL_NATIVE_PORT=$(echo $MODULE_INFO_OUTPUT | jq -r '.native_port')
    ACTUAL_NATIVE_CLIENT_ID=$(echo $MODULE_INFO_OUTPUT | jq -r '.native_client_id')
    ACTUAL_COUNTERPARTY_CHANNEL=$(echo $MODULE_INFO_OUTPUT | jq -r '.counterparty_channel')
    ACTUAL_COUNTERPARTY_PORT=$(echo $MODULE_INFO_OUTPUT | jq -r '.counterparty_port')
    ACTUAL_COUNTERPARTY_CLIENT_ID=$(echo $MODULE_INFO_OUTPUT | jq -r '.counterparty_client_id')
    ACTUAL_ENABLE=$(echo $MODULE_INFO_OUTPUT | jq -r '.token_wrapper_enabled // "false"')
    ACTUAL_OPERATOR_ADDRESS=$(echo $MODULE_INFO_OUTPUT | jq -r '.operator_address')
    ACTUAL_PROPOSED_OPERATOR_ADDRESS=$(echo $MODULE_INFO_OUTPUT | jq -r '.proposed_operator_address // "zig1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqk9whvl"')
    ACTUAL_MODULE_ADDRESS=$(echo $MODULE_INFO_OUTPUT | jq -r '.module_address')

    # Printing actual vs expected values for debugging
    echo "🔍 Module Info:"
    echo "  Decimal Difference: Expected=$EXPECTED_DECIMAL_DIFFERENCE, Actual=$ACTUAL_DECIMAL_DIFFERENCE"
    echo "  Denom: Expected=$EXPECTED_DENOM, Actual=$ACTUAL_DENOM"
    echo "  Native Channel: Expected=$EXPECTED_NATIVE_CHANNEL, Actual=$ACTUAL_NATIVE_CHANNEL"
    echo "  Native Port: Expected=$EXPECTED_NATIVE_PORT, Actual=$ACTUAL_NATIVE_PORT"
    echo "  Native Client ID: Expected=$EXPECTED_NATIVE_CLIENT_ID, Actual=$ACTUAL_NATIVE_CLIENT_ID"
    echo "  Counterparty Channel: Expected=$EXPECTED_COUNTERPARTY_CHANNEL, Actual=$ACTUAL_COUNTERPARTY_CHANNEL"
    echo "  Counterparty Port: Expected=$EXPECTED_COUNTERPARTY_PORT, Actual=$ACTUAL_COUNTERPARTY_PORT"
    echo "  Counterparty Client ID: Expected=$EXPECTED_COUNTERPARTY_CLIENT_ID, Actual=$ACTUAL_COUNTERPARTY_CLIENT_ID"
    echo "  Token Wrapper Enabled: Expected=$EXPECTED_ENABLE, Actual=$ACTUAL_ENABLE"
    echo "  Operator Address: Expected=$EXPECTED_OPERATOR_ADDRESS, Actual=$ACTUAL_OPERATOR_ADDRESS"
    echo "  Proposed Operator Address: Expected=$EXPECTED_PROPOSED_OPERATOR_ADDRESS, Actual=$ACTUAL_PROPOSED_OPERATOR_ADDRESS"
    echo "  Module Address: Expected=$EXPECTED_MODULE_ADDRESS, Actual=$ACTUAL_MODULE_ADDRESS"
    echo ""


    # Validate all fields match expected values
    [ "$ACTUAL_DECIMAL_DIFFERENCE" = "$EXPECTED_DECIMAL_DIFFERENCE" ] && \
    [ "$ACTUAL_DENOM" = "$EXPECTED_DENOM" ] && \
    [ "$ACTUAL_NATIVE_CHANNEL" = "$EXPECTED_NATIVE_CHANNEL" ] && \
    [ "$ACTUAL_NATIVE_PORT" = "$EXPECTED_NATIVE_PORT" ] && \
    [ "$ACTUAL_NATIVE_CLIENT_ID" = "$EXPECTED_NATIVE_CLIENT_ID" ] && \
    [ "$ACTUAL_COUNTERPARTY_CHANNEL" = "$EXPECTED_COUNTERPARTY_CHANNEL" ] && \
    [ "$ACTUAL_COUNTERPARTY_PORT" = "$EXPECTED_COUNTERPARTY_PORT" ] && \
    [ "$ACTUAL_COUNTERPARTY_CLIENT_ID" = "$EXPECTED_COUNTERPARTY_CLIENT_ID" ] && \
    [ "$ACTUAL_ENABLE" = "$EXPECTED_ENABLE" ] && \
    [ "$ACTUAL_OPERATOR_ADDRESS" = "$EXPECTED_OPERATOR_ADDRESS" ] && \
    [ "$ACTUAL_PROPOSED_OPERATOR_ADDRESS" = "$EXPECTED_PROPOSED_OPERATOR_ADDRESS" ] && \
    [ "$ACTUAL_MODULE_ADDRESS" = "$EXPECTED_MODULE_ADDRESS" ]
}

# Test 001: Check token wrapper module info and initial state
run_test "TW-001" "Verify token wrapper module configuration and initial state" "test_001_module_info"

test_002_unfunded_transfer() {
    echo '🔄 Testing IBC transfer when module is not funded...'

    # Setup test environment
    setup_test_env

    # Check module wallet balances using common utilities
    echo '🔍 Checking module wallet balances...'
    MODULE_INFO=$(get_module_info)
    HAS_BALANCES=$(echo $MODULE_INFO | jq 'has("balances")')

    # If EXPECTED_BALANCES is set to true and module has balances, test succeeds immediately
    if [[ "${EXPECTED_BALANCES:-false}" == 'true' && $HAS_BALANCES == 'true' ]]; then
        echo '✅ Module wallet has expected balances therefore no IBC vouchers should be received - test passes'
        return 0
    fi

    if [[ $HAS_BALANCES != 'false' ]]; then
        echo "❌ Test failed: Module wallet should have no balances for this test"
        return 1
    fi

    # Validate accounts exist
    if ! validate_accounts "$TEST_ACCOUNT_NAME"; then
        return 1
    fi

    # Get account addresses
    ZIGCHAIN_TEST_ADDRESS=$(zigchaind keys show $TEST_ACCOUNT_NAME -a)
    AXELAR_TEST_ADDRESS=$(axelard keys show $TEST_ACCOUNT_NAME -a 2>/dev/null)

    echo "📍 Using $TEST_ACCOUNT_NAME account on zigchain: $ZIGCHAIN_TEST_ADDRESS"
    echo "📍 Using $TEST_ACCOUNT_NAME account on axelar: $AXELAR_TEST_ADDRESS"

    # Get initial balances and transfer stats before the transfer
    echo '📊 Getting initial balances and transfer statistics...'
    INITIAL_ZIGCHAIN_BALANCES=$(get_zigchain_balance "$ZIGCHAIN_TEST_ADDRESS")
    INITIAL_AXELAR_BALANCES=$(get_axelar_balance "$AXELAR_TEST_ADDRESS")
    INITIAL_TRANSFERRED_IN=$(get_total_transferred_in)
    INITIAL_TRANSFERRED_OUT=$(get_total_transferred_out)
    echo "📈 Initial transfer stats - In: $INITIAL_TRANSFERRED_IN, Out: $INITIAL_TRANSFERRED_OUT"

    echo "📊 Initial ZIGChain $TEST_ACCOUNT_NAME balances:"
    echo "$INITIAL_ZIGCHAIN_BALANCES" | jq -r '.balances[] | "  \(.amount) \(.denom)"'
    echo "📊 Initial Axelar $TEST_ACCOUNT_NAME balances:"
    echo "$INITIAL_AXELAR_BALANCES" | jq -r '.balances[] | "  \(.amount) \(.denom)"'

    # Validate sufficient balance for transfer using common utility
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    if ! validate_sufficient_balance "axelar" "$TEST_ACCOUNT_NAME" "$TRANSFER_AMOUNT_NUMERIC" "unit-zig"; then
        return 1
    fi

    # Perform IBC transfer from axelar to zigchain
    echo '🚀 Performing IBC transfer from axelar to zigchain...'
    TRANSFER_CMD="axelard tx ibc-transfer transfer transfer $AXELAR_ZIGCHAIN_CHANNEL_ID $ZIGCHAIN_TEST_ADDRESS $AXELAR_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $AXELAR_FEES --chain-id $AXELAR_CHAIN_ID -y -o json 2>/dev/null"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "IBC transfer from Axelar to ZIGChain")

    # Wait for ibc transfer transaction to be processed
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"

    # Get final balances and transfer stats after the transfer
    echo '📊 Getting final balances and transfer statistics...'
    FINAL_ZIGCHAIN_BALANCES=$(get_zigchain_balance "$ZIGCHAIN_TEST_ADDRESS")
    FINAL_AXELAR_BALANCES=$(get_axelar_balance "$AXELAR_TEST_ADDRESS")
    FINAL_TRANSFERRED_IN=$(get_total_transferred_in)
    FINAL_TRANSFERRED_OUT=$(get_total_transferred_out)
    echo "📈 Final transfer stats - In: $FINAL_TRANSFERRED_IN, Out: $FINAL_TRANSFERRED_OUT"

    echo "📊 Final ZIGChain $TEST_ACCOUNT_NAME balances:"
    echo "$FINAL_ZIGCHAIN_BALANCES" | jq -r '.balances[] | "  \(.amount) \(.denom)"'
    echo "📊 Final Axelar $TEST_ACCOUNT_NAME balances:"
    echo "$FINAL_AXELAR_BALANCES" | jq -r '.balances[] | "  \(.amount) \(.denom)"'

    # Verify balance changes
    echo '🔍 Verifying balance changes...'

    # Check axelar balance decreased by the transfer amount
    INITIAL_AXELAR_AMOUNT=$(extract_balance_amount "$INITIAL_AXELAR_BALANCES" "unit-zig")
    FINAL_AXELAR_AMOUNT=$(extract_balance_amount "$FINAL_AXELAR_BALANCES" "unit-zig")
    EXPECTED_AXELAR_AMOUNT=$(echo "$INITIAL_AXELAR_AMOUNT - $TRANSFER_AMOUNT_NUMERIC" | bc)

    # Check that zigchain received IBC vouchers (since module is unfunded)
    TRANSFER_AMOUNT='10000000000000'
    EXPECTED_ZIGCHAIN_IBC_DENOM=$(get_ibc_hash)
    INITIAL_ZIGCHAIN_IBC_AMOUNT=$(extract_balance_amount "$INITIAL_ZIGCHAIN_BALANCES" "$EXPECTED_ZIGCHAIN_IBC_DENOM")
    INITIAL_ZIGCHAIN_IBC_AMOUNT=${INITIAL_ZIGCHAIN_IBC_AMOUNT:-0}
    EXPECTED_ZIGCHAIN_IBC_BALANCE_AMOUNT=$(echo "$INITIAL_ZIGCHAIN_IBC_AMOUNT + $TRANSFER_AMOUNT" | bc)

    # Log balance comparisons
    log_balance_comparison "Axelar $TEST_ACCOUNT_NAME" "$INITIAL_AXELAR_AMOUNT" "$FINAL_AXELAR_AMOUNT" "$EXPECTED_AXELAR_AMOUNT"
    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME IBC" "$INITIAL_ZIGCHAIN_IBC_AMOUNT" "$(extract_balance_amount "$FINAL_ZIGCHAIN_BALANCES" "$EXPECTED_ZIGCHAIN_IBC_DENOM")" "$EXPECTED_ZIGCHAIN_IBC_BALANCE_AMOUNT"

    # Verify the changes (transfer statistics should remain unchanged for unfunded transfer)
    [[ "$FINAL_AXELAR_AMOUNT" == "$EXPECTED_AXELAR_AMOUNT" ]] && \
    [[ $(extract_balance_amount "$FINAL_ZIGCHAIN_BALANCES" "$EXPECTED_ZIGCHAIN_IBC_DENOM") == "$EXPECTED_ZIGCHAIN_IBC_BALANCE_AMOUNT" ]] && \
    [[ "$FINAL_TRANSFERRED_IN" == "$INITIAL_TRANSFERRED_IN" ]] && \
    [[ "$FINAL_TRANSFERRED_OUT" == "$INITIAL_TRANSFERRED_OUT" ]]
}

# Test 002: Send tokens from axelar to zigchain when token wrapper is not funded
run_test "TW-002" "IBC transfer when token wrapper is unfunded (should receive IBC vouchers)" "test_002_unfunded_transfer"

test_003_fund_module() {
    echo '💰 Funding module wallet with native tokens...'

    # Setup test environment
    setup_test_env

    # Get initial module balances using common utilities
    echo '📊 Retrieving initial module balances...'
    INITIAL_UZIG_BALANCE=$(get_module_uzig_balance)
    echo "📈 Initial uzig balance: $INITIAL_UZIG_BALANCE"

    # Validate operator account exists
    if ! validate_zigchain_account "$OP_ACCOUNT_NAME"; then
        return 1
    fi

    # Validate module state
    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    # Fund module address with native tokens
    echo '🚀 Funding module address with native tokens...'
    FUND_CMD="zigchaind tx tokenwrapper fund-module-wallet $MODULE_FUND_AMOUNT --from $OP_ACCOUNT_NAME --chain-id $ZIGCHAIN_CHAIN_ID --fees $ZIGCHAIN_FEES -y -o json"
    OUTPUT=$(execute_tx "$FUND_CMD" "Fund module wallet")

    # Wait for transaction to be processed
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify module info and compare balances using common utilities
    echo '📊 Module address balance after funding:'
    FINAL_UZIG_BALANCE=$(get_module_uzig_balance)

    echo "📈 Final uzig balance: $FINAL_UZIG_BALANCE"

    # Calculate expected final balance using common utilities
    FUND_AMOUNT_NUMERIC=$(amount_to_numeric "$MODULE_FUND_AMOUNT")
    EXPECTED_FINAL_UZIG_BALANCE=$(echo "$INITIAL_UZIG_BALANCE + $FUND_AMOUNT_NUMERIC" | bc)

    # Log balance comparison
    log_balance_comparison "Module uzig" "$INITIAL_UZIG_BALANCE" "$FINAL_UZIG_BALANCE" "$EXPECTED_FINAL_UZIG_BALANCE"

    # Validate funding was successful
    [[ "$FINAL_UZIG_BALANCE" == "$EXPECTED_FINAL_UZIG_BALANCE" ]]
}

# Test 003: Fund module address
run_test "TW-003" "Fund token wrapper module wallet" "test_003_fund_module"

test_004_funded_transfer() {
    echo '🚀 Testing IBC transfer from Axelar to ZIGChain when module is funded...'

    # Setup test environment
    setup_test_env

    # Pre-transfer validation
    if ! pre_transfer_validation "Axelar → ZIGChain (funded)"; then
        return 1
    fi

    # Verify module is funded before proceeding using common utilities
    echo '🔍 Verifying module is funded...'
    CURRENT_UZIG_BALANCE=$(get_module_uzig_balance)
    TOKEN_WRAPPER_ENABLED=$(is_token_wrapper_enabled)
    EXPECTED_FUND_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")

    if ! validate_module_uzig_balance "$EXPECTED_FUND_AMOUNT" || ! $TOKEN_WRAPPER_ENABLED; then
        echo "❌ Module not properly funded or token wrapper not enabled"
        return 1
    fi

    # Capture balances and transfer stats BEFORE transfer
    echo '📊 Capturing balances and transfer statistics before transfer...'

    # Get balances using common utilities
    ZIGCHAIN_BOB_UZIG_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_BEFORE=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "unit-zig")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance before: $ZIGCHAIN_BOB_UZIG_BEFORE"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance before: $AXELAR_BOB_AMOUNT_BEFORE unit-zig"

    # Validate sufficient balance for transfer using common utility
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    if ! validate_sufficient_balance "axelar" "$TEST_ACCOUNT_NAME" "$TRANSFER_AMOUNT_NUMERIC" "unit-zig"; then
        if [[ "$ZIGCHAIN_NODE" == *"testnet"* || "$AXELAR_NODE" == *"testnet"* ]]; then
            echo ""
            echo "💡 TIP: If you are using TestNet, you can transfer uzig from your ZIGChain account to your Axelar account to receive unit-zig on Axelar."
            echo "      Example:"
            echo "        zigchaind tx ibc-transfer transfer transfer channel-0 $(axelard keys show $TEST_ACCOUNT_NAME -a 2>/dev/null) $AXELAR_TRANSFER_AMOUNT uzig --from $TEST_ACCOUNT_NAME -y"
            echo ""
        fi
        return 1
    fi

    # Get module balances before transfer using common utilities
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)
    IBC_HASH=$(get_ibc_hash)
    MODULE_IBC_BEFORE=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance before: $MODULE_UZIG_BEFORE"
    echo "📈 Module IBC balance before: $MODULE_IBC_BEFORE"

    # Get transfer statistics before transfer using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # Send tokens from axelar to zigchain
    echo '🚀 Sending tokens from axelar to zigchain...'
    ZIGCHAIN_TEST_ADDRESS=$(zigchaind keys show $TEST_ACCOUNT_NAME -a)
    TRANSFER_CMD="axelard tx ibc-transfer transfer transfer $AXELAR_ZIGCHAIN_CHANNEL_ID $ZIGCHAIN_TEST_ADDRESS $AXELAR_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $AXELAR_FEES --chain-id $AXELAR_CHAIN_ID -y -o json 2>/dev/null"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "IBC transfer from Axelar to ZIGChain")

    # Wait for ibc transfer transaction to be processed
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"

    # Capture balances and transfer stats AFTER transfer
    echo '📊 Capturing balances and transfer statistics after transfer...'

    # Get balances after transfer using common utilities
    ZIGCHAIN_BOB_UZIG_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_AFTER=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "unit-zig")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance after: $ZIGCHAIN_BOB_UZIG_AFTER"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance after: $AXELAR_BOB_AMOUNT_AFTER unit-zig"

    # Get module balances after transfer using common utilities
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)
    MODULE_IBC_AFTER=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance after: $MODULE_UZIG_AFTER"
    echo "📈 Module IBC balance after: $MODULE_IBC_AFTER"

    # Get transfer statistics after transfer using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    echo "📊 Transfer stats after: In=$TRANSFERRED_IN_AFTER, Out=$TRANSFERRED_OUT_AFTER"

    # Calculate expected values using common utilities
    EXPECTED_ZIGCHAIN_BOB_UZIG=$(echo "$ZIGCHAIN_BOB_UZIG_BEFORE + 10" | bc)
    EXPECTED_AXELAR_BOB_AMOUNT=$(echo "$AXELAR_BOB_AMOUNT_BEFORE - $TRANSFER_AMOUNT_NUMERIC" | bc)
    EXPECTED_MODULE_UZIG=$(echo "$MODULE_UZIG_BEFORE - 10" | bc)
    EXPECTED_MODULE_IBC=$(echo "$MODULE_IBC_BEFORE + $TRANSFER_AMOUNT_NUMERIC" | bc)
    EXPECTED_TRANSFERRED_IN=$(echo "$TRANSFERRED_IN_BEFORE + 10" | bc)
    EXPECTED_TRANSFERRED_OUT="$TRANSFERRED_OUT_BEFORE"

    # Log balance comparisons using common utility
    echo '🔍 Validating balance changes...'
    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME uzig" "$ZIGCHAIN_BOB_UZIG_BEFORE" "$ZIGCHAIN_BOB_UZIG_AFTER" "$EXPECTED_ZIGCHAIN_BOB_UZIG"
    log_balance_comparison "Axelar $TEST_ACCOUNT_NAME" "$AXELAR_BOB_AMOUNT_BEFORE" "$AXELAR_BOB_AMOUNT_AFTER" "$EXPECTED_AXELAR_BOB_AMOUNT"
    log_balance_comparison "Module uzig" "$MODULE_UZIG_BEFORE" "$MODULE_UZIG_AFTER" "$EXPECTED_MODULE_UZIG"
    log_balance_comparison "Module IBC" "$MODULE_IBC_BEFORE" "$MODULE_IBC_AFTER" "$EXPECTED_MODULE_IBC"
    log_balance_comparison "Transferred in" "$TRANSFERRED_IN_BEFORE" "$TRANSFERRED_IN_AFTER" "$EXPECTED_TRANSFERRED_IN"
    log_balance_comparison "Transferred out" "$TRANSFERRED_OUT_BEFORE" "$TRANSFERRED_OUT_AFTER" "$EXPECTED_TRANSFERRED_OUT"

    # Validate all balance changes and transfer statistics
    [[ "$ZIGCHAIN_BOB_UZIG_AFTER" == "$EXPECTED_ZIGCHAIN_BOB_UZIG" ]] && \
    [[ "$AXELAR_BOB_AMOUNT_AFTER" == "$EXPECTED_AXELAR_BOB_AMOUNT" ]] && \
    [[ "$MODULE_UZIG_AFTER" == "$EXPECTED_MODULE_UZIG" ]] && \
    [[ "$MODULE_IBC_AFTER" == "$EXPECTED_MODULE_IBC" ]] && \
    [[ "$TRANSFERRED_IN_AFTER" == "$EXPECTED_TRANSFERRED_IN" ]] && \
    [[ "$TRANSFERRED_OUT_AFTER" == "$EXPECTED_TRANSFERRED_OUT" ]]
}

# Test 004: Axelar → ZIGChain IBC transfer when Token Wrapper is funded
run_test "TW-004" "IBC transfer from Axelar to ZIGChain when token wrapper is funded (should receive native tokens)" "test_004_funded_transfer"

test_005_return_transfer() {
    echo '↩️ Testing return transfer from ZIGChain to Axelar...'

    # Setup test environment
    setup_test_env

    # Pre-transfer validation
    if ! pre_transfer_validation "ZIGChain → Axelar (return)"; then
        return 1
    fi

    # Verify module is funded with sufficient uzig tokens using common utilities
    echo '🔍 Verifying module wallet has sufficient tokens for return transfer...'
    REQUIRED_UZIG_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    REQUIRED_IBC_AMOUNT=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    IBC_HASH=$(get_ibc_hash)

    if ! validate_module_uzig_balance "$REQUIRED_UZIG_AMOUNT"; then
        return 1
    fi

    # Verify module has sufficient IBC tokens
    MODULE_IBC_BALANCE=$(get_module_ibc_balance "$IBC_HASH")
    if ! [[ $(echo "$MODULE_IBC_BALANCE >= $REQUIRED_IBC_AMOUNT" | bc -l) -eq 1 ]]; then
        echo "❌ Insufficient IBC balance in module wallet."
        echo "   Required: >= $REQUIRED_IBC_AMOUNT IBC tokens"
        echo "   Current: $MODULE_IBC_BALANCE IBC tokens"
        return 1
    fi

    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before return transfer...'
    ZIGCHAIN_BOB_UZIG_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_BEFORE=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "unit-zig")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance before: $ZIGCHAIN_BOB_UZIG_BEFORE"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance before: $AXELAR_BOB_AMOUNT_BEFORE unit-zig"

    # Verify account has sufficient uzig balance for return transfer (including fees)
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    FEES_AMOUNT_NUMERIC=$(get_fee_numeric "$ZIGCHAIN_FEES")
    REQUIRED_ZIG_AMOUNT=$(echo "$TRANSFER_AMOUNT_NUMERIC + $FEES_AMOUNT_NUMERIC" | bc)

    if ! validate_sufficient_balance "zigchain" "$TEST_ACCOUNT_NAME" "$REQUIRED_ZIG_AMOUNT" "uzig"; then
        # Check if running on TestNet and provide faucet information
        if [[ "$ZIGCHAIN_NODE" == *"testnet"* || "$AXELAR_NODE" == *"testnet"* ]]; then
            echo "💡 TIP: You can request ZIG tokens from the faucet available at https://faucet.zigchain.com/"
        fi
        return 1
    fi

    # Get module balances before transfer using common utilities
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)
    MODULE_IBC_BEFORE=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance before: $MODULE_UZIG_BEFORE"
    echo "📈 Module IBC balance before: $MODULE_IBC_BEFORE"

    # Get transfer statistics before transfer using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # Send tokens back from ZIGChain to Axelar
    echo '🚀 Sending tokens back from ZIGChain to Axelar...'
    AXELAR_TEST_ADDRESS=$(axelard keys show $TEST_ACCOUNT_NAME -a 2>/dev/null)
    TRANSFER_CMD="zigchaind tx ibc-transfer transfer transfer $ZIGCHAIN_AXELAR_CHANNEL_ID $AXELAR_TEST_ADDRESS $ZIG_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "Return transfer from ZIGChain to Axelar")

    # Wait for return transfer transaction to be processed
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Get balances after transfer using common utilities
    echo '📊 Getting balances after return transfer...'
    ZIGCHAIN_BOB_UZIG_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_AFTER=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "unit-zig")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance after: $ZIGCHAIN_BOB_UZIG_AFTER"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance after: $AXELAR_BOB_AMOUNT_AFTER unit-zig"

    # Get module balances after transfer using common utilities
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)
    MODULE_IBC_AFTER=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance after: $MODULE_UZIG_AFTER"
    echo "📈 Module IBC balance after: $MODULE_IBC_AFTER"

    # Get transfer statistics after transfer using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    echo "📊 Transfer stats after: In=$TRANSFERRED_IN_AFTER, Out=$TRANSFERRED_OUT_AFTER"

    # Calculate expected values using common utilities
    ZIG_TRANSFER_NUMERIC="$TRANSFER_AMOUNT_NUMERIC"
    AXELAR_TRANSFER_NUMERIC="$REQUIRED_IBC_AMOUNT"
    EXPECTED_ZIGCHAIN_BOB_UZIG=$(echo "$ZIGCHAIN_BOB_UZIG_BEFORE - $ZIG_TRANSFER_NUMERIC - $FEES_AMOUNT_NUMERIC" | bc)
    EXPECTED_AXELAR_BOB_AMOUNT=$(echo "$AXELAR_BOB_AMOUNT_BEFORE + $AXELAR_TRANSFER_NUMERIC" | bc)
    EXPECTED_MODULE_UZIG=$(echo "$MODULE_UZIG_BEFORE + $ZIG_TRANSFER_NUMERIC" | bc)
    EXPECTED_MODULE_IBC=$(echo "$MODULE_IBC_BEFORE - $AXELAR_TRANSFER_NUMERIC" | bc)
    EXPECTED_TRANSFERRED_IN="$TRANSFERRED_IN_BEFORE"
    EXPECTED_TRANSFERRED_OUT=$(echo "$TRANSFERRED_OUT_BEFORE + $ZIG_TRANSFER_NUMERIC" | bc)

    # Log balance comparisons using common utility
    echo '🔍 Validating balance changes...'
    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME uzig" "$ZIGCHAIN_BOB_UZIG_BEFORE" "$ZIGCHAIN_BOB_UZIG_AFTER" "$EXPECTED_ZIGCHAIN_BOB_UZIG"
    log_balance_comparison "Axelar $TEST_ACCOUNT_NAME" "$AXELAR_BOB_AMOUNT_BEFORE" "$AXELAR_BOB_AMOUNT_AFTER" "$EXPECTED_AXELAR_BOB_AMOUNT"
    log_balance_comparison "Module uzig" "$MODULE_UZIG_BEFORE" "$MODULE_UZIG_AFTER" "$EXPECTED_MODULE_UZIG"
    log_balance_comparison "Module IBC" "$MODULE_IBC_BEFORE" "$MODULE_IBC_AFTER" "$EXPECTED_MODULE_IBC"
    log_balance_comparison "Transferred in" "$TRANSFERRED_IN_BEFORE" "$TRANSFERRED_IN_AFTER" "$EXPECTED_TRANSFERRED_IN"
    log_balance_comparison "Transferred out" "$TRANSFERRED_OUT_BEFORE" "$TRANSFERRED_OUT_AFTER" "$EXPECTED_TRANSFERRED_OUT"

    # Validate all balance changes and transfer statistics
    [[ "$ZIGCHAIN_BOB_UZIG_AFTER" == "$EXPECTED_ZIGCHAIN_BOB_UZIG" ]] && \
    [[ "$AXELAR_BOB_AMOUNT_AFTER" == "$EXPECTED_AXELAR_BOB_AMOUNT" ]] && \
    [[ "$MODULE_UZIG_AFTER" == "$EXPECTED_MODULE_UZIG" ]] && \
    [[ "$MODULE_IBC_AFTER" == "$EXPECTED_MODULE_IBC" ]] && \
    [[ "$TRANSFERRED_IN_AFTER" == "$EXPECTED_TRANSFERRED_IN" ]] && \
    [[ "$TRANSFERRED_OUT_AFTER" == "$EXPECTED_TRANSFERRED_OUT" ]]
}

# Test 005: Send tokens back to axelar
run_test "TW-005" "Return transfer from ZIGChain to Axelar" "test_005_return_transfer"

test_006_recovery() {
    echo '🔄 Testing native ZIG token recovery process...'

    # Setup test environment
    setup_test_env

    # Validate operator account exists
    if ! validate_zigchain_account "$OP_ACCOUNT_NAME"; then
        return 1
    fi

    # Validate test account exists
    if ! validate_zigchain_account "$TEST_ACCOUNT_NAME"; then
        return 1
    fi

    # Get initial balances using common utilities
    echo '📊 Checking account balances before recovery...'
    BOB_NATIVE_BALANCE_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    IBC_HASH=$(get_ibc_hash)
    BOB_IBC_BALANCE_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")

    echo "📈 Bob native balance before: $BOB_NATIVE_BALANCE_BEFORE uzig"
    echo "📈 Bob IBC balance before: $BOB_IBC_BALANCE_BEFORE ibc/..."

    # If Bob has no IBC balance, withdraw from module wallet first
    if [[ -z "$BOB_IBC_BALANCE_BEFORE" || "$BOB_IBC_BALANCE_BEFORE" == "0" ]]; then
        echo "🔄 Bob has no IBC balance, withdrawing from module wallet..."

        # Check module IBC balance
        MODULE_IBC_BALANCE_BEFORE=$(get_module_ibc_balance "$IBC_HASH")
        REQUIRED_IBC_AMOUNT=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")

        if ! [[ $(echo "$MODULE_IBC_BALANCE_BEFORE >= $REQUIRED_IBC_AMOUNT" | bc -l) -eq 1 ]]; then
            echo "❌ Module IBC balance ($MODULE_IBC_BALANCE_BEFORE) is less than required ($REQUIRED_IBC_AMOUNT)"
            return 1
        fi

        # Withdraw from module wallet
        WITHDRAW_AMOUNT="${REQUIRED_IBC_AMOUNT}${IBC_HASH}"
        echo "💸 Withdrawing $WITHDRAW_AMOUNT from module wallet..."
        WITHDRAW_CMD="zigchaind tx tokenwrapper withdraw-from-module-wallet $WITHDRAW_AMOUNT --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
        OUTPUT=$(execute_tx "$WITHDRAW_CMD" "Withdraw IBC tokens from module wallet")

        # Wait for withdrawal to process
        wait_for_tx "$TRANSACTION_TIMEOUT"

        # Update balances after withdrawal
        BOB_IBC_BALANCE_BEFORE=$(get_account_balance "zigchain" "$OP_ACCOUNT_NAME" "$IBC_HASH")
        BOB_NATIVE_BALANCE_BEFORE=$(get_account_balance "zigchain" "$OP_ACCOUNT_NAME" "uzig")

        echo "📈 Updated Bob IBC balance after withdrawal: $BOB_IBC_BALANCE_BEFORE ibc/..."

        if ! [[ $(echo "$BOB_IBC_BALANCE_BEFORE >= $REQUIRED_IBC_AMOUNT" | bc -l) -eq 1 ]]; then
            echo "❌ Bob IBC balance after withdrawal ($BOB_IBC_BALANCE_BEFORE) is insufficient"
            return 1
        fi
    fi

    # Calculate equivalent uzig amount (IBC amount / 10^12)
    EQUIVALENT_UZIG=$(echo "$BOB_IBC_BALANCE_BEFORE / 1000000000000" | bc)
    echo "🔢 IBC amount: $BOB_IBC_BALANCE_BEFORE"
    echo "🔢 Equivalent uzig amount: $EQUIVALENT_UZIG"

    # Validate module has sufficient uzig balance for recovery
    echo '🔍 Validating module balances before recovery...'
    MODULE_IBC_BALANCE_BEFORE=$(get_module_ibc_balance "$IBC_HASH")
    MODULE_NATIVE_BALANCE_BEFORE=$(get_module_uzig_balance)

    echo "📈 Module native balance before: $MODULE_NATIVE_BALANCE_BEFORE uzig"
    echo "📈 Module IBC balance before: $MODULE_IBC_BALANCE_BEFORE ibc/..."

    if ! validate_module_uzig_balance "$EQUIVALENT_UZIG"; then
        echo "💡 TIP: Fund the module wallet using:"
        echo "     zigchaind tx tokenwrapper fund-module-wallet ${EQUIVALENT_UZIG}uzig --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y"
        return 1
    fi

    # Get transfer statistics before recovery using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # Execute recovery transaction
    echo '🔄 Executing ZIG token recovery...'
    ZIGCHAIN_OP_ADDRESS=$(zigchaind keys show $TEST_ACCOUNT_NAME -a)
    RECOVERY_CMD="zigchaind tx tokenwrapper recover-zig $ZIGCHAIN_OP_ADDRESS --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$RECOVERY_CMD" "ZIG token recovery")

    # Wait for recovery to process
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Calculate expected values after recovery
    FEES_AMOUNT_NUMERIC=$(get_fee_numeric "$ZIGCHAIN_FEES")
    EXPECTED_BOB_NATIVE_AFTER=$(echo "$BOB_NATIVE_BALANCE_BEFORE + $EQUIVALENT_UZIG" | bc)
    EXPECTED_BOB_IBC_AFTER=0
    EXPECTED_MODULE_IBC_AFTER=$(echo "$MODULE_IBC_BALANCE_BEFORE + $BOB_IBC_BALANCE_BEFORE" | bc)
    EXPECTED_MODULE_NATIVE_AFTER=$(echo "$MODULE_NATIVE_BALANCE_BEFORE - $EQUIVALENT_UZIG" | bc)
    EXPECTED_TRANSFERRED_IN_AFTER=$(echo "$TRANSFERRED_IN_BEFORE + $EQUIVALENT_UZIG" | bc)
    EXPECTED_TRANSFERRED_OUT_AFTER="$TRANSFERRED_OUT_BEFORE"

    # Get balances after recovery using common utilities
    echo '📊 Checking balances after recovery...'
    BOB_NATIVE_BALANCE_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    BOB_IBC_BALANCE_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")
    MODULE_IBC_BALANCE_AFTER=$(get_module_ibc_balance "$IBC_HASH")
    MODULE_NATIVE_BALANCE_AFTER=$(get_module_uzig_balance)

    # Get transfer statistics after recovery using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)

    # Log balance comparisons using common utility
    echo '🔍 Validating balance changes...'
    log_balance_comparison "Bob native" "$BOB_NATIVE_BALANCE_BEFORE" "$BOB_NATIVE_BALANCE_AFTER" "$EXPECTED_BOB_NATIVE_AFTER"
    log_balance_comparison "Bob IBC" "$BOB_IBC_BALANCE_BEFORE" "$BOB_IBC_BALANCE_AFTER" "$EXPECTED_BOB_IBC_AFTER"
    log_balance_comparison "Module IBC" "$MODULE_IBC_BALANCE_BEFORE" "$MODULE_IBC_BALANCE_AFTER" "$EXPECTED_MODULE_IBC_AFTER"
    log_balance_comparison "Module native" "$MODULE_NATIVE_BALANCE_BEFORE" "$MODULE_NATIVE_BALANCE_AFTER" "$EXPECTED_MODULE_NATIVE_AFTER"
    log_balance_comparison "Transferred in" "$TRANSFERRED_IN_BEFORE" "$TRANSFERRED_IN_AFTER" "$EXPECTED_TRANSFERRED_IN_AFTER"
    log_balance_comparison "Transferred out" "$TRANSFERRED_OUT_BEFORE" "$TRANSFERRED_OUT_AFTER" "$EXPECTED_TRANSFERRED_OUT_AFTER"

    # Validate all balance changes and transfer statistics
    [[ "$BOB_IBC_BALANCE_AFTER" == "$EXPECTED_BOB_IBC_AFTER" ]] && \
    [[ "$BOB_NATIVE_BALANCE_AFTER" == "$EXPECTED_BOB_NATIVE_AFTER" ]] && \
    [[ "$MODULE_IBC_BALANCE_AFTER" == "$EXPECTED_MODULE_IBC_AFTER" ]] && \
    [[ "$MODULE_NATIVE_BALANCE_AFTER" == "$EXPECTED_MODULE_NATIVE_AFTER" ]] && \
    [[ "$TRANSFERRED_IN_AFTER" == "$EXPECTED_TRANSFERRED_IN_AFTER" ]] && \
    [[ "$TRANSFERRED_OUT_AFTER" == "$EXPECTED_TRANSFERRED_OUT_AFTER" ]]
}

# Test 006: Test native ZIG token recovery process
run_test "TW-006" "Test native ZIG token recovery process" "test_006_recovery"

test_007_disabled_transfer() {
    echo '🚫 Testing IBC transfer when token wrapper is disabled...'

    # Setup test environment
    setup_test_env

    # Validate all required accounts exist
    echo '🔍 Validating accounts...'
    if ! validate_accounts "$TEST_ACCOUNT_NAME" || ! validate_zigchain_account "$OP_ACCOUNT_NAME"; then
        return 1
    fi

    # Validate module has sufficient uzig balance for the transfer
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    REQUIRED_UZIG=$(echo "$TRANSFER_AMOUNT_NUMERIC / 1000000000000" | bc)

    if ! validate_module_uzig_balance "$REQUIRED_UZIG"; then
        return 1
    fi

    # Validate axelar account has sufficient balance
    if ! validate_sufficient_balance "axelar" "$TEST_ACCOUNT_NAME" "$TRANSFER_AMOUNT_NUMERIC" "unit-zig"; then
        return 1
    fi

    # Disable the token wrapper module
    echo '🚫 Disabling token wrapper module...'
    DISABLE_CMD="zigchaind tx tokenwrapper disable-token-wrapper --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$DISABLE_CMD" "Disable token wrapper")

    # Wait for disable transaction
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify token wrapper is disabled
    if is_token_wrapper_enabled; then
        echo "❌ Token wrapper should be disabled but is still enabled"
        return 1
    fi
    echo "✅ Token wrapper module disabled successfully"

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before IBC transfer...'
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)
    IBC_HASH=$(get_ibc_hash)
    MODULE_IBC_BEFORE=$(get_module_ibc_balance "$IBC_HASH")
    ZIGCHAIN_BOB_IBC_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")

    # Perform IBC transfer from axelar to zigchain
    echo '🚀 Performing IBC transfer from axelar to zigchain...'
    ZIGCHAIN_TEST_ADDRESS=$(zigchaind keys show $TEST_ACCOUNT_NAME -a)
    TRANSFER_CMD="axelard tx ibc-transfer transfer transfer $AXELAR_ZIGCHAIN_CHANNEL_ID $ZIGCHAIN_TEST_ADDRESS $AXELAR_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $AXELAR_FEES --chain-id $AXELAR_CHAIN_ID -y -o json 2>/dev/null"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "IBC transfer from Axelar to ZIGChain (disabled TW)")

    # Wait for IBC transfer
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"

    # Check that module wallet balances remain unchanged
    echo '🔍 Checking module wallet balances after transfer...'
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)
    MODULE_IBC_AFTER=$(get_module_ibc_balance "$IBC_HASH")

    echo "📊 Module uzig balance - before: $MODULE_UZIG_BEFORE, after: $MODULE_UZIG_AFTER"
    echo "📊 Module IBC balance - before: $MODULE_IBC_BEFORE, after: $MODULE_IBC_AFTER"

    if [[ "$MODULE_UZIG_BEFORE" != "$MODULE_UZIG_AFTER" ]]; then
        echo "❌ Module uzig balance changed from $MODULE_UZIG_BEFORE to $MODULE_UZIG_AFTER"
        return 1
    fi
    if [[ "$MODULE_IBC_BEFORE" != "$MODULE_IBC_AFTER" ]]; then
        echo "❌ Module IBC balance changed from $MODULE_IBC_BEFORE to $MODULE_IBC_AFTER"
        return 1
    fi
    echo "✅ Module wallet balances unchanged"

    # Check that zigchain account received IBC vouchers
    echo '🔍 Checking zigchain account IBC balance...'
    ZIGCHAIN_BOB_IBC_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")
    EXPECTED_ZIGCHAIN_BOB_IBC=$(echo "$ZIGCHAIN_BOB_IBC_BEFORE + $TRANSFER_AMOUNT_NUMERIC" | bc)

    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME IBC" "$ZIGCHAIN_BOB_IBC_BEFORE" "$ZIGCHAIN_BOB_IBC_AFTER" "$EXPECTED_ZIGCHAIN_BOB_IBC"

    if [[ "$ZIGCHAIN_BOB_IBC_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_IBC" ]]; then
        echo "❌ Zigchain account IBC balance should be $EXPECTED_ZIGCHAIN_BOB_IBC but got $ZIGCHAIN_BOB_IBC_AFTER"
        return 1
    fi
    echo "✅ Zigchain account IBC balance increased correctly"

    # Re-enable the token wrapper module for cleanup
    echo '🔄 Re-enabling token wrapper module...'
    ENABLE_CMD="zigchaind tx tokenwrapper enable-token-wrapper --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$ENABLE_CMD" "Re-enable token wrapper")

    # Wait for enable transaction
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify token wrapper is enabled again
    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper should be re-enabled but is still disabled"
        return 1
    fi
    echo "✅ Token wrapper module re-enabled successfully"
}

# Test 007: IBC transfer when token wrapper is disabled
run_test "TW-007" "IBC transfer when token wrapper is disabled (should receive IBC vouchers)" "test_007_disabled_transfer"

test_008_wrong_ibc_settings_wrong_native_client_id() {
    echo '⚙️ Testing IBC transfer with wrong IBC settings - wrong native client id...'

    # Setup test environment
    setup_test_env

    # Validate all required accounts exist
    echo '🔍 Validating accounts...'
    if ! validate_accounts "$TEST_ACCOUNT_NAME" || ! validate_zigchain_account "$OP_ACCOUNT_NAME"; then
        return 1
    fi

    # Validate module has sufficient uzig balance
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    REQUIRED_UZIG=$(echo "$TRANSFER_AMOUNT_NUMERIC / 1000000000000" | bc)

    if ! validate_module_uzig_balance "$REQUIRED_UZIG"; then
        return 1
    fi

    # Validate axelar account has sufficient balance
    if ! validate_sufficient_balance "axelar" "$TEST_ACCOUNT_NAME" "$TRANSFER_AMOUNT_NUMERIC" "unit-zig"; then
        return 1
    fi

    # Store expected IBC settings for restoration
    EXPECTED_NATIVE_CLIENT_ID=${EXPECTED_NATIVE_CLIENT_ID:-'07-tendermint-1'}
    EXPECTED_COUNTERPARTY_CLIENT_ID=${EXPECTED_COUNTERPARTY_CLIENT_ID:-'07-tendermint-0'}
    EXPECTED_NATIVE_PORT=${EXPECTED_NATIVE_PORT:-'transfer'}
    EXPECTED_COUNTERPARTY_PORT=${EXPECTED_COUNTERPARTY_PORT:-'transfer'}
    EXPECTED_NATIVE_CHANNEL=${EXPECTED_NATIVE_CHANNEL:-'channel-1'}
    EXPECTED_COUNTERPARTY_CHANNEL=${EXPECTED_COUNTERPARTY_CHANNEL:-'channel-0'}
    EXPECTED_DENOM=${EXPECTED_DENOM:-$AXELAR_DENOM}
    EXPECTED_DECIMAL_DIFFERENCE=${EXPECTED_DECIMAL_DIFFERENCE:-'12'}

    # Update IBC settings with wrong native client ID
    echo '⚙️ Setting wrong native client id...'
    WRONG_NATIVE_CLIENT_ID="07-tendermint-999"

    UPDATE_CMD="zigchaind tx tokenwrapper update-ibc-settings $WRONG_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $EXPECTED_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$UPDATE_CMD" "Update IBC settings with wrong client ID")

    # Wait for settings update
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were updated with wrong client id
    UPDATED_NATIVE_CLIENT_ID=$(echo "$(get_module_info)" | jq -r '.native_client_id')
    if [[ "$UPDATED_NATIVE_CLIENT_ID" != "$WRONG_NATIVE_CLIENT_ID" ]]; then
        echo "❌ Native client ID should be $WRONG_NATIVE_CLIENT_ID but got $UPDATED_NATIVE_CLIENT_ID"
        return 1
    fi
    echo "✅ IBC settings updated with wrong native client id"

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before IBC transfer...'
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)
    # Use the IBC hash based on expected settings
    IBC_HASH="ibc/$(calculate_ibc_hash "$EXPECTED_NATIVE_PORT" "$EXPECTED_NATIVE_CHANNEL" "unit-zig")"
    MODULE_IBC_BEFORE=$(get_module_ibc_balance "$IBC_HASH")
    ZIGCHAIN_BOB_IBC_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")

    # Perform IBC transfer from axelar to zigchain
    echo '🚀 Performing IBC transfer from axelar to zigchain...'
    ZIGCHAIN_TEST_ADDRESS=$(zigchaind keys show $TEST_ACCOUNT_NAME -a)
    TRANSFER_CMD="axelard tx ibc-transfer transfer transfer $AXELAR_ZIGCHAIN_CHANNEL_ID $ZIGCHAIN_TEST_ADDRESS $AXELAR_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $AXELAR_FEES --chain-id $AXELAR_CHAIN_ID -y -o json 2>/dev/null"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "IBC transfer with wrong IBC settings")

    # Wait for IBC transfer
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"

    # Check that module wallet balances remain unchanged
    echo '🔍 Checking module wallet balances after transfer...'
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)
    MODULE_IBC_AFTER=$(get_module_ibc_balance "$IBC_HASH")

    echo "📊 Module uzig balance - before: $MODULE_UZIG_BEFORE, after: $MODULE_UZIG_AFTER"
    echo "📊 Module IBC balance - before: $MODULE_IBC_BEFORE, after: $MODULE_IBC_AFTER"

    if [[ "$MODULE_UZIG_BEFORE" != "$MODULE_UZIG_AFTER" ]]; then
        echo "❌ Module uzig balance changed from $MODULE_UZIG_BEFORE to $MODULE_UZIG_AFTER"
        return 1
    fi
    if [[ "$MODULE_IBC_BEFORE" != "$MODULE_IBC_AFTER" ]]; then
        echo "❌ Module IBC balance changed from $MODULE_IBC_BEFORE to $MODULE_IBC_AFTER"
        return 1
    fi
    echo "✅ Module wallet balances unchanged"

    # Check that zigchain account received IBC vouchers
    echo '🔍 Checking zigchain account IBC balance...'
    ZIGCHAIN_BOB_IBC_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")
    EXPECTED_ZIGCHAIN_BOB_IBC=$(echo "$ZIGCHAIN_BOB_IBC_BEFORE + $TRANSFER_AMOUNT_NUMERIC" | bc)

    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME IBC" "$ZIGCHAIN_BOB_IBC_BEFORE" "$ZIGCHAIN_BOB_IBC_AFTER" "$EXPECTED_ZIGCHAIN_BOB_IBC"

    if [[ "$ZIGCHAIN_BOB_IBC_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_IBC" ]]; then
        echo "❌ Zigchain account IBC balance should be $EXPECTED_ZIGCHAIN_BOB_IBC but got $ZIGCHAIN_BOB_IBC_AFTER"
        return 1
    fi
    echo "✅ Zigchain account IBC balance increased correctly"

    # Restore IBC settings to correct values
    echo '🔄 Restoring IBC settings to correct values...'
    RESTORE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $EXPECTED_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$RESTORE_CMD" "Restore correct IBC settings")

    # Wait for restoration
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were restored
    FINAL_NATIVE_CLIENT_ID=$(echo "$(get_module_info)" | jq -r '.native_client_id')
    if [[ "$FINAL_NATIVE_CLIENT_ID" != "$EXPECTED_NATIVE_CLIENT_ID" ]]; then
        echo "❌ Native client ID should be restored to $EXPECTED_NATIVE_CLIENT_ID but got $FINAL_NATIVE_CLIENT_ID"
        return 1
    fi
    echo "✅ IBC settings restored to correct values"
}

# Test 008: IBC transfer with wrong IBC settings
run_test "TW-008" "IBC transfer with wrong IBC settings - wrong native client id (should receive IBC vouchers)" "test_008_wrong_ibc_settings_wrong_native_client_id"

test_009_wrong_ibc_settings_wrong_native_channel_id() {
    echo '⚙️ Testing IBC transfer with wrong IBC settings - wrong native channel id...'

    # Setup test environment
    setup_test_env

    # Validate all required accounts exist
    echo '🔍 Validating accounts...'
    if ! validate_accounts "$TEST_ACCOUNT_NAME" || ! validate_zigchain_account "$OP_ACCOUNT_NAME"; then
        return 1
    fi

    # Validate module has sufficient uzig balance
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    REQUIRED_UZIG=$(echo "$TRANSFER_AMOUNT_NUMERIC / 1000000000000" | bc)

    if ! validate_module_uzig_balance "$REQUIRED_UZIG"; then
        return 1
    fi

    # Validate axelar account has sufficient balance
    if ! validate_sufficient_balance "axelar" "$TEST_ACCOUNT_NAME" "$TRANSFER_AMOUNT_NUMERIC" "unit-zig"; then
        return 1
    fi

    # Use the correct IBC hash
    IBC_HASH=$(get_ibc_hash)

    # Store expected IBC settings for restoration
    EXPECTED_NATIVE_CLIENT_ID=${EXPECTED_NATIVE_CLIENT_ID:-'07-tendermint-1'}
    EXPECTED_COUNTERPARTY_CLIENT_ID=${EXPECTED_COUNTERPARTY_CLIENT_ID:-'07-tendermint-0'}
    EXPECTED_NATIVE_PORT=${EXPECTED_NATIVE_PORT:-'transfer'}
    EXPECTED_COUNTERPARTY_PORT=${EXPECTED_COUNTERPARTY_PORT:-'transfer'}
    EXPECTED_NATIVE_CHANNEL=${EXPECTED_NATIVE_CHANNEL:-'channel-1'}
    EXPECTED_COUNTERPARTY_CHANNEL=${EXPECTED_COUNTERPARTY_CHANNEL:-'channel-0'}
    EXPECTED_DENOM=${EXPECTED_DENOM:-$AXELAR_DENOM}
    EXPECTED_DECIMAL_DIFFERENCE=${EXPECTED_DECIMAL_DIFFERENCE:-'12'}

    # Update IBC settings with wrong native channel ID
    echo '⚙️ Setting wrong native channel id...'
    WRONG_NATIVE_CHANNEL_ID="channel-999"

    UPDATE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $WRONG_NATIVE_CHANNEL_ID $EXPECTED_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$UPDATE_CMD" "Update IBC settings with wrong native channel ID")

    # Wait for settings update
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were updated with wrong native channel id
    UPDATED_NATIVE_CHANNEL_ID=$(echo "$(get_module_info)" | jq -r '.native_channel')
    if [[ "$UPDATED_NATIVE_CHANNEL_ID" != "$WRONG_NATIVE_CHANNEL_ID" ]]; then
        echo "❌ Native channel ID should be $WRONG_NATIVE_CHANNEL_ID but got $UPDATED_NATIVE_CHANNEL_ID"
        return 1
    fi
    echo "✅ IBC settings updated with wrong native channel id"

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before IBC transfer...'
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)
    MODULE_IBC_BEFORE=$(get_module_ibc_balance "$IBC_HASH")
    ZIGCHAIN_BOB_IBC_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")

    # Perform IBC transfer from axelar to zigchain
    echo '🚀 Performing IBC transfer from axelar to zigchain...'
    ZIGCHAIN_TEST_ADDRESS=$(zigchaind keys show $TEST_ACCOUNT_NAME -a)
    TRANSFER_CMD="axelard tx ibc-transfer transfer transfer $AXELAR_ZIGCHAIN_CHANNEL_ID $ZIGCHAIN_TEST_ADDRESS $AXELAR_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $AXELAR_FEES --chain-id $AXELAR_CHAIN_ID -y -o json 2>/dev/null"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "IBC transfer with wrong IBC settings")

    # Wait for IBC transfer
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"

    # Check that module wallet balances remain unchanged
    echo '🔍 Checking module wallet balances after transfer...'
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)
    MODULE_IBC_AFTER=$(get_module_ibc_balance "$IBC_HASH")

    echo "📊 Module uzig balance - before: $MODULE_UZIG_BEFORE, after: $MODULE_UZIG_AFTER"
    echo "📊 Module IBC balance - before: $MODULE_IBC_BEFORE, after: $MODULE_IBC_AFTER"

    if [[ "$MODULE_UZIG_BEFORE" != "$MODULE_UZIG_AFTER" ]]; then
        echo "❌ Module uzig balance changed from $MODULE_UZIG_BEFORE to $MODULE_UZIG_AFTER"
        return 1
    fi
    if [[ "$MODULE_IBC_BEFORE" != "$MODULE_IBC_AFTER" ]]; then
        echo "❌ Module IBC balance changed from $MODULE_IBC_BEFORE to $MODULE_IBC_AFTER"
        return 1
    fi
    echo "✅ Module wallet balances unchanged"

    # Check that zigchain account received IBC vouchers
    echo '🔍 Checking zigchain account IBC balance...'
    ZIGCHAIN_BOB_IBC_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")
    EXPECTED_ZIGCHAIN_BOB_IBC=$(echo "$ZIGCHAIN_BOB_IBC_BEFORE + $TRANSFER_AMOUNT_NUMERIC" | bc)

    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME IBC" "$ZIGCHAIN_BOB_IBC_BEFORE" "$ZIGCHAIN_BOB_IBC_AFTER" "$EXPECTED_ZIGCHAIN_BOB_IBC"

    if [[ "$ZIGCHAIN_BOB_IBC_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_IBC" ]]; then
        echo "❌ Zigchain account IBC balance should be $EXPECTED_ZIGCHAIN_BOB_IBC but got $ZIGCHAIN_BOB_IBC_AFTER"
        return 1
    fi
    echo "✅ Zigchain account IBC balance increased correctly"

    # Restore IBC settings to correct values
    echo '🔄 Restoring IBC settings to correct values...'
    RESTORE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $EXPECTED_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$RESTORE_CMD" "Restore correct IBC settings")

    # Wait for restoration
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were restored
    FINAL_NATIVE_CHANNEL_ID=$(echo "$(get_module_info)" | jq -r '.native_channel')
    if [[ "$FINAL_NATIVE_CHANNEL_ID" != "$EXPECTED_NATIVE_CHANNEL" ]]; then
        echo "❌ Native channel ID should be restored to $EXPECTED_NATIVE_CHANNEL but got $FINAL_NATIVE_CHANNEL_ID"
        return 1
    fi
    echo "✅ IBC settings restored to correct values"
}

# Test 009: IBC transfer with wrong IBC settings - wrong native channel id
run_test "TW-009" "IBC transfer with wrong IBC settings - wrong native channel id (should receive IBC vouchers)" "test_009_wrong_ibc_settings_wrong_native_channel_id"

test_010_wrong_ibc_settings_wrong_counterparty_channel_id() {
    echo '⚙️ Testing IBC transfer with wrong IBC settings - wrong counterparty channel id...'

    # Setup test environment
    setup_test_env

    # Validate all required accounts exist
    echo '🔍 Validating accounts...'
    if ! validate_accounts "$TEST_ACCOUNT_NAME" || ! validate_zigchain_account "$OP_ACCOUNT_NAME"; then
        return 1
    fi

    # Validate module has sufficient uzig balance
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    REQUIRED_UZIG=$(echo "$TRANSFER_AMOUNT_NUMERIC / 1000000000000" | bc)

    if ! validate_module_uzig_balance "$REQUIRED_UZIG"; then
        return 1
    fi

    # Validate axelar account has sufficient balance
    if ! validate_sufficient_balance "axelar" "$TEST_ACCOUNT_NAME" "$TRANSFER_AMOUNT_NUMERIC" "unit-zig"; then
        return 1
    fi

    # Use the correct IBC hash
    IBC_HASH=$(get_ibc_hash)

    # Store expected IBC settings for restoration
    EXPECTED_NATIVE_CLIENT_ID=${EXPECTED_NATIVE_CLIENT_ID:-'07-tendermint-1'}
    EXPECTED_COUNTERPARTY_CLIENT_ID=${EXPECTED_COUNTERPARTY_CLIENT_ID:-'07-tendermint-0'}
    EXPECTED_NATIVE_PORT=${EXPECTED_NATIVE_PORT:-'transfer'}
    EXPECTED_COUNTERPARTY_PORT=${EXPECTED_COUNTERPARTY_PORT:-'transfer'}
    EXPECTED_NATIVE_CHANNEL=${EXPECTED_NATIVE_CHANNEL:-'channel-1'}
    EXPECTED_COUNTERPARTY_CHANNEL=${EXPECTED_COUNTERPARTY_CHANNEL:-'channel-0'}
    EXPECTED_DENOM=${EXPECTED_DENOM:-$AXELAR_DENOM}
    EXPECTED_DECIMAL_DIFFERENCE=${EXPECTED_DECIMAL_DIFFERENCE:-'12'}

    # Update IBC settings with wrong counterparty channel ID
    echo '⚙️ Setting wrong counterparty channel id...'
    WRONG_COUNTERPARTY_CHANNEL_ID="channel-999"

    UPDATE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $WRONG_COUNTERPARTY_CHANNEL_ID $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$UPDATE_CMD" "Update IBC settings with wrong counterparty channel ID")

    # Wait for settings update
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were updated with wrong counterparty channel id
    UPDATED_COUNTERPARTY_CHANNEL_ID=$(echo "$(get_module_info)" | jq -r '.counterparty_channel')
    if [[ "$UPDATED_COUNTERPARTY_CHANNEL_ID" != "$WRONG_COUNTERPARTY_CHANNEL_ID" ]]; then
        echo "❌ Counterparty channel ID should be $WRONG_COUNTERPARTY_CHANNEL_ID but got $UPDATED_COUNTERPARTY_CHANNEL_ID"
        return 1
    fi
    echo "✅ IBC settings updated with wrong counterparty channel id"

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before IBC transfer...'
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)
    MODULE_IBC_BEFORE=$(get_module_ibc_balance "$IBC_HASH")
    ZIGCHAIN_BOB_IBC_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")

    # Perform IBC transfer from axelar to zigchain
    echo '🚀 Performing IBC transfer from axelar to zigchain...'
    ZIGCHAIN_TEST_ADDRESS=$(zigchaind keys show $TEST_ACCOUNT_NAME -a)
    TRANSFER_CMD="axelard tx ibc-transfer transfer transfer $AXELAR_ZIGCHAIN_CHANNEL_ID $ZIGCHAIN_TEST_ADDRESS $AXELAR_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $AXELAR_FEES --chain-id $AXELAR_CHAIN_ID -y -o json 2>/dev/null"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "IBC transfer with wrong IBC settings")

    # Wait for IBC transfer
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"

    # Check that module wallet balances remain unchanged
    echo '🔍 Checking module wallet balances after transfer...'
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)
    MODULE_IBC_AFTER=$(get_module_ibc_balance "$IBC_HASH")

    echo "📊 Module uzig balance - before: $MODULE_UZIG_BEFORE, after: $MODULE_UZIG_AFTER"
    echo "📊 Module IBC balance - before: $MODULE_IBC_BEFORE, after: $MODULE_IBC_AFTER"

    if [[ "$MODULE_UZIG_BEFORE" != "$MODULE_UZIG_AFTER" ]]; then
        echo "❌ Module uzig balance changed from $MODULE_UZIG_BEFORE to $MODULE_UZIG_AFTER"
        return 1
    fi
    if [[ "$MODULE_IBC_BEFORE" != "$MODULE_IBC_AFTER" ]]; then
        echo "❌ Module IBC balance changed from $MODULE_IBC_BEFORE to $MODULE_IBC_AFTER"
        return 1
    fi
    echo "✅ Module wallet balances unchanged"

    # Check that zigchain account received IBC vouchers
    echo '🔍 Checking zigchain account IBC balance...'
    ZIGCHAIN_BOB_IBC_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")
    EXPECTED_ZIGCHAIN_BOB_IBC=$(echo "$ZIGCHAIN_BOB_IBC_BEFORE + $TRANSFER_AMOUNT_NUMERIC" | bc)

    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME IBC" "$ZIGCHAIN_BOB_IBC_BEFORE" "$ZIGCHAIN_BOB_IBC_AFTER" "$EXPECTED_ZIGCHAIN_BOB_IBC"

    if [[ "$ZIGCHAIN_BOB_IBC_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_IBC" ]]; then
        echo "❌ Zigchain account IBC balance should be $EXPECTED_ZIGCHAIN_BOB_IBC but got $ZIGCHAIN_BOB_IBC_AFTER"
        return 1
    fi
    echo "✅ Zigchain account IBC balance increased correctly"

    # Restore IBC settings to correct values
    echo '🔄 Restoring IBC settings to correct values...'
    RESTORE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $EXPECTED_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$RESTORE_CMD" "Restore correct IBC settings")

    # Wait for restoration
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were restored
    FINAL_COUNTERPARTY_CHANNEL_ID=$(echo "$(get_module_info)" | jq -r '.counterparty_channel')
    if [[ "$FINAL_COUNTERPARTY_CHANNEL_ID" != "$EXPECTED_COUNTERPARTY_CHANNEL" ]]; then
        echo "❌ Counterparty channel ID should be restored to $EXPECTED_COUNTERPARTY_CHANNEL but got $FINAL_COUNTERPARTY_CHANNEL_ID"
        return 1
    fi
    echo "✅ IBC settings restored to correct values"
}

# Test 010: IBC transfer with wrong IBC settings - wrong counterparty channel id
run_test "TW-010" "IBC transfer with wrong IBC settings - wrong counterparty channel id (should receive IBC vouchers)" "test_010_wrong_ibc_settings_wrong_counterparty_channel_id"

test_011_return_transfer_module_disabled() {
    echo '↩️ Testing return transfer from ZIGChain to Axelar with module disabled...'

    # Setup test environment
    setup_test_env

    # Pre-transfer validation
    if ! pre_transfer_validation "ZIGChain → Axelar (module disabled)"; then
        return 1
    fi

    # Verify module is funded with sufficient uzig tokens using common utilities
    echo '🔍 Verifying module wallet has sufficient tokens for return transfer...'
    REQUIRED_UZIG_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    REQUIRED_IBC_AMOUNT=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    IBC_HASH=$(get_ibc_hash)

    if ! validate_module_uzig_balance "$REQUIRED_UZIG_AMOUNT"; then
        return 1
    fi

    # Verify module has sufficient IBC tokens
    MODULE_IBC_BALANCE=$(get_module_ibc_balance "$IBC_HASH")
    if ! [[ $(echo "$MODULE_IBC_BALANCE >= $REQUIRED_IBC_AMOUNT" | bc -l) -eq 1 ]]; then
        echo "❌ Insufficient IBC balance in module wallet."
        echo "   Required: >= $REQUIRED_IBC_AMOUNT IBC tokens"
        echo "   Current: $MODULE_IBC_BALANCE IBC tokens"
        return 1
    fi

    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before return transfer...'
    ZIGCHAIN_BOB_UZIG_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_BEFORE=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "unit-zig")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance before: $ZIGCHAIN_BOB_UZIG_BEFORE"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance before: $AXELAR_BOB_AMOUNT_BEFORE unit-zig"

    # Verify account has sufficient uzig balance for return transfer (including fees)
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    FEES_AMOUNT_NUMERIC=$(get_fee_numeric "$ZIGCHAIN_FEES")
    REQUIRED_ZIG_AMOUNT=$(echo "$TRANSFER_AMOUNT_NUMERIC + $FEES_AMOUNT_NUMERIC" | bc)

    if ! validate_sufficient_balance "zigchain" "$TEST_ACCOUNT_NAME" "$REQUIRED_ZIG_AMOUNT" "uzig"; then
        return 1
    fi

    # Get module balances before transfer using common utilities
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)
    MODULE_IBC_BEFORE=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance before: $MODULE_UZIG_BEFORE"
    echo "📈 Module IBC balance before: $MODULE_IBC_BEFORE"

    # Get transfer statistics before transfer using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # DISABLE the token wrapper module before attempting transfer
    echo '🚫 Disabling token wrapper module before transfer...'
    DISABLE_CMD="zigchaind tx tokenwrapper disable-token-wrapper --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$DISABLE_CMD" "Disable token wrapper")

    # Wait for disable transaction
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify token wrapper is disabled
    if is_token_wrapper_enabled; then
        echo "❌ Token wrapper should be disabled but is still enabled"
        return 1
    fi
    echo "✅ Token wrapper module disabled successfully"

    # Send tokens back from ZIGChain to Axelar
    echo '🚀 Attempting return transfer from ZIGChain to Axelar with module disabled...'
    AXELAR_TEST_ADDRESS=$(axelard keys show $TEST_ACCOUNT_NAME -a 2>/dev/null)
    TRANSFER_CMD="zigchaind tx ibc-transfer transfer transfer $ZIGCHAIN_AXELAR_CHANNEL_ID $AXELAR_TEST_ADDRESS $ZIG_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "Return transfer from ZIGChain to Axelar (module disabled)")

    # Wait for return transfer transaction to be processed
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Get balances after transfer using common utilities
    echo '📊 Getting balances after return transfer...'
    ZIGCHAIN_BOB_UZIG_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_AFTER=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "unit-zig")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance after: $ZIGCHAIN_BOB_UZIG_AFTER"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance after: $AXELAR_BOB_AMOUNT_AFTER unit-zig"

    # Get module balances after transfer using common utilities
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)
    MODULE_IBC_AFTER=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance after: $MODULE_UZIG_AFTER"
    echo "📈 Module IBC balance after: $MODULE_IBC_AFTER"

    # Get transfer statistics after transfer using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    echo "📊 Transfer stats after: In=$TRANSFERRED_IN_AFTER, Out=$TRANSFERRED_OUT_AFTER"

    # Check that module wallet balances remain unchanged
    echo '🔍 Checking module wallet balances after transfer...'
    if [[ "$MODULE_UZIG_BEFORE" != "$MODULE_UZIG_AFTER" ]]; then
        echo "❌ Module uzig balance changed from $MODULE_UZIG_BEFORE to $MODULE_UZIG_AFTER"
        return 1
    fi
    if [[ "$MODULE_IBC_BEFORE" != "$MODULE_IBC_AFTER" ]]; then
        echo "❌ Module IBC balance changed from $MODULE_IBC_BEFORE to $MODULE_IBC_AFTER"
        return 1
    fi
    echo "✅ Module wallet balances unchanged"

    # Check that zigchain account IBC balance remains unchanged (since module is disabled)
    echo '🔍 Checking zigchain account IBC balance...'
    ZIGCHAIN_BOB_IBC_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")
    ZIGCHAIN_BOB_IBC_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")

    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME IBC" "$ZIGCHAIN_BOB_IBC_BEFORE" "$ZIGCHAIN_BOB_IBC_AFTER" "$ZIGCHAIN_BOB_IBC_BEFORE"

    if [[ "$ZIGCHAIN_BOB_IBC_AFTER" != "$ZIGCHAIN_BOB_IBC_BEFORE" ]]; then
        echo "❌ Zigchain account IBC balance should remain unchanged but got $ZIGCHAIN_BOB_IBC_AFTER"
        return 1
    fi
    echo "✅ Zigchain account IBC balance unchanged"

    # Check that zigchain account uzig balance decreased by the fees
    EXPECTED_ZIGCHAIN_BOB_UZIG=$(echo "$ZIGCHAIN_BOB_UZIG_BEFORE - $FEES_AMOUNT_NUMERIC" | bc)
    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME uzig" "$ZIGCHAIN_BOB_UZIG_BEFORE" "$ZIGCHAIN_BOB_UZIG_AFTER" "$EXPECTED_ZIGCHAIN_BOB_UZIG"
    if [[ "$ZIGCHAIN_BOB_UZIG_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_UZIG" ]]; then
        echo "❌ Zigchain account uzig balance should decrease by the fees but got $ZIGCHAIN_BOB_UZIG_AFTER"
        return 1
    fi
    echo "✅ Zigchain account uzig balance decreased by the fees"

    # Check that axelar account balance remains unchanged
    log_balance_comparison "Axelar $TEST_ACCOUNT_NAME" "$AXELAR_BOB_AMOUNT_BEFORE" "$AXELAR_BOB_AMOUNT_AFTER" "$AXELAR_BOB_AMOUNT_BEFORE"

    if [[ "$AXELAR_BOB_AMOUNT_BEFORE" != "$AXELAR_BOB_AMOUNT_AFTER" ]]; then
        echo "❌ Axelar account balance should remain unchanged but changed from $AXELAR_BOB_AMOUNT_BEFORE to $AXELAR_BOB_AMOUNT_AFTER"
        return 1
    fi
    echo "✅ Axelar account balance unchanged"

    # Re-enable the token wrapper module for cleanup
    echo '🔄 Re-enabling token wrapper module...'
    ENABLE_CMD="zigchaind tx tokenwrapper enable-token-wrapper --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$ENABLE_CMD" "Re-enable token wrapper")

    # Wait for enable transaction
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify token wrapper is enabled again
    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper should be re-enabled but is still disabled"
        return 1
    fi
    echo "✅ Token wrapper module re-enabled successfully"

    # Validate that transfer statistics remain unchanged (since transfer failed completely)
    [[ "$TRANSFERRED_IN_BEFORE" == "$TRANSFERRED_IN_AFTER" ]] && \
    [[ "$TRANSFERRED_OUT_BEFORE" == "$TRANSFERRED_OUT_AFTER" ]]
}

# Test 011: Return transfer from ZIGChain to Axelar with module disabled
run_test "TW-011" "Return transfer from ZIGChain to Axelar with module disabled (balances should remain unchanged)" "test_011_return_transfer_module_disabled"

test_012_return_transfer_insufficient_ibc() {
    echo '↩️ Testing return transfer from ZIGChain to Axelar with insufficient IBC balance...'

    # Setup test environment
    setup_test_env

    # Pre-transfer validation
    if ! pre_transfer_validation "ZIGChain → Axelar (insufficient IBC)"; then
        return 1
    fi

    # Verify module is funded with sufficient uzig tokens using common utilities
    echo '🔍 Verifying module wallet has sufficient tokens for return transfer...'
    REQUIRED_UZIG_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    REQUIRED_IBC_AMOUNT=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    IBC_HASH=$(get_ibc_hash)

    if ! validate_module_uzig_balance "$REQUIRED_UZIG_AMOUNT"; then
        return 1
    fi

    # Verify module has sufficient IBC tokens
    MODULE_IBC_BALANCE=$(get_module_ibc_balance "$IBC_HASH")
    if ! [[ $(echo "$MODULE_IBC_BALANCE >= $REQUIRED_IBC_AMOUNT" | bc -l) -eq 1 ]]; then
        echo "❌ Insufficient IBC balance in module wallet."
        echo "   Required: >= $REQUIRED_IBC_AMOUNT IBC tokens"
        echo "   Current: $MODULE_IBC_BALANCE IBC tokens"
        return 1
    fi

    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before return transfer...'
    ZIGCHAIN_BOB_UZIG_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_BEFORE=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "unit-zig")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance before: $ZIGCHAIN_BOB_UZIG_BEFORE"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance before: $AXELAR_BOB_AMOUNT_BEFORE unit-zig"

    # Verify account has sufficient uzig balance for return transfer (including fees)
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    FEES_AMOUNT_NUMERIC=$(get_fee_numeric "$ZIGCHAIN_FEES")
    REQUIRED_ZIG_AMOUNT=$(echo "$TRANSFER_AMOUNT_NUMERIC + $FEES_AMOUNT_NUMERIC" | bc)

    if ! validate_sufficient_balance "zigchain" "$TEST_ACCOUNT_NAME" "$REQUIRED_ZIG_AMOUNT" "uzig"; then
        return 1
    fi

    # Get module balances before transfer using common utilities
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)
    MODULE_IBC_BEFORE=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance before: $MODULE_UZIG_BEFORE"
    echo "📈 Module IBC balance before: $MODULE_IBC_BEFORE"

    # Get transfer statistics before transfer using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # WITHDRAW ALL IBC tokens from module to simulate insufficient balance
    echo '💸 Withdrawing ALL IBC tokens from module wallet to simulate insufficient balance...'
    TOTAL_MODULE_IBC_BALANCE=$(get_module_ibc_balance "$IBC_HASH")
    WITHDRAW_AMOUNT="${TOTAL_MODULE_IBC_BALANCE}${IBC_HASH}"
    echo "💸 Withdrawing $WITHDRAW_AMOUNT from module wallet (all available IBC tokens)..."
    WITHDRAW_CMD="zigchaind tx tokenwrapper withdraw-from-module-wallet $WITHDRAW_AMOUNT --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$WITHDRAW_CMD" "Withdraw all IBC tokens from module wallet")

    # Wait for withdrawal to process
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify module now has insufficient IBC balance (should be 0)
    MODULE_IBC_AFTER_WITHDRAWAL=$(get_module_ibc_balance "$IBC_HASH")
    if [[ $(echo "$MODULE_IBC_AFTER_WITHDRAWAL > 0" | bc -l) -eq 1 ]]; then
        echo "❌ Module should have zero IBC balance after withdrawing all tokens"
        echo "   Expected: 0 IBC tokens"
        echo "   Current: $MODULE_IBC_AFTER_WITHDRAWAL IBC tokens"
        return 1
    fi
    echo "✅ Module now has zero IBC balance (insufficient for transfer)"

    # Send tokens back from ZIGChain to Axelar
    echo '🚀 Attempting return transfer from ZIGChain to Axelar with insufficient IBC balance...'
    AXELAR_TEST_ADDRESS=$(axelard keys show $TEST_ACCOUNT_NAME -a 2>/dev/null)
    TRANSFER_CMD="zigchaind tx ibc-transfer transfer transfer $ZIGCHAIN_AXELAR_CHANNEL_ID $AXELAR_TEST_ADDRESS $ZIG_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"

    # Execute the transfer and expect it to fail
    if execute_tx_expect_failure "$TRANSFER_CMD" "Return transfer from ZIGChain to Axelar (insufficient IBC)"; then
        echo "✅ Transfer correctly failed due to insufficient IBC balance"
    else
        echo "❌ Transfer should have failed but succeeded"
        return 1
    fi

    # Wait for any processing that might have occurred
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Get balances after transfer using common utilities
    echo '📊 Getting balances after return transfer...'
    ZIGCHAIN_BOB_UZIG_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_AFTER=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "unit-zig")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance after: $ZIGCHAIN_BOB_UZIG_AFTER"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance after: $AXELAR_BOB_AMOUNT_AFTER unit-zig"

    # Get module balances after transfer using common utilities
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)
    MODULE_IBC_AFTER=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance after: $MODULE_UZIG_AFTER"
    echo "📈 Module IBC balance after: $MODULE_IBC_AFTER"

    # Get transfer statistics after transfer using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    echo "📊 Transfer stats after: In=$TRANSFERRED_IN_AFTER, Out=$TRANSFERRED_OUT_AFTER"

    # Check that sender balance decreased only by fees (transfer should fail)
    echo '🔍 Checking sender balance after failed transfer...'
    EXPECTED_ZIGCHAIN_BOB_UZIG_AFTER=$(echo "$ZIGCHAIN_BOB_UZIG_BEFORE - $FEES_AMOUNT_NUMERIC" | bc)

    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME uzig" "$ZIGCHAIN_BOB_UZIG_BEFORE" "$ZIGCHAIN_BOB_UZIG_AFTER" "$EXPECTED_ZIGCHAIN_BOB_UZIG_AFTER"

    if [[ "$ZIGCHAIN_BOB_UZIG_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_UZIG_AFTER" ]]; then
        echo "❌ Sender should only lose fees ($FEES_AMOUNT_NUMERIC uzig) but balance changed from $ZIGCHAIN_BOB_UZIG_BEFORE to $ZIGCHAIN_BOB_UZIG_AFTER"
        return 1
    fi
    echo "✅ Sender balance decreased only by transaction fees"

    # Check that receiver balance remains unchanged
    log_balance_comparison "Axelar $TEST_ACCOUNT_NAME" "$AXELAR_BOB_AMOUNT_BEFORE" "$AXELAR_BOB_AMOUNT_AFTER" "$AXELAR_BOB_AMOUNT_BEFORE"

    if [[ "$AXELAR_BOB_AMOUNT_BEFORE" != "$AXELAR_BOB_AMOUNT_AFTER" ]]; then
        echo "❌ Receiver balance should remain unchanged but changed from $AXELAR_BOB_AMOUNT_BEFORE to $AXELAR_BOB_AMOUNT_AFTER"
        return 1
    fi
    echo "✅ Receiver balance unchanged"

    # Check that module wallet balances remain unchanged
    echo '🔍 Checking module wallet balances after failed transfer...'
    if [[ "$MODULE_UZIG_BEFORE" != "$MODULE_UZIG_AFTER" ]]; then
        echo "❌ Module uzig balance changed from $MODULE_UZIG_BEFORE to $MODULE_UZIG_AFTER"
        return 1
    fi
    if [[ "$MODULE_IBC_AFTER_WITHDRAWAL" != "$MODULE_IBC_AFTER" ]]; then
        echo "❌ Module IBC balance changed from $MODULE_IBC_AFTER_WITHDRAWAL to $MODULE_IBC_AFTER"
        return 1
    fi
    echo "✅ Module wallet balances unchanged"

    # RESTORE the IBC balance back to the module wallet
    echo '🔄 Restoring IBC balance to module wallet...'
    ZIGCHAIN_OP_ADDRESS=$(zigchaind keys show $OP_ACCOUNT_NAME -a)
    RESTORE_AMOUNT="${MODULE_IBC_BEFORE}${IBC_HASH}"
    echo "💰 Restoring $RESTORE_AMOUNT to module wallet (original balance)..."
    FUND_CMD="zigchaind tx tokenwrapper fund-module-wallet $RESTORE_AMOUNT --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$FUND_CMD" "Restore IBC tokens to module wallet")

    # Wait for restoration to process
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC balance is restored
    MODULE_IBC_AFTER_RESTORE=$(get_module_ibc_balance "$IBC_HASH")
    if [[ "$MODULE_IBC_AFTER_RESTORE" != "$MODULE_IBC_BEFORE" ]]; then
        echo "❌ Module IBC balance should be restored to $MODULE_IBC_BEFORE but got $MODULE_IBC_AFTER_RESTORE"
        return 1
    fi
    echo "✅ Module IBC balance restored successfully"

    # Validate that transfer statistics remain unchanged (since transfer failed completely)
    [[ "$TRANSFERRED_IN_BEFORE" == "$TRANSFERRED_IN_AFTER" ]] && \
    [[ "$TRANSFERRED_OUT_BEFORE" == "$TRANSFERRED_OUT_AFTER" ]]
}

# Test 012: Return transfer from ZIGChain to Axelar with insufficient IBC balance
run_test "TW-012" "Return transfer from ZIGChain to Axelar with insufficient IBC balance (should fail entirely)" "test_012_return_transfer_insufficient_ibc"

test_013_return_transfer_wrong_native_client_id() {
    echo '🚫 Testing return transfer from ZIGChain to Axelar with wrong native client id (should fail)...'

    # Setup test environment
    setup_test_env

    # Pre-transfer validation
    if ! pre_transfer_validation "ZIGChain → Axelar (wrong native client id)"; then
        return 1
    fi

    # Verify module is funded with sufficient uzig tokens using common utilities
    echo '🔍 Verifying module wallet has sufficient tokens for return transfer...'
    REQUIRED_UZIG_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    REQUIRED_IBC_AMOUNT=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    IBC_HASH=$(get_ibc_hash)

    if ! validate_module_uzig_balance "$REQUIRED_UZIG_AMOUNT"; then
        return 1
    fi

    # Verify module has sufficient IBC tokens
    MODULE_IBC_BALANCE=$(get_module_ibc_balance "$IBC_HASH")
    if ! [[ $(echo "$MODULE_IBC_BALANCE >= $REQUIRED_IBC_AMOUNT" | bc -l) -eq 1 ]]; then
        echo "❌ Insufficient IBC balance in module wallet."
        echo "   Required: >= $REQUIRED_IBC_AMOUNT IBC tokens"
        echo "   Current: $MODULE_IBC_BALANCE IBC tokens"
        return 1
    fi

    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before return transfer...'
    ZIGCHAIN_BOB_UZIG_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    ZIGCHAIN_BOB_IBC_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")
    AXELAR_BOB_AMOUNT_BEFORE=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "unit-zig")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance before: $ZIGCHAIN_BOB_UZIG_BEFORE"
    echo "📈 Zigchain $TEST_ACCOUNT_NAME IBC balance before: $ZIGCHAIN_BOB_IBC_BEFORE"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance before: $AXELAR_BOB_AMOUNT_BEFORE unit-zig"

    # Verify account has sufficient uzig balance for return transfer (including fees)
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    FEES_AMOUNT_NUMERIC=$(get_fee_numeric "$ZIGCHAIN_FEES")
    REQUIRED_ZIG_AMOUNT=$(echo "$TRANSFER_AMOUNT_NUMERIC + $FEES_AMOUNT_NUMERIC" | bc)

    if ! validate_sufficient_balance "zigchain" "$TEST_ACCOUNT_NAME" "$REQUIRED_ZIG_AMOUNT" "uzig"; then
        return 1
    fi

    # Get module balances before transfer using common utilities
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)
    MODULE_IBC_BEFORE=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance before: $MODULE_UZIG_BEFORE"
    echo "📈 Module IBC balance before: $MODULE_IBC_BEFORE"

    # Get transfer statistics before transfer using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # Store expected IBC settings for restoration
    EXPECTED_NATIVE_CLIENT_ID=${EXPECTED_NATIVE_CLIENT_ID:-'07-tendermint-1'}
    EXPECTED_COUNTERPARTY_CLIENT_ID=${EXPECTED_COUNTERPARTY_CLIENT_ID:-'07-tendermint-0'}
    EXPECTED_NATIVE_PORT=${EXPECTED_NATIVE_PORT:-'transfer'}
    EXPECTED_COUNTERPARTY_PORT=${EXPECTED_COUNTERPARTY_PORT:-'transfer'}
    EXPECTED_NATIVE_CHANNEL=${EXPECTED_NATIVE_CHANNEL:-'channel-1'}
    EXPECTED_COUNTERPARTY_CHANNEL=${EXPECTED_COUNTERPARTY_CHANNEL:-'channel-0'}
    EXPECTED_DENOM=${EXPECTED_DENOM:-$AXELAR_DENOM}
    EXPECTED_DECIMAL_DIFFERENCE=${EXPECTED_DECIMAL_DIFFERENCE:-'12'}

    # Update IBC settings with wrong native client ID
    echo '⚙️ Setting wrong native client id...'
    WRONG_NATIVE_CLIENT_ID="07-tendermint-999"

    UPDATE_CMD="zigchaind tx tokenwrapper update-ibc-settings $WRONG_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $EXPECTED_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$UPDATE_CMD" "Update IBC settings with wrong native client ID")

    # Wait for settings update
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were updated with wrong native client id
    UPDATED_NATIVE_CLIENT_ID=$(echo "$(get_module_info)" | jq -r '.native_client_id')
    if [[ "$UPDATED_NATIVE_CLIENT_ID" != "$WRONG_NATIVE_CLIENT_ID" ]]; then
        echo "❌ Native client ID should be $WRONG_NATIVE_CLIENT_ID but got $UPDATED_NATIVE_CLIENT_ID"
        return 1
    fi
    echo "✅ IBC settings updated with wrong native client id"

    # Attempt to send tokens back from ZIGChain to Axelar with wrong native client id (should fail)
    echo '🚫 Attempting return transfer from ZIGChain to Axelar with wrong native client id (should fail)...'
    AXELAR_TEST_ADDRESS=$(axelard keys show $TEST_ACCOUNT_NAME -a 2>/dev/null)
    TRANSFER_CMD="zigchaind tx ibc-transfer transfer transfer $ZIGCHAIN_AXELAR_CHANNEL_ID $AXELAR_TEST_ADDRESS $ZIG_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    if ! execute_tx_expect_failure "$TRANSFER_CMD" "Return transfer from ZIGChain to Axelar (wrong native client id - should fail)"; then
        echo "❌ IBC transfer should have failed but succeeded"
        return 1
    fi

    # Wait for any processing that might have occurred
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"

    # Get balances after transfer using common utilities
    echo '📊 Getting balances after return transfer...'
    ZIGCHAIN_BOB_UZIG_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_AFTER=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "unit-zig")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance after: $ZIGCHAIN_BOB_UZIG_AFTER"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance after: $AXELAR_BOB_AMOUNT_AFTER unit-zig"

    # Get module balances after transfer using common utilities
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)
    MODULE_IBC_AFTER=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance after: $MODULE_UZIG_AFTER"
    echo "📈 Module IBC balance after: $MODULE_IBC_AFTER"

    # Get transfer statistics after transfer using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    echo "📊 Transfer stats after: In=$TRANSFERRED_IN_AFTER, Out=$TRANSFERRED_OUT_AFTER"

    # Check that module wallet balances remain unchanged
    echo '🔍 Checking module wallet balances after transfer...'
    if [[ "$MODULE_UZIG_BEFORE" != "$MODULE_UZIG_AFTER" ]]; then
        echo "❌ Module uzig balance changed from $MODULE_UZIG_BEFORE to $MODULE_UZIG_AFTER"
        return 1
    fi
    if [[ "$MODULE_IBC_BEFORE" != "$MODULE_IBC_AFTER" ]]; then
        echo "❌ Module IBC balance changed from $MODULE_IBC_BEFORE to $MODULE_IBC_AFTER"
        return 1
    fi
    echo "✅ Module wallet balances unchanged"

    # Check that zigchain account lost only transaction fees (transfer should have failed)
    echo '🔍 Checking zigchain account balance after failed transfer...'
    EXPECTED_ZIGCHAIN_BOB_UZIG=$(echo "$ZIGCHAIN_BOB_UZIG_BEFORE - $FEES_AMOUNT_NUMERIC" | bc)

    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME uzig" "$ZIGCHAIN_BOB_UZIG_BEFORE" "$ZIGCHAIN_BOB_UZIG_AFTER" "$EXPECTED_ZIGCHAIN_BOB_UZIG"

    if [[ "$ZIGCHAIN_BOB_UZIG_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_UZIG" ]]; then
        echo "❌ Zigchain account should have lost only fees ($FEES_AMOUNT_NUMERIC uzig) but balance changed from $ZIGCHAIN_BOB_UZIG_BEFORE to $ZIGCHAIN_BOB_UZIG_AFTER"
        return 1
    fi
    echo "✅ Zigchain account lost only transaction fees"

    # Check that zigchain account IBC balance did not change (transfer failed completely)
    echo '🔍 Checking zigchain account IBC balance after failed transfer...'
    ZIGCHAIN_BOB_IBC_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_HASH")

    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME IBC" "$ZIGCHAIN_BOB_IBC_BEFORE" "$ZIGCHAIN_BOB_IBC_AFTER" "$ZIGCHAIN_BOB_IBC_BEFORE"

    if [[ "$ZIGCHAIN_BOB_IBC_AFTER" != "$ZIGCHAIN_BOB_IBC_BEFORE" ]]; then
        echo "❌ Zigchain account IBC balance should remain unchanged since transfer failed, but changed from $ZIGCHAIN_BOB_IBC_BEFORE to $ZIGCHAIN_BOB_IBC_AFTER"
        return 1
    fi
    echo "✅ Zigchain account IBC balance unchanged"

    # Check that axelar account balance remains unchanged
    log_balance_comparison "Axelar $TEST_ACCOUNT_NAME" "$AXELAR_BOB_AMOUNT_BEFORE" "$AXELAR_BOB_AMOUNT_AFTER" "$AXELAR_BOB_AMOUNT_BEFORE"

    if [[ "$AXELAR_BOB_AMOUNT_BEFORE" != "$AXELAR_BOB_AMOUNT_AFTER" ]]; then
        echo "❌ Axelar account balance should remain unchanged but changed from $AXELAR_BOB_AMOUNT_BEFORE to $AXELAR_BOB_AMOUNT_AFTER"
        return 1
    fi
    echo "✅ Axelar account balance unchanged"

    # Restore IBC settings to correct values
    echo '🔄 Restoring IBC settings to correct values...'
    RESTORE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $EXPECTED_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$RESTORE_CMD" "Restore correct IBC settings")

    # Wait for restoration
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were restored
    FINAL_NATIVE_CLIENT_ID=$(echo "$(get_module_info)" | jq -r '.native_client_id')
    if [[ "$FINAL_NATIVE_CLIENT_ID" != "$EXPECTED_NATIVE_CLIENT_ID" ]]; then
        echo "❌ Native client ID should be restored to $EXPECTED_NATIVE_CLIENT_ID but got $FINAL_NATIVE_CLIENT_ID"
        return 1
    fi
    echo "✅ IBC settings restored to correct values"

    # Validate that transfer statistics remain unchanged (since transfer failed completely)
    [[ "$TRANSFERRED_IN_BEFORE" == "$TRANSFERRED_IN_AFTER" ]] && \
    [[ "$TRANSFERRED_OUT_BEFORE" == "$TRANSFERRED_OUT_AFTER" ]]
}

# Test 013: Return transfer from ZIGChain to Axelar with wrong native client id
run_test "TW-013" "Return transfer from ZIGChain to Axelar with wrong native client id (should fail completely)" "test_013_return_transfer_wrong_native_client_id"

test_014_return_transfer_wrong_native_channel_id() {
    echo '↩️ Testing return transfer from ZIGChain to Axelar with wrong native channel id...'

    # Setup test environment
    setup_test_env

    # Pre-transfer validation
    if ! pre_transfer_validation "ZIGChain → Axelar (wrong native channel id)"; then
        return 1
    fi

    # Verify module is funded with sufficient uzig tokens using common utilities
    echo '🔍 Verifying module wallet has sufficient tokens for return transfer...'
    REQUIRED_UZIG_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    REQUIRED_IBC_AMOUNT=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    IBC_HASH=$(get_ibc_hash)

    if ! validate_module_uzig_balance "$REQUIRED_UZIG_AMOUNT"; then
        return 1
    fi

    # Verify module has sufficient IBC tokens
    MODULE_IBC_BALANCE=$(get_module_ibc_balance "$IBC_HASH")
    if ! [[ $(echo "$MODULE_IBC_BALANCE >= $REQUIRED_IBC_AMOUNT" | bc -l) -eq 1 ]]; then
        echo "❌ Insufficient IBC balance in module wallet."
        echo "   Required: >= $REQUIRED_IBC_AMOUNT $IBC_HASH tokens"
        echo "   Current: $MODULE_IBC_BALANCE $IBC_HASH tokens"
        return 1
    fi

    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    # Use the IBC hash based on expected settings
    IBC_VOUCHER_DENOM="ibc/$(calculate_ibc_hash "transfer" "$AXELAR_ZIGCHAIN_CHANNEL_ID" "uzig")"

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before return transfer...'
    ZIGCHAIN_BOB_UZIG_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_BEFORE=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "$IBC_VOUCHER_DENOM")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance before: $ZIGCHAIN_BOB_UZIG_BEFORE"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance before: $AXELAR_BOB_AMOUNT_BEFORE $IBC_VOUCHER_DENOM"

    # Verify account has sufficient uzig balance for return transfer (including fees)
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    FEES_AMOUNT_NUMERIC=$(get_fee_numeric "$ZIGCHAIN_FEES")
    REQUIRED_ZIG_AMOUNT=$(echo "$TRANSFER_AMOUNT_NUMERIC + $FEES_AMOUNT_NUMERIC" | bc)

    if ! validate_sufficient_balance "zigchain" "$TEST_ACCOUNT_NAME" "$REQUIRED_ZIG_AMOUNT" "uzig"; then
        return 1
    fi

    # Get module balances before transfer using common utilities
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)
    MODULE_IBC_BEFORE=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance before: $MODULE_UZIG_BEFORE"
    echo "📈 Module IBC balance before: $MODULE_IBC_BEFORE"

    # Get transfer statistics before transfer using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # Store expected IBC settings for restoration
    EXPECTED_NATIVE_CLIENT_ID=${EXPECTED_NATIVE_CLIENT_ID:-'07-tendermint-1'}
    EXPECTED_COUNTERPARTY_CLIENT_ID=${EXPECTED_COUNTERPARTY_CLIENT_ID:-'07-tendermint-0'}
    EXPECTED_NATIVE_PORT=${EXPECTED_NATIVE_PORT:-'transfer'}
    EXPECTED_COUNTERPARTY_PORT=${EXPECTED_COUNTERPARTY_PORT:-'transfer'}
    EXPECTED_NATIVE_CHANNEL=${EXPECTED_NATIVE_CHANNEL:-'channel-1'}
    EXPECTED_COUNTERPARTY_CHANNEL=${EXPECTED_COUNTERPARTY_CHANNEL:-'channel-0'}
    EXPECTED_DENOM=${EXPECTED_DENOM:-$AXELAR_DENOM}
    EXPECTED_DECIMAL_DIFFERENCE=${EXPECTED_DECIMAL_DIFFERENCE:-'12'}

    # Update IBC settings with wrong native channel ID
    echo '⚙️ Setting wrong native channel id...'
    WRONG_NATIVE_CHANNEL_ID="channel-999"

    UPDATE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $WRONG_NATIVE_CHANNEL_ID $EXPECTED_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$UPDATE_CMD" "Update IBC settings with wrong native channel ID")

    # Wait for settings update
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were updated with wrong native channel id
    UPDATED_NATIVE_CHANNEL_ID=$(echo "$(get_module_info)" | jq -r '.native_channel')
    if [[ "$UPDATED_NATIVE_CHANNEL_ID" != "$WRONG_NATIVE_CHANNEL_ID" ]]; then
        echo "❌ Native channel ID should be $WRONG_NATIVE_CHANNEL_ID but got $UPDATED_NATIVE_CHANNEL_ID"
        return 1
    fi
    echo "✅ IBC settings updated with wrong native channel id"

    # Attempt to send tokens back from ZIGChain to Axelar with wrong native channel id (should succeed as regular IBC transfer)
    echo '✅ Attempting return transfer from ZIGChain to Axelar with wrong native channel id (should succeed as regular IBC transfer)...'
    AXELAR_TEST_ADDRESS=$(axelard keys show $TEST_ACCOUNT_NAME -a 2>/dev/null)
    TRANSFER_CMD="zigchaind tx ibc-transfer transfer transfer $ZIGCHAIN_AXELAR_CHANNEL_ID $AXELAR_TEST_ADDRESS $ZIG_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "Return transfer from ZIGChain to Axelar (wrong native channel id - should succeed as regular IBC transfer)")

    # Wait for any processing that might have occurred
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"

    # Get balances after transfer using common utilities
    echo '📊 Getting balances after return transfer...'
    ZIGCHAIN_BOB_UZIG_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_AFTER=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "$IBC_VOUCHER_DENOM")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance after: $ZIGCHAIN_BOB_UZIG_AFTER"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance after: $AXELAR_BOB_AMOUNT_AFTER $IBC_VOUCHER_DENOM"

    # Get module balances after transfer using common utilities
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)
    MODULE_IBC_AFTER=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance after: $MODULE_UZIG_AFTER"
    echo "📈 Module IBC balance after: $MODULE_IBC_AFTER"

    # Get transfer statistics after transfer using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    echo "📊 Transfer stats after: In=$TRANSFERRED_IN_AFTER, Out=$TRANSFERRED_OUT_AFTER"

    # Check that module wallet balances remain unchanged
    echo '🔍 Checking module wallet balances after transfer...'
    if [[ "$MODULE_UZIG_BEFORE" != "$MODULE_UZIG_AFTER" ]]; then
        echo "❌ Module uzig balance changed from $MODULE_UZIG_BEFORE to $MODULE_UZIG_AFTER"
        return 1
    fi
    if [[ "$MODULE_IBC_BEFORE" != "$MODULE_IBC_AFTER" ]]; then
        echo "❌ Module IBC balance changed from $MODULE_IBC_BEFORE to $MODULE_IBC_AFTER"
        return 1
    fi
    echo "✅ Module wallet balances unchanged"

    # Check that zigchain account lost transfer amount + transaction fees (transfer should have succeeded)
    echo '🔍 Checking zigchain account balance after successful transfer...'
    EXPECTED_ZIGCHAIN_BOB_UZIG=$(echo "$ZIGCHAIN_BOB_UZIG_BEFORE - $TRANSFER_AMOUNT_NUMERIC - $FEES_AMOUNT_NUMERIC" | bc)

    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME uzig" "$ZIGCHAIN_BOB_UZIG_BEFORE" "$ZIGCHAIN_BOB_UZIG_AFTER" "$EXPECTED_ZIGCHAIN_BOB_UZIG"

    if [[ "$ZIGCHAIN_BOB_UZIG_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_UZIG" ]]; then
        echo "❌ Zigchain account should have lost transfer amount ($TRANSFER_AMOUNT_NUMERIC uzig) + fees ($FEES_AMOUNT_NUMERIC uzig) but balance changed from $ZIGCHAIN_BOB_UZIG_BEFORE to $ZIGCHAIN_BOB_UZIG_AFTER"
        return 1
    fi
    echo "✅ Zigchain account lost transfer amount + transaction fees"

    # Check that axelar account balance increases by IBC vouchers
    # Since native channel ID is wrong, this should be a regular IBC transfer creating vouchers
    EXPECTED_AXELAR_BOB_VOUCHER=$(echo "$AXELAR_BOB_AMOUNT_BEFORE + $TRANSFER_AMOUNT_NUMERIC" | bc)

    log_balance_comparison "Axelar $TEST_ACCOUNT_NAME IBC vouchers ($IBC_VOUCHER_DENOM)" "$AXELAR_BOB_AMOUNT_BEFORE" "$AXELAR_BOB_AMOUNT_AFTER" "$EXPECTED_AXELAR_BOB_VOUCHER"

    if [[ "$AXELAR_BOB_AMOUNT_AFTER" != "$EXPECTED_AXELAR_BOB_VOUCHER" ]]; then
        echo "❌ Axelar account should have received IBC vouchers ($TRANSFER_AMOUNT_NUMERIC $IBC_VOUCHER_DENOM) but balance is $AXELAR_BOB_AMOUNT_AFTER"
        return 1
    fi
    echo "✅ Axelar account received IBC vouchers"

    # Restore IBC settings to correct values
    echo '🔄 Restoring IBC settings to correct values...'
    RESTORE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $EXPECTED_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$RESTORE_CMD" "Restore correct IBC settings")

    # Wait for restoration
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were restored
    FINAL_NATIVE_CHANNEL_ID=$(echo "$(get_module_info)" | jq -r '.native_channel')
    if [[ "$FINAL_NATIVE_CHANNEL_ID" != "$EXPECTED_NATIVE_CHANNEL" ]]; then
        echo "❌ Native channel ID should be restored to $EXPECTED_NATIVE_CHANNEL but got $FINAL_NATIVE_CHANNEL_ID"
        return 1
    fi
    echo "✅ IBC settings restored to correct values"

    # Validate that transfer statistics remain unchanged (since transfer failed completely)
    [[ "$TRANSFERRED_IN_BEFORE" == "$TRANSFERRED_IN_AFTER" ]] && \
    [[ "$TRANSFERRED_OUT_BEFORE" == "$TRANSFERRED_OUT_AFTER" ]]
}

# Test 014: Return transfer from ZIGChain to Axelar with wrong native channel id
run_test "TW-014" "Return transfer from ZIGChain to Axelar with wrong native channel id (should succeed as regular IBC transfer)" "test_014_return_transfer_wrong_native_channel_id"

test_015_return_transfer_wrong_counterparty_channel() {
    echo '↩️ Testing return transfer from ZIGChain to Axelar with wrong counterparty channel...'

    # Setup test environment
    setup_test_env

    # Pre-transfer validation
    if ! pre_transfer_validation "ZIGChain → Axelar (wrong counterparty channel)"; then
        return 1
    fi

    # Verify module is funded with sufficient uzig tokens using common utilities
    echo '🔍 Verifying module wallet has sufficient tokens for return transfer...'
    REQUIRED_UZIG_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    REQUIRED_IBC_AMOUNT=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    IBC_HASH=$(get_ibc_hash)

    if ! validate_module_uzig_balance "$REQUIRED_UZIG_AMOUNT"; then
        return 1
    fi

    # Verify module has sufficient IBC tokens
    MODULE_IBC_BALANCE=$(get_module_ibc_balance "$IBC_HASH")
    if ! [[ $(echo "$MODULE_IBC_BALANCE >= $REQUIRED_IBC_AMOUNT" | bc -l) -eq 1 ]]; then
        echo "❌ Insufficient IBC balance in module wallet."
        echo "   Required: >= $REQUIRED_IBC_AMOUNT IBC tokens"
        echo "   Current: $MODULE_IBC_BALANCE IBC tokens"
        return 1
    fi

    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before return transfer...'
    ZIGCHAIN_BOB_UZIG_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_BEFORE=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "unit-zig")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance before: $ZIGCHAIN_BOB_UZIG_BEFORE"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance before: $AXELAR_BOB_AMOUNT_BEFORE unit-zig"

    # Verify account has sufficient uzig balance for return transfer (including fees)
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    FEES_AMOUNT_NUMERIC=$(get_fee_numeric "$ZIGCHAIN_FEES")
    REQUIRED_ZIG_AMOUNT=$(echo "$TRANSFER_AMOUNT_NUMERIC + $FEES_AMOUNT_NUMERIC" | bc)

    if ! validate_sufficient_balance "zigchain" "$TEST_ACCOUNT_NAME" "$REQUIRED_ZIG_AMOUNT" "uzig"; then
        return 1
    fi

    # Get module balances before transfer using common utilities
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)
    MODULE_IBC_BEFORE=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance before: $MODULE_UZIG_BEFORE"
    echo "📈 Module IBC balance before: $MODULE_IBC_BEFORE"

    # Get transfer statistics before transfer using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # Store expected IBC settings for restoration
    EXPECTED_NATIVE_CLIENT_ID=${EXPECTED_NATIVE_CLIENT_ID:-'07-tendermint-1'}
    EXPECTED_COUNTERPARTY_CLIENT_ID=${EXPECTED_COUNTERPARTY_CLIENT_ID:-'07-tendermint-0'}
    EXPECTED_NATIVE_PORT=${EXPECTED_NATIVE_PORT:-'transfer'}
    EXPECTED_COUNTERPARTY_PORT=${EXPECTED_COUNTERPARTY_PORT:-'transfer'}
    EXPECTED_NATIVE_CHANNEL=${EXPECTED_NATIVE_CHANNEL:-'channel-1'}
    EXPECTED_COUNTERPARTY_CHANNEL=${EXPECTED_COUNTERPARTY_CHANNEL:-'channel-0'}
    EXPECTED_DENOM=${EXPECTED_DENOM:-$AXELAR_DENOM}
    EXPECTED_DECIMAL_DIFFERENCE=${EXPECTED_DECIMAL_DIFFERENCE:-'12'}

    # Update IBC settings with wrong counterparty channel
    echo '⚙️ Setting wrong counterparty channel...'
    WRONG_COUNTERPARTY_CHANNEL="channel-999"

    UPDATE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $WRONG_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$UPDATE_CMD" "Update IBC settings with wrong counterparty channel")

    # Wait for settings update
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were updated with wrong counterparty channel
    UPDATED_COUNTERPARTY_CHANNEL=$(echo "$(get_module_info)" | jq -r '.counterparty_channel')
    if [[ "$UPDATED_COUNTERPARTY_CHANNEL" != "$WRONG_COUNTERPARTY_CHANNEL" ]]; then
        echo "❌ Counterparty channel should be $WRONG_COUNTERPARTY_CHANNEL but got $UPDATED_COUNTERPARTY_CHANNEL"
        return 1
    fi
    echo "✅ IBC settings updated with wrong counterparty channel"

    # Send tokens back from ZIGChain to Axelar
    echo '🚀 Attempting return transfer from ZIGChain to Axelar with wrong counterparty channel...'
    AXELAR_TEST_ADDRESS=$(axelard keys show $TEST_ACCOUNT_NAME -a 2>/dev/null)
    TRANSFER_CMD="zigchaind tx ibc-transfer transfer transfer $ZIGCHAIN_AXELAR_CHANNEL_ID $AXELAR_TEST_ADDRESS $ZIG_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "Return transfer from ZIGChain to Axelar (wrong counterparty channel)")

    # Wait for return transfer transaction to be processed
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Get balances after transfer using common utilities
    echo '📊 Getting balances after return transfer...'
    ZIGCHAIN_BOB_UZIG_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_AFTER=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "unit-zig")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance after: $ZIGCHAIN_BOB_UZIG_AFTER"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance after: $AXELAR_BOB_AMOUNT_AFTER unit-zig"

    # Get module balances after transfer using common utilities
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)
    MODULE_IBC_AFTER=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance after: $MODULE_UZIG_AFTER"
    echo "📈 Module IBC balance after: $MODULE_IBC_AFTER"

    # Get transfer statistics after transfer using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    echo "📊 Transfer stats after: In=$TRANSFERRED_IN_AFTER, Out=$TRANSFERRED_OUT_AFTER"

    # Check that module wallet balances remain unchanged
    echo '🔍 Checking module wallet balances after transfer...'
    if [[ "$MODULE_UZIG_BEFORE" != "$MODULE_UZIG_AFTER" ]]; then
        echo "❌ Module uzig balance changed from $MODULE_UZIG_BEFORE to $MODULE_UZIG_AFTER"
        return 1
    fi
    if [[ "$MODULE_IBC_BEFORE" != "$MODULE_IBC_AFTER" ]]; then
        echo "❌ Module IBC balance changed from $MODULE_IBC_BEFORE to $MODULE_IBC_AFTER"
        return 1
    fi
    echo "✅ Module wallet balances unchanged"

    # Check that sender balance only decreased by fees (transfer should fail completely)
    echo '🔍 Checking sender balance after failed transfer...'
    EXPECTED_ZIGCHAIN_BOB_UZIG=$(echo "$ZIGCHAIN_BOB_UZIG_BEFORE - $FEES_AMOUNT_NUMERIC" | bc)

    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME uzig" "$ZIGCHAIN_BOB_UZIG_BEFORE" "$ZIGCHAIN_BOB_UZIG_AFTER" "$EXPECTED_ZIGCHAIN_BOB_UZIG"

    if [[ "$ZIGCHAIN_BOB_UZIG_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_UZIG" ]]; then
        echo "❌ Sender balance should only decrease by fees ($FEES_AMOUNT_NUMERIC) but changed from $ZIGCHAIN_BOB_UZIG_BEFORE to $ZIGCHAIN_BOB_UZIG_AFTER"
        return 1
    fi
    echo "✅ Sender balance only decreased by transaction fees"

    # Check that axelar account balance remains unchanged
    log_balance_comparison "Axelar $TEST_ACCOUNT_NAME" "$AXELAR_BOB_AMOUNT_BEFORE" "$AXELAR_BOB_AMOUNT_AFTER" "$AXELAR_BOB_AMOUNT_BEFORE"

    if [[ "$AXELAR_BOB_AMOUNT_BEFORE" != "$AXELAR_BOB_AMOUNT_AFTER" ]]; then
        echo "❌ Axelar account balance should remain unchanged but changed from $AXELAR_BOB_AMOUNT_BEFORE to $AXELAR_BOB_AMOUNT_AFTER"
        return 1
    fi
    echo "✅ Axelar account balance unchanged"

    # Restore IBC settings to correct values
    echo '🔄 Restoring IBC settings to correct values...'
    RESTORE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $EXPECTED_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$RESTORE_CMD" "Restore correct IBC settings")

    # Wait for restoration
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were restored
    FINAL_COUNTERPARTY_CHANNEL=$(echo "$(get_module_info)" | jq -r '.counterparty_channel')
    if [[ "$FINAL_COUNTERPARTY_CHANNEL" != "$EXPECTED_COUNTERPARTY_CHANNEL" ]]; then
        echo "❌ Counterparty channel should be restored to $EXPECTED_COUNTERPARTY_CHANNEL but got $FINAL_COUNTERPARTY_CHANNEL"
        return 1
    fi
    echo "✅ IBC settings restored to correct values"

    # Validate that transfer statistics remain unchanged (since transfer failed completely)
    [[ "$TRANSFERRED_IN_BEFORE" == "$TRANSFERRED_IN_AFTER" ]] && \
    [[ "$TRANSFERRED_OUT_BEFORE" == "$TRANSFERRED_OUT_AFTER" ]]
}

# Test 015: Return transfer from ZIGChain to Axelar with wrong counterparty channel
run_test "TW-015" "Return transfer from ZIGChain to Axelar with wrong counterparty channel (should fail completely)" "test_015_return_transfer_wrong_counterparty_channel"

test_016_return_transfer_wrong_module_denom() {
    echo '↩️ Testing return transfer from ZIGChain to Axelar with wrong module denom...'

    # Setup test environment
    setup_test_env

    # Pre-transfer validation
    if ! pre_transfer_validation "ZIGChain → Axelar (wrong module denom)"; then
        return 1
    fi

    # Verify module is funded with sufficient uzig tokens using common utilities
    echo '🔍 Verifying module wallet has sufficient tokens for return transfer...'
    REQUIRED_UZIG_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    REQUIRED_IBC_AMOUNT=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    IBC_HASH=$(get_ibc_hash)

    if ! validate_module_uzig_balance "$REQUIRED_UZIG_AMOUNT"; then
        return 1
    fi

    # Verify module has sufficient IBC tokens
    MODULE_IBC_BALANCE=$(get_module_ibc_balance "$IBC_HASH")
    if ! [[ $(echo "$MODULE_IBC_BALANCE >= $REQUIRED_IBC_AMOUNT" | bc -l) -eq 1 ]]; then
        echo "❌ Insufficient IBC balance in module wallet."
        echo "   Required: >= $REQUIRED_IBC_AMOUNT IBC tokens"
        echo "   Current: $MODULE_IBC_BALANCE IBC tokens"
        return 1
    fi

    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    # Use the IBC hash based on expected settings
    IBC_VOUCHER_DENOM="ibc/$(calculate_ibc_hash "$EXPECTED_COUNTERPARTY_PORT" "$EXPECTED_COUNTERPARTY_CHANNEL" "uzig")"

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before return transfer...'
    ZIGCHAIN_BOB_UZIG_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_BEFORE=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "$IBC_VOUCHER_DENOM")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance before: $ZIGCHAIN_BOB_UZIG_BEFORE"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance before: $AXELAR_BOB_AMOUNT_BEFORE $IBC_VOUCHER_DENOM"

    # Verify account has sufficient uzig balance for return transfer (including fees)
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    FEES_AMOUNT_NUMERIC=$(get_fee_numeric "$ZIGCHAIN_FEES")
    REQUIRED_ZIG_AMOUNT=$(echo "$TRANSFER_AMOUNT_NUMERIC + $FEES_AMOUNT_NUMERIC" | bc)

    if ! validate_sufficient_balance "zigchain" "$TEST_ACCOUNT_NAME" "$REQUIRED_ZIG_AMOUNT" "uzig"; then
        return 1
    fi

    # Get module balances before transfer using common utilities
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)
    MODULE_IBC_BEFORE=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance before: $MODULE_UZIG_BEFORE"
    echo "📈 Module IBC balance before: $MODULE_IBC_BEFORE"

    # Get transfer statistics before transfer using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # Store expected IBC settings for restoration
    EXPECTED_NATIVE_CLIENT_ID=${EXPECTED_NATIVE_CLIENT_ID:-'07-tendermint-1'}
    EXPECTED_COUNTERPARTY_CLIENT_ID=${EXPECTED_COUNTERPARTY_CLIENT_ID:-'07-tendermint-0'}
    EXPECTED_NATIVE_PORT=${EXPECTED_NATIVE_PORT:-'transfer'}
    EXPECTED_COUNTERPARTY_PORT=${EXPECTED_COUNTERPARTY_PORT:-'transfer'}
    EXPECTED_NATIVE_CHANNEL=${EXPECTED_NATIVE_CHANNEL:-'channel-1'}
    EXPECTED_COUNTERPARTY_CHANNEL=${EXPECTED_COUNTERPARTY_CHANNEL:-'channel-0'}
    EXPECTED_DENOM=${EXPECTED_DENOM:-$AXELAR_DENOM}
    EXPECTED_DECIMAL_DIFFERENCE=${EXPECTED_DECIMAL_DIFFERENCE:-'12'}

    # Update IBC settings with wrong module denom
    echo '⚙️ Setting wrong module denom...'
    WRONG_MODULE_DENOM="wrong-denom"

    UPDATE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $EXPECTED_COUNTERPARTY_CHANNEL $WRONG_MODULE_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$UPDATE_CMD" "Update IBC settings with wrong module denom")

    # Wait for settings update
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were updated with wrong module denom
    UPDATED_MODULE_DENOM=$(echo "$(get_module_info)" | jq -r '.denom')
    if [[ "$UPDATED_MODULE_DENOM" != "$WRONG_MODULE_DENOM" ]]; then
        echo "❌ Module denom should be $WRONG_MODULE_DENOM but got $UPDATED_MODULE_DENOM"
        return 1
    fi
    echo "✅ IBC settings updated with wrong module denom"

    # Attempt to send tokens back from ZIGChain to Axelar with wrong module denom (should fail)
    echo '🚫 Attempting return transfer from ZIGChain to Axelar with wrong module denom (should fail)...'
    AXELAR_TEST_ADDRESS=$(axelard keys show $TEST_ACCOUNT_NAME -a 2>/dev/null)
    TRANSFER_CMD="zigchaind tx ibc-transfer transfer transfer $ZIGCHAIN_AXELAR_CHANNEL_ID $AXELAR_TEST_ADDRESS $ZIG_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    
    # Get ibc hash with wrong module denom
    WRONG_IBC_HASH="ibc/$(calculate_ibc_hash "$EXPECTED_NATIVE_PORT" "$EXPECTED_NATIVE_CHANNEL" "$WRONG_MODULE_DENOM")"

    # Construct expected error message
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$AXELAR_TRANSFER_AMOUNT")
    EXPECTED_ERROR_MSG="failed to execute message; message index: 0: module does not have enough balance of ${TRANSFER_AMOUNT_NUMERIC}${WRONG_IBC_HASH}"
    
    # Execute transfer expecting failure with specific error message
    if ! execute_tx_expect_failure_and_verify "zigchain" "$TRANSFER_CMD" "Return transfer from ZIGChain to Axelar (wrong module denom - should fail)" "$EXPECTED_ERROR_MSG"; then
        echo "❌ Transfer should have failed with expected error message but didn't"
        return 1
    fi
    echo "✅ Transfer failed with expected error message"

    # Get balances after transfer using common utilities
    echo '📊 Getting balances after return transfer...'
    ZIGCHAIN_BOB_UZIG_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    AXELAR_BOB_AMOUNT_AFTER=$(get_account_balance "axelar" "$TEST_ACCOUNT_NAME" "$IBC_VOUCHER_DENOM")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance after: $ZIGCHAIN_BOB_UZIG_AFTER"
    echo "📈 Axelar $TEST_ACCOUNT_NAME balance after: $AXELAR_BOB_AMOUNT_AFTER $IBC_VOUCHER_DENOM"

    # Get module balances after transfer using common utilities
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)
    MODULE_IBC_AFTER=$(get_module_ibc_balance "$IBC_HASH")

    echo "📈 Module uzig balance after: $MODULE_UZIG_AFTER"
    echo "📈 Module IBC balance after: $MODULE_IBC_AFTER"

    # Get transfer statistics after transfer using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    echo "📊 Transfer stats after: In=$TRANSFERRED_IN_AFTER, Out=$TRANSFERRED_OUT_AFTER"

    # Check that module wallet balances remain unchanged
    echo '🔍 Checking module wallet balances after transfer...'
    if [[ "$MODULE_UZIG_BEFORE" != "$MODULE_UZIG_AFTER" ]]; then
        echo "❌ Module uzig balance changed from $MODULE_UZIG_BEFORE to $MODULE_UZIG_AFTER"
        return 1
    fi
    if [[ "$MODULE_IBC_BEFORE" != "$MODULE_IBC_AFTER" ]]; then
        echo "❌ Module IBC balance changed from $MODULE_IBC_BEFORE to $MODULE_IBC_AFTER"
        return 1
    fi
    echo "✅ Module wallet balances unchanged"

    # Check that zigchain account only lost transaction fees (transfer failed)
    echo '🔍 Checking zigchain account balance after failed transfer...'
    EXPECTED_ZIGCHAIN_BOB_UZIG=$(echo "$ZIGCHAIN_BOB_UZIG_BEFORE - $FEES_AMOUNT_NUMERIC" | bc)

    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME uzig" "$ZIGCHAIN_BOB_UZIG_BEFORE" "$ZIGCHAIN_BOB_UZIG_AFTER" "$EXPECTED_ZIGCHAIN_BOB_UZIG"

    if [[ "$ZIGCHAIN_BOB_UZIG_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_UZIG" ]]; then
        echo "❌ Zigchain account should have only lost fees ($FEES_AMOUNT_NUMERIC uzig) but balance changed from $ZIGCHAIN_BOB_UZIG_BEFORE to $ZIGCHAIN_BOB_UZIG_AFTER"
        return 1
    fi
    echo "✅ Zigchain account only lost transaction fees"

    # Check that axelar account balance remains unchanged (transfer failed)
    echo '🔍 Checking axelar account balance after failed transfer...'
    if [[ "$AXELAR_BOB_AMOUNT_AFTER" != "$AXELAR_BOB_AMOUNT_BEFORE" ]]; then
        echo "❌ Axelar account balance should remain unchanged but changed from $AXELAR_BOB_AMOUNT_BEFORE to $AXELAR_BOB_AMOUNT_AFTER"
        return 1
    fi
    echo "✅ Axelar account balance unchanged"

    # Restore IBC settings to correct values
    echo '🔄 Restoring IBC settings to correct values...'
    RESTORE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $EXPECTED_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$RESTORE_CMD" "Restore correct IBC settings")

    # Wait for restoration
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were restored
    FINAL_MODULE_DENOM=$(echo "$(get_module_info)" | jq -r '.denom')
    if [[ "$FINAL_MODULE_DENOM" != "$EXPECTED_DENOM" ]]; then
        echo "❌ Module denom should be restored to $EXPECTED_DENOM but got $FINAL_MODULE_DENOM"
        return 1
    fi
    echo "✅ IBC settings restored to correct values"

    # Validate that transfer statistics remain unchanged (since transfer failed completely)
    [[ "$TRANSFERRED_IN_BEFORE" == "$TRANSFERRED_IN_AFTER" ]] && \
    [[ "$TRANSFERRED_OUT_BEFORE" == "$TRANSFERRED_OUT_AFTER" ]]
}

# Test 016: Return transfer from ZIGChain to Axelar with wrong module denom
run_test "TW-016" "Return transfer from ZIGChain to Axelar with wrong module denom (should fail completely)" "test_016_return_transfer_wrong_module_denom"

test_017_zigchain_to_cosmos() {
    echo '🌍 Testing transfer from ZIGChain to Cosmos...'
    echo '   This test verifies that the receiver on Cosmos receives IBC vouchers equivalent to the uzig sent from ZIGChain'

    # Setup test environment
    setup_test_env

    # Pre-transfer validation
    if ! pre_transfer_validation "ZIGChain → Cosmos"; then
        return 1
    fi

    # Verify module is funded with sufficient uzig tokens using common utilities
    echo '🔍 Verifying module wallet has sufficient tokens for transfer...'
    REQUIRED_UZIG_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")

    if ! validate_module_uzig_balance "$REQUIRED_UZIG_AMOUNT"; then
        return 1
    fi

    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    # Calculate IBC voucher denomination for the transfer
    IBC_VOUCHER_DENOM="ibc/$(calculate_ibc_hash "transfer" "$COSMOS_ZIGCHAIN_CHANNEL_ID" "uzig")"
    echo "🔍 IBC voucher denomination on Cosmos: $IBC_VOUCHER_DENOM"

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before transfer...'
    ZIGCHAIN_BOB_UZIG_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    COSMOS_BOB_UZIG_BEFORE=$(get_account_balance "cosmos" "$TEST_ACCOUNT_NAME" "uzig")
    COSMOS_BOB_IBC_VOUCHER_BEFORE=$(get_account_balance "cosmos" "$TEST_ACCOUNT_NAME" "$IBC_VOUCHER_DENOM")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance before: $ZIGCHAIN_BOB_UZIG_BEFORE"
    echo "📈 Cosmos $TEST_ACCOUNT_NAME uzig balance before: $COSMOS_BOB_UZIG_BEFORE"
    echo "📈 Cosmos $TEST_ACCOUNT_NAME IBC voucher balance before: $COSMOS_BOB_IBC_VOUCHER_BEFORE $IBC_VOUCHER_DENOM"

    # Verify account has sufficient uzig balance for transfer (including fees)
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    FEES_AMOUNT_NUMERIC=$(get_fee_numeric "$ZIGCHAIN_FEES")
    REQUIRED_ZIG_AMOUNT=$(echo "$TRANSFER_AMOUNT_NUMERIC + $FEES_AMOUNT_NUMERIC" | bc)

    if ! validate_sufficient_balance "zigchain" "$TEST_ACCOUNT_NAME" "$REQUIRED_ZIG_AMOUNT" "uzig"; then
        return 1
    fi

    # Get module balances before transfer using common utilities
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)

    echo "📈 Module uzig balance before: $MODULE_UZIG_BEFORE"

    # Get transfer statistics before transfer using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # Send tokens from ZIGChain to Cosmos
    echo '🚀 Sending tokens from ZIGChain to Cosmos...'
    COSMOS_TEST_ADDRESS=$(cosmosd keys show $TEST_ACCOUNT_NAME -a 2>/dev/null)
    TRANSFER_CMD="zigchaind tx ibc-transfer transfer transfer $ZIGCHAIN_COSMOS_CHANNEL_ID $COSMOS_TEST_ADDRESS $ZIG_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "Transfer from ZIGChain to Cosmos")

    # Wait for transfer transaction to be processed
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Get balances after transfer using common utilities
    echo '📊 Getting balances after transfer...'
    ZIGCHAIN_BOB_UZIG_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    COSMOS_BOB_UZIG_AFTER=$(get_account_balance "cosmos" "$TEST_ACCOUNT_NAME" "uzig")
    COSMOS_BOB_IBC_VOUCHER_AFTER=$(get_account_balance "cosmos" "$TEST_ACCOUNT_NAME" "$IBC_VOUCHER_DENOM")

    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance after: $ZIGCHAIN_BOB_UZIG_AFTER"
    echo "📈 Cosmos $TEST_ACCOUNT_NAME uzig balance after: $COSMOS_BOB_UZIG_AFTER"
    echo "📈 Cosmos $TEST_ACCOUNT_NAME IBC voucher balance after: $COSMOS_BOB_IBC_VOUCHER_AFTER $IBC_VOUCHER_DENOM"

    # Get module balances after transfer using common utilities
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)

    echo "📈 Module uzig balance after: $MODULE_UZIG_AFTER"

    # Get transfer statistics after transfer using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    echo "📊 Transfer stats after: In=$TRANSFERRED_IN_AFTER, Out=$TRANSFERRED_OUT_AFTER"

    # Calculate expected values using common utilities
    ZIG_TRANSFER_NUMERIC="$TRANSFER_AMOUNT_NUMERIC"
    EXPECTED_ZIGCHAIN_BOB_UZIG=$(echo "$ZIGCHAIN_BOB_UZIG_BEFORE - $ZIG_TRANSFER_NUMERIC - $FEES_AMOUNT_NUMERIC" | bc)
    EXPECTED_MODULE_UZIG="$MODULE_UZIG_BEFORE"  # Module should remain the same
    EXPECTED_TRANSFERRED_IN="$TRANSFERRED_IN_BEFORE"
    EXPECTED_TRANSFERRED_OUT="$TRANSFERRED_OUT_BEFORE"
    
    # Calculate expected IBC voucher balance on Cosmos
    EXPECTED_COSMOS_BOB_IBC_VOUCHER=$(echo "$COSMOS_BOB_IBC_VOUCHER_BEFORE + $ZIG_TRANSFER_NUMERIC" | bc)

    # Log balance comparisons using common utility
    echo '🔍 Validating balance changes...'
    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME uzig" "$ZIGCHAIN_BOB_UZIG_BEFORE" "$ZIGCHAIN_BOB_UZIG_AFTER" "$EXPECTED_ZIGCHAIN_BOB_UZIG"
    log_balance_comparison "Module uzig" "$MODULE_UZIG_BEFORE" "$MODULE_UZIG_AFTER" "$EXPECTED_MODULE_UZIG"
    log_balance_comparison "Transferred in" "$TRANSFERRED_IN_BEFORE" "$TRANSFERRED_IN_AFTER" "$EXPECTED_TRANSFERRED_IN"
    log_balance_comparison "Transferred out" "$TRANSFERRED_OUT_BEFORE" "$TRANSFERRED_OUT_AFTER" "$EXPECTED_TRANSFERRED_OUT"
    log_balance_comparison "Cosmos $TEST_ACCOUNT_NAME IBC vouchers ($IBC_VOUCHER_DENOM)" "$COSMOS_BOB_IBC_VOUCHER_BEFORE" "$COSMOS_BOB_IBC_VOUCHER_AFTER" "$EXPECTED_COSMOS_BOB_IBC_VOUCHER"

    # Validate all balance changes and transfer statistics
    if [[ "$COSMOS_BOB_IBC_VOUCHER_AFTER" != "$EXPECTED_COSMOS_BOB_IBC_VOUCHER" ]]; then
        echo "❌ Cosmos account should have received IBC vouchers ($ZIG_TRANSFER_NUMERIC $IBC_VOUCHER_DENOM) but balance is $COSMOS_BOB_IBC_VOUCHER_AFTER"
        return 1
    fi
    echo "✅ Cosmos account received IBC vouchers equivalent to the uzig sent from ZIGChain"
    
    [[ "$ZIGCHAIN_BOB_UZIG_AFTER" == "$EXPECTED_ZIGCHAIN_BOB_UZIG" ]] && \
    [[ "$MODULE_UZIG_AFTER" == "$EXPECTED_MODULE_UZIG" ]] && \
    [[ "$TRANSFERRED_IN_AFTER" == "$EXPECTED_TRANSFERRED_IN" ]] && \
    [[ "$TRANSFERRED_OUT_AFTER" == "$EXPECTED_TRANSFERRED_OUT" ]]
}

# Test 017: Send uzig amount from ZIGChain to Cosmos
run_test "TW-017" "Send uzig amount from ZIGChain to Cosmos" "test_017_zigchain_to_cosmos"

test_018_cosmos_to_zigchain() {
    echo '🌍 Testing transfer from Cosmos to ZIGChain...'
    echo '   This test verifies that IBC vouchers are sent from Cosmos to ZIGChain and properly converted to native uzig tokens'

    # Setup test environment
    setup_test_env

    # Pre-transfer validation
    if ! pre_transfer_validation "Cosmos → ZIGChain"; then
        return 1
    fi

    # Verify module is funded with sufficient uzig tokens using common utilities
    echo '🔍 Verifying module wallet has sufficient tokens for transfer...'
    REQUIRED_UZIG_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")

    if ! validate_module_uzig_balance "$REQUIRED_UZIG_AMOUNT"; then
        return 1
    fi

    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    # Calculate IBC voucher denomination for the transfer
    IBC_VOUCHER_DENOM="ibc/$(calculate_ibc_hash "transfer" "$COSMOS_ZIGCHAIN_CHANNEL_ID" "uzig")"
    echo "🔍 IBC voucher denomination on Cosmos: $IBC_VOUCHER_DENOM"

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before transfer...'
    COSMOS_BOB_IBC_VOUCHER_BEFORE=$(get_account_balance "cosmos" "$TEST_ACCOUNT_NAME" "$IBC_VOUCHER_DENOM")
    COSMOS_BOB_UATOM_BEFORE=$(get_account_balance "cosmos" "$TEST_ACCOUNT_NAME" "uatom")
    ZIGCHAIN_BOB_UZIG_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")

    echo "📈 Cosmos $TEST_ACCOUNT_NAME IBC voucher balance before: $COSMOS_BOB_IBC_VOUCHER_BEFORE $IBC_VOUCHER_DENOM"
    echo "📈 Cosmos $TEST_ACCOUNT_NAME uatom balance before: $COSMOS_BOB_UATOM_BEFORE"
    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance before: $ZIGCHAIN_BOB_UZIG_BEFORE"

    # Verify account has sufficient IBC voucher balance for transfer (including fees)
    REQUIRED_COSMOS_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    REQUIRED_UATOM_AMOUNT=$(get_fee_numeric "$COSMOS_FEES")

    echo "📈 Required Cosmos amount: $REQUIRED_COSMOS_AMOUNT"
    echo "📈 Required uatom amount: $REQUIRED_UATOM_AMOUNT"

    if ! validate_sufficient_balance "cosmos" "$TEST_ACCOUNT_NAME" "$REQUIRED_COSMOS_AMOUNT" "$IBC_VOUCHER_DENOM"; then
        return 1
    fi

    if ! validate_sufficient_balance "cosmos" "$TEST_ACCOUNT_NAME" "$REQUIRED_UATOM_AMOUNT" "uatom"; then
        return 1
    fi

    # Get module balances before transfer using common utilities
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)

    echo "📈 Module uzig balance before: $MODULE_UZIG_BEFORE"

    # Get transfer statistics before transfer using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # Send tokens from Cosmos to ZIGChain
    echo '🚀 Sending tokens from Cosmos to ZIGChain...'
    ZIGCHAIN_TEST_ADDRESS=$(zigchaind keys show $TEST_ACCOUNT_NAME -a)
    TRANSFER_CMD="cosmosd tx ibc-transfer transfer transfer $COSMOS_ZIGCHAIN_CHANNEL_ID $ZIGCHAIN_TEST_ADDRESS ${REQUIRED_COSMOS_AMOUNT}${IBC_VOUCHER_DENOM} --from $TEST_ACCOUNT_NAME --fees $COSMOS_FEES --chain-id $COSMOS_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "Transfer from Cosmos to ZIGChain")

    # Wait for transfer transaction to be processed
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Get balances after transfer using common utilities
    echo '📊 Getting balances after transfer...'
    COSMOS_BOB_IBC_VOUCHER_AFTER=$(get_account_balance "cosmos" "$TEST_ACCOUNT_NAME" "$IBC_VOUCHER_DENOM")
    COSMOS_BOB_UATOM_AFTER=$(get_account_balance "cosmos" "$TEST_ACCOUNT_NAME" "uatom")
    ZIGCHAIN_BOB_UZIG_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")

    echo "📈 Cosmos $TEST_ACCOUNT_NAME IBC voucher balance after: $COSMOS_BOB_IBC_VOUCHER_AFTER $IBC_VOUCHER_DENOM"
    echo "📈 Cosmos $TEST_ACCOUNT_NAME uatom balance after: $COSMOS_BOB_UATOM_AFTER"
    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance after: $ZIGCHAIN_BOB_UZIG_AFTER"

    # Get module balances after transfer using common utilities
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)

    echo "📈 Module uzig balance after: $MODULE_UZIG_AFTER"

    # Get transfer statistics after transfer using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    echo "📊 Transfer stats after: In=$TRANSFERRED_IN_AFTER, Out=$TRANSFERRED_OUT_AFTER"

    # Calculate expected values using common utilities
    COSMOS_TRANSFER_NUMERIC=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")
    EXPECTED_COSMOS_BOB_IBC_VOUCHER=$(echo "$COSMOS_BOB_IBC_VOUCHER_BEFORE - $COSMOS_TRANSFER_NUMERIC" | bc)
    EXPECTED_ZIGCHAIN_BOB_UZIG=$(echo "$ZIGCHAIN_BOB_UZIG_BEFORE + $COSMOS_TRANSFER_NUMERIC" | bc)
    EXPECTED_MODULE_UZIG="$MODULE_UZIG_BEFORE" # Module should remain the same
    EXPECTED_TRANSFERRED_IN="$TRANSFERRED_IN_BEFORE" # Transferred in should remain the same
    EXPECTED_TRANSFERRED_OUT="$TRANSFERRED_OUT_BEFORE"

    # Log balance comparisons using common utility
    echo '🔍 Validating balance changes...'
    log_balance_comparison "Cosmos $TEST_ACCOUNT_NAME IBC vouchers ($IBC_VOUCHER_DENOM)" "$COSMOS_BOB_IBC_VOUCHER_BEFORE" "$COSMOS_BOB_IBC_VOUCHER_AFTER" "$EXPECTED_COSMOS_BOB_IBC_VOUCHER"
    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME uzig" "$ZIGCHAIN_BOB_UZIG_BEFORE" "$ZIGCHAIN_BOB_UZIG_AFTER" "$EXPECTED_ZIGCHAIN_BOB_UZIG"
    log_balance_comparison "Module uzig" "$MODULE_UZIG_BEFORE" "$MODULE_UZIG_AFTER" "$EXPECTED_MODULE_UZIG"
    log_balance_comparison "Transferred in" "$TRANSFERRED_IN_BEFORE" "$TRANSFERRED_IN_AFTER" "$EXPECTED_TRANSFERRED_IN"
    log_balance_comparison "Transferred out" "$TRANSFERRED_OUT_BEFORE" "$TRANSFERRED_OUT_AFTER" "$EXPECTED_TRANSFERRED_OUT"

    # Validate all balance changes and transfer statistics
    if [[ "$COSMOS_BOB_IBC_VOUCHER_AFTER" != "$EXPECTED_COSMOS_BOB_IBC_VOUCHER" ]]; then
        echo "❌ Cosmos account should have burned IBC vouchers ($COSMOS_TRANSFER_NUMERIC $IBC_VOUCHER_DENOM) but balance is $COSMOS_BOB_IBC_VOUCHER_AFTER"
        return 1
    fi
    echo "✅ Cosmos account burned IBC vouchers as expected"
    
    if [[ "$ZIGCHAIN_BOB_UZIG_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_UZIG" ]]; then
        echo "❌ ZIGChain account should have received native uzig tokens ($COSMOS_TRANSFER_NUMERIC uzig) but balance is $ZIGCHAIN_BOB_UZIG_AFTER"
        return 1
    fi
    echo "✅ ZIGChain account received native uzig tokens as expected"
    
    [[ "$MODULE_UZIG_AFTER" == "$EXPECTED_MODULE_UZIG" ]] && \
    [[ "$TRANSFERRED_IN_AFTER" == "$EXPECTED_TRANSFERRED_IN" ]] && \
    [[ "$TRANSFERRED_OUT_AFTER" == "$EXPECTED_TRANSFERRED_OUT" ]]
}

# Test 018: Send IBC uzig vouchers from Cosmos to ZIGChain
run_test "TW-018" "Send IBC uzig vouchers from Cosmos to ZIGChain" "test_018_cosmos_to_zigchain"

test_019_cosmos_to_zigchain_atom() {
    echo '🌍 Testing transfer of uatom from Cosmos to ZIGChain...'
    echo '   This test verifies that uatom is sent from Cosmos to ZIGChain and the receiver gets IBC vouchers (not native uzig tokens)'

    # Setup test environment
    setup_test_env

    # Pre-transfer validation
    if ! pre_transfer_validation "Cosmos → ZIGChain (uatom)"; then
        return 1
    fi

    # Verify module is funded with sufficient uzig tokens using common utilities
    echo '🔍 Verifying module wallet has sufficient tokens for transfer...'
    REQUIRED_UZIG_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")

    if ! validate_module_uzig_balance "$REQUIRED_UZIG_AMOUNT"; then
        return 1
    fi

    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    # Calculate IBC voucher denomination for uatom transfer
    IBC_VOUCHER_DENOM="ibc/$(calculate_ibc_hash "transfer" "$ZIGCHAIN_COSMOS_CHANNEL_ID" "uatom")"
    echo "🔍 IBC voucher denomination for uatom on ZIGChain: $IBC_VOUCHER_DENOM"

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before transfer...'
    COSMOS_BOB_UATOM_BEFORE=$(get_account_balance "cosmos" "$TEST_ACCOUNT_NAME" "uatom")
    ZIGCHAIN_BOB_UZIG_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    ZIGCHAIN_BOB_IBC_VOUCHER_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_VOUCHER_DENOM")

    echo "📈 Cosmos $TEST_ACCOUNT_NAME uatom balance before: $COSMOS_BOB_UATOM_BEFORE"
    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance before: $ZIGCHAIN_BOB_UZIG_BEFORE"
    echo "📈 Zigchain $TEST_ACCOUNT_NAME IBC voucher balance before: $ZIGCHAIN_BOB_IBC_VOUCHER_BEFORE $IBC_VOUCHER_DENOM"

    # Verify account has sufficient uatom balance for transfer (including fees)
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$COSMOS_TRANSFER_AMOUNT")
    FEES_AMOUNT_NUMERIC=$(get_fee_numeric "$COSMOS_FEES")
    REQUIRED_COSMOS_AMOUNT=$(echo "$TRANSFER_AMOUNT_NUMERIC + $FEES_AMOUNT_NUMERIC" | bc)

    if ! validate_sufficient_balance "cosmos" "$TEST_ACCOUNT_NAME" "$REQUIRED_COSMOS_AMOUNT" "uatom"; then
        return 1
    fi

    # Get module balances before transfer using common utilities
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)

    echo "📈 Module uzig balance before: $MODULE_UZIG_BEFORE"

    # Get transfer statistics before transfer using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # Send uatom tokens from Cosmos to ZIGChain
    echo '🚀 Sending uatom tokens from Cosmos to ZIGChain...'
    ZIGCHAIN_TEST_ADDRESS=$(zigchaind keys show $TEST_ACCOUNT_NAME -a)
    TRANSFER_CMD="cosmosd tx ibc-transfer transfer transfer $COSMOS_ZIGCHAIN_CHANNEL_ID $ZIGCHAIN_TEST_ADDRESS $COSMOS_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $COSMOS_FEES --chain-id $COSMOS_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "Transfer uatom from Cosmos to ZIGChain")

    # Wait for transfer transaction to be processed
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Get balances after transfer using common utilities
    echo '📊 Getting balances after transfer...'
    COSMOS_BOB_UATOM_AFTER=$(get_account_balance "cosmos" "$TEST_ACCOUNT_NAME" "uatom")
    ZIGCHAIN_BOB_UZIG_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    ZIGCHAIN_BOB_IBC_VOUCHER_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_VOUCHER_DENOM")

    echo "📈 Cosmos $TEST_ACCOUNT_NAME uatom balance after: $COSMOS_BOB_UATOM_AFTER"
    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance after: $ZIGCHAIN_BOB_UZIG_AFTER"
    echo "📈 Zigchain $TEST_ACCOUNT_NAME IBC voucher balance after: $ZIGCHAIN_BOB_IBC_VOUCHER_AFTER $IBC_VOUCHER_DENOM"

    # Get module balances after transfer using common utilities
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)

    echo "📈 Module uzig balance after: $MODULE_UZIG_AFTER"

    # Get transfer statistics after transfer using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    echo "📊 Transfer stats after: In=$TRANSFERRED_IN_AFTER, Out=$TRANSFERRED_OUT_AFTER"

    # Calculate expected values using common utilities
    COSMOS_TRANSFER_NUMERIC=$(amount_to_numeric "$COSMOS_TRANSFER_AMOUNT")
    EXPECTED_COSMOS_BOB_UATOM=$(echo "$COSMOS_BOB_UATOM_BEFORE - $COSMOS_TRANSFER_NUMERIC - $FEES_AMOUNT_NUMERIC" | bc)
    EXPECTED_ZIGCHAIN_BOB_UZIG="$ZIGCHAIN_BOB_UZIG_BEFORE"  # uzig balance should remain unchanged
    EXPECTED_ZIGCHAIN_BOB_IBC_VOUCHER=$(echo "$ZIGCHAIN_BOB_IBC_VOUCHER_BEFORE + $COSMOS_TRANSFER_NUMERIC" | bc)  # IBC vouchers should increase
    EXPECTED_MODULE_UZIG="$MODULE_UZIG_BEFORE"  # Module should remain the same (no token wrapper processing for uatom)
    EXPECTED_TRANSFERRED_IN="$TRANSFERRED_IN_BEFORE"  # No token wrapper processing
    EXPECTED_TRANSFERRED_OUT="$TRANSFERRED_OUT_BEFORE"

    # Log balance comparisons using common utility
    echo '🔍 Validating balance changes...'
    log_balance_comparison "Cosmos $TEST_ACCOUNT_NAME uatom" "$COSMOS_BOB_UATOM_BEFORE" "$COSMOS_BOB_UATOM_AFTER" "$EXPECTED_COSMOS_BOB_UATOM"
    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME uzig" "$ZIGCHAIN_BOB_UZIG_BEFORE" "$ZIGCHAIN_BOB_UZIG_AFTER" "$EXPECTED_ZIGCHAIN_BOB_UZIG"
    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME IBC vouchers ($IBC_VOUCHER_DENOM)" "$ZIGCHAIN_BOB_IBC_VOUCHER_BEFORE" "$ZIGCHAIN_BOB_IBC_VOUCHER_AFTER" "$EXPECTED_ZIGCHAIN_BOB_IBC_VOUCHER"
    log_balance_comparison "Module uzig" "$MODULE_UZIG_BEFORE" "$MODULE_UZIG_AFTER" "$EXPECTED_MODULE_UZIG"
    log_balance_comparison "Transferred in" "$TRANSFERRED_IN_BEFORE" "$TRANSFERRED_IN_AFTER" "$EXPECTED_TRANSFERRED_IN"
    log_balance_comparison "Transferred out" "$TRANSFERRED_OUT_BEFORE" "$TRANSFERRED_OUT_AFTER" "$EXPECTED_TRANSFERRED_OUT"

    # Validate all balance changes and transfer statistics
    if [[ "$ZIGCHAIN_BOB_UZIG_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_UZIG" ]]; then
        echo "❌ ZIGChain account uzig balance should remain unchanged but changed from $ZIGCHAIN_BOB_UZIG_BEFORE to $ZIGCHAIN_BOB_UZIG_AFTER"
        return 1
    fi
    echo "✅ ZIGChain account uzig balance remained unchanged as expected"
    
    if [[ "$ZIGCHAIN_BOB_IBC_VOUCHER_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_IBC_VOUCHER" ]]; then
        echo "❌ ZIGChain account should have received IBC vouchers ($COSMOS_TRANSFER_NUMERIC $IBC_VOUCHER_DENOM) but balance is $ZIGCHAIN_BOB_IBC_VOUCHER_AFTER"
        return 1
    fi
    echo "✅ ZIGChain account received IBC vouchers for uatom as expected"
    
    [[ "$COSMOS_BOB_UATOM_AFTER" == "$EXPECTED_COSMOS_BOB_UATOM" ]] && \
    [[ "$MODULE_UZIG_AFTER" == "$EXPECTED_MODULE_UZIG" ]] && \
    [[ "$TRANSFERRED_IN_AFTER" == "$EXPECTED_TRANSFERRED_IN" ]] && \
    [[ "$TRANSFERRED_OUT_AFTER" == "$EXPECTED_TRANSFERRED_OUT" ]]
}

# Test 019: Send uatom amount from Cosmos to ZIGChain
run_test "TW-019" "Send uatom amount from Cosmos to ZIGChain" "test_019_cosmos_to_zigchain_atom"

test_020_dummy_to_zigchain() {
    echo '🌍 Testing transfer of unit-zig from Dummy to ZIGChain...'
    echo '   This test verifies that unit-zig is sent from Dummy to ZIGChain and the receiver gets IBC vouchers (not native uzig tokens)'

    # Setup test environment
    setup_test_env

    # Pre-transfer validation
    if ! pre_transfer_validation "Dummy → ZIGChain"; then
        return 1
    fi

    # Verify module is funded with sufficient uzig tokens using common utilities
    echo '🔍 Verifying module wallet has sufficient tokens for transfer...'
    REQUIRED_UZIG_AMOUNT=$(amount_to_numeric "$ZIG_TRANSFER_AMOUNT")

    if ! validate_module_uzig_balance "$REQUIRED_UZIG_AMOUNT"; then
        return 1
    fi

    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    # Calculate IBC voucher denomination for unit-zig transfer
    IBC_VOUCHER_DENOM="ibc/$(calculate_ibc_hash "transfer" "$ZIGCHAIN_DUMMY_CHANNEL_ID" "unit-zig")"
    echo "🔍 IBC voucher denomination for unit-zig on ZIGChain: $IBC_VOUCHER_DENOM"

    # Get balances before transfer using common utilities
    echo '📊 Getting balances before transfer...'
    DUMMY_BOB_UNIT_ZIG_BEFORE=$(get_account_balance "dummy" "$TEST_ACCOUNT_NAME" "unit-zig")
    ZIGCHAIN_BOB_UZIG_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    ZIGCHAIN_BOB_IBC_VOUCHER_BEFORE=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_VOUCHER_DENOM")

    echo "📈 Dummy $TEST_ACCOUNT_NAME unit-zig balance before: $DUMMY_BOB_UNIT_ZIG_BEFORE"
    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance before: $ZIGCHAIN_BOB_UZIG_BEFORE"
    echo "📈 Zigchain $TEST_ACCOUNT_NAME IBC voucher balance before: $ZIGCHAIN_BOB_IBC_VOUCHER_BEFORE $IBC_VOUCHER_DENOM"

    # Verify account has sufficient unit-zig balance for transfer (including fees)
    TRANSFER_AMOUNT_NUMERIC=$(amount_to_numeric "$DUMMY_TRANSFER_AMOUNT")
    FEES_AMOUNT_NUMERIC=$(get_fee_numeric "$DUMMY_FEES")
    REQUIRED_DUMMY_AMOUNT=$(echo "$TRANSFER_AMOUNT_NUMERIC + $FEES_AMOUNT_NUMERIC" | bc)

    echo "📈 Required Dummy amount: $REQUIRED_DUMMY_AMOUNT"
    echo "📈 Fees amount: $FEES_AMOUNT_NUMERIC"

    if ! validate_sufficient_balance "dummy" "$TEST_ACCOUNT_NAME" "$REQUIRED_DUMMY_AMOUNT" "unit-zig"; then
        return 1
    fi

    # Get module balances before transfer using common utilities
    MODULE_UZIG_BEFORE=$(get_module_uzig_balance)

    echo "📈 Module uzig balance before: $MODULE_UZIG_BEFORE"

    # Get transfer statistics before transfer using common utilities
    TRANSFERRED_IN_BEFORE=$(get_total_transferred_in)
    TRANSFERRED_OUT_BEFORE=$(get_total_transferred_out)
    echo "📊 Transfer stats before: In=$TRANSFERRED_IN_BEFORE, Out=$TRANSFERRED_OUT_BEFORE"

    # Send unit-zig tokens from Dummy to ZIGChain
    echo '🚀 Sending unit-zig tokens from Dummy to ZIGChain...'
    ZIGCHAIN_TEST_ADDRESS=$(zigchaind keys show $TEST_ACCOUNT_NAME -a)
    TRANSFER_CMD="dummyd tx ibc-transfer transfer transfer $DUMMY_ZIGCHAIN_CHANNEL_ID $ZIGCHAIN_TEST_ADDRESS $DUMMY_TRANSFER_AMOUNT --from $TEST_ACCOUNT_NAME --fees $DUMMY_FEES --chain-id $DUMMY_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "Transfer unit-zig from Dummy to ZIGChain")

    # Wait for transfer transaction to be processed
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Get balances after transfer using common utilities
    echo '📊 Getting balances after transfer...'
    DUMMY_BOB_UNIT_ZIG_AFTER=$(get_account_balance "dummy" "$TEST_ACCOUNT_NAME" "unit-zig")
    ZIGCHAIN_BOB_UZIG_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "uzig")
    ZIGCHAIN_BOB_IBC_VOUCHER_AFTER=$(get_account_balance "zigchain" "$TEST_ACCOUNT_NAME" "$IBC_VOUCHER_DENOM")

    echo "📈 Dummy $TEST_ACCOUNT_NAME unit-zig balance after: $DUMMY_BOB_UNIT_ZIG_AFTER"
    echo "📈 Zigchain $TEST_ACCOUNT_NAME uzig balance after: $ZIGCHAIN_BOB_UZIG_AFTER"
    echo "📈 Zigchain $TEST_ACCOUNT_NAME IBC voucher balance after: $ZIGCHAIN_BOB_IBC_VOUCHER_AFTER $IBC_VOUCHER_DENOM"

    # Get module balances after transfer using common utilities
    MODULE_UZIG_AFTER=$(get_module_uzig_balance)

    echo "📈 Module uzig balance after: $MODULE_UZIG_AFTER"

    # Get transfer statistics after transfer using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    echo "📊 Transfer stats after: In=$TRANSFERRED_IN_AFTER, Out=$TRANSFERRED_OUT_AFTER"

    # Calculate expected values using common utilities
    DUMMY_TRANSFER_NUMERIC=$(amount_to_numeric "$DUMMY_TRANSFER_AMOUNT")
    EXPECTED_DUMMY_BOB_UNIT_ZIG=$(echo "$DUMMY_BOB_UNIT_ZIG_BEFORE - $DUMMY_TRANSFER_NUMERIC - $FEES_AMOUNT_NUMERIC" | bc)
    EXPECTED_ZIGCHAIN_BOB_UZIG="$ZIGCHAIN_BOB_UZIG_BEFORE"  # uzig balance should remain unchanged
    EXPECTED_ZIGCHAIN_BOB_IBC_VOUCHER=$(echo "$ZIGCHAIN_BOB_IBC_VOUCHER_BEFORE + $DUMMY_TRANSFER_NUMERIC" | bc)  # IBC vouchers should increase
    EXPECTED_MODULE_UZIG="$MODULE_UZIG_BEFORE"  # Module should remain the same (no token wrapper processing for unit-zig)
    EXPECTED_TRANSFERRED_IN="$TRANSFERRED_IN_BEFORE"  # No token wrapper processing
    EXPECTED_TRANSFERRED_OUT="$TRANSFERRED_OUT_BEFORE"

    # Log balance comparisons using common utility
    echo '🔍 Validating balance changes...'
    log_balance_comparison "Dummy $TEST_ACCOUNT_NAME unit-zig" "$DUMMY_BOB_UNIT_ZIG_BEFORE" "$DUMMY_BOB_UNIT_ZIG_AFTER" "$EXPECTED_DUMMY_BOB_UNIT_ZIG"
    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME uzig" "$ZIGCHAIN_BOB_UZIG_BEFORE" "$ZIGCHAIN_BOB_UZIG_AFTER" "$EXPECTED_ZIGCHAIN_BOB_UZIG"
    log_balance_comparison "Zigchain $TEST_ACCOUNT_NAME IBC vouchers ($IBC_VOUCHER_DENOM)" "$ZIGCHAIN_BOB_IBC_VOUCHER_BEFORE" "$ZIGCHAIN_BOB_IBC_VOUCHER_AFTER" "$EXPECTED_ZIGCHAIN_BOB_IBC_VOUCHER"
    log_balance_comparison "Module uzig" "$MODULE_UZIG_BEFORE" "$MODULE_UZIG_AFTER" "$EXPECTED_MODULE_UZIG"
    log_balance_comparison "Transferred in" "$TRANSFERRED_IN_BEFORE" "$TRANSFERRED_IN_AFTER" "$EXPECTED_TRANSFERRED_IN"
    log_balance_comparison "Transferred out" "$TRANSFERRED_OUT_BEFORE" "$TRANSFERRED_OUT_AFTER" "$EXPECTED_TRANSFERRED_OUT"

    # Validate all balance changes and transfer statistics
    if [[ "$ZIGCHAIN_BOB_UZIG_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_UZIG" ]]; then
        echo "❌ ZIGChain account uzig balance should remain unchanged but changed from $ZIGCHAIN_BOB_UZIG_BEFORE to $ZIGCHAIN_BOB_UZIG_AFTER"
        return 1
    fi
    echo "✅ ZIGChain account uzig balance remained unchanged as expected"
    
    if [[ "$ZIGCHAIN_BOB_IBC_VOUCHER_AFTER" != "$EXPECTED_ZIGCHAIN_BOB_IBC_VOUCHER" ]]; then
        echo "❌ ZIGChain account should have received IBC vouchers ($DUMMY_TRANSFER_NUMERIC $IBC_VOUCHER_DENOM) but balance is $ZIGCHAIN_BOB_IBC_VOUCHER_AFTER"
        return 1
    fi
    echo "✅ ZIGChain account received IBC vouchers for unit-zig as expected"
    
    [[ "$DUMMY_BOB_UNIT_ZIG_AFTER" == "$EXPECTED_DUMMY_BOB_UNIT_ZIG" ]] && \
    [[ "$MODULE_UZIG_AFTER" == "$EXPECTED_MODULE_UZIG" ]] && \
    [[ "$TRANSFERRED_IN_AFTER" == "$EXPECTED_TRANSFERRED_IN" ]] && \
    [[ "$TRANSFERRED_OUT_AFTER" == "$EXPECTED_TRANSFERRED_OUT" ]]
}

# Test 020: Send unit-zig amount from Dummy to ZIGChain (only runs on localnet)
if is_localnet; then
    run_test "TW-020" "Send unit-zig amount from Dummy to ZIGChain" "test_020_dummy_to_zigchain"
else
    echo "⏭️  Skipping TW-020 (Dummy to ZIGChain test) - only runs against localnet"
fi

test_990_timeout() {
    echo '⏰ Testing IBC timeout mechanism...'

    # Setup test environment
    setup_test_env

    # First, ensure the module has enough IBC tokens for the timeout test
    echo '🔄 Ensuring module has enough IBC tokens for timeout test...'

    # Send tokens from axelar to zigchain to ensure module has IBC tokens
    echo '🚀 Sending tokens to zigchain for timeout test...'
    ZIGCHAIN_BOB_ADDRESS=$(zigchaind keys show bob -a)
    TRANSFER_CMD="axelard tx ibc-transfer transfer transfer $AXELAR_ZIGCHAIN_CHANNEL_ID $ZIGCHAIN_BOB_ADDRESS $TIMEOUT_TEST_AMOUNT --from bob --fees $AXELAR_FEES --chain-id $AXELAR_CHAIN_ID -y -o json 2>/dev/null"
    OUTPUT=$(execute_tx "$TRANSFER_CMD" "Setup IBC tokens for timeout test")

    # Wait for the transfer to be processed
    wait_for_tx "$IBC_TRANSFER_TRANSACTION_TIMEOUT"

    # Verify the module has the IBC tokens using common utilities
    echo '🔍 Verifying module has IBC tokens...'
    IBC_HASH=$(get_ibc_hash)
    IBC_BALANCE_BEFORE=$(get_module_ibc_balance "$IBC_HASH")
    NATIVE_BALANCE_BEFORE=$(get_module_uzig_balance)

    # Validate expected balances (these are specific test values)
    [[ $IBC_BALANCE_BEFORE == '11000000000000' ]] && [[ $NATIVE_BALANCE_BEFORE == '989' ]]

    # Stop the relayer to prevent packet delivery
    echo '🛑 Stopping relayer to trigger timeout...'
    kill $RELAYER_PID1 $RELAYER_PID2 || true
    sleep 2

    # Send a transfer with a very short timeout (1 second = 1000000000 nanoseconds)
    echo '⏰ Sending transfer with 1-second timeout...'
    AXELAR_BOB_ADDRESS=$(axelard keys show bob -a 2>/dev/null)
    TIMEOUT_CMD="zigchaind tx ibc-transfer transfer transfer $ZIGCHAIN_AXELAR_CHANNEL_ID $AXELAR_BOB_ADDRESS $TIMEOUT_TEST_NATIVE_AMOUNT --packet-timeout-timestamp $TIMEOUT_TEST_DURATION --from bob --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$TIMEOUT_CMD" "IBC transfer with short timeout")

    # Wait for the timeout to occur (wait longer than the timeout)
    echo '⏳ Waiting for timeout to occur...'
    sleep 2

    # Restart the relayer
    echo '🔄 Restarting relayer...'
    $IGNITE_CMD relayer hermes start $AXELAR_CHAIN_ID $ZIGCHAIN_CHAIN_ID > $WORK_DIR/logs/relayer.log 2>&1 &
    RELAYER_PID=$!

    # Wait for relayer to start
    wait_for_tx "$RELAYER_START_TIMEOUT"

    # Check that the timeout was triggered and refund occurred
    echo '🔍 Checking timeout and refund mechanism...'
    # The refund mechanism should have:
    # 1. Locked the IBC tokens that were sent back to the module
    # 2. Unlocked the native tokens to the sender (bob)
    # 3. The module should have the same IBC balance as before (since the refund locks the IBC tokens back)

    # Get balances after timeout using common utilities
    IBC_BALANCE_AFTER=$(get_module_ibc_balance "$IBC_HASH")
    NATIVE_BALANCE_AFTER=$(get_module_uzig_balance)

    # Validate expected balances after timeout
    [[ $IBC_BALANCE_BEFORE == '11000000000000' ]] && [[ $NATIVE_BALANCE_AFTER == '989' ]]

    # Check bob's balance to see if he received the refund
    BOB_NATIVE_BALANCE=$(get_account_balance "zigchain" "bob" "uzig")
    EXPECTED_BOB_BALANCE_AFTER_TIMEOUT='499999011'  # Same as before timeout test
    [[ $BOB_NATIVE_BALANCE == $EXPECTED_BOB_BALANCE_AFTER_TIMEOUT ]]

    # Check total transfers to see if the timeout was recorded using common utilities
    TRANSFERRED_IN_AFTER=$(get_total_transferred_in)
    TRANSFERRED_OUT_AFTER=$(get_total_transferred_out)
    EXPECTED_IN='21'
    EXPECTED_OUT='10'
    [[ $TRANSFERRED_IN_AFTER == "$EXPECTED_IN" ]] && [[ $TRANSFERRED_OUT_AFTER == "$EXPECTED_OUT" ]]
}

# Test 990: Test IBC timeout mechanism
run_test "TW-990" "Test IBC timeout mechanism" "test_990_timeout"

test_998_set_original_ibc_settings() {
    echo '🔧 Setting original module IBC settings...'

    # Setup test environment
    setup_test_env

    # Extract expected values (with defaults)
    EXPECTED_DECIMAL_DIFFERENCE=${EXPECTED_DECIMAL_DIFFERENCE:-'12'}
    EXPECTED_DENOM=${EXPECTED_DENOM:-$AXELAR_DENOM}
    EXPECTED_NATIVE_CHANNEL=${EXPECTED_NATIVE_CHANNEL:-'channel-1'}
    EXPECTED_NATIVE_PORT=${EXPECTED_NATIVE_PORT:-'transfer'}
    EXPECTED_NATIVE_CLIENT_ID=${EXPECTED_NATIVE_CLIENT_ID:-'07-tendermint-1'}
    EXPECTED_COUNTERPARTY_CHANNEL=${EXPECTED_COUNTERPARTY_CHANNEL:-'channel-0'}
    EXPECTED_COUNTERPARTY_PORT=${EXPECTED_COUNTERPARTY_PORT:-'transfer'}
    EXPECTED_COUNTERPARTY_CLIENT_ID=${EXPECTED_COUNTERPARTY_CLIENT_ID:-'07-tendermint-0'}

    # Update IBC settings to original configuration
    UPDATE_CMD="zigchaind tx tokenwrapper update-ibc-settings $EXPECTED_NATIVE_CLIENT_ID $EXPECTED_COUNTERPARTY_CLIENT_ID $EXPECTED_NATIVE_PORT $EXPECTED_COUNTERPARTY_PORT $EXPECTED_NATIVE_CHANNEL $EXPECTED_COUNTERPARTY_CHANNEL $EXPECTED_DENOM $EXPECTED_DECIMAL_DIFFERENCE --from $OP_ACCOUNT_NAME --fees $ZIGCHAIN_FEES --chain-id $ZIGCHAIN_CHAIN_ID -y -o json"
    OUTPUT=$(execute_tx "$UPDATE_CMD" "Set original IBC settings")

    # Wait for settings update
    wait_for_tx "$TRANSACTION_TIMEOUT"

    # Verify IBC settings were set to original values
    MODULE_INFO=$(get_module_info)

    ACTUAL_NATIVE_CLIENT_ID=$(echo "$MODULE_INFO" | jq -r '.native_client_id')
    ACTUAL_COUNTERPARTY_CLIENT_ID=$(echo "$MODULE_INFO" | jq -r '.counterparty_client_id')
    ACTUAL_NATIVE_PORT=$(echo "$MODULE_INFO" | jq -r '.native_port')
    ACTUAL_COUNTERPARTY_PORT=$(echo "$MODULE_INFO" | jq -r '.counterparty_port')
    ACTUAL_NATIVE_CHANNEL=$(echo "$MODULE_INFO" | jq -r '.native_channel')
    ACTUAL_COUNTERPARTY_CHANNEL=$(echo "$MODULE_INFO" | jq -r '.counterparty_channel')
    ACTUAL_DENOM=$(echo "$MODULE_INFO" | jq -r '.denom')
    ACTUAL_DECIMAL_DIFFERENCE=$(echo "$MODULE_INFO" | jq -r '.decimal_difference')

    # Validate all settings match expected configuration
    [ "$ACTUAL_NATIVE_CLIENT_ID" = "$EXPECTED_NATIVE_CLIENT_ID" ] && \
    [ "$ACTUAL_COUNTERPARTY_CLIENT_ID" = "$EXPECTED_COUNTERPARTY_CLIENT_ID" ] && \
    [ "$ACTUAL_NATIVE_PORT" = "$EXPECTED_NATIVE_PORT" ] && \
    [ "$ACTUAL_COUNTERPARTY_PORT" = "$EXPECTED_COUNTERPARTY_PORT" ] && \
    [ "$ACTUAL_NATIVE_CHANNEL" = "$EXPECTED_NATIVE_CHANNEL" ] && \
    [ "$ACTUAL_COUNTERPARTY_CHANNEL" = "$EXPECTED_COUNTERPARTY_CHANNEL" ] && \
    [ "$ACTUAL_DENOM" = "$EXPECTED_DENOM" ] && \
    [ "$ACTUAL_DECIMAL_DIFFERENCE" = "$EXPECTED_DECIMAL_DIFFERENCE" ]
}

# Test 998: Set original module IBC settings
run_test "TW-998" "Set original module IBC settings" "test_998_set_original_ibc_settings"

test_999_cleanup() {
    echo '🧹 Cleaning up token wrapper module balances...'

    # Setup test environment
    setup_test_env

    # Get module info using common utilities
    MODULE_INFO=$(get_module_info)

    # If no balances, skip cleanup
    if [[ $(echo $MODULE_INFO | jq 'has("balances")') == 'false' ]]; then
        echo '✅ No balances to withdraw - cleanup complete'
        return 0
    fi

    # Withdraw each balance individually
    echo "$MODULE_INFO" | jq -r '.balances[] | "\(.amount)\(.denom)"' | while read -r balance; do
        echo "💸 Withdrawing balance: $balance"
        WITHDRAW_CMD="zigchaind tx tokenwrapper withdraw-from-module-wallet $balance --from $OP_ACCOUNT_NAME -y -o json"
        OUTPUT=$(execute_tx "$WITHDRAW_CMD" "Withdraw $balance from module wallet")

        # Wait for each withdrawal to be processed
        wait_for_tx "$TRANSACTION_TIMEOUT"
    done

    # Verify that balances are empty
    FINAL_MODULE_INFO=$(get_module_info)
    if [[ $(echo $FINAL_MODULE_INFO | jq 'has("balances")') == 'false' ]]; then
        echo '✅ All balances withdrawn successfully - cleanup complete'
        return 0
    else
        echo '❌ Some balances remain after cleanup'
        return 1
    fi
}

# Test 999: Clean up token wrapper module balances
run_test "TW-999" "Clean up token wrapper module balances" "test_999_cleanup"

# Example: To add more tests that should be skipped by default, you can:
# 1. Add them to the SKIP_BY_DEFAULT_TESTS array at the top of the script, or
# 2. Use add_skip_by_default_test() function before defining the test:
#    add_skip_by_default_test "TW-998"
#    run_test "TW-998" "Some other cleanup test" "..."


# Generate final report
generate_report

# Store the exit code
EXIT_CODE=$?

# Cleanup (only if all tests passed AND running all tests AND not running against testnet)
if [ $EXIT_CODE -eq 0 ] && [ "$RUN_SPECIFIC_TESTS" = false ] && [[ "$ZIGCHAIN_NODE" != *"testnet"* && "$AXELAR_NODE" != *"testnet"* ]]; then
    echo ''
    echo '🧹 Cleaning up...'
    kill $AXELAR_PID $ZIGCHAIN_PID $DUMMY_PID $RELAYER_PID1 $RELAYER_PID2 $RELAYER_PID || true
    pkill hermes zigchaind axelard dummyd || true
    echo '✅ Cleanup completed'
elif [[ "$ZIGCHAIN_NODE" == *"testnet"* || "$AXELAR_NODE" == *"testnet"* ]]; then
    echo ''
    echo 'ℹ️  Running against testnet - no cleanup required.'
    echo '   Testnet environment will remain active for additional testing.'
elif [ "$RUN_SPECIFIC_TESTS" = true ]; then
    echo ''
    echo 'ℹ️  Partial test execution completed. Environment preserved for additional testing.'
    echo '   To clean up manually, run: kill $AXELAR_PID $ZIGCHAIN_PID $DUMMY_PID $RELAYER_PID1 $RELAYER_PID2 $RELAYER_PID || true'
    echo '   And: pkill hermes zigchaind axelard dummyd || true'
fi

echo '🏁 Test execution completed!'

# Exit with the appropriate code
exit $EXIT_CODE
