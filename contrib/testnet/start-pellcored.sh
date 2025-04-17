#!/bin/bash
set -e

logt() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

function load_defaults {
  export DAEMON_HOME=${DAEMON_HOME:=~/.pellcored}
  export COSMOVISOR_VERSION=${COSMOVISOR_VERSION:=v1.5.0}
  export CHAIN_ID=${CHAIN_ID:=ignite_186-1}
  export COSMOVISOR_CHECKSUM=${COSMOVISOR_CHECKSUM:=7f4bebfb18a170bff1c725f13dda326e0158132deef9f037ab0c2a48727c3077}
  export VISOR_NAME=${VISOR_NAME:=${DAEMON_HOME}/cosmovisor/cosmovisor}
  export DAEMON_NAME=${DAEMON_NAME:=pellcored}
  export DAEMON_ALLOW_DOWNLOAD_BINARIES=${DAEMON_ALLOW_DOWNLOAD_BINARIES:=true}
  export DAEMON_RESTART_AFTER_UPGRADE=${DAEMON_RESTART_AFTER_UPGRADE:=true}
  export MONIKER=${MONIKER:=local-test}
  export FAST_SYNC=${FAST_SYNC:=true}
  export WASMVM_VERSION=${WASMVM_VERSION:=v2.1.2}
  export LD_LIBRARY_PATH=${LD_LIBRARY_PATH:=~/.pellcored/lib}
  export DEBUG=${DEBUG:=false}
  export LOG_LEVEL=${LOG_LEVEL:=info}
  export SEEDS=${SEEDS:=""}
  export RPC_SERVERS=${RPC_SERVERS:=""}
  export IS_ARCHIVE=${IS_ARCHIVE:=false}


  # TESTNET
  export BINARY_LIST_TESTNET=${BINARY_LIST_TESTNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/refs/heads/main/testnet/binary_list.json}
  export RPC_NODE_RPC_FILE_TESTNET=${RPC_NODE_RPC_FILE_TESTNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/refs/heads/main/testnet/rpc_node}
  export APP_TOML_FILE_TESTNET=${APP_TOML_FILE_TESTNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/refs/heads/main/testnet/app.toml}
  export CONFIG_TOML_FILE_TESTNET=${CONFIG_TOML_FILE_TESTNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/refs/heads/main/testnet/config.toml}
  export CLIENT_TOML_FILE_TESTNET=${CLIENT_TOML_FILE_TESTNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/refs/heads/main/testnet/client.toml}
  export GENESIS_FILE_TESTNET=${GENESIS_FILE_TESTNET:=https://raw.githubusercontent.com/0xPellNetwork/network-config/refs/heads/main/testnet/genesis.json}
}

function install_cosmovisor {
    if ! command -v cosmovisor &> /dev/null
    then
        logt "cosmovisor not found, installing..."
        wget https://github.com/cosmos/cosmos-sdk/releases/download/cosmovisor%2F${COSMOVISOR_VERSION}/cosmovisor-${COSMOVISOR_VERSION}-linux-amd64.tar.gz
        echo "${COSMOVISOR_CHECKSUM}  cosmovisor-${COSMOVISOR_VERSION}-linux-amd64.tar.gz" | sha256sum -c -
        mkdir -p ${DAEMON_HOME}/cosmovisor
        tar -xzf cosmovisor-${COSMOVISOR_VERSION}-linux-amd64.tar.gz -C ${DAEMON_HOME}/cosmovisor
        rm cosmovisor-${COSMOVISOR_VERSION}-linux-amd64.tar.gz
    else
        logt "cosmovisor is already installed."
    fi
}

download_and_verify_wasmvm() {
    mkdir -p "$LD_LIBRARY_PATH"
    local checksums_tmp="./checksums.txt"
    wget "https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm.$(uname -m).so" \
        -O "$LD_LIBRARY_PATH/libwasmvm.$(uname -m).so"
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to download libwasmvm library."
        return 1
    fi

    wget "https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt" \
        -O "$checksums_tmp"
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to download checksums file."
        return 1
    fi

    local lib_checksum=$(sha256sum "$LD_LIBRARY_PATH/libwasmvm.$(uname -m).so" | cut -d ' ' -f 1)
    local expected_checksum=$(grep "libwasmvm.$(uname -m)" "$checksums_tmp" | cut -d ' ' -f 1)

    if [[ "$lib_checksum" != "$expected_checksum" ]]; then
        echo "Error: Checksum verification failed for libwasmvm.$(uname -m).so."
        return 1
    fi
    echo "libwasmvm.$(uname -m).so downloaded and verified successfully."
    return 0
}

download_and_move_binaries() {
  logt "Downloading binary list from ${BINARY_LIST_TESTNET}"
  wget -q -O binary_list.json "${BINARY_LIST_TESTNET}" || {
    logt "ERROR: Failed to download binary list from ${BINARY_LIST_TESTNET}. Please check the URL or your network connection."
    return 1
  }

  DOWNLOAD_BINARIES=$(cat binary_list.json | tr -d '\n')
  rm -rf binary_list.json
  logt "BINARY_LIST: ${DOWNLOAD_BINARIES}"

  if [ -z "${DOWNLOAD_BINARIES}" ]; then
    logt "DOWNLOAD_BINARIES is not set. Please set it before proceeding."
    return 1
  fi

  # --------------------------------------------------------------------------------
  # If FAST_SYNC is enabled, only download the *last* binary from the JSON array.
  # Otherwise, download all binaries.
  # --------------------------------------------------------------------------------
  if [ "${FAST_SYNC}" = "true" ]; then
    logt "FAST_SYNC is enabled. Will only download the last binary to save resources."

    # Extract the last binary object from the array
    local last_binary=$(echo "${DOWNLOAD_BINARIES}" | jq -c '.binaries | last')

    if [ -n "${last_binary}" ] && [ "${last_binary}" != "null" ]; then
      # Extract download URL and target location
      local download_url=$(echo "${last_binary}" | jq -r '.download_url')
      local binary_location=$(echo "${last_binary}" | jq -r '.binary_location')

      if [ -n "${download_url}" ] && [ -n "${binary_location}" ]; then
        # Download and move the single last binary
        process_single_binary "${download_url}" "${binary_location}"
        # After downloading, overwrite the genesis binary (if needed) since FAST_SYNC=true
        overwrite_genesis_binary "${binary_location}"
      else
        logt "Invalid last JSON entry: ${last_binary}"
      fi
    else
      logt "No last binary object found in JSON."
    fi

  else
    # FAST_SYNC != true, download all binaries as before
    logt "FAST_SYNC is disabled. Downloading all binaries."

    local last_downloaded_binary_location=""

    # Loop through each binary in the JSON array
    while read -r binary; do
      # Extract the download URL and target location
      local download_url=$(echo "${binary}" | jq -r '.download_url')
      local binary_location=$(echo "${binary}" | jq -r '.binary_location')

      if [ -z "${download_url}" ] || [ -z "${binary_location}" ]; then
        logt "Invalid JSON entry: ${binary}"
        continue
      fi

      # Process each binary
      if process_single_binary "${download_url}" "${binary_location}"; then
        last_downloaded_binary_location="${binary_location}"
      fi
    done < <(echo "${DOWNLOAD_BINARIES}" | jq -c '.binaries[]')

    # When FAST_SYNC=false, we don't need to overwrite, but if you still want the final 
    # binary to be the "active" one, you could do that below:
    logt "Last downloaded binary location: ${last_downloaded_binary_location}"
  fi

  # --------------------------------------------------------------------------------
  # Debug logic remains the same: if DEBUG=true, do the extra steps
  # --------------------------------------------------------------------------------
  if [ "${DEBUG}" = "true" ] && [ -n "${binary_location}" ]; then
    logt "DEBUG is enabled. Listing files in ${DAEMON_HOME}"
    ls -al "${DAEMON_HOME}"
    cp /usr/local/bin/pellcored "${DAEMON_HOME}/${binary_location}"
    logt "DEBUG is enabled. Overwrote ${DAEMON_HOME}/${binary_location} with /usr/local/bin/pellcored"
  fi
}

# --------------------------------------------------------------------------------
# Helper function to download and move a single binary
# --------------------------------------------------------------------------------
process_single_binary() {
  local download_url="$1"
  local binary_location="$2"

  # Extract the file name from the URL
  local binary_file=$(basename "${download_url}")

  logt "Downloading binary from ${download_url}"
  wget -q -O "${binary_file}" "${download_url}"
  if [ $? -ne 0 ]; then
    logt "Failed to download binary from ${download_url}"
    return 1
  fi

  # Set executable permissions
  chmod +x "${binary_file}"

  # Create the target directory
  local target_dir="${DAEMON_HOME}/${binary_location%/*}"
  mkdir -p "${target_dir}" || logt "Directory already exists or failed to create: ${target_dir}"

  # Move the file to the target location
  mv "${binary_file}" "${DAEMON_HOME}/${binary_location}"
  if [ $? -eq 0 ]; then
    logt "Successfully downloaded and moved the binary to ${DAEMON_HOME}/${binary_location}"
  else
    logt "Failed to move binary to ${DAEMON_HOME}/${binary_location}"
    return 1
  fi
}

# --------------------------------------------------------------------------------
# Helper function to overwrite the genesis binary if FAST_SYNC is enabled
# --------------------------------------------------------------------------------
overwrite_genesis_binary() {
  local binary_location="$1"
  local last_downloaded_binary_path="${DAEMON_HOME}/${binary_location}"
  local first_binary_path="${DAEMON_HOME}/cosmovisor/genesis/bin/pellcored"

  mkdir -p "$(dirname "${first_binary_path}")"
  if [ -f "${last_downloaded_binary_path}" ]; then
    cp -f "${last_downloaded_binary_path}" "${first_binary_path}"
    logt "FAST_SYNC is enabled. Overwrote ${first_binary_path} with ${last_downloaded_binary_path}"
  else
    logt "FAST_SYNC enabled but last downloaded binary not found: ${last_downloaded_binary_path}"
  fi
}

function init_chain {
  if [ -d "${DAEMON_HOME}/config" ]; then
      logt "${DAEMON_NAME} home directory already initialized."
  else
      logt "${DAEMON_NAME} home directory not initialized."
      logt "MONIKER: ${MONIKER}"
      logt "DAEMON_HOME: ${DAEMON_HOME}"
      ${DAEMON_HOME}/cosmovisor/genesis/bin/pellcored init ${MONIKER} --home ${DAEMON_HOME} --chain-id ${CHAIN_ID}
  fi
}

function download_configs {
  wget -q ${APP_TOML_FILE_TESTNET} -O ${DAEMON_HOME}/config/app.toml
  wget -q ${CONFIG_TOML_FILE_TESTNET} -O ${DAEMON_HOME}/config/config.toml
  wget -q ${GENESIS_FILE_TESTNET} -O ${DAEMON_HOME}/config/genesis.json
}

function change_config_values {
  export EXTERNAL_IP=$(curl -4 icanhazip.com)
  logt "******* DEBUG STATE SYNC VALUES *******"
  logt "EXTERNAL_IP: ${EXTERNAL_IP}"
  logt "FAST_SYNC: ${FAST_SYNC}"

  logt "SED Change Config Files."
  sed -i -e "s/^enable = .*/enable = \"${FAST_SYNC}\"/" ${DAEMON_HOME}/config/config.toml
  sed '/^\[statesync\]/,/^\[/ s/enable = "true"/enable = "false"/' ${DAEMON_HOME}/config/config.toml
  sed -i -e "s/^moniker = .*/moniker = \"${MONIKER}\"/" ${DAEMON_HOME}/config/config.toml
  sed -i -e "s/^external_address = .*/external_address = \"${EXTERNAL_IP}:26656\"/" ${DAEMON_HOME}/config/config.toml

  if [ "${FAST_SYNC}" == "true" ]; then
    logt "Configuring fast sync..."

    # Read the base RPC address from rpc_node
    RPC=$(curl -s "$RPC_NODE_RPC_FILE_TESTNET")

    # Get the latest block height and trust hash
    LATEST_HEIGHT=$(curl -s $RPC/block | jq -r .result.block.header.height)
    LATEST_HEIGHT=$((LATEST_HEIGHT - 2000))
    if [ "$LATEST_HEIGHT" -lt 0 ]; then
        LATEST_HEIGHT=0
    fi
    TRUST_HASH=$(curl -s "$RPC/block?height=$LATEST_HEIGHT" | jq -r .result.block_id.hash)

    # Update the configuration file
    sed -i.bak -E "s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$LATEST_HEIGHT|" ${DAEMON_HOME}/config/config.toml
    sed -i.bak -E "s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"|" ${DAEMON_HOME}/config/config.toml

    if [ ! -z "${RPC_SERVERS}" ]; then
        sed -i.bak -E "s|^(rpc_servers[[:space:]]+=[[:space:]]+)\".*\"|\1\"${RPC_SERVERS}\"|" "${DAEMON_HOME}/config/config.toml"
    fi

    if [ ! -z "${SEEDS}" ]; then
      sed -i.bak -E "s|^(seeds[[:space:]]+=[[:space:]]+)\".*\"|\1\"${SEEDS}\"|" "${DAEMON_HOME}/config/config.toml"
      sed -i.bak -E "s|^(persistent_peers[[:space:]]+=[[:space:]]+)\".*\"|\1\"${SEEDS}\"|" "${DAEMON_HOME}/config/config.toml"        
    fi

    logt "Fast sync configured."
  fi

  if [ "${IS_ARCHIVE}" == "true" ]; then
    sed -i.bak -E "s|^(pruning[[:space:]]+=[[:space:]]+).*$|\1\"nothing\"|" "${DAEMON_HOME}/config/app.toml" # pruning nothing
    sed -i.bak -E "s|^(min-retain-blocks[[:space:]]+=[[:space:]]+).*$|\0|" "${DAEMON_HOME}/config/app.toml" # keep all blocks
    sed -i.bak -E "s|^(snapshot-interval[[:space:]]+=[[:space:]]+).*$|\1 1000|" "${DAEMON_HOME}/config/app.toml" # 1000 blocks generate a snapshot
    sed -i.bak -E "s|^(snapshot-keep-recent[[:space:]]+=[[:space:]]+).*$|\0|" "${DAEMON_HOME}/config/app.toml" # keep all snapshots
  fi
}

function start_network {
  ${VISOR_NAME} version
  ${VISOR_NAME} run start --home ${DAEMON_HOME} \
    --log_level ${LOG_LEVEL} \
    --moniker ${MONIKER} \
    --rpc.laddr tcp://0.0.0.0:26657 \
    --minimum-gas-prices 1.0apell \
    --grpc.enable=true \
    --chain-id ${CHAIN_ID}
}

logt "Load Default Values for ENV Vars if not set."
load_defaults

logt "Install Cosmovisor"
install_cosmovisor

logt "Download and Verify WASMVM"
download_and_verify_wasmvm

logt "Download and Move Binaries"
download_and_move_binaries

logt "Init Chain"
init_chain

logt "Download Configs"
download_configs

logt "Modify Chain Configs"
change_config_values

logt "Start Network"
start_network
