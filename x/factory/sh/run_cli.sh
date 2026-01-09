#!/usr/bin/env zsh
# Runs common commands to test the contract on-fly during development
# not meant to replace unit or integration tests

MODULE_NAME="factory"

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
  eval "printf '%s=%s\\n' \"\$var_name\" \"\${$var_name}\" >> /tmp/.$MODULE_NAME.mod.env"

  echo "# - - - - - - - - - - - - - - - - - - - -"
  echo
}

# --------------------------------------------------------------------------------------------
# SETUP VARIABLES
# --------------------------------------------------------------------------------------------

ACCOUNT=z
pv "z account" ACCOUNT

ACCOUNT_ADDRESS=$(zigchaind keys show $ACCOUNT -a)
pv "z account address" ACCOUNT_ADDRESS

# ADD a random 10-chars suffix to denom name to avoid conflicts on repeated runs
RANDOM_APPEND=$(openssl rand -hex 5 | tr -dc 'a-fA-F')
SUBDENOM_NAME="panda$RANDOM_APPEND"

# General flags used in commands
pv "TX_FLAGS" $TX_FLAGS


# --------------------------------------------------------------------------------------------
# CREATING TOKEN
# --------------------------------------------------------------------------------------------

SHA256=$(printf '%s' "{"a":1,"b":2}" | sha256sum | awk '{print $1}')

echo "\n 🚀 EXEC: creating $SUBDENOM_NAME token" | cb
CREATE_DENOM_JSON_OUTPUT=$(
  zigchaind tx factory create-denom $SUBDENOM_NAME 1000000000 true \
  'ipfs://bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi' \
  $SHA256 \
  --from $ACCOUNT $TX_FLAGS --yes -o json
  )

echo "Extract: create denom tx id" | cb
CREATE_DENOM_TX_ID=$(echo $CREATE_DENOM_JSON_OUTPUT | jq -r '.txhash')

pv "CREATE_DENOM_TX_ID" CREATE_DENOM_TX_ID

echo "Waiting for: create denom tx to be included in a block" | cb
sleep $SLEEP
echo

echo "Extract: create denom tx id" | cb
CREATE_DENOM_TX_JSON_OUTPUT=$(q tx "$CREATE_DENOM_TX_ID")

echo "Extract: full denom id" | cb
FULL_DENOM_ID=$(echo "$CREATE_DENOM_TX_JSON_OUTPUT" | jq -r ".events[] | select(.type == \"denom_created\") | .attributes[] | select(.key == \"denom\") | .value")

pv "FULL_DENOM_ID" FULL_DENOM_ID

# --------------------------------------------------------------------------------------------
# MINTING TOKEN
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: Minting 1000 $SUBDENOM_NAME tokens" | cb
MINT_JSON_OUTPUT=$(
  zigchaind tx factory mint 1000$FULL_DENOM_ID $ACCOUNT_ADDRESS \
  --from $ACCOUNT $TX_FLAGS --yes -o json
  )

echo "Extract: mint tx id" | cb
MINT_TX_ID=$(echo $MINT_JSON_OUTPUT | jq -r '.txhash')

pv "MINT_TX_ID" MINT_TX_ID

echo "Waiting for: mint tx to be included in a block" | cb
sleep $SLEEP
echo

echo "Extract: mint tx id" | cb
MINT_TX_JSON_OUTPUT=$(q tx "$MINT_TX_ID")

echo "Check: account balance" | cb
zigchaind q bank balance $ACCOUNT_ADDRESS $FULL_DENOM_ID $TX_Q_FLAGS | cb

echo
echo "- - - - - - - - - - - - - -"
echo

# --------------------------------------------------------------------------------------------
# BURNING TOKEN
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: Burning 250 $SUBDENOM_NAME tokens" | cb
BURN_JSON_OUTPUT=$(
  zigchaind tx factory burn 250$FULL_DENOM_ID \
  --from $ACCOUNT $TX_FLAGS --yes -o json
  )

echo "Extract: burn tx id" | cb
BURN_TX_ID=$(echo $BURN_JSON_OUTPUT | jq -r '.txhash')

pv "BURN_TX_ID" BURN_TX_ID

echo "Waiting for: burn tx to be included in a block" | cb
sleep $SLEEP
echo

echo "Check: account balance" | cb
TOKEN_BANK_INFO=$(zigchaind q bank balance $ACCOUNT_ADDRESS $FULL_DENOM_ID $TX_Q_FLAGS)
echo $TOKEN_BANK_INFO | cb

# Check that the factory balance is updated
TOKEN_FACTORY_INFO=$(zigchaind q factory show-denom $FULL_DENOM_ID $TX_Q_FLAGS --output json)
echo $TOKEN_FACTORY_INFO | cb
MINTING_CAP=$(echo $TOKEN_FACTORY_INFO | jq -r '.minting_cap')
TOTAL_MINTED=$(echo $TOKEN_FACTORY_INFO | jq -r '.total_minted')
TOTAL_SUPPLY=$(echo $TOKEN_FACTORY_INFO | jq -r '.total_supply')

echo "Check: minting cap, total minted and total supply" | cb
if [ "$MINTING_CAP" != "1000000000" ] || [ "$TOTAL_MINTED" != "1000" ] || [ "$TOTAL_SUPPLY" != "750" ]; then
  echo "‼️ Failed to burn tokens" | cb
  q tx $BURN_TX_ID | grep raw_log
  exit 1
fi

# --------------------------------------------------------------------------------------------
# UPDATE MINTING CAP
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: Updating minting cap to a higher value from current supply" | cb
UPDATE_MINTING_CAP_JSON_OUTPUT=$(
  zigchaind tx factory update-minting-cap $FULL_DENOM_ID 200000000000 true \
  --from $ACCOUNT $TX_FLAGS --yes -o json
  )

echo "Extract: update minting cap tx id" | cb
UPDATE_MINTING_CAP_TX_ID=$(echo $UPDATE_MINTING_CAP_JSON_OUTPUT | jq -r '.txhash')

pv "UPDATE_MINTING_CAP_TX_ID" UPDATE_MINTING_CAP_TX_ID

echo "Waiting for: update minting cap tx to be included in a block" | cb
sleep $SLEEP

echo "\n ✨INFO: get denom" | cb
zigchaind q factory show-denom $FULL_DENOM_ID $TX_Q_FLAGS | cb


# --------------------------------------------------------------------------------------------
# UPDATE DENOM METADATA AUTH
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: Updating denom metadata auth (admins)" | cb

echo "\n Info: Check first the current admin tokens" | cb
DENOMS_BY_ADMIN_JSON_OUTPUT=$(zigchaind query factory denoms-by-admin $ACCOUNT_ADDRESS $TX_Q_FLAGS -o json)

# Check if in DENOMS_BY_ADMIN_JSON_OUTPUT there is a key "denoms"
if echo $DENOMS_BY_ADMIN_JSON_OUTPUT | jq -e '.denoms' >/dev/null; then
  echo "Some denoms found owned by $ACCOUNT, checking if the current one is there" | cb
  DENOMS_BY_ADMIN=$(echo $DENOMS_BY_ADMIN_JSON_OUTPUT | jq -r '.denoms[]')
  echo "DENOMS_BY_ADMIN: $DENOMS_BY_ADMIN" | cb

  if [[ ! $DENOMS_BY_ADMIN =~ $FULL_DENOM_ID ]]; then
    echo "‼️ Failed to find the new denom in the list of denoms by admin" | cb
    echo "DENOMS_BY_ADMIN: $DENOMS_BY_ADMIN" | cb
    exit 1
  fi
else
  echo "No denoms found owned by $ACCOUNT, as expected" | cb
fi

echo "\t ✅ Denom is in the list of denoms by admin, as expected" | cb

ACCOUNT_ADMIN=$ACCOUNT
ACCOUNT_ADDRESS_ADMIN=$ACCOUNT_ADDRESS

ACCOUNT_METADATA=zuser2
ACCOUNT_ADDRESS_METADATA=$(zigchaind keys show $ACCOUNT_METADATA -a)
pv "meta account address" ACCOUNT_ADDRESS_METADATA

UPDATE_DENOM_METADATA_AUTH_JSON_OUTPUT=$(
  zigchaind tx factory update-denom-metadata-auth $FULL_DENOM_ID \
  $ACCOUNT_ADDRESS_METADATA \
  --from $ACCOUNT $TX_FLAGS_NO_GAS_ADJ --gas-adjustment 1.5 --yes -o json
  )

UPDATE_DENOM_METADATA_AUTH_TX_ID=$(echo $UPDATE_DENOM_METADATA_AUTH_JSON_OUTPUT | jq -r '.txhash')
pv "UPDATE_DENOM_METADATA_AUTH_TX_ID" UPDATE_DENOM_METADATA_AUTH_TX_ID

echo "Waiting for: update denom metadata auth tx to be included in a block" | cb
sleep $SLEEP

echo "\n ✨INFO: get denom" | cb
zigchaind q factory show-denom $FULL_DENOM_ID $TX_Q_FLAGS | cb

# Check that the new admins are set
DENOM_JSON_OUTPUT=$(zigchaind query factory denom-auth $FULL_DENOM_ID $TX_Q_FLAGS -o json)
echo DENOM_JSON_OUTPUT
BANK_ADMIN=$(echo $DENOM_JSON_OUTPUT | jq -r '.denom_auth.bank_admin')
METADATA_ADMIN=$(echo $DENOM_JSON_OUTPUT | jq -r '.denom_auth.metadata_admin')

if [ "$BANK_ADMIN" != "$ACCOUNT_ADDRESS_ADMIN" ] || [ "$METADATA_ADMIN" != "$ACCOUNT_ADDRESS_METADATA" ]; then
  echo "‼️ Failed to update denom auth" | cb
  echo "BANK_ADMIN is $BANK_ADMIN and it should be $ACCOUNT_ADDRESS_ADMIN" | cb
  echo "METADATA_ADMIN is $METADATA_ADMIN and it should be $ACCOUNT_ADDRESS_METADATA" | cb
  q tx $UPDATE_DENOM_AUTH_TX_ID
  exit 1
fi

# Check that quering the list of denoms by admin returns the new denom for bank admin and for metadata admin
echo "\n ✨INFO: get denoms by admin to ensure that it is there" | cb
DENOMS_BY_ADMIN_JSON_OUTPUT=$(zigchaind query factory denoms-by-admin $ACCOUNT_ADDRESS_ADMIN $TX_Q_FLAGS -o json)
echo $DENOMS_BY_ADMIN_JSON_OUTPUT | jq
DENOMS_BY_ADMIN=$(echo $DENOMS_BY_ADMIN_JSON_OUTPUT | jq -r '.denoms[]')
echo "DENOMS_BY_ADMIN: $DENOMS_BY_ADMIN" | cb
if [[ ! $DENOMS_BY_ADMIN =~ $FULL_DENOM_ID ]]; then
  echo "‼️ Failed to find the new denom in the list of denoms by admin" | cb
  echo "DENOMS_BY_ADMIN: $DENOMS_BY_ADMIN" | cb
  q tx $UPDATE_DENOM_AUTH_TX_ID
  exit 1
fi
echo "\t ✅ Denom is in the list of denoms by admin, as expected" | cb

# Check that quering the list of denoms by metadata admin returns the new denom for metadata admin
echo "\n ✨INFO: get denoms by metadata admin to ensure that it is there" | cb
DENOMS_BY_META_JSON_OUTPUT=$(zigchaind query factory denoms-by-admin $ACCOUNT_ADDRESS_METADATA $TX_Q_FLAGS -o json)
echo $DENOMS_BY_META_JSON_OUTPUT | jq
DENOMS_BY_META=$(echo $DENOMS_BY_META_JSON_OUTPUT | jq -r '.denoms[]')
echo "DENOMS_BY_META: $DENOMS_BY_META" | cb
if [[ ! $DENOMS_BY_META =~ $FULL_DENOM_ID ]]; then
  echo "‼️ Failed to find the new denom in the list of denoms by metadata admin" | cb
  echo "DENOMS_BY_META: $DENOMS_BY_META" | cb
  q tx $UPDATE_DENOM_METADATA_AUTH_TX_ID
  exit 1
fi

echo "\t ✅ Denom is in the list of denoms by metadata admin, as expected" | cb


# --------------------------------------------------------------------------------------------
# UPDATE DENOM ADMIN AUTH (REQUIRES PROPOSE AND CLAIM)
# --------------------------------------------------------------------------------------------

ACCOUNT_ADMIN=zuser1

ACCOUNT_ADDRESS_ADMIN=$(zigchaind keys show $ACCOUNT_ADMIN -a)
pv "admin account address" ACCOUNT_ADDRESS_ADMIN

ACCOUNT_METADATA_NEW=zuser4
ACCOUNT_ADDRESS_METADATA_NEW=$(zigchaind keys show $ACCOUNT_METADATA_NEW -a)

echo "\n 🚀 EXEC: Updating denom ADMIN auth to $ACCOUNT_ADMIN with address $ACCOUNT_ADDRESS_ADMIN" | cb

PROPOSE_DENOM_ADMIN_AUTH_JSON_OUTPUT=$(
  zigchaind tx factory propose-denom-admin $FULL_DENOM_ID \
  $ACCOUNT_ADDRESS_ADMIN \
  $ACCOUNT_ADDRESS_METADATA_NEW \
  --from $ACCOUNT $TX_FLAGS_NO_GAS_ADJ --gas-adjustment 1.5 --yes -o json
  )

PROPOSE_DENOM_ADMIN_AUTH_TX_ID=$(echo $PROPOSE_DENOM_ADMIN_AUTH_JSON_OUTPUT | jq -r '.txhash')
pv "PROPOSE_DENOM_ADMIN_AUTH_TX_ID" PROPOSE_DENOM_ADMIN_AUTH_TX_ID

echo "Waiting for: propose denom metadata auth tx to be included in a block" | cb
sleep $SLEEP

echo "\n ✨INFO: get denom and admins should remain the same" | cb
zigchaind q factory show-denom $FULL_DENOM_ID $TX_Q_FLAGS | cb

# Check that the admins are still the same
DENOM_JSON_OUTPUT=$(zigchaind query factory denom-auth $FULL_DENOM_ID $TX_Q_FLAGS -o json)
echo DENOM_JSON_OUTPUT
BANK_ADMIN=$(echo $DENOM_JSON_OUTPUT | jq -r '.denom_auth.bank_admin')
METADATA_ADMIN=$(echo $DENOM_JSON_OUTPUT | jq -r '.denom_auth.metadata_admin')

if [ "$BANK_ADMIN" != "$ACCOUNT_ADDRESS" ] || [ "$METADATA_ADMIN" != "$ACCOUNT_ADDRESS_METADATA" ]; then
  echo "‼️ Failed to update denom auth" | cb
  echo "BANK_ADMIN is $BANK_ADMIN and it should be $ACCOUNT_ADDRESS" | cb
  echo "METADATA_ADMIN is $METADATA_ADMIN and it should be $ACCOUNT_ADDRESS_METADATA" | cb
  q tx $UPDATE_DENOM_AUTH_TX_ID
  exit 1
fi

echo "\n 🚀 EXEC: Claiming the new denom admin auth" | cb

CLAIM_DENOM_ADMIN_AUTH_JSON_OUTPUT=$(
  zigchaind tx factory claim-denom-admin $FULL_DENOM_ID \
  --from $ACCOUNT_ADMIN $TX_FLAGS_NO_GAS_ADJ --gas-adjustment 1.5 --yes -o json
  )

CLAIM_DENOM_ADMIN_AUTH_TX_ID=$(echo $CLAIM_DENOM_ADMIN_AUTH_JSON_OUTPUT | jq -r '.txhash')
pv "CLAIM_DENOM_ADMIN_AUTH_TX_ID" CLAIM_DENOM_ADMIN_AUTH_TX_ID

echo "Waiting for: claim denom admin auth tx to be included in a block" | cb
sleep $SLEEP

echo "\n ✨INFO: get denom" | cb
zigchaind q factory show-denom $FULL_DENOM_ID $TX_Q_FLAGS | cb

# Check that the new admins are set
DENOM_JSON_OUTPUT=$(zigchaind query factory denom-auth $FULL_DENOM_ID $TX_Q_FLAGS -o json)
echo $DENOM_JSON_OUTPUT

BANK_ADMIN=$(echo $DENOM_JSON_OUTPUT | jq -r '.denom_auth.bank_admin')
METADATA_ADMIN=$(echo $DENOM_JSON_OUTPUT | jq -r '.denom_auth.metadata_admin')

if [ "$BANK_ADMIN" != "$ACCOUNT_ADDRESS_ADMIN" ] || [ "$METADATA_ADMIN" != "$ACCOUNT_ADDRESS_METADATA_NEW" ]; then
  echo "‼️ Failed to update denom auth to $ACCOUNT_ADDRESS_ADMIN and $ACCOUNT_ADDRESS_METADATA_NEW" | cb
  echo "BANK_ADMIN is $BANK_ADMIN and it should be $ACCOUNT_ADDRESS_ADMIN" | cb
  echo "METADATA_ADMIN is $METADATA_ADMIN and it should be $ACCOUNT_ADDRESS_METADATA_NEW" | cb
  q tx $CLAIM_DENOM_ADMIN_AUTH_TX_ID
  exit 1
fi

echo "✅ Successfully updated denom $FULL_DENOM_ID auth to Bank admin: $ACCOUNT_ADDRESS_ADMIN \n
and Meta admin $ACCOUNT_ADDRESS_METADATA" | cb

# Query the list of denom by admin
echo "\n ✨INFO: get denoms by old admin to ensure that it is not longer there" | cb

# Check that the denom is not in the list of denoms by the old admin
DENOMS_BY_OLD_ADMIN_JSON_OUTPUT=$(zigchaind query factory denoms-by-admin $ACCOUNT_ADDRESS $TX_Q_FLAGS -o json)
# Check if in DENOMS_BY_OLD_ADMIN_JSON_OUTPUT there is a key "denoms"
if echo $DENOMS_BY_OLD_ADMIN_JSON_OUTPUT | jq -e '.denoms' >/dev/null; then
  echo "Some denoms found owned by $ACCOUNT, checking if the current one is there" | cb
  DENOMS_BY_OLD_ADMIN=$(echo $DENOMS_BY_OLD_ADMIN_JSON_OUTPUT | jq -r '.denoms[]')
  echo "DENOMS_BY_OLD_ADMIN: $DENOMS_BY_OLD_ADMIN" | cb

  if [[ $DENOMS_BY_OLD_ADMIN =~ $FULL_DENOM_ID ]]; then
    echo "‼️ Failed to find the new denom in the list of denoms by the old admin" | cb
    echo "DENOMS_BY_OLD_ADMIN: $DENOMS_BY_OLD_ADMIN" | cb
    q tx $CLAIM_DENOM_ADMIN_AUTH_TX_ID
    exit 1
  fi
  echo "\t ✅ Denoms is not in the list of denoms by the old admin, as expected"
else
  echo "\t ✅ No denoms found owned by $ACCOUNT, as expected"
fi

echo "\t ✅ Denom is not in the list of denoms by the old admin, as expected"

# Check that the new denom is in the list for new admin
echo "\n ✨INFO: get denoms by new admin to ensure that it is there" | cb
DENOMS_BY_ADMIN_JSON_OUTPUT=$(zigchaind query factory denoms-by-admin $ACCOUNT_ADDRESS_ADMIN $TX_Q_FLAGS -o json)
DENOMS_BY_ADMIN=$(echo $DENOMS_BY_ADMIN_JSON_OUTPUT | jq -r '.denoms[]')
echo "DENOMS_BY_ADMIN: $DENOMS_BY_ADMIN" | cb

if [[ ! $DENOMS_BY_ADMIN =~ $FULL_DENOM_ID ]]; then
  echo "‼️ Failed to find the new denom in the list of denoms by admin" | cb
  echo "DENOMS_BY_ADMIN: $DENOMS_BY_ADMIN" | cb
  q tx $CLAIM_DENOM_ADMIN_AUTH_TX_ID
  exit 1
fi
echo "\t ✅ Denom is in the list of denoms by the new admin, as expected"


# check that for the metadata admin the denom is also there
echo "\n ✨INFO: get denoms by metadata admin to ensure that it is there" | cb
DENOMS_BY_META_JSON_OUTPUT=$(zigchaind query factory denoms-by-admin $ACCOUNT_ADDRESS_METADATA_NEW $TX_Q_FLAGS -o json)
DENOMS_BY_META=$(echo $DENOMS_BY_META_JSON_OUTPUT | jq -r '.denoms[]')
echo "DENOMS_BY_META: $DENOMS_BY_META" | cb
if [[ ! $DENOMS_BY_META =~ $FULL_DENOM_ID ]]; then
  echo "‼️ Failed to find the new denom in the list of denoms by metadata admin" | cb
  echo "DENOMS_BY_META: $DENOMS_BY_META" | cb
  q tx $CLAIM_DENOM_ADMIN_AUTH_TX_ID
  exit 1
fi

echo "\t ✅ Denom is in the list of denoms by new metadata admin, as expected"

# ensure that the old metadata admin is not in the list of denoms by metadata admin
echo "\n ✨INFO: get denoms by old metadata admin to ensure that it is not there" | cb
DENOMS_BY_OLD_META_JSON_OUTPUT=$(zigchaind query factory denoms-by-admin $ACCOUNT_ADDRESS_METADATA $TX_Q_FLAGS -o json)
# Check if in DENOMS_BY_OLD_META_JSON_OUTPUT there is a key "denoms"
if echo $DENOMS_BY_OLD_META_JSON_OUTPUT | jq -e '.denoms' >/dev/null; then
  echo "Some denoms found owned by $ACCOUNT_METADATA, checking if the current one is there" | cb
  DENOMS_BY_OLD_META=$(echo $DENOMS_BY_OLD_META_JSON_OUTPUT | jq -r '.denoms[]')
  echo "DENOMS_BY_OLD_META: $DENOMS_BY_OLD_META" | cb

  if [[ $DENOMS_BY_OLD_META =~ $FULL_DENOM_ID ]]; then
    echo "‼️ Failed to find the new denom in the list of denoms by the old metadata admin" | cb
    echo "DENOMS_BY_OLD_META: $DENOMS_BY_OLD_META" | cb
    q tx $CLAIM_DENOM_ADMIN_AUTH_TX_ID
    exit 1
  fi
  echo "\t ✅ Denom is not in the list of denoms by the old metadata admin, as expected"
else
  echo "\t ✅ No denoms found owned by $ACCOUNT_METADATA, as expected"
fi

echo "\t ✅ Denom is not in the list of denoms by the old metadata admin, as expected"

# --------------------------------------------------------------------------------------------
# META ADMIN UPDATES URI AND URI HASH
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: Meta Admin updates URI to new_uri and URI hash to $SHA256" | cb
URI="ipfs://bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi_new"
SHA256=$(printf '%s' "{"a":1,"b":2}" | sha256sum | awk '{print $1}')

UPDATE_URI_JSON_OUTPUT=$(
  zigchaind tx factory update-uri $FULL_DENOM_ID \
  $URI \
  $SHA256 \
  --from $ACCOUNT_ADDRESS_METADATA_NEW $TX_FLAGS_NO_GAS_ADJ --gas-adjustment 1.5 --yes -o json
  )

echo "Extract: update uri tx id" | cb
UPDATE_URI_TX_ID=$(echo $UPDATE_URI_JSON_OUTPUT | jq -r '.txhash')

pv "UPDATE_URI_TX_ID" UPDATE_URI_TX_ID

echo "Waiting for: update uri tx to be included in a block" | cb
sleep $SLEEP

echo "\n ✨INFO: get denom-metadata" | cb
zigchaind query bank denom-metadata $FULL_DENOM_ID $TX_Q_FLAGS | cb

# Check that the URI and URI hash are updated
DENOM_METADATA_JSON_OUTPUT=$(zigchaind query bank denom-metadata $FULL_DENOM_ID $TX_Q_FLAGS -o json)
URI_OUTPUT=$(echo $DENOM_METADATA_JSON_OUTPUT | jq -r '.metadata.uri')
URI_HASH_OUTPUT=$(echo $DENOM_METADATA_JSON_OUTPUT | jq -r '.metadata.uri_hash')

if [ "$URI_OUTPUT" != $URI ] || [ "$URI_HASH_OUTPUT" != $SHA256 ]; then
  echo "‼️ Failed to update URI and URI hash" | cb
  q tx $UPDATE_URI_TX_ID
  exit 1
fi


# --------------------------------------------------------------------------------------------
# UPDATE DENOM METADATA AUTH (ADMIN)
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: Metadata Admin Updates the metadata denom auth" | cb

ACCOUNT_NEW_METADATA=zuser3
ACCOUNT_NEW_ADDRESS_METADATA=$(zigchaind keys show $ACCOUNT_NEW_METADATA -a)
pv "meta account address" ACCOUNT_NEW_ADDRESS_METADATA

UPDATE_DENOM_AUTH_JSON_OUTPUT=$(
  zigchaind tx factory update-denom-metadata-auth $FULL_DENOM_ID \
  $ACCOUNT_NEW_ADDRESS_METADATA \
  --from $ACCOUNT_ADDRESS_METADATA_NEW $TX_FLAGS_NO_GAS_ADJ --gas-adjustment 1.5 --yes -o json
  )

UPDATE_DENOM_AUTH_TX_ID=$(echo $UPDATE_DENOM_AUTH_JSON_OUTPUT | jq -r '.txhash')
pv "UPDATE_DENOM_AUTH_TX_ID" UPDATE_DENOM_AUTH_TX_ID

echo "Waiting for: update denom auth tx to be included in a block" | cb
sleep $SLEEP

echo "\n ✨INFO: get denom" | cb
zigchaind q factory show-denom $FULL_DENOM_ID $TX_Q_FLAGS | cb

# Check that the new admins are set
DENOM_JSON_OUTPUT=$(zigchaind query factory denom-auth $FULL_DENOM_ID $TX_Q_FLAGS -o json)
METADATA_ADMIN=$(echo $DENOM_JSON_OUTPUT | jq -r '.denom_auth.metadata_admin')

if [ "$METADATA_ADMIN" != "$ACCOUNT_NEW_ADDRESS_METADATA" ]; then
  echo "‼️ Failed to update metadata denom auth to $ACCOUNT_NEW_ADDRESS_METADATA" | cb
  q tx $UPDATE_DENOM_AUTH_TX_ID
  exit 1
fi

# Check that the new denom is in the list
echo "\n ✨INFO: get denoms by new meta admin $ACCOUNT_NEW_METADATA to ensure that it is there" | cb
DENOMS_BY_META_JSON_OUTPUT=$(zigchaind query factory denoms-by-admin $ACCOUNT_NEW_ADDRESS_METADATA $TX_Q_FLAGS -o json)
echo $DENOMS_BY_META_JSON_OUTPUT | jq

DENOMS_BY_META=$(echo $DENOMS_BY_META_JSON_OUTPUT | jq -r '.denoms[]')
echo "DENOMS_BY_META: $DENOMS_BY_META" | cb

if [[ ! $DENOMS_BY_META =~ $FULL_DENOM_ID ]]; then
  echo "‼️ Failed to find the denom in the list of denoms by meta" | cb
  echo "DENOMS_BY_META: $DENOMS_BY_META" | cb
  exit 1
fi
echo "✅ Denom is in the list of denoms by the new meta admin, as expected" | cb

# --------------------------------------------------------------------------------------------
# UPDATE DENOM METADATA
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: Metadata Admin Updates the metadata" | cb
URI="ipfs://bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi_new_uri_6"
SHA256=$(printf '%s' "{"a":1,"b":2}" | sha256sum | awk '{print $1}')

METADATA_DATA=$(
  echo '{
      "description": "Updated metadata",
      "denom_units": [
        {
          "denom": "'"$FULL_DENOM_ID"'",
          "exponent": 0
        },
        {
          "denom": "m'"$SUBDENOM_NAME"'",
          "exponent": 3
        }
      ],
      "base": "'"$FULL_DENOM_ID"'",
      "display": "m'"$SUBDENOM_NAME"'",
      "name": "pandacdcf",
      "symbol": "'"$FULL_DENOM_ID"'",
      "uri": "'"$URI"'",
      "uri_hash": "'"$SHA256"'"
  }'
)
echo $METADATA_DATA | jq

UPDATE_METADATA_JSON_OUTPUT=$(
  zigchaind tx factory set-denom-metadata \
  $METADATA_DATA \
  --from $ACCOUNT_NEW_METADATA $TX_FLAGS_NO_GAS_ADJ --gas-adjustment 1.5 --yes -o json
)

UPDATE_METADATA_TX_ID=$(echo $UPDATE_METADATA_JSON_OUTPUT | jq -r '.txhash')
pv "UPDATE_METADATA_TX_ID" UPDATE_METADATA_TX_ID

echo "Waiting for: update metadata tx to be included in a block" | cb
sleep $SLEEP

echo "\n ✨INFO: get denom-metadata" | cb
zigchaind query bank denom-metadata $FULL_DENOM_ID $TX_Q_FLAGS | cb

# Check that the metadata is updated
DENOM_METADATA_JSON_OUTPUT=$(zigchaind query bank denom-metadata $FULL_DENOM_ID $TX_Q_FLAGS -o json)
DESCRIPTION=$(echo $DENOM_METADATA_JSON_OUTPUT | jq -r '.metadata.description')

if [ "$DESCRIPTION" != "Updated metadata" ]; then
  echo "‼️ Failed to update metadata" | cb
  q tx $UPDATE_METADATA_TX_ID
  exit 1
fi

# --------------------------------------------------------------------------------------------
# DISABLE MINTING CAP
# --------------------------------------------------------------------------------------------

echo "\n 🚀 EXEC: New Admin Updating minting cap - disable" | cb
DISABLE_MINTING_CAP_JSON_OUTPUT=$(
  zigchaind tx factory update-minting-cap $FULL_DENOM_ID 100000000000 false \
  --from $ACCOUNT_ADMIN $TX_FLAGS_NO_GAS_ADJ --gas-adjustment 1.5 --yes -o json
  )

DISABLE_MINTING_CAP_TX_ID=$(echo $DISABLE_MINTING_CAP_JSON_OUTPUT | jq -r '.txhash')

pv "DISABLE_MINTING_CAP_TX_ID" DISABLE_MINTING_CAP_TX_ID

echo "Waiting for: disable minting cap tx to be included in a block" | cb
sleep $SLEEP
echo

echo "\n ✨INFO: get denom" | cb
zigchaind q factory show-denom $FULL_DENOM_ID | cb

# Check that canChangeMintingCap is false
DENOM_JSON_OUTPUT=$(zigchaind q factory show-denom $FULL_DENOM_ID $TX_Q_FLAGS $TX_Q_FLAGS -o json)
CAN_CHANGE_MINTING_CAP=$(echo $DENOM_JSON_OUTPUT | jq -r '.can_change_minting_cap')

if [ "$CAN_CHANGE_MINTING_CAP" != "false" ]; then
  echo "‼️ Failed change the minting cap" | cb
  q tx $DISABLE_MINTING_CAP_TX_ID
  exit 1
fi

echo "\n 🎉 SUCCESS: All checks passed! 🎉" | cb