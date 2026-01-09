#!/bin/bash

# Common Test Utilities for TokenWrapper Tests
#
# This file contains common utility functions to reduce redundancy
# across the tokenwrapper test suite.

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

# Calculate IBC hash dynamically
# Usage: calculate_ibc_hash "$native_port" "$native_channel" "$denom"
calculate_ibc_hash() {
    local native_port="$1"
    local native_channel="$2"
    local denom="$3"
    echo -n "${native_port}/${native_channel}/${denom}" | shasum -a 256 | awk '{print $1}' | tr '[:lower:]' '[:upper:]'
}

# =============================================================================
# CHAIN UTILITY FUNCTIONS
# =============================================================================

# Get the appropriate binary for the specified chain
# Usage: get_chain_binary "$chain"
get_chain_binary() {
    local chain="$1"
    
    case "$chain" in
        "zigchain")
            echo "zigchaind"
            ;;
        "axelar")
            echo "axelard"
            ;;
        "cosmos")
            echo "cosmosd"
            ;;
        "dummy")
            echo "dummyd"
            ;;
        *)
            echo "❌ Error: Invalid chain '$chain'. Use 'zigchain' or 'axelar'" >&2
            return 1
            ;;
    esac
}

# =============================================================================
# ACCOUNT VALIDATION FUNCTIONS
# =============================================================================

# Validate that an account exists in zigchain keyring
# Usage: validate_zigchain_account "$account_name"
validate_zigchain_account() {
    local account_name="$1"
    if ! zigchaind keys show "$account_name" > /dev/null 2>&1; then
        echo "❌ Error: Account '$account_name' not found in zigchain keyring"
        echo "   Please add it first: zigchaind keys add $account_name"
        return 1
    fi
    echo "✅ Account '$account_name' exists in zigchain keyring"
    return 0
}

# Validate that an account exists in axelar keyring
# Usage: validate_axelar_account "$account_name"
validate_axelar_account() {
    local account_name="$1"
    if ! axelard keys show "$account_name" > /dev/null 2>&1; then
        echo "❌ Error: Account '$account_name' not found in axelar keyring"
        echo "   Please add it first: axelard keys add $account_name"
        return 1
    fi
    echo "✅ Account '$account_name' exists in axelar keyring"
    return 0
}

# Validate that accounts exist in both chains
# Usage: validate_accounts "$account_name"
validate_accounts() {
    local account_name="$1"
    echo "🔍 Validating account '$account_name' exists in both keyrings..."

    if ! validate_zigchain_account "$account_name"; then
        return 1
    fi

    if ! validate_axelar_account "$account_name"; then
        return 1
    fi

    echo "✅ Account '$account_name' validated in both keyrings"
    return 0
}

# =============================================================================
# BALANCE CHECKING FUNCTIONS
# =============================================================================

# Get module info (cached to avoid repeated calls)
# Usage: get_module_info
get_module_info() {
    zigchaind q tokenwrapper module-info -o json
}

# Get account balance from zigchain
# Usage: get_zigchain_balance "$account_name"
get_zigchain_balance() {
    local account_name="$1"
    zigchaind q bank balances "$account_name" -o json
}

# Get account balance from axelar
# Usage: get_axelar_balance "$account_name"
get_axelar_balance() {
    local account_name="$1"
    axelard q bank balances "$account_name" -o json 2>/dev/null
}

# Get account balance from cosmos
# Usage: get_cosmos_balance "$account_name"
get_cosmos_balance() {
    local account_name="$1"
    cosmosd q bank balances "$account_name" -o json 2>/dev/null
}

# Get account balance from dummy
# Usage: get_dummy_balance "$account_name"
get_dummy_balance() {
    local account_name="$1"
    dummyd q bank balances "$account_name" -o json 2>/dev/null
}

# Extract specific balance amount from balance JSON
# Usage: extract_balance_amount "$balance_json" "$denom"
extract_balance_amount() {
    local balance_json="$1"
    local denom="$2"
    echo "$balance_json" | jq -r --arg denom "$denom" '((.balances // [])[] | select(.denom == $denom) | .amount) // "0"'
}

# Get specific balance for account and denom
# Usage: get_account_balance "$chain" "$account_name" "$denom"
get_account_balance() {
    local chain="$1"
    local account_name="$2"
    local denom="$3"

    case "$chain" in
        "zigchain")
            local balance_json=$(get_zigchain_balance "$account_name")
            ;;
        "axelar")
            local balance_json=$(get_axelar_balance "$account_name")
            ;;
        "cosmos")
            local balance_json=$(get_cosmos_balance "$account_name")
            ;;
        "dummy")
            local balance_json=$(get_dummy_balance "$account_name")
            ;;
        *)
            echo "❌ Error: Invalid chain '$chain'. Use 'zigchain' or 'axelar'"
            return 1
            ;;
    esac

    extract_balance_amount "$balance_json" "$denom"
}

# =============================================================================
# MODULE STATE FUNCTIONS
# =============================================================================

# Get module uzig balance
# Usage: get_module_uzig_balance
get_module_uzig_balance() {
    local module_info=$(get_module_info)
    extract_balance_amount "$module_info" "uzig"
}

# Get module IBC balance for specific IBC hash
# Usage: get_module_ibc_balance "$ibc_hash"
get_module_ibc_balance() {
    local ibc_hash="$1"
    local module_info=$(get_module_info)
    extract_balance_amount "$module_info" "$ibc_hash"
}

# Get IBC hash from module info
# Usage: get_ibc_hash
get_ibc_hash() {
    local module_info=$(get_module_info)
    local native_port=$(echo "$module_info" | jq -r '.native_port')
    local native_channel=$(echo "$module_info" | jq -r '.native_channel')
    echo "ibc/$(calculate_ibc_hash "$native_port" "$native_channel" "unit-zig")"
}

# Check if token wrapper is enabled
# Usage: is_token_wrapper_enabled
is_token_wrapper_enabled() {
    local module_info=$(get_module_info)
    local enabled=$(echo "$module_info" | jq -r '.token_wrapper_enabled // "false"')
    [ "$enabled" = "true" ]
}

# =============================================================================
# AMOUNT CONVERSION FUNCTIONS
# =============================================================================

# Convert amount string to numeric (remove denom suffix)
# Usage: amount_to_numeric "$amount_with_denom"
amount_to_numeric() {
    local amount="$1"
    # Remove common denom suffixes
    echo "$amount" | sed 's/uzig$//' | sed 's/unit-zig$//' | sed 's/udummy$//' | sed 's/uatom$//' | sed 's/ibc\///'
}

# Convert numeric amount to uzig denom
# Usage: numeric_to_uzig "$numeric_amount"
numeric_to_uzig() {
    local amount="$1"
    echo "${amount}uzig"
}

# Convert numeric amount to unit-zig denom
# Usage: numeric_to_unit_zig "$numeric_amount"
numeric_to_unit_zig() {
    local amount="$1"
    echo "${amount}unit-zig"
}

# =============================================================================
# FEE CALCULATION FUNCTIONS
# =============================================================================

# Get fee amount as numeric (default to 0 if empty)
# Usage: get_fee_numeric "$fee_string"
get_fee_numeric() {
    local fee="${1:-0uzig}"
    amount_to_numeric "$fee"
}

# =============================================================================
# TRANSFER STATISTICS FUNCTIONS
# =============================================================================

# Get transfer statistics
# Usage: get_transfer_stats
get_transfer_stats() {
    zigchaind q tokenwrapper total-transfers -o json
}

# Get total transferred in
# Usage: get_total_transferred_in
get_total_transferred_in() {
    local stats=$(get_transfer_stats)
    echo "$stats" | jq -r '.total_transferred_in // "0"'
}

# Get total transferred out
# Usage: get_total_transferred_out
get_total_transferred_out() {
    local stats=$(get_transfer_stats)
    echo "$stats" | jq -r '.total_transferred_out // "0"'
}

# =============================================================================
# VALIDATION FUNCTIONS
# =============================================================================

# Validate sufficient balance for transfer
# Usage: validate_sufficient_balance "$chain" "$account" "$required_amount" "$denom"
validate_sufficient_balance() {
    local chain="$1"
    local account="$2"
    local required_amount="$3"
    local denom="$4"

    local current_balance=$(get_account_balance "$chain" "$account" "$denom")

    if ! [[ $(echo "$current_balance >= $required_amount" | bc -l) -eq 1 ]]; then
        echo "❌ Insufficient balance for $account on $chain."
        echo "   Required: >= $required_amount $denom"
        echo "   Current: $current_balance $denom"
        return 1
    fi

    echo "✅ Sufficient balance validated for $account on $chain: $current_balance $denom"
    return 0
}

# Validate module has sufficient uzig balance
# Usage: validate_module_uzig_balance "$required_amount"
validate_module_uzig_balance() {
    local required_amount="$1"
    local current_balance=$(get_module_uzig_balance)

    if ! [[ $(echo "$current_balance >= $required_amount" | bc -l) -eq 1 ]]; then
        echo "❌ Insufficient uzig balance in module wallet."
        echo "   Required: >= $required_amount uzig"
        echo "   Current: $current_balance uzig"
        return 1
    fi

    echo "✅ Module has sufficient uzig balance: $current_balance uzig"
    return 0
}

# =============================================================================
# TRANSACTION FUNCTIONS
# =============================================================================

# Execute transaction and capture output
# Usage: execute_tx "$command" "$description"
execute_tx() {
    local command="$1"
    local description="$2"

    echo "🔄 Executing: $description"
    local output=$(eval "$command")
    local raw_logs=$(echo "$output" | jq -r '.raw_log')
    local txhash=$(echo "$output" | jq -r '.txhash')

    # Display logs and hash to stderr so they appear on terminal
    # even when output is captured with command substitution
    echo "📝 Raw logs: $raw_logs" >&2
    echo "🔗 Tx hash: $txhash" >&2

    # Return the output for further processing
    echo "$output"
}

# Execute transaction and expect it to fail
# Usage: execute_tx_expect_failure "$command" "$description"
# Returns: 0 if transaction failed (success), 1 if transaction succeeded (failure)
execute_tx_expect_failure() {
    local command="$1"
    local description="$2"

    echo "🔄 Executing (expecting failure): $description"
    local output=$(eval "$command" 2>&1)
    local exit_code=$?

    # Check if command failed at execution level
    if [[ $exit_code -ne 0 ]]; then
        echo "✅ Command failed at execution level as expected" >&2
        return 0
    fi

    # Parse JSON output to check for transaction failure
    local raw_logs=$(echo "$output" | jq -r '.raw_log' 2>/dev/null || echo "")
    local txhash=$(echo "$output" | jq -r '.txhash' 2>/dev/null || echo "")

    # Display logs and hash to stderr
    echo "📝 Raw logs: $raw_logs" >&2
    echo "🔗 Tx hash: $txhash" >&2

    # Check if transaction failed (raw_log contains error)
    if [[ -n "$raw_logs" && "$raw_logs" != "null" && "$raw_logs" != *"failed"* && "$raw_logs" != *"error"* ]]; then
        echo "❌ Transaction succeeded when it should have failed" >&2
        return 1
    else
        echo "✅ Transaction failed as expected" >&2
        return 0
    fi
}

# Wait for transaction processing
# Usage: wait_for_tx "$timeout_var"
wait_for_tx() {
    local timeout="${1:-$TRANSACTION_TIMEOUT}"
    echo "⏳ Waiting $timeout seconds for transaction to be processed..."
    sleep "$timeout"
    echo "✅ Transaction processing wait completed"
}

# Execute transaction expecting failure, wait for processing, and verify error message
# Usage: execute_tx_expect_failure_and_verify "$chain" "$command" "$description" "$expected_error_message"
# Returns: 0 if transaction failed with expected message, 1 otherwise
execute_tx_expect_failure_and_verify() {
    local chain="$1"
    local command="$2"
    local description="$3"
    local expected_error_message="$4"
    
    # Get the appropriate binary for the chain
    local binary=$(get_chain_binary "$chain")
    if [[ $? -ne 0 ]]; then
        return 1
    fi

    echo "🔄 Executing (expecting failure with specific message): $description"
    echo "📋 Expected error message: $expected_error_message"

    # Step 1: Execute transaction and expect it to fail
    local output=$(eval "$command" 2>&1)
    local exit_code=$?

    # Check if command failed at execution level
    if [[ $exit_code -ne 0 ]]; then
        echo "✅ Command failed at execution level as expected" >&2
        # Extract txhash if available even from failed command
        local txhash=$(echo "$output" | jq -r '.txhash' 2>/dev/null || echo "")
        if [[ -n "$txhash" && "$txhash" != "null" ]]; then
            echo "🔗 Tx hash: $txhash" >&2
            # Step 2: Wait for transaction processing
            wait_for_tx
            
            # Step 3: Retrieve transaction info and verify error message
            local tx_info=$($binary query tx "$txhash" -o json 2>/dev/null)
            if [[ -n "$tx_info" ]]; then
                local raw_log=$(echo "$tx_info" | jq -r '.raw_log' 2>/dev/null || echo "")
                echo "📝 Raw logs from query: $raw_log" >&2
                
                if [[ "$raw_log" == "$expected_error_message" ]]; then
                    echo "✅ Transaction failed with expected error message" >&2
                    return 0
                else
                    echo "❌ Transaction failed but with unexpected error message" >&2
                    echo "   Expected: $expected_error_message" >&2
                    echo "   Got: $raw_log" >&2
                    return 1
                fi
            else
                echo "❌ Could not retrieve transaction info" >&2
                return 1
            fi
        else
            echo "❌ No transaction hash found in output" >&2
            return 1
        fi
    fi

    # Parse JSON output to check for transaction failure
    local raw_logs=$(echo "$output" | jq -r '.raw_log' 2>/dev/null || echo "")
    local txhash=$(echo "$output" | jq -r '.txhash' 2>/dev/null || echo "")

    # Display logs and hash to stderr
    echo "📝 Raw logs: $raw_logs" >&2
    echo "🔗 Tx hash: $txhash" >&2

    # Check if transaction failed (raw_log contains error)
    if [[ -n "$raw_logs" && "$raw_logs" != "null" && "$raw_logs" != *"failed"* && "$raw_logs" != *"error"* ]]; then
        echo "❌ Transaction succeeded when it should have failed" >&2
        return 1
    else
        echo "✅ Transaction failed as expected" >&2
        
        # Step 2: Wait for transaction processing
        if [[ -n "$txhash" && "$txhash" != "null" ]]; then
            wait_for_tx
            
            # Step 3: Retrieve transaction info and verify error message
            local tx_info=$($binary query tx "$txhash" -o json 2>/dev/null)
            if [[ -n "$tx_info" ]]; then
                local raw_log=$(echo "$tx_info" | jq -r '.raw_log' 2>/dev/null || echo "")
                echo "📝 Raw logs from query: $raw_log" >&2
                
                if [[ "$raw_log" == "$expected_error_message" ]]; then
                    echo "✅ Transaction failed with expected error message" >&2
                    return 0
                else
                    echo "❌ Transaction failed but with unexpected error message" >&2
                    echo "   Expected: $expected_error_message" >&2
                    echo "   Got: $raw_log" >&2
                    return 1
                fi
            else
                echo "❌ Could not retrieve transaction info" >&2
                return 1
            fi
        else
            echo "❌ No transaction hash found in output" >&2
            return 1
        fi
    fi
}

# =============================================================================
# LOGGING FUNCTIONS
# =============================================================================

# Log balance comparison
# Usage: log_balance_comparison "$label" "$before" "$after" "$expected"
log_balance_comparison() {
    local label="$1"
    local before="$2"
    local after="$3"
    local expected="$4"

    echo "   $label balance - Before: $before, After: $after, Expected: $expected"
}

# =============================================================================
# TEST FRAMEWORK FUNCTIONS
# =============================================================================

# Test execution control variables
RUN_SPECIFIC_TESTS=false
SPECIFIED_TESTS=()

# Test tracking variables
TEST_IDS=()
TEST_RESULTS=()
TEST_DESCRIPTIONS=()

# Tests that should be skipped by default unless explicitly specified
SKIP_BY_DEFAULT_TESTS=("TW-999" "TW-990" "TW-998")

# Check if a command exists
# Usage: check_command "$command_name"
check_command() {
    if ! command -v "$1" &>/dev/null; then
        echo "Error: $1 is not installed"
        exit 1
    fi
}

# Parse command line arguments for selective test execution
# Usage: parse_test_args "$@"
parse_test_args() {
    if [ $# -gt 0 ]; then
        RUN_SPECIFIC_TESTS=true
        for arg in "$@"; do
            # Support comma-separated or space-separated test IDs
            IFS=',' read -ra TEST_LIST <<< "$arg"
            for test_spec in "${TEST_LIST[@]}"; do
                # Check if it's a range (e.g., TW-001-TW-005)
                if [[ "$test_spec" =~ ^(TW-[0-9]+)-(TW-[0-9]+)$ ]]; then
                    start_test="${BASH_REMATCH[1]}"
                    end_test="${BASH_REMATCH[2]}"
                    # Extract numbers
                    start_num=$(echo "$start_test" | sed 's/TW-//')
                    end_num=$(echo "$end_test" | sed 's/TW-//')
                    # Generate test IDs in range
                    for ((i=start_num; i<=end_num; i++)); do
                        SPECIFIED_TESTS+=("TW-$(printf "%03d" $i)")
                    done
                else
                    # Single test ID
                    SPECIFIED_TESTS+=("$test_spec")
                fi
            done
        done
    fi
}

# Add a test to the skip-by-default list
# Usage: add_skip_by_default_test "$test_id"
add_skip_by_default_test() {
    local test_id="$1"
    SKIP_BY_DEFAULT_TESTS+=("$test_id")
}

# Check if a test is in the skip by default list
# Usage: is_skip_by_default_test "$test_id"
is_skip_by_default_test() {
    local test_id="$1"
    for skip_test in "${SKIP_BY_DEFAULT_TESTS[@]}"; do
        if [ "$test_id" = "$skip_test" ]; then
            return 0
        fi
    done
    return 1
}

# Check if a test should be run
# Usage: should_run_test "$test_id"
should_run_test() {
    local test_id="$1"

    # Check if this is a test that should be skipped by default
    if is_skip_by_default_test "$test_id"; then
        # Only run if specific tests were requested and this test is in the list
        if [ "$RUN_SPECIFIC_TESTS" = true ]; then
            for specified_test in "${SPECIFIED_TESTS[@]}"; do
                if [ "$test_id" = "$specified_test" ]; then
                    return 0
                fi
            done
        fi
        return 1
    fi

    # If no specific tests requested, run all tests
    if [ "$RUN_SPECIFIC_TESTS" = false ]; then
        return 0
    fi

    # Check if test_id is in the specified tests list
    for specified_test in "${SPECIFIED_TESTS[@]}"; do
        if [ "$test_id" = "$specified_test" ]; then
            return 0
        fi
    done

    return 1
}

# Run a test with tracking
# Usage: run_test "$test_id" "$test_description" "$test_command"
run_test() {
    local test_id="$1"
    local test_description="$2"
    local test_command="$3"

    # Check if this test should be run
    if ! should_run_test "$test_id"; then
        echo "⏭️  Skipping test $test_id: $test_description"
        return 0
    fi

    TEST_IDS+=("$test_id")
    TEST_DESCRIPTIONS+=("$test_description")

    echo ""
    echo "================================================================="
    echo "🧪 TEST $test_id: $test_description"
    echo "================================================================="

    if eval "$test_command"; then
        echo "✅ TEST $test_id PASSED"
        TEST_RESULTS+=("PASS")
        # Beep once when test completes
        osascript -e 'beep 1' 2>/dev/null || true
        return 0
    else
        echo "❌ TEST $test_id FAILED"
        TEST_RESULTS+=("FAIL")
        # Beep once when test completes
        osascript -e 'beep 1' 2>/dev/null || true
        return 1
    fi
}

# Generate final test report
# Usage: generate_report
generate_report() {
    echo ""
    echo "================================================================="
    echo "📊 TEST EXECUTION REPORT"
    echo "================================================================="

    local passed=0
    local failed=0
    local total_tests=${#TEST_RESULTS[@]}

    local i=0
    while [ $i -lt $total_tests ]; do
        local status="${TEST_RESULTS[$i]}"
        local description="${TEST_DESCRIPTIONS[$i]}"
        local test_id="${TEST_IDS[$i]}"

        if [ "$status" = "PASS" ]; then
            echo "✅ TEST $test_id: $description"
            passed=$((passed + 1))
        else
            echo "❌ TEST $test_id: $description"
            failed=$((failed + 1))
        fi
        i=$((i + 1))
    done

    echo ""
    echo "================================================================="
    echo "📈 SUMMARY: $passed passed, $failed failed out of $total_tests tests"
    echo "================================================================="

    if [ $failed -eq 0 ]; then
        echo "🎉 ALL TESTS PASSED!"
        return 0
    else
        echo "💥 SOME TESTS FAILED!"
        return 1
    fi
}

# =============================================================================
# COMMON TEST SETUP PATTERNS
# =============================================================================

# Setup test environment variables with defaults
# Usage: setup_test_env
setup_test_env() {
    # Default values
    export TEST_ACCOUNT_NAME=${TEST_ACCOUNT_NAME:-"bob"}
    export OP_ACCOUNT_NAME=${OP_ACCOUNT_NAME:-"operator"}

    export ZIGCHAIN_FEES=${ZIGCHAIN_FEES:-"0uzig"}
    export DUMMY_FEES=${DUMMY_FEES:-"0unit-zig"}
    export AXELAR_FEES=${AXELAR_FEES:-"0unit-zig"}
    export COSMOS_FEES=${COSMOS_FEES:-"0uatom"}

    export DUMMY_ZIGCHAIN_CHANNEL_ID=${DUMMY_ZIGCHAIN_CHANNEL_ID:-"channel-0"}
    export ZIGCHAIN_DUMMY_CHANNEL_ID=${ZIGCHAIN_DUMMY_CHANNEL_ID:-"channel-0"}

    export AXELAR_ZIGCHAIN_CHANNEL_ID=${AXELAR_ZIGCHAIN_CHANNEL_ID:-"channel-0"}
    export ZIGCHAIN_AXELAR_CHANNEL_ID=${ZIGCHAIN_AXELAR_CHANNEL_ID:-"channel-1"}

    export COSMOS_ZIGCHAIN_CHANNEL_ID=${COSMOS_ZIGCHAIN_CHANNEL_ID:-"channel-0"}
    export ZIGCHAIN_COSMOS_CHANNEL_ID=${ZIGCHAIN_COSMOS_CHANNEL_ID:-"channel-2"}

    echo "🔧 Test environment configured:"
    echo "   Test Account: $TEST_ACCOUNT_NAME"
    echo "   Operator Account: $OP_ACCOUNT_NAME"
    echo "   ZIGChain Fees: $ZIGCHAIN_FEES"
    echo "   Axelar Fees: $AXELAR_FEES"
    echo "   Cosmos Fees: $COSMOS_FEES"
    echo "   Dummy Fees: $DUMMY_FEES"
}

# Common pre-transfer validation
# Usage: pre_transfer_validation "$transfer_type"
pre_transfer_validation() {
    local transfer_type="$1"

    echo "🔍 Performing pre-transfer validation for: $transfer_type"

    # Validate accounts exist
    if ! validate_accounts "$TEST_ACCOUNT_NAME"; then
        return 1
    fi

    # Validate module state
    if ! is_token_wrapper_enabled; then
        echo "❌ Token wrapper is not enabled"
        return 1
    fi

    echo "✅ Pre-transfer validation passed"
    return 0
}

# =============================================================================
# ENVIRONMENT DETECTION
# =============================================================================

# Check if running against localnet environment
# Returns 0 (true) if localnet, 1 (false) if not
is_localnet() {
    # Check if running against testnet (opposite of localnet)
    if [[ "$ZIGCHAIN_NODE" == *"testnet"* || "$AXELAR_NODE" == *"testnet"* ]]; then
        return 1  # Not localnet
    else
        return 0  # Is localnet
    fi
}