#!/usr/bin/env zsh
# Example how to get gas_estimate using JSON file

# Stop the script if any command fails or if an unset variable is used
set -o errexit -o nounset -o pipefail

# Source environment variables (adjust a path as necessary)
# shellcheck source=/dev/null
. $HOME/src/zigchain/sh/env.sh

ACCOUNT=z
ACCOUNT_ADDRESS=$(zigchaind keys show $ACCOUNT -a)
CHAIN_ID=zigchain

# Generate a unique subdenom name to avoid conflicts
SUBDENOM_NAME="abc.$(openssl rand -hex 5)"

echo "ACCOUNT_ADDRESS=$ACCOUNT_ADDRESS"

# Prepare the transaction JSON
cat << EOF > /tmp/tx_create_denom.json
{
  "body": {
    "messages": [
      {
        "@type": "/zigchain.factory.MsgCreateDenom",
        "creator": "$ACCOUNT_ADDRESS",
        "subDenom": "$SUBDENOM_NAME",
        "mintingCap": "1000000000",
        "canChangeMintingCap": true,
        "URI": "ipfs://bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
        "URIHash": "49878d925682dfbc0afaad192f4bcfd1"
      }
    ],
    "memo": "Create denom simulation",
    "timeout_height": "0"
  },
  "auth_info": {
    "signer_infos": [],
    "fee": {
      "amount": [],
      "gas_limit": "200000"
    }
  },
  "signatures": []
}
EOF

# Simulate the transaction
echo "Simulating transaction for denom creation..."
SIMULATION_OUTPUT=$(zigchaind tx simulate /tmp/tx_create_denom.json --from $ACCOUNT --yes -o json)

# Extract gas estimate from simulation
GAS_ESTIMATE=$(echo $SIMULATION_OUTPUT | jq -r '.gas_info.gas_used')

echo "Gas estimate: $GAS_ESTIMATE"

echo "Simulating transaction for denom creation using --dry-run..."
SIMULATION_OUTPUT_2=$(zigchaind tx factory create-denom $SUBDENOM_NAME 1000000000 true \
  'ipfs://bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi' \
  '49878d925682dfbc0afaad192f4bcfd1' \
  --from $ACCOUNT_ADDRESS --yes --dry-run -o json
)

# Extract gas estimate from simulation
echo $SIMULATION_OUTPUT_2