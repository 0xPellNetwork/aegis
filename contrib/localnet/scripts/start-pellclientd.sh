#!/bin/bash

# This script is used to start PellClient for the localnet
# An optional argument can be passed and can have the following value:
# background: start the PellClient in the background, this prevent the image from being stopped when PellClient must be restarted

/usr/sbin/sshd

HOSTNAME=$(hostname)
OPTION=$1

# clear null endpoint evm chain config
clearConfig() {
  jq '.EVMChainConfigs |= with_entries(select(.value.Endpoint != ""))' /root/.pellcored/config/pellclient_config.json > tmp.json && mv tmp.json /root/.pellcored/config/pellclient_config.json
}

add_chain() {
  chain_id=$1
  network=$2
  network_type=$3
  endpoint=$4
  start_height=$5

  if [ -z "$endpoint" ]; then
        echo "empty endpoint"
        return 0
  fi

  jq --arg chain_id "$chain_id" \
     --arg network "$network" \
     --arg network_type "$network_type" \
     --arg endpoint "$endpoint" \
     --arg start_height "$start_height" \
   '.EVMChainConfigs += {($chain_id): {"Chain": {"network": $network | tonumber, "network_type": $network_type | tonumber,"vm_type": 1, "id": $chain_id | tonumber}, "Endpoint": $endpoint, "ForceStartHeight": $start_height | tonumber}}' \
   /root/.pellcored/config/pellclient_config.json > tmp.json && mv tmp.json /root/.pellcored/config/pellclient_config.json
}

updateChainCfg() {
  clearConfig
  add_chain "1337" "0" 1 "http://eth:8545" 0 # bsc testnet

  # add_chain "97" "4" 1 $BSC_TESTNET_EXTERNAL_RPC_URL   # bsc testnet
  # add_chain "5003" "14" 1 $MANTLE_TESTNET_EXTERNAL_RPC_URL   # mantle testnet
  # add_chain "1115" "13" 1 $CORE_TESTNET_EXTERNAL_RPC_URL   # core testnet
}

# read HOTKEY_BACKEND env var for hotkey keyring backend and set default to test
BACKEND="test"
if [ "$HOTKEY_BACKEND" == "file" ]; then
    BACKEND="file"
fi

cp  /root/preparams/PreParams_$HOSTNAME.json /root/preParams.json
num=$(echo $HOSTNAME | tr -dc '0-9')
node="pellcore$num"

if [ -f "/root/.pellcored/config/pellclient_config.json" ]; then
    # setting support chain endpoint
    # add_chain "97" "10" \"$BSC_TESTNET_EXTERNAL_RPC_URL\"        #bsc testnet
    # add_chain "1115" "21" \"$CORE_TESTNET_EXTERNAL_RPC_URL\"     #core testnet
    updateChainCfg

    # start pellclient
    pellclientd start < /root/password.file
fi
echo "Wait for pellcore to exchange genesis file"
#  pause nodes other than pellcore0 to wait for pellcore0 to create genesis.json
#  additional pause time is needed for importing data into the genesis as the export file is read into memory
if [ "$OPTION" != "import-data" ]; then
  sleep 10
else
  sleep 510
fi

while [ ! -f $HOME/.pellcored/os.json ]; do
    echo "Waiting for pellcore to exchange os.json file..."
    sleep 1
done

operator=$(cat $HOME/.pellcored/os.json | jq '.ObserverAddress' )
operatorAddress=$(echo "$operator" | tr -d '"')
echo "operatorAddress: $operatorAddress"
echo "Start pellclientd"
if [ $HOSTNAME == "pellclient0" ]
then
    rm ~/.tss/*
    MYIP=$(/sbin/ip -o -4 addr list eth0 | awk '{print $4}' | cut -d/ -f1)
    pellclientd init --pellcore-url pellcore0 --chain-id ignite_186-1 --operator "$operatorAddress" --log-format=text --public-ip "$MYIP" --keyring-backend "$BACKEND"

    # check if the option is additional-evm
   # in this case, the additional evm is represented with the sepolia chain, we set manually the eth2 endpoint to the sepolia chain (11155111 -> http://eth2:8545)
    # in /root/.pellcored/config/pellclient_config.json
    if [ "$OPTION" == "additional-evm" ]; then
     set_sepolia_endpoint
    fi

    updateChainCfg

    pellclientd start < /root/password.file
else
  num=$(echo $HOSTNAME | tr -dc '0-9')
  node="pellcore$num"
  MYIP=$(/sbin/ip -o -4 addr list eth0 | awk '{print $4}' | cut -d/ -f1)
  SEED=""
  while [ -z "$SEED" ]
  do
    SEED=$(curl --retry 10 --retry-delay 5 --retry-connrefused  -s pellclient0:8123/p2p)
  done
  rm ~/.tss/*
  CLIENT0IP=$(ssh pellclient0 /sbin/ip -o -4 addr list eth0 | awk '{print $4}' | cut -d/ -f1)
  pellclientd init --peer /ip4/$CLIENT0IP/tcp/6668/p2p/"$SEED" --pellcore-url "$node" --chain-id ignite_186-1 --operator "$operatorAddress" --log-format=text --public-ip "$MYIP" --log-level 1 --keyring-backend "$BACKEND"

  # check if the option is additional-evm
  # in this case, the additional evm is represented with the sepolia chain, we set manually the eth2 endpoint to the sepolia chain (11155111 -> http://eth2:8545)
  # in /root/.pellcored/config/pellclient_config.json
  if [ "$OPTION" == "additional-evm" ]; then
   set_sepolia_endpoint
  fi

  updateChainCfg
  pellclientd start < /root/password.file
fi

# check if the option is background
# in this case, we tail the pellclientd log file
if [ "$OPTION" == "background" ]; then
    sleep 3
    tail -f $HOME/pellclient.log
fi
