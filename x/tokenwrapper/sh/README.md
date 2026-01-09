# IBC TokenWrapper Test Scripts

This directory contains scripts to set up a localnet environment and run tests for the Token Wrapper module's IBC functionality. Tests can be run against different environments including both localnet and testnet setups.

## Scripts Overview

### `setup_localnet.sh`
Sets up the localnet environment with:
- 3 chains (Axelar, ZIGChain, Dummy)
- IBC relayers between Axelar<->ZIGChain and Dummy<->ZIGChain
- TokenWrapper module configuration on ZIGChain

### `run_tests.sh`
Runs comprehensive tests against configured blockchain environments (localnet or testnet):
- Each test has a unique incremental ID
- Tests are tracked with descriptions and pass/fail status
- Final report shows all test results with ✅/❌ indicators
- Supports selective test execution (ranges, comma-separated, skip-by-default tests)
- Uses common utility functions from `common_test_utils.sh`

### `common_test_utils.sh`
Provides shared utility functions for the test suite:
- Chain interaction functions (balance checking, transaction execution)
- Account validation and setup functions
- Module state management functions
- Test framework functions (test tracking, reporting)
- IBC hash calculation and validation utilities

## Usage

### Step 1: Setup Localnet Environment
```bash
./x/tokenwrapper/sh/setup_localnet.sh
```

This will:
- Generate chain configurations
- Start all chains in the background
- Configure and start IBC relayers
- Create an environment file with all necessary variables

### Step 2: Run Tests
```bash
# Source the environment file created by setup_localnet.sh
source /path/to/test_env.sh  # (path shown in setup output)

# Run all tests (skips TW-990, TW-998, TW-999 by default)
./x/tokenwrapper/sh/run_tests.sh

# Run specific tests
./x/tokenwrapper/sh/run_tests.sh TW-001 TW-003

# Run a range of tests
./x/tokenwrapper/sh/run_tests.sh TW-001-TW-005

# Run comma-separated tests
./x/tokenwrapper/sh/run_tests.sh TW-001,TW-003,TW-005

# Run skip-by-default tests (timeout, cleanup, etc.)
./x/tokenwrapper/sh/run_tests.sh TW-990 TW-999

# Run all tests including skip-by-default ones
./x/tokenwrapper/sh/run_tests.sh TW-001-TW-999
```

### Running Tests Against Testnet Environment

You can run tests against a pre-configured testnet environment using the provided environment file:

```bash
# Run all tests
source ./x/tokenwrapper/sh/testnet_env.sh && ./x/tokenwrapper/sh/run_tests.sh

# Run specific test by ID
source ./x/tokenwrapper/sh/testnet_env.sh && ./x/tokenwrapper/sh/run_tests.sh TW-001

# Run range of tests
source ./x/tokenwrapper/sh/testnet_env.sh && ./x/tokenwrapper/sh/run_tests.sh TW-001-TW-010

# Run skip-by-default tests (cleanup, timeout, etc.)
source ./x/tokenwrapper/sh/testnet_env.sh && ./x/tokenwrapper/sh/run_tests.sh TW-990 TW-999
```

**Note**: When running against testnet environments:
- No cleanup is performed automatically (testnet remains active for additional testing)
- Environment variables are pre-configured for the testnet setup
- Tests run against live testnet infrastructure rather than local chains

## Test Cases

The test suite includes 20+ core tests plus few special tests:

### Core Tests (TW-001 through TW-020)
- **TW-001: Token Wrapper Module Configuration** - Verify module settings and initial state
- **TW-002: IBC Transfer (Unfunded Module)** - Test behavior when module lacks funds (should receive IBC vouchers)
- **TW-003: Fund Module Wallet** - Add native tokens to module
- **TW-004: IBC Transfer (Funded Module)** - Test successful token wrapping (should receive native tokens)
- **TW-005: Return Transfer** - Test ZIGChain → Axelar transfer
- **TW-006: Native ZIG Recovery** - Test IBC voucher recovery process
- **TW-007: IBC Transfer (Disabled Module)** - Test behavior when module is disabled (should receive IBC vouchers)
- **TW-008: Wrong IBC Settings (Native Client ID)** - Test with wrong native client ID (should receive IBC vouchers)
- **TW-009: Wrong IBC Settings (Native Channel ID)** - Test with wrong native channel ID (should receive IBC vouchers)
- **TW-010: Wrong IBC Settings (Counterparty Channel ID)** - Test with wrong counterparty channel ID (should receive IBC vouchers)
- **TW-011: Return Transfer (Module Disabled)** - Test return transfer when module is disabled (balances should remain unchanged)
- **TW-012: Return Transfer (Insufficient IBC)** - Test return transfer with insufficient IBC balance (should fail entirely)
- **TW-013: Return Transfer (Wrong Native Client ID)** - Test return transfer with wrong native client ID (should fail completely)
- **TW-014: Return Transfer (Wrong Native Channel ID)** - Test return transfer with wrong native channel ID (should succeed as regular IBC transfer)
- **TW-015: Return Transfer (Wrong Counterparty Channel)** - Test return transfer with wrong counterparty channel (should fail completely)
- **TW-016: Return Transfer (Wrong Module Denom)** - Test return transfer with wrong module denom (should fail completely)
- **TW-017: ZIGChain to Cosmos** - Send uzig amount from ZIGChain to Cosmos
- **TW-018: Cosmos to ZIGChain** - Send IBC uzig vouchers from Cosmos to ZIGChain
- **TW-019: Cosmos to ZIGChain (ATOM)** - Send uatom amount from Cosmos to ZIGChain
- **TW-020: Dummy to ZIGChain** - Send unit-zig amount from Dummy to ZIGChain (only runs when running against localnet)

### Special Tests (Skip by Default)
- **TW-990: IBC Timeout Mechanism** - Test timeout handling (skip by default)
- **TW-998: Set Original IBC Settings** - Restore original module IBC settings (skip by default)
- **TW-999: Clean Up Module Balances** - Clean up module balances (skip by default)

## Configuration

Environment variables can be customized by setting them before running `setup_localnet.sh`:

```bash
export AXELAR_CHAIN_ID="my-axelar"
export ZIGCHAIN_RPC_PORT="26660"
# ... other variables
./x/tokenwrapper/sh/setup_localnet.sh
```

## Output

- **Setup**: Creates working directory with logs and chain data (localnet only)
- **Tests**: Real-time test execution with fixed IDs (TW-001, TW-002, etc.) and status
- **Report**: Final summary with all test results and visual indicators (✅/❌)

## Cleanup

### Localnet Environment
- **Full test suite**: Environment is automatically cleaned up on completion
- **Partial tests**: Environment is preserved for additional testing
- **Skip-by-default tests**: Environment is preserved (TW-990, TW-998, TW-999)

Manual cleanup (when needed):
```bash
# Kill all processes
pkill hermes zigchaind axelard dummyd

# Or use the specific PIDs (shown in test output)
kill $AXELAR_PID $ZIGCHAIN_PID $DUMMY_PID $RELAYER_PID1 $RELAYER_PID2 $RELAYER_PID
```

### Testnet Environment
- **No automatic cleanup**: Testnet environments remain active for additional testing
- **External management**: Testnet infrastructure is managed externally
- **Preservation**: All testnet resources are preserved between test runs

## Logs

### Localnet Environment
All logs are stored in the working directory created by `setup_localnet.sh`:
- `logs/axelar.log` - Axelar chain logs
- `logs/zigchain.log` - ZIGChain logs
- `logs/dummy.log` - Dummy chain logs
- `logs/relayer1.log` - Dummy<->ZIGChain relayer
- `logs/relayer2.log` - Axelar<->ZIGChain relayer

### Testnet Environment
Logs are managed by the respective testnet infrastructure and are not controlled by these scripts.
