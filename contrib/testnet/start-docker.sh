#!/bin/bash

# Function: Get user input with default value
get_input() {
    local prompt=$1
    local default_value=$2
    local value

    read -p "${prompt} (default: ${default_value}): " value
    echo ${value:-$default_value}
}

# Check command line arguments
if [ $# -eq 3 ]; then
    NETWORK_NAME=$1
    NODE_NAME=$2
    FAST_SYNC=$3
else
    # Interactive input
    echo "Please enter the following parameters:"
    NETWORK_NAME=$(get_input "Enter network name (mainnet/testnet)" "mainnet")
    NODE_NAME=$(get_input "Enter node name" "pellcored-local")
    FAST_SYNC=$(get_input "Enable fast sync (true/false)" "false")
fi

# Validate network name
case $NETWORK_NAME in
    "mainnet"|"testnet")
        echo "Using network: $NETWORK_NAME"
        ;;
    *)
        echo "Error: Network name must be 'mainnet' or 'testnet'"
        exit 1
        ;;
esac

# Validate fast sync parameter
case $FAST_SYNC in
    "true"|"false")
        echo "Fast sync: $FAST_SYNC"
        ;;
    *)
        echo "Error: Fast sync parameter must be 'true' or 'false'"
        exit 1
        ;;
esac

# Display configuration information
echo "Configuration:"
echo "Network: $NETWORK_NAME"
echo "Node name: $NODE_NAME"
echo "Fast sync: $FAST_SYNC"

# Confirm whether to continue
read -p "Continue? (y/n): " confirm
if [[ $confirm != "y" && $confirm != "Y" ]]; then
    echo "Operation cancelled"
    exit 0
fi

# Set environment variables
export NETWORK=$NETWORK_NAME
export NODE_NAME=$NODE_NAME
export FAST_SYNC=$FAST_SYNC

# First, clone the config repository
echo "Cloning network configuration..."
git clone -b add_testnet_config https://github.com/0xPellNetwork/network-config.git

# Change to the network-specific directory
cd network-config/${NETWORK_NAME}

# Configure fast sync if enabled
fastsync_config

# Copy necessary config files to the build context
cp app.toml ../../
cp client.toml ../../
cp config.toml ../../
cp genesis.json ../../

# Go back to the original directory where Dockerfile is located
cd ../..

# Build the Docker image
echo "Building Docker image..."
docker build -t pellcored-node .

# Start Docker container
echo "Starting Docker container..."
docker run -d \
    --name pell-node \
    -e NETWORK=${NETWORK_NAME} \
    -e NODE_NAME=${NODE_NAME} \
    -e FAST_SYNC=${FAST_SYNC} \
    -v "$(pwd)/network-config/${NETWORK_NAME}:/network-config" \
    -v "$(pwd)/network-config/${NETWORK_NAME}/data/${NETWORK_NAME}:/root/.pellcored/${NETWORK_NAME}" \
    -p 26656:26656 \
    -p 26657:26657 \
    -p 1317:1317 \
    --restart unless-stopped \
    pellcored-node

# Check if container started successfully
if [ $? -eq 0 ]; then
    echo "Container started successfully!"
    echo "Container ID: $(docker ps -q -f name=pell-node)"
    echo "View logs: docker logs pell-node"
else
    echo "Container failed to start!"
    exit 1
fi

fastsync_config() {
    if [ "${FAST_SYNC}" == "true" ]; then
        RPC=$(cat state_sync_node | sed 's/26656/26657/')
        
        LATEST_HEIGHT=$(curl -s $RPC/block | jq -r .result.block.header.height)
        TRUST_HASH=$(curl -s "$RPC/block?height=$LATEST_HEIGHT" | jq -r .result.block_id.hash)
        sed -i.bak -E "s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$LATEST_HEIGHT|" config.toml
        sed -i.bak -E "s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"|" config.toml

        echo "Fast sync: $FAST_SYNC"
    fi
}