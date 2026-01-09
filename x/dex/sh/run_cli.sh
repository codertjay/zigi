#!/usr/bin/env zsh
# Runs common commands to test the contract on-fly during development
# not meant to replace unit or integration tests

MODULE_NAME="dex"

# source the environment variables
source sh/utils/prep.env.sh

# stop a script if at any point anything fails (or a variable is missed)
set -o errexit -o nounset -o pipefail

# Function to handle exit
final_print() {
  local exit_code=$?
  if ((exit_code != 0)); then
    printf "Script failed with error code: %d\n" "$exit_code" >&2
    echo "To reload the environment variables, run:"
    echo "source /tmp/.$MODULE_NAME.mod.env" | cb bash
  fi
}

# run clean in case of exit
trap final_print EXIT

echo "# ENV for $MODULE_NAME script" >/tmp/.$MODULE_NAME.mod.env

# print var name and value in a pretty way
# also save it to a file, so we can load it later
pv() {
  local title="$1"
  local var_name="$2"

  echo
  echo "# - - - - - - - - - - - - - - - - - - - -"

  echo "$title" | cb
  # Make it work for both bash and zsh
  eval "printf '%s=%s\\n' \"\$var_name\" \"\${$var_name}\"" | cb bash
  echo "# $title" >>/tmp/.$MODULE_NAME.mod.env
  eval "printf '%s=\"%s\"\\n' \"\$var_name\" \"\${$var_name}\" >> /tmp/.$MODULE_NAME.mod.env"
  echo "# - - - - - - - - - - - - - - - - - - - -"
  echo
}

# --------------------------------------------------------------------------------------------
# SETUP VARIABLES
# --------------------------------------------------------------------------------------------

ACCOUNT="z"
pv "z account" ACCOUNT

ACCOUNT_ADDRESS=$(zigchaind keys show $ACCOUNT -a)
pv "z account address" ACCOUNT_ADDRESS

# ADD a random 10-chars suffix to denom name to avoid conflicts on repeated runs
RANDOM_APPEND=$(openssl rand -hex 5 | tr -dc 'a-fA-F')

# General flags used in commands
#pv "TX_FLAGS" TX_FLAGS

BASE="panda$RANDOM_APPEND"
pv "BASE: panda token" BASE
FULL_BASE="coin.$ACCOUNT_ADDRESS.$BASE"
pv "FULL_BASE: panda factory version" FULL_BASE

QUOTE="quote$RANDOM_APPEND"
pv "QUOTE: quote" QUOTE
FULL_QUOTE="coin.$ACCOUNT_ADDRESS.$QUOTE"
pv "FULL_QUOTE: quote factory version" FULL_QUOTE

FULL_THIRD="uzig"
pv "FULL_THIRD" FULL_THIRD

multiplier=1000000

# --------------------------------------------------------------------------------------------
# CREATING 2 TOKENS
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: creating $BASE token and getting the tx id" | cb
SHA256=$(printf '%s' "{"a":1,"b":2}" | sha256sum | awk '{print $1}')
echo $TX_FLAGS
CREATE_BASE_TX_ID=$(
  zigchaind tx factory create-denom $BASE 1000000000 true \
    "ipfs://ipfs.io/$BASE" $SHA256 \
    --from $ACCOUNT $TX_FLAGS --yes -o json | tx_id
)
pv "CREATE_BASE_TX_ID" CREATE_BASE_TX_ID

echo "Waiting for: create $BASE token tx to be included in a block" | cb
sleep $SLEEP
echo

echo "\n 🚀 EXEC: creating $QUOTE token and getting the tx id" | cb
CREATE_QUOTE_TX_ID=$(
  zigchaind tx factory create-denom $QUOTE 1000000000 true \
    "ipfs://ipfs.io/$QUOTE" $SHA256 --yes \
    --from $ACCOUNT $TX_FLAGS -o json | tx_id
)

pv "CREATE_QUOTE_TX_ID" CREATE_QUOTE_TX_ID

echo "Waiting for: create $QUOTE token tx to be included in a block" | cb
sleep $SLEEP
echo

# Check the balance of the new tokens
zigchaind q factory denom $FULL_BASE $TX_Q_FLAGS | cb
BASE_JSON_OUTPUT=$(zigchaind q factory denom $FULL_BASE $TX_Q_FLAGS -o json)

CONFIRM_BASE=$(echo $BASE_JSON_OUTPUT | jq -r '.denom')

# check as we have created the token
if [ "$CONFIRM_BASE" != "$FULL_BASE" ]; then
  echo "Failed to create $FULL_BASE token"
  q tx $CREATE_BASE_TX_ID
  exit 1
fi

zigchaind q factory denom $FULL_QUOTE $TX_Q_FLAGS | cb
QUOTE_JSON_OUTPUT=$(zigchaind q factory denom $FULL_QUOTE $TX_Q_FLAGS -o json)

CONFIRM_QUOTE=$(echo $QUOTE_JSON_OUTPUT | jq -r '.denom')

# check as we have created the token
if [ "$CONFIRM_QUOTE" != "$FULL_QUOTE" ]; then
  echo "Failed to create $FULL_QUOTE token"
  q tx $CREATE_QUOTE_TX_ID
  exit 1
fi

# --------------------------------------------------------------------------------------------
# MINTING TOKENS
# --------------------------------------------------------------------------------------------

BASE_MINT=$((100 * $multiplier))
echo "\n 🚀 EXEC: Minting 1,000,000 $FULL_BASE" | cb
MINT_BASE_TX_ID=$(
  zigchaind tx factory mint-and-send-tokens $BASE_MINT$FULL_BASE $ACCOUNT_ADDRESS \
    -y --from $ACCOUNT $TX_FLAGS -o json | tx_id
)

pv "MINT_BASE_TX_ID" MINT_BASE_TX_ID

echo "Waiting for: minting $FULL_BASE token tx to be included in a block" | cb
sleep $SLEEP
echo

echo "\n 🚀 EXEC: minting 1,000,000 $FULL_QUOTE" | cb
BASE_MINT=$((100 * $multiplier))
MINT_QUOTE_TX_ID=$(
  zigchaind tx factory mint-and-send-tokens 1000000$FULL_QUOTE $ACCOUNT_ADDRESS \
    -y --from $ACCOUNT $TX_FLAGS -o json | tx_id
)
pv "MINT_QUOTE_TX_ID" MINT_QUOTE_TX_ID

echo "Waiting for: minting $FULL_QUOTE token tx to be included in a block" | cb
sleep $SLEEP
echo

# check new tokens balances
BASE_BALANCE=$(q bank balance $ACCOUNT_ADDRESS $FULL_BASE $TX_Q_FLAGS -o json | jq -r '.balance.amount')
if [ "$BASE_BALANCE" != "100000000" ]; then
  echo "Failed to mint $BASE token"
  exit
fi

echo "\t ✅ $BASE token minted successfully"

QUOTE_BALANCE=$(q bank balance $ACCOUNT_ADDRESS $FULL_QUOTE $TX_Q_FLAGS -o json | jq -r '.balance.amount')
if [ "$QUOTE_BALANCE" != "1000000" ]; then
  echo "Failed to mint $QUOTE token"
  exit
fi

echo "\t ✅ $QUOTE token minted successfully"

# --------------------------------------------------------------------------------------------
# CREATE A NEW POOL
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: Creating LP1 with 10000 $FULL_BASE & 40000 $FULL_QUOTE" | cb

TX_POOL_1_TX_ID=$(
  zigchaind tx dex create-pool 10000$FULL_BASE 40000$FULL_QUOTE -y \
    --from $ACCOUNT $TX_FLAGS -o json | tx_id
)
pv "TX_POOL_1_TX_ID" TX_POOL_1_TX_ID

echo "Waiting for: pool 1 to be created" | cb
sleep $SLEEP
echo

POOL_1_TX_JSON_OUTPUT=$(zigchaind q tx $TX_POOL_1_TX_ID $TX_Q_FLAGS -o json)

POOL_1_ID=$(echo $POOL_1_TX_JSON_OUTPUT | jq -r '.events[-1] .attributes[2] .value')
pv "POOL_1_ID: for $BASE and $QUOTE" POOL_1_ID

POOL_1_JSON_OUTPUT=$(zigchaind q dex get-pool $POOL_1_ID $TX_Q_FLAGS -o json)
POOL_1_LP_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .lp_token .amount')
POOL_1_BASE_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .coins[0] .amount')
POOL_1_QUOTE_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .coins[1] .amount')

# sqrt(10000 * 40000) = 20000
if [ "$POOL_1_LP_AMOUNT" != "20000" ]; then
  echo "Failed to create LP1 $FULL_BASE and $FULL_QUOTE pool"
  echo "POOL_1_AMOUNT: $POOL_1_LP_AMOUNT and we expected 20000"
  exit
fi

echo "\t ✅ LP1 $FULL_BASE and $FULL_QUOTE pool $POOL_1_ID created successfully"
echo "\t ✅ Amount of created LP token: $POOL_1_LP_AMOUNT"

if [ "$POOL_1_BASE_AMOUNT" != "10000" ]; then
  echo "Failed to create LP1 $FULL_BASE and $FULL_QUOTE pool"
  echo "POOL_1_COIN_1_AMOUNT: $POOL_1_BASE_AMOUNT and we expected 10000"
  exit
fi

echo "\t ✅ Amount of $FULL_BASE in pool LP1 $POOL_1_ID: $POOL_1_BASE_AMOUNT"

if [ "$POOL_1_QUOTE_AMOUNT" != "40000" ]; then
  echo "Failed to create LP1 $FULL_BASE and $FULL_QUOTE pool"
  echo "POOL_1_COIN_2_AMOUNT: $POOL_1_QUOTE_AMOUNT and we expected 40000"
  exit
fi

echo "\t ✅ Amount of $FULL_QUOTE in pool LP1 $POOL_1_ID: $POOL_1_QUOTE_AMOUNT"

ACCOUNT_POOL_1_BALANCE=$(q bank balance $ACCOUNT_ADDRESS $POOL_1_ID $TX_Q_FLAGS -o json | jq -r '.balance .amount')

# The number of LP tokens minted should be 20000 but user receives 20000 - minimum lock which is 1000
# 20000 - 1000 = 19000
if [ "$ACCOUNT_POOL_1_BALANCE" != "19000" ]; then
  echo "Failed to mint $POOL_1_ID token"
  echo "ACCOUNT_POOL_1_BALANCE: $ACCOUNT_POOL_1_BALANCE and we expected 20000"
  exit
fi

echo "\t ✅ Account: $ACCOUNT has received $ACCOUNT_POOL_1_BALANCE LP Tokens"

# POOL 2 --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: Creating LP2 with 10000$FULL_BASE & 40000$FULL_THIRD" | cb
TX_POOL_2_TX_ID=$(
  zigchaind tx dex create-pool 10000$FULL_BASE 40000$FULL_THIRD -y \
    --from $ACCOUNT $TX_FLAGS -o json | tx_id
)
pv "TX_POOL_2_TX_ID" TX_POOL_2_TX_ID

echo "Waiting for: LP2 to be created" | cb
sleep $SLEEP
echo

POOL_2_TX_JSON_OUTPUT=$(zigchaind q tx $TX_POOL_2_TX_ID $TX_Q_FLAGS -o json)

POOL_2_ID=$(echo $POOL_2_TX_JSON_OUTPUT | jq -r '.events[-1] .attributes[2] .value')
pv "POOL_2_ID: for $BASE and $FULL_THIRD" POOL_2_ID

POOL_2_JSON_OUTPUT=$(zigchaind q dex get-pool $POOL_2_ID $TX_Q_FLAGS -o json)
POOL_2_LP_AMOUNT=$(echo $POOL_2_JSON_OUTPUT | jq -r '.pool .lp_token .amount')
POOL_2_BASE_AMOUNT=$(echo $POOL_2_JSON_OUTPUT | jq -r '.pool .coins[0] .amount')
POOL_2_QUOTE_AMOUNT=$(echo $POOL_2_JSON_OUTPUT | jq -r '.pool .coins[1] .amount')

if [ "$POOL_2_LP_AMOUNT" != "20000" ]; then
  echo "Failed to create LP2 $FULL_BASE and $FULL_THIRD pool"
  echo "POOL_2_AMOUNT: $POOL_2_LP_AMOUNT and we expected 2000"
  exit
fi

echo "\t ✅ LP2 $FULL_BASE and $FULL_THIRD pool created successfully"
echo "\t ✅ Amount of created LP2 token: $POOL_2_LP_AMOUNT"

if [ "$POOL_2_BASE_AMOUNT" != "10000" ]; then
  echo "Failed to create LP2 $FULL_BASE and $FULL_THIRD pool"
  echo "POOL_2_COIN_1_AMOUNT: $POOL_2_BASE_AMOUNT and we expected 10000"
  exit
fi

echo "\t ✅ Amount of $FULL_BASE in pool LP2 $POOL_2_ID: $POOL_2_BASE_AMOUNT"

if [ "$POOL_2_QUOTE_AMOUNT" != "40000" ]; then
  echo "Failed to create LP2 $FULL_BASE and $FULL_THIRD pool"
  echo "POOL_2_COIN_2_AMOUNT: $POOL_2_QUOTE_AMOUNT and we expected 40000"
  exit
fi

echo "\t ✅ Amount of $FULL_THIRD in pool LP2 $POOL_2_ID: $POOL_2_QUOTE_AMOUNT"

ACCOUNT_POOL_2_BALANCE=$(q bank balance $ACCOUNT_ADDRESS $POOL_2_ID $TX_Q_FLAGS -o json | jq -r '.balance .amount')

# The number of LP tokens minted should be 20000 but user receives 20000 - minimum lock which is 1000
# 20000 - 1000 = 19000
if [ "$ACCOUNT_POOL_2_BALANCE" != "19000" ]; then
  echo "Failed to mint $POOL_2_ID token"
  echo "ACCOUNT_POOL_2_BALANCE: $ACCOUNT_POOL_2_BALANCE and we expected 1900"
  exit
fi

echo "\t ✅ From LP2 | Account: $ACCOUNT has received $ACCOUNT_POOL_2_BALANCE LP Tokens"

# POOL 3 --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: Creating LP3 with $FULL_QUOTE & $FULL_THIRD" | cb
TX_POOL_3_TX_ID=$(
  zigchaind tx dex create-pool 10000$FULL_QUOTE 40000$FULL_THIRD -y \
    --from $ACCOUNT $TX_FLAGS -o json | tx_id
)
pv "TX_POOL_3_TX_ID" TX_POOL_3_TX_ID

echo "Waiting for: LP3 to be created" | cb
sleep $SLEEP
echo

POOL_3_TX_JSON_OUTPUT=$(zigchaind q tx $TX_POOL_3_TX_ID $TX_Q_FLAGS -o json)

POOL_3_ID=$(echo $POOL_3_TX_JSON_OUTPUT | jq -r '.events[-1] .attributes[2] .value')
pv "POOL_3_ID: for $BASE and $FULL_THIRD" POOL_3_ID

POOL_3_JSON_OUTPUT=$(zigchaind q dex get-pool $POOL_3_ID $TX_Q_FLAGS -o json)
POOL_3_LP_AMOUNT=$(echo $POOL_3_JSON_OUTPUT | jq -r '.pool .lp_token .amount')
POOL_3_BASE_AMOUNT=$(echo $POOL_3_JSON_OUTPUT | jq -r '.pool .coins[0] .amount')
POOL_3_QUOTE_AMOUNT=$(echo $POOL_3_JSON_OUTPUT | jq -r '.pool .coins[1] .amount')

if [ "$POOL_3_LP_AMOUNT" != "20000" ]; then
  echo "Failed to create $FULL_QUOTE and $FULL_THIRD pool for LP3"
  echo "POOL_3_AMOUNT: $POOL_3_LP_AMOUNT and we expected 20000"
  exit
fi

echo "\t ✅ LP3 $FULL_QUOTE and $FULL_THIRD pool created successfully"
echo "\t ✅ LP3 Amount of created LP token: $POOL_3_LP_AMOUNT"

if [ "$POOL_3_BASE_AMOUNT" != "10000" ]; then
  echo "Failed to create LP3 $FULL_QUOTE and $FULL_THIRD pool"
  echo "POOL_3_COIN_1_AMOUNT: $POOL_3_BASE_AMOUNT and we expected 10000"
  exit
fi

echo "\t ✅ Amount of $FULL_QUOTE in pool LP3 $POOL_3_ID: $POOL_3_BASE_AMOUNT"

if [ "$POOL_3_QUOTE_AMOUNT" != "40000" ]; then
  echo "Failed to create LP3 $FULL_QUOTE and $FULL_THIRD pool"
  echo "POOL_3_COIN_2_AMOUNT: $POOL_3_QUOTE_AMOUNT and we expected 40000"
  exit
fi

echo "\t ✅ Amount of $FULL_THIRD in pool LP3 $POOL_3_ID: $POOL_3_QUOTE_AMOUNT"

ACCOUNT_POOL_3_BALANCE=$(q bank balance $ACCOUNT_ADDRESS $POOL_3_ID $TX_Q_FLAGS $TX_Q_FLAGS -o json | jq -r '.balance .amount')

# The number of LP tokens minted should be 20000 but user receives 20000 - minimum lock which is 1000
# 20000 - 1000 = 19000
if [ "$ACCOUNT_POOL_3_BALANCE" != "19000" ]; then
  echo "Failed to mint $POOL_3_ID token"
  echo "ACCOUNT_POOL_3_BALANCE: $ACCOUNT_POOL_3_BALANCE and we expected 19000"
  exit
fi

echo "\t ✅ Account: $ACCOUNT has received $ACCOUNT_POOL_3_BALANCE LP Tokens"

# --------------------------------------------------------------------------------------------
# ADD LIQUIDITY TO NEW POOLS - Same Ratio
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: add liquidity to pool $POOL_1_ID 100$FULL_BASE 400$FULL_QUOTE" | cb
ADD_LIQUIDITY_TX_ID=$(
  zigchaind tx dex add-liquidity $POOL_1_ID 100$FULL_BASE 400$FULL_QUOTE -y \
    --from $ACCOUNT $TX_FLAGS -o json | tx_id
)
pv "ADD_LIQUIDITY_TX_ID" ADD_LIQUIDITY_TX_ID

echo "Waiting for: to allow for the liquidity to be added to pool $POOL_1_ID" | cb
sleep $SLEEP
echo

# check pool's new tokens balances
POOL_1_JSON_OUTPUT=$(zigchaind q dex pool $POOL_1_ID $TX_Q_FLAGS -o json)
POOL_1_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .lp_token .amount')
POOL_1_BASE_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .coins[0] .amount')
POOL_1_QUOTE_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .coins[1] .amount')

if [ "$POOL_1_AMOUNT" != "20200" ]; then
  echo "Failed to add liquidity to $FULL_BASE and $FULL_QUOTE pool"
  echo "POOL_1_AMOUNT: $POOL_1_AMOUNT expected 20200"
  exit
fi

echo "\t ✅ Liquidity added to pool ZP1 $POOL_1_ID: $FULL_BASE and $FULL_QUOTE successfully"
echo "\t ✅ Current amount of LP token: $POOL_1_AMOUNT"

if [ "$POOL_1_BASE_AMOUNT" != "10100" ]; then
  echo "Failed to add liquidity to $FULL_BASE and $FULL_QUOTE pool"
  echo "POOL_1_COIN_1_AMOUNT: $POOL_1_BASE_AMOUNT expected 10100"
  exit
fi

echo "\t ✅ Amount of $FULL_BASE in pool ZP1 $POOL_1_ID: $POOL_1_BASE_AMOUNT"

if [ "$POOL_1_QUOTE_AMOUNT" != "40400" ]; then
  echo "Failed to add liquidity to $FULL_BASE and $FULL_QUOTE pool"
  echo "POOL_1_COIN_2_AMOUNT: $POOL_1_QUOTE_AMOUNT expected 40400"
  exit
fi

echo "\t ✅ Amount of $FULL_QUOTE in pool ZP1 $POOL_1_ID: $POOL_1_QUOTE_AMOUNT"

ACCOUNT_POOL_1_BALANCE=$(q bank balance $ACCOUNT_ADDRESS $POOL_1_ID $TX_Q_FLAGS -o json | jq -r '.balance .amount')

# The number of LP tokens that the user has is the ones minted before 19000 + 200
if [ "$ACCOUNT_POOL_1_BALANCE" != "19200" ]; then
  echo "Failed to mint $POOL_1_ID token"
  echo "ACCOUNT_POOL_1_BALANCE: $ACCOUNT_POOL_1_BALANCE and we expected 19200"
  exit
fi

echo "\t ✅ Account: $ACCOUNT has $ACCOUNT_POOL_1_BALANCE $POOL_1_ID Tokens"

# --------------------------------------------------------------------------------------------
# ADD LIQUIDITY TO NEW POOLS - Different Ratio
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: add unbalanced liquidity to LP2 pool $POOL_2_ID 501$FULL_BASE 2000$FULL_THIRD. To keep the ratio, actually will be added 500$FULL_BASE 2000$FULL_THIRD" | cb
ADD_LIQUIDITY_TX_ID=$(
  zigchaind tx dex add-liquidity $POOL_2_ID 501$FULL_BASE 2000$FULL_THIRD -y \
    --from $ACCOUNT $TX_FLAGS -o json | tx_id
)

pv "ADD_LIQUIDITY_TX_ID" ADD_LIQUIDITY_TX_ID

echo "Waiting for: to allow for the unbalanced liquidity LP2 to be added to pool $POOL_1_ID" | cb
sleep $SLEEP
echo

# check pool's new tokens balances
POOL_2_JSON_OUTPUT=$(zigchaind q dex pool $POOL_2_ID $TX_Q_FLAGS -o json)
POOL_2_AMOUNT=$(echo $POOL_2_JSON_OUTPUT | jq -r '.pool .lp_token .amount')
POOL_2_BASE_AMOUNT=$(echo $POOL_2_JSON_OUTPUT | jq -r '.pool .coins[0] .amount')
POOL_2_QUOTE_AMOUNT=$(echo $POOL_2_JSON_OUTPUT | jq -r '.pool .coins[1] .amount')

# Before adding liquidity the pool had 20000 LP tokens
# the liquidity added is sqrt(50*200) = 100
if [ "$POOL_2_AMOUNT" != "21000" ]; then
  echo "Failed to add liquidity to LP2 pool $POOL_2_ID with $FULL_BASE and $FULL_THIRD"
  echo "POOL_2_AMOUNT: $POOL_2_AMOUNT expected 21000"
  exit
fi

echo "\t ✅ Unbalanced Liquidity added to pool LP2 $POOL_2_ID: $FULL_BASE and $FULL_THIRD successfully"
echo "\t ✅ Current amount of LP token: $POOL_2_AMOUNT"

# the liquidity added is 50$FULL_BASE and 200$FULL_THIRD
# As there was 10000 and we are adding only 50 -> the new amount is 10050
if [ "$POOL_2_BASE_AMOUNT" != "10500" ]; then
  echo "Failed to add liquidity to LP2 pool $POOL_2_ID with $FULL_BASE and $FULL_THIRD"
  echo "POOL_2_BASE_AMOUNT: $POOL_2_BASE_AMOUNT expected 10500"
  exit
fi

echo "\t ✅ Amount of $FULL_BASE in pool ZP1 $POOL_2_ID: $POOL_2_BASE_AMOUNT"

if [ "$POOL_2_QUOTE_AMOUNT" != "42000" ]; then
  echo "Failed to add liquidity to LP2 pool $POOL_2_ID with $FULL_BASE and $FULL_THIRD"
  echo "POOL_2_QUOTE_AMOUNT: $POOL_2_QUOTE_AMOUNT expected 42000"
  exit
fi

echo "\t ✅ Amount of $FULL_THIRD in pool LP2 $POOL_2_ID: $POOL_2_QUOTE_AMOUNT"

ACCOUNT_POOL_2_BALANCE=$(q bank balance $ACCOUNT_ADDRESS $POOL_2_ID $TX_Q_FLAGS -o json | jq -r '.balance .amount')

# The number of LP tokens that the user has is the ones minted before 19000 + 100
if [ "$ACCOUNT_POOL_2_BALANCE" != "20000" ]; then
  echo "Failed to mint $POOL_2_ID token"
  echo "ACCOUNT_POOL_1_BALANCE: $ACCOUNT_POOL_2_BALANCE and we expected 20000"
  exit
fi

echo "\t ✅ Account: $ACCOUNT has $ACCOUNT_POOL_2_BALANCE $POOL_2_ID Tokens"

# --------------------------------------------------------------------------------------------
# SWAP COINS
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: swap-exact-in 10$FULL_BASE for 36$FULL_QUOTE" | cb

ACCOUNT_BALANCES_JSON_OUTPUT=$(zigchaind q bank balances $ACCOUNT_ADDRESS $TX_Q_FLAGS -o json)
ACCOUNT_BASE_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_BASE'") .amount')
ACCOUNT_QUOTE_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_QUOTE'") .amount')
ACCOUNT_POOL_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$POOL_1_ID'") .amount')
echo "\t Account $ACCOUNT has $ACCOUNT_BASE_BALANCE $FULL_BASE before swap"
echo "\t Account $ACCOUNT has $ACCOUNT_QUOTE_BALANCE $FULL_QUOTE before swap"
echo "\t Account $ACCOUNT has $ACCOUNT_POOL_BALANCE $POOL_1_ID before swap"

SWAP_IN_TX_ID=$(zigchaind tx dex swap-exact-in $POOL_1_ID 10$FULL_BASE $TX_FLAGS -y --from $ACCOUNT $TX_FLAGS -o json | tx_id)
pv "SWAP_IN_TX_ID" SWAP_IN_TX_ID

echo "Waiting for: to allow swap to be added" | cb
sleep $SLEEP
echo

echo "\n Info: checking Account $ACCOUNT balances after swap-exact-in" | cb
ACCOUNT_BALANCES_JSON_OUTPUT=$(zigchaind q bank balances $ACCOUNT_ADDRESS $TX_Q_FLAGS -o json)
ACCOUNT_BASE_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_BASE'") .amount')
ACCOUNT_QUOTE_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_QUOTE'") .amount')
ACCOUNT_POOL_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$POOL_1_ID'") .amount')

# The balance of Base token before swap is 99979400.
# After swap should be 99979400 - 10 = 99979390
if [ "$ACCOUNT_BASE_BALANCE" != "99979390" ]; then
  echo "ACCOUNT_BASE_BALANCE: $ACCOUNT_BASE_BALANCE and we expect 99979390" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_BASE token" | cb
  exit
fi

echo "\t ✅ Account $ACCOUNT has $ACCOUNT_BASE_BALANCE $FULL_BASE after swap-exact-in (99979400 - 10 = 99979390)"

# Balance before swap is 949600
if [ "$ACCOUNT_QUOTE_BALANCE" != "949636" ]; then
  echo "ACCOUNT_QUOTE_BALANCE: $ACCOUNT_QUOTE_BALANCE and we expect 949636" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_QUOTE token" | cb
  exit
fi

echo "\t ✅ Account $ACCOUNT has $ACCOUNT_QUOTE_BALANCE $FULL_QUOTE after swap-exact-in (949636)"

# The balance of LP token before swap is 19200 as it is a swap it should not change
if [ "$ACCOUNT_POOL_BALANCE" != "19200" ]; then
  echo "ACCOUNT_POOL 1 BALANCE: $ACCOUNT_POOL_BALANCE and we expect 19200" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_BASE token" | cb
  exit
fi

echo "\t ✅ Account $ACCOUNT has $ACCOUNT_POOL_BALANCE $POOL_1_ID after swap-exact-in (19200 same as before)"

echo "\n Info: checking LP1 $POOL_1_ID balances after swap-exact-in" | cb

# check pool's new tokens balances
POOL_1_JSON_OUTPUT=$(zigchaind q dex pool $POOL_1_ID $TX_Q_FLAGS -o json)
POOL_1_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .lp_token .amount')
POOL_1_BASE_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .coins[0] .amount')
POOL_1_QUOTE_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .coins[1] .amount')

# The amount of LP token should not change
if [ "$POOL_1_AMOUNT" != "20200" ]; then
  echo "POOL_1_AMOUNT: $POOL_1_AMOUNT and we expect 20200" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_BASE and $FULL_QUOTE pool" | cb
  exit
fi

echo "\t ✅ LP1 $POOL_1_ID pool has $POOL_1_AMOUNT LP tokens after swap-exact-in (20200 same as before)"

# The amount of Base token in the pool should be 10100 + 10 = 10110
if [ "$POOL_1_BASE_AMOUNT" != "10110" ]; then
  echo "POOL_1_COIN_1_AMOUNT: $POOL_1_BASE_AMOUNT and we expect 10110" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_BASE and $FULL_QUOTE pool" | cb
  exit
fi

echo "\t ✅ LP1 $POOL_1_ID pool has $POOL_1_BASE_AMOUNT $FULL_BASE after swap-exact-in (10100 + 10 = 10110)"

# The amount of Quote token in the pool should be 40400 - 36 = 40364
if [ "$POOL_1_QUOTE_AMOUNT" != "40364" ]; then
  echo "POOL_1_COIN_2_AMOUNT: $POOL_1_QUOTE_AMOUNT and we expect 40364" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_BASE and $FULL_QUOTE pool" | cb
  exit
fi

echo "\t ✅ LP1 $POOL_1_ID pool has $POOL_1_QUOTE_AMOUNT $FULL_QUOTE after swap-exact-in (40400 - 36 = 40364)"

# --------------------------------------------------------------------------------------------
# SWAP COINS RECEIVER ADDRESS
# --------------------------------------------------------------------------------------------

RECEIVER="zuser1"
RECEIVER_ADDRESS=$(zigchaind keys show $RECEIVER -a)
pv "RECEIVER" RECEIVER
pv "RECEIVER_ADDRESS" RECEIVER_ADDRESS

echo "\n 🚀 EXEC: swapping with receiver address: $RECEIVER_ADDRESS" | cb

ACCOUNT_BALANCES_JSON_OUTPUT=$(zigchaind q bank balances $ACCOUNT_ADDRESS $TX_Q_FLAGS -o json)
ACCOUNT_BASE_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_BASE'") .amount')
ACCOUNT_QUOTE_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_QUOTE'") .amount')
ACCOUNT_POOL_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$POOL_1_ID'") .amount')
echo "\t Account $ACCOUNT has $ACCOUNT_BASE_BALANCE $FULL_BASE before swap"
echo "\t Account $ACCOUNT has $ACCOUNT_QUOTE_BALANCE $FULL_QUOTE before swap"
echo "\t Account $ACCOUNT has $ACCOUNT_POOL_BALANCE $POOL_1_ID before swap"

RECEIVER_BALANCES_JSON_OUTPUT=$(zigchaind q bank balances $RECEIVER_ADDRESS $TX_Q_FLAGS -o json)
RECEIVER_BASE_BALANCE=$(echo $RECEIVER_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_BASE'") .amount')
RECEIVER_QUOTE_BALANCE=$(echo $RECEIVER_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_QUOTE'") .amount')
RECEIVER_POOL_BALANCE=$(echo $RECEIVER_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$POOL_1_ID'") .amount')
echo "\t Receiver $RECEIVER has $RECEIVER_BASE_BALANCE $FULL_BASE before swap"
echo "\t Receiver $RECEIVER has $RECEIVER_QUOTE_BALANCE $FULL_QUOTE before swap"
echo "\t Receiver $RECEIVER has $RECEIVER_POOL_BALANCE $POOL_1_ID before swap"

echo "\n 🚀 EXEC: swap-exact-in 10$FULL_BASE for 36$FULL_QUOTE and send to $RECEIVER_ADDRESS" | cb
zigchaind tx dex swap-exact-in $POOL_1_ID 10$FULL_BASE --receiver $RECEIVER_ADDRESS $TX_FLAGS -y --from $ACCOUNT | grep "txhash" | cb

echo "Waiting for: to allow swap to be added" | cb
sleep $SLEEP

# ------------------------------------------------------------------------------------------------

echo "\n Info: checking Account $ACCOUNT balances after swap-exact-in" | cb
ACCOUNT_BALANCES_JSON_OUTPUT=$(zigchaind q bank balances $ACCOUNT_ADDRESS $TX_Q_FLAGS -o json)
ACCOUNT_BASE_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_BASE'") .amount')
ACCOUNT_QUOTE_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_QUOTE'") .amount')
ACCOUNT_POOL_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$POOL_1_ID'") .amount')

# The balance of Base token before swap is 99979390.
# After swap should be 99979390 - 10 = 99979380
if [ "$ACCOUNT_BASE_BALANCE" != "99979380" ]; then
  echo "ACCOUNT_BASE_BALANCE: $ACCOUNT_BASE_BALANCE and we expect 99979380" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_BASE token" | cb
  exit
fi

echo "\t ✅ Account $ACCOUNT has $ACCOUNT_BASE_BALANCE $FULL_BASE after swap-exact-in (99979390 - 10 = 99979380)"

# Balance before swap is 949636 and it should not change as it is sent to the receiver
if [ "$ACCOUNT_QUOTE_BALANCE" != "949636" ]; then
  echo "ACCOUNT_QUOTE_BALANCE: $ACCOUNT_QUOTE_BALANCE and we expect 949636" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_QUOTE token" | cb
  exit
fi

echo "\t ✅ Account $ACCOUNT has $ACCOUNT_QUOTE_BALANCE $FULL_QUOTE after swap-exact-in (949636)"

# The balance of LP token before swap is 19200 as it is a swap it should not change
if [ "$ACCOUNT_POOL_BALANCE" != "19200" ]; then
  echo "ACCOUNT_POOL 1 BALANCE: $ACCOUNT_POOL_BALANCE and we expect 19200" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_BASE token" | cb
  exit
fi

echo "\t ✅ Account $ACCOUNT has $ACCOUNT_POOL_BALANCE $POOL_1_ID after swap-exact-in (19200 same as before)"

# ------------------------------------------------------------------------------------------------

echo "\n Info: checking Receiver $RECEIVER balances after swap-exact-in" | cb
RECEIVER_BALANCES_JSON_OUTPUT=$(zigchaind q bank balances $RECEIVER_ADDRESS $TX_Q_FLAGS -o json)
RECEIVER_BASE_BALANCE=$(echo $RECEIVER_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_BASE'") .amount')
RECEIVER_QUOTE_BALANCE=$(echo $RECEIVER_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_QUOTE'") .amount')
RECEIVER_POOL_BALANCE=$(echo $RECEIVER_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$POOL_1_ID'") .amount')

# The balance of Base token before swap is 0 and it shouldn't change as it is sent to the receiver
if [ -n "$RECEIVER_BASE_BALANCE" ] && [ "$RECEIVER_BASE_BALANCE" != "0" ]; then
  echo "RECEIVER_BASE_BALANCE: $RECEIVER_BASE_BALANCE and we expect 0" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_BASE token" | cb
  exit
fi

echo "\t ✅ Receiver $RECEIVER has $RECEIVER_BASE_BALANCE $FULL_BASE after swap-exact-in (0)"

# Balance before swap is 0
if [ -n "$RECEIVER_QUOTE_BALANCE" ] && [ "$RECEIVER_QUOTE_BALANCE" != "36" ]; then
  echo "RECEIVER_QUOTE_BALANCE: $RECEIVER_QUOTE_BALANCE and we expect 36" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_QUOTE token" | cb
  exit
fi

echo "\t ✅ Receiver $RECEIVER has $RECEIVER_QUOTE_BALANCE $FULL_QUOTE after swap-exact-in (36)"

# The balance of LP token before swap is 0 as it is a swap it should not change
if [ -n "$RECEIVER_POOL_BALANCE" ] && [ "$RECEIVER_POOL_BALANCE" != "0" ]; then
  echo "RECEIVER_POOL 1 BALANCE: $RECEIVER_POOL_BALANCE and we expect 0" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_BASE token" | cb
  exit
fi

echo "\t ✅ Receiver $RECEIVER has $RECEIVER_POOL_BALANCE $POOL_1_ID after swap-exact-in (0 same as before)"

echo "\n Info: checking LP1 $POOL_1_ID balances after swap-exact-in" | cb

# check pool's new tokens balances
POOL_1_JSON_OUTPUT=$(zigchaind q dex pool $POOL_1_ID $TX_Q_FLAGS -o json)
POOL_1_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .lp_token .amount')
POOL_1_BASE_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .coins[0] .amount')
POOL_1_QUOTE_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .coins[1] .amount')

# The amount of LP token should not change
if [ "$POOL_1_AMOUNT" != "20200" ]; then
  echo "POOL_1_AMOUNT: $POOL_1_AMOUNT and we expect 20200" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_BASE and $FULL_QUOTE pool" | cb
  exit
fi

echo "\t ✅ LP1 $POOL_1_ID pool has $POOL_1_AMOUNT LP tokens after swap-exact-in (20200 same as before)"

# The amount of Base token in the pool should be 10110 + 10 = 10120
if [ "$POOL_1_BASE_AMOUNT" != "10120" ]; then
  echo "POOL_1_COIN_1_AMOUNT: $POOL_1_BASE_AMOUNT and we expect 10120" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_BASE and $FULL_QUOTE pool" | cb
  exit
fi

echo "\t ✅ LP1 $POOL_1_ID pool has $POOL_1_BASE_AMOUNT $FULL_BASE after swap-exact-in (10110 + 10 = 10120)"

# The amount of Quote token in the pool should be 40364 - 36 = 40328
if [ "$POOL_1_QUOTE_AMOUNT" != "40328" ]; then
  echo "POOL_1_COIN_2_AMOUNT: $POOL_1_QUOTE_AMOUNT and we expect 40328" | cb
  echo "FAIL: Failed to swap-exact-in $FULL_BASE and $FULL_QUOTE pool" | cb
  exit
fi

echo "\t ✅ LP1 $POOL_1_ID pool has $POOL_1_QUOTE_AMOUNT $FULL_QUOTE after swap-exact-in (40364 - 36 = 40328)"


# --------------------------------------------------------------------------------------------
# REMOVE LIQUIDITY FROM PL1
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: remove liquidity: 200$POOL_1_ID" | cb

echo " Info: LP1 $POOL_1_ID before removing liquidity" | cb
POOL_1_JSON_OUTPUT=$(zigchaind q dex pool $POOL_1_ID $TX_Q_FLAGS -o json)
POOL_1_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .lp_token .amount')
POOL_1_BASE_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .coins[0] .amount')
POOL_1_QUOTE_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .coins[1] .amount')

echo "\t LP1 $POOL_1_ID pool has $POOL_1_AMOUNT LP tokens before removing liquidity"
echo "\t LP1 $POOL_1_ID pool has $POOL_1_BASE_AMOUNT $FULL_BASE before removing liquidity"
echo "\t LP1 $POOL_1_ID pool has $POOL_1_QUOTE_AMOUNT $FULL_QUOTE before removing liquidity"

echo " Info: Account $ACCOUNT before removing liquidity" | cb
ACCOUNT_BALANCES_JSON_OUTPUT=$(zigchaind q bank balances $ACCOUNT_ADDRESS $TX_Q_FLAGS -o json)
ACCOUNT_POOL_1_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$POOL_1_ID'") .amount')
ACCOUNT_BASE_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_BASE'") .amount')
ACCOUNT_QUOTE_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_QUOTE'") .amount')

echo "\t Account $ACCOUNT has $ACCOUNT_POOL_1_BALANCE $POOL_1_ID before removing liquidity"
echo "\t Account $ACCOUNT has $ACCOUNT_BASE_BALANCE $FULL_BASE before removing liquidity"
echo "\t Account $ACCOUNT has $ACCOUNT_QUOTE_BALANCE $FULL_QUOTE before removing liquidity"

# Remove liquidity from the pool
echo "\n 🚀 EXEC: removing liquidity: 200$POOL_1_ID" | cb
zigchaind tx dex remove-liquidity 200$POOL_1_ID $TX_FLAGS -y --from $ACCOUNT | grep "txhash" | cb

echo "Waiting for: to allow for the liquidity to be removed" | cb
sleep $SLEEP

echo "\n Info: LP1 $POOL_1_ID after removing liquidity" | cb
POOL_1_JSON_OUTPUT=$(zigchaind q dex pool $POOL_1_ID $TX_Q_FLAGS -o json)
POOL_1_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .lp_token .amount')
POOL_1_BASE_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .coins[0] .amount')
POOL_1_QUOTE_AMOUNT=$(echo $POOL_1_JSON_OUTPUT | jq -r '.pool .coins[1] .amount')

# LP Tokens should decrease by 200.
# 20200 - 200 = 20000
if [ "$POOL_1_AMOUNT" != "20000" ]; then
  echo "Failed to remove liquidity from LP1 $POOL_1_ID - $FULL_BASE and $FULL_QUOTE pool"
  echo "POOL_1_AMOUNT: $POOL_1_AMOUNT expected 20000"
  exit
fi

echo "\t ✅ LP1 $POOL_1_ID pool has $POOL_1_AMOUNT LP tokens after removing liquidity (20000)"

# When removing liquidity, the account should receive from base
# 200 * (10120 / 20200) = 100
# Pool Base amount should decrease by 100 -> 10120 - 100 = 10020
if [ "$POOL_1_BASE_AMOUNT" != "10020" ]; then
  echo "Failed to remove liquidity from LP1 $POOL_1_ID - $FULL_BASE and $FULL_QUOTE pool"
  echo "POOL_1_COIN_1_AMOUNT: $POOL_1_BASE_AMOUNT expected 10020"
  exit
fi

echo "\t ✅ LP1 $POOL_1_ID pool has $POOL_1_BASE_AMOUNT $FULL_BASE after removing liquidity (10120 - 100 = 10020)"

# When removing liquidity, the account should receive from quote
# 200 * (40328 / 20200) = 400
# Pool Quote amount should decrease by 400 -> 40328 - 400 ~= 39929
if [ "$POOL_1_QUOTE_AMOUNT" != "39929" ]; then
  echo "Failed to remove liquidity from LP1 $POOL_1_ID - $FULL_BASE and $FULL_QUOTE pool"
  echo "POOL_1_COIN_2_AMOUNT: $POOL_1_QUOTE_AMOUNT expected 39929"
  exit
fi

echo "\t ✅ LP1 $POOL_1_ID pool has $POOL_1_QUOTE_AMOUNT $FULL_QUOTE after removing liquidity (40328 - 400 ~= 39929)"

echo "\n Info: Account $ACCOUNT after removing liquidity" | cb
ACCOUNT_BALANCES_JSON_OUTPUT=$(zigchaind q bank balances $ACCOUNT_ADDRESS $TX_Q_FLAGS -o json)
ACCOUNT_POOL_1_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$POOL_1_ID'") .amount')
ACCOUNT_BASE_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_BASE'") .amount')
ACCOUNT_QUOTE_BALANCE=$(echo $ACCOUNT_BALANCES_JSON_OUTPUT | jq -r '.balances[] | select(.denom == "'$FULL_QUOTE'") .amount')

# The number of LP tokens that the user has is the ones minted before 19200 - 200
if [ "$ACCOUNT_POOL_1_BALANCE" != "19000" ]; then
  echo "Failed to mint $POOL_1_ID token"
  echo "ACCOUNT_POOL_1_BALANCE: $ACCOUNT_POOL_1_BALANCE and we expected 19000"
  exit
fi

echo "\t ✅ Account: $ACCOUNT has $ACCOUNT_POOL_1_BALANCE $POOL_1_ID Tokens after removing liquidity (19200 - 200 = 19000)"

# The balance of Base token before removing liquidity is 99979380.
# After removing liquidity should be 99979380 + 100 = 99979480
if [ "$ACCOUNT_BASE_BALANCE" != "99979480" ]; then
  echo "ACCOUNT_BASE_BALANCE: $ACCOUNT_BASE_BALANCE and we expect 99979480" | cb
  echo "FAIL: Failed to remove liquidity from $FULL_BASE" | cb
  exit
fi

echo "\t ✅ Account $ACCOUNT has $ACCOUNT_BASE_BALANCE $FULL_BASE after removing liquidity (99979380 + 100 = 99979480)"

# The balance of Quote token before removing liquidity is 950035.
# After removing liquidity should be 949636 + 399 ~= 950035
if [ "$ACCOUNT_QUOTE_BALANCE" != "950035" ]; then
  echo "ACCOUNT_QUOTE_BALANCE: $ACCOUNT_QUOTE_BALANCE and we expect 950035" | cb
  echo "FAIL: Failed to remove liquidity from $FULL_QUOTE" | cb
  exit
fi

echo "\t ✅ Account $ACCOUNT has $ACCOUNT_QUOTE_BALANCE $FULL_QUOTE after removing liquidity (949636 + 400 = 950036)"





# --------------------------------------------------------------------------------------------
# SIMULATE SWAP
# --------------------------------------------------------------------------------------------

# Iterate over the pool with different swap amounts

#eth_price=2500
#TOTAL=0
#multiplier=1
#
#pv "Multiplier" multiplier
#pv "ETH price" eth_price
#
#declare -i amount=0
# Loop with a step
#for ((i=1; i<=45; i+=1)); do
#  amount=$((0.5*$multiplier))
#  echo "-----------------------------------"
#  echo "❓ EXEC: query swap $amount$FULL_BASE ($(echo "scale=0 ; $amount / $multiplier" | bc))" | cb
#  echo "-----------------------------------"
#  SWAP_OUT=$(q dex swap-in $POOL_1_ID $amount$FULL_BASE -o json)
#  RECEIVED_AMOUNT=$(echo $SWAP_OUT | jq -r '.out .amount')
#
#  EARN_AMOUNT=$(($RECEIVED_AMOUNT - ($amount * $eth_price)))
#  echo "$RECEIVED_AMOUNT - ($amount * $eth_price) = $EARN_AMOUNT"
#  echo "$(echo "scale=2 ; $RECEIVED_AMOUNT / $multiplier" | bc) - ($(echo "scale=2 ; $amount / $multiplier" | bc) * $eth_price) = $(echo "scale=2 ; $EARN_AMOUNT / $multiplier" | bc)"
#  echo "-----------------------------------"
#  echo "🚀 EXEC: query swap $amount$FULL_BASE ($(echo "scale=0 ; $amount / $multiplier" | bc))" | cb
#  SWAP_OUT=$(tx dex swap $POOL_1_ID $amount$FULL_BASE -y -o json)
#  RECEIVED_AMOUNT=$(echo $SWAP_OUT | jq -r '.out .amount')
#
#  EARN_AMOUNT_SHORT=$(($EARN_AMOUNT / $multiplier))
#  TOTAL=$(($TOTAL + EARN_AMOUNT_SHORT))
#
#  if (( EARN_AMOUNT > 0 )); then
#      echo "Total: $TOTAL so far"
#  else
#      echo "Total: $TOTAL earned in $i swaps"
#      exit 0
#  fi
#
#  sleep 1
#
#  BASE_BALANCE=$(q bank balance $ACCOUNT_ADDRESS $FULL_BASE -o json | jq  -r '.balance .amount')
#  QUOTE_BALANCE=$(q bank balance $ACCOUNT_ADDRESS $FULL_QUOTE -o json | jq  -r '.balance .amount')
#done

echo "-----------------------------------"
echo "-----------------------------------"

# Iterate over the pool with different swap amounts
# Define the start, end, and step

#start=$((10*$multiplier))
#end=$((40*$multiplier))
#step=$((10*$multiplier))
#
#TOTAL=0
## Loop with a step
#for ((i=1; i<=45; i+=1)); do
#  amount=$((100*$multiplier))
#  echo "-----------------------------------"
#  echo "❓ EXEC: query swap $amount$FULL_QUOTE ($(echo "scale=0 ; $amount / $multiplier" | bc))" | cb
#  echo "-----------------------------------"
#  SWAP_OUT=$(q dex swap-in $POOL_1_ID $amount$FULL_QUOTE -o json)
#  RECEIVED_AMOUNT=$(echo $SWAP_OUT | jq -r '.coinOut .amount')
#  EARN_AMOUNT=$(($RECEIVED_AMOUNT * $eth_price - $amount))
#  echo "$RECEIVED_AMOUNT * $eth_price - $amount = $EARN_AMOUNT"
#  echo "$(echo "scale=2 ; $RECEIVED_AMOUNT / $multiplier" | bc) * $eth_price - $(echo "scale=2 ; $amount / $multiplier" | bc) = $(echo "scale=2 ; $EARN_AMOUNT / $multiplier" | bc)"
#  echo "-----------------------------------"
#  echo "🚀 EXEC: query swap $amount$FULL_QUOTE ($(echo "scale=0 ; $amount / $multiplier" | bc))" | cb
#  SWAP_OUT=$(tx dex swap-exact-in $POOL_1_ID $amount$FULL_QUOTE -y -o json)
#  RECEIVED_AMOUNT=$(echo $SWAP_OUT | jq -r '.coinOut .amount')
#
#  EARN_AMOUNT_SHORT=$(($EARN_AMOUNT / $multiplier))
#  TOTAL=$(($TOTAL + EARN_AMOUNT_SHORT))
#
#  if (( EARN_AMOUNT > 0 )); then
#      echo "Total: $TOTAL so far"
#  else
#      echo "Total: $TOTAL earned in $i swaps"
#      exit 0
#  fi
#
#  sleep 1
#
#  BASE_BALANCE=$(q bank balance $ACCOUNT_ADDRESS $FULL_BASE -o json | jq  -r '.balance .amount')
#  QUOTE_BALANCE=$(q bank balance $ACCOUNT_ADDRESS $FULL_QUOTE -o json | jq  -r '.balance .amount')
#done

echo "🎉 SUCCESS: All checks passed! 🎉" | cb

echo "To load the environment variables, run:"
echo "source $HOME/src/zigchain/sh/env.sh" | cb bash
echo "source /tmp/.$MODULE_NAME.mod.env" | cb bash
