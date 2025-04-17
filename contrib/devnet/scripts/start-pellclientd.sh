#!/bin/bash

# This script is used to start PellClient for the localnet
# An optional argument can be passed and can have the following value:
# background: start the PellClient in the background, this prevent the image from being stopped when PellClient must be restarted

# sepolia is used in chain migration tests, this functions set the sepolia endpoint in the pellclient_config.json

## XXX: Must be an archive node's rpc
set_bsc_endpoint() {
  jq ".EVMChainConfigs.\"97\".Endpoint = \"$BSC_EXTERNAL_RPC_URL\"" /root/.pellcored/config/pellclient_config.json > tmp.json && mv tmp.json /root/.pellcored/config/pellclient_config.json
}

cat << EOF > /root/password.file
${CLIENT_HOTKEY_PASSWORD}
${CLIENT_TSS_PASSWORD}
EOF

set_bsc_endpoint
pellclientd start < /root/password.file