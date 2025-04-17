#!/bin/bash

# This script is used to start the pellcored nodes
# It initializes the nodes and creates the genesis.json file
# It also starts the nodes
# The number of nodes is passed as an first argument to the script
# The second argument is optional and can have the following value:
# 1. upgrade : This is used to test the upgrade process, a proposal is created for the upgrade and the nodes are started using cosmovisor

logt() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

function load_defaults {
  export DAEMON_HOME=${DAEMON_HOME:=/root/.pellcored}
  export NETWORK=${NETWORK:=devnet}
  # export RESTORE_TYPE=${RESTORE_TYPE:=statesync}
  # export TRUST_HEIGHT_DIFFERENCE_STATE_SYNC=${TRUST_HEIGHT_DIFFERENCE_STATE_SYNC:=40000}
  export COSMOVISOR_VERSION=${COSMOVISOR_VERSION:=v1.5.0}
  export CHAIN_ID=${CHAIN_ID:=ignite_186-1}
  export COSMOVISOR_CHECKSUM=${COSMOVISOR_CHECKSUM:=626dfc58c266b85f84a7ed8e2fe0e2346c15be98cfb9f9b88576ba899ed78cdc}
  export VISOR_NAME=${VISOR_NAME:=cosmovisor}
  export DAEMON_NAME=${DAEMON_NAME:=pellcored}
  export DAEMON_ALLOW_DOWNLOAD_BINARIES=${DAEMON_ALLOW_DOWNLOAD_BINARIES:=false}
  export DAEMON_RESTART_AFTER_UPGRADE=${DAEMON_RESTART_AFTER_UPGRADE:=true}
  export UNSAFE_SKIP_BACKUP=${UNSAFE_SKIP_BACKUP:=true}
  export CLIENT_DAEMON_NAME=${CLIENT_DAEMON_NAME:=pellclientd}
  export CLIENT_DAEMON_ARGS=${CLIENT_DAEMON_ARGS:""}
  export CLIENT_SKIP_UPGRADE=${CLIENT_SKIP_UPGRADE:=true}
  export CLIENT_START_PROCESS=${CLIENT_START_PROCESS:=false}
  export MONIKER=${MONIKER:=local-test}
  export RE_DO_START_SEQUENCE=${RE_DO_START_SEQUENCE:=false}

  #DEVNET
  export BINARY_LIST_DEVNET=${BINARY_LIST_DEVNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/main/devnet/binary_list.json}
  export STATE_SYNC_RPC_NODE_FILE_DEVNET=${STATE_SYNC_RPC_NODE_FILE_DEVNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/main/devnet/state_sync_node}
  export RPC_STATE_SYNC_RPC_LIST_FILE_DEVNET=${RPC_STATE_SYNC_RPC_LIST_FILE_DEVNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/main/devnet/rpc_state_sync_nodes}
  export APP_TOML_FILE_DEVNET=${APP_TOML_FILE_DEVNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/main/devnet/app.toml}
  export CONFIG_TOML_FILE_DEVNET=${CONFIG_TOML_FILE_DEVNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/main/devnet/config.toml}
  export CLIENT_TOML_FILE_DEVNET=${CLIENT_TOML_FILE_DEVNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/main/devnet/client.toml}
  export GENESIS_FILE_DEVNET=${GENESIS_FILE_DEVNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/main/devnet/genesis.json}

}

function init_chain {
  if [ -d "${DAEMON_HOME}/config" ]; then
      logt "${DAEMON_NAME} home directory already initialized."
  else
      logt "${DAEMON_NAME} home directory not initialized."
      logt "MONIKER: ${MONIKER}"
      logt "DAEMON_HOME: ${DAEMON_HOME}"
      ${DAEMON_NAME} init ${MONIKER} --home ${DAEMON_HOME} --chain-id ${CHAIN_ID}
  fi
}

function download_configs {
  wget -q ${APP_TOML_FILE_DEVNET} -O ${DAEMON_HOME}/config/app.toml
  wget -q ${CONFIG_TOML_FILE_DEVNET} -O ${DAEMON_HOME}/config/config.toml
  # wget -q ${CLIENT_TOML_FILE_DEVNET} -O ${DAEMON_HOME}/config/client.toml
  wget -q ${GENESIS_FILE_DEVNET} -O ${DAEMON_HOME}/config/genesis.json
  # wget -q ${BINARY_LIST_DEVNET}
  # export DOWNLOAD_BINARIES=$(cat binary_list.json | tr -d '\n')
  # rm -rf binary_list.json
  # logt "BINARY_LIST: ${DOWNLOAD_BINARIES}"
}

function change_config_values {
  export EXTERNAL_IP=$(curl -4 icanhazip.com)
  logt "******* DEBUG STATE SYNC VALUES *******"
  logt "EXTERNAL_IP: ${EXTERNAL_IP}"
  logt "SED Change Config Files."
  sed -i -e "s/^enable = .*/enable = \"true\"/" ${DAEMON_HOME}/config/config.toml
  sed '/^\[statesync\]/,/^\[/ s/enable = "true"/enable = "false"/' ${DAEMON_HOME}/config/config.toml
  sed -i -e "s/^moniker = .*/moniker = \"${MONIKER}\"/" ${DAEMON_HOME}/config/config.toml
  sed -i -e "s/^external_address = .*/external_address = \"${EXTERNAL_IP}:26656\"/" ${DAEMON_HOME}/config/config.toml
}

function setup_basic_keyring {
  if ${DAEMON_NAME} keys show "$MONIKER" --keyring-backend test > /dev/null 2>&1; then
    echo "Key $MONIKER already exists."
  else
    ${DAEMON_NAME} keys add "$MONIKER" --keyring-backend test
    echo "Key $MONIKER created."
  fi
}

function move_pellcored_binaries {
  mkdir -p ${DAEMON_HOME}/cosmovisor || logt "Directory already exists ${DAEMON_HOME}/cosmovisor"
  mkdir -p ${DAEMON_HOME}/cosmovisor/genesis || logt "Directory already exists ${DAEMON_HOME}/cosmovisor/genesis"
  mkdir -p ${DAEMON_HOME}/cosmovisor/genesis/bin || logt "Directory already exists ${DAEMON_HOME}/cosmovisor/genesis/bin"
  cp /usr/local/bin/pellcored ${DAEMON_HOME}/cosmovisor/genesis/bin/pellcored
}

function start_network {
  ${VISOR_NAME} version
  ${VISOR_NAME} run start --home ${DAEMON_HOME} \
    --log_level info \
    --moniker ${MONIKER} \
    --rpc.laddr tcp://0.0.0.0:26657 \
    --minimum-gas-prices 1.0apell "--grpc.enable=true"
}

logt "Load Default Values for ENV Vars if not set."
load_defaults

logt "Init Chain"
init_chain

logt "Download Configs"
download_configs

logt "Download Historical Binaries"
download_binary_version

# logt "Setup Restore Type: ${RESTORE_TYPE}"
# setup_restore_type

logt "Modify Chain Configs"
change_config_values

logt "Move root binaries to current"
move_pellcored_binaries

logt "Start sequence has completed, echo into file so on restart it doesn't download snapshots again."
echo "START_SEQUENCE_COMPLETE" >> ${DAEMON_HOME}/start_sequence_status

logt "Start Network"
start_network