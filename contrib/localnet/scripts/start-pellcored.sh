#!/bin/bash

# This script is used to start the pellcored nodes
# It initializes the nodes and creates the genesis.json file
# It also starts the nodes
# The number of nodes is passed as an first argument to the script
# The second argument is optional and can have the following value:
# 1. upgrade : This is used to test the upgrade process, a proposal is created for the upgrade and the nodes are started using cosmovisor

set -e

set -x

/usr/sbin/sshd

if [ $# -lt 1 ]
then
  echo "Usage: genesis.sh <num of nodes> [option]"
  exit 1
fi
NUMOFNODES=2
OPTION=$2

# create keys
CHAINID="ignite_186-1"
KEYRING="test"
HOSTNAME=$(hostname)
INDEX=${HOSTNAME:0-1}

# Environment variables used for upgrade testing
export DAEMON_HOME=$HOME/.pellcored
export DAEMON_NAME=pellcored
export DAEMON_ALLOW_DOWNLOAD_BINARIES=false
export DAEMON_RESTART_AFTER_UPGRADE=true
export CLIENT_DAEMON_NAME=pellclientd
export CLIENT_DAEMON_ARGS="-enable-chains,GOERLI,-val,operator"
export DAEMON_DATA_BACKUP_DIR=$DAEMON_HOME
export CLIENT_SKIP_UPGRADE=true
export CLIENT_START_PROCESS=false
export UNSAFE_SKIP_BACKUP=true

# import data option
export NETWORK_TYPE=${NETWORK_TYPE:-"testnet"}
export NETWORK_SNAPSHOT_URL=${NETWORK_SNAPSHOT_URL:-"host.docker.internal:8283"}

# upgrade name used for upgrade testing
export UpgradeName=${UPGRADE_VERSION}
echo "UpgradeName: $UpgradeName"

# generate node list
START=1
# shellcheck disable=SC2100
END=$((NUMOFNODES - 1))
NODELIST=()
for i in $(eval echo "{$START..$END}")
do
  NODELIST+=("pellcore$i")
done

echo "HOSTNAME: $HOSTNAME"

# Init a new node to generate genesis file .
# Copy config files from existing folders which get copied via Docker Copy when building images
mkdir -p ~/.backup/config
if [ -f "/root/.pellcored/config/genesis.json" ]; then
    pellcored start --pruning=nothing --minimum-gas-prices=0.0001apell --json-rpc.api eth,txpool,personal,net,debug,web3,miner --api.enable --home /root/.pellcored --chain-id=$CHAINID  --log_format=text
fi
pellcored init Pellnode-Localnet --chain-id=$CHAINID
rm -rf ~/.pellcored/config/app.toml
rm -rf ~/.pellcored/config/client.toml
rm -rf ~/.pellcored/config/config.toml
cp -r ~/pellcored/common/app.toml ~/.pellcored/config/
if [ $HOSTNAME == "pellcore0" ]
then
  cp -r ~/pellcored/pellcore0/app.toml ~/.pellcored/config/
fi
cp -r ~/pellcored/common/client.toml ~/.pellcored/config/
cp -r ~/pellcored/common/config.toml ~/.pellcored/config/
sed -i -e "/moniker =/s/=.*/= \"$HOSTNAME\"/" "$HOME"/.pellcored/config/config.toml

# This script modifies the genesis.json configuration file for pellcored by replacing 
# all occurrences of the string "stake" with "apell". The -i option is used to edit 
# the file in place.
# TODO: Update this script to handle other denominations if needed in the future.
sed -i 's/stake/apell/g' ~/.pellcored/config/genesis.json

# Add two new keys for operator and hotkey and create the required json structure for os_info
source ~/add-keys.sh

# Pause other nodes so that the primary can node can do the genesis creation
if [ $HOSTNAME != "pellcore0" ]
then
  while [ ! -f ~/.pellcored/config/init_complete ]; do
    echo "Waiting for init_complete file to exist..."
    sleep 1
  done
  # need to wait for pellcore0 to be up
  while ! curl -s -o /dev/null pellcore0:26657/status ; do
    echo "Waiting for pellcore0 rpc"
    sleep 1
done
fi

# Genesis creation following steps
# 1. Accumulate all the os_info files from other nodes on zetcacore0 and create a genesis.json
# 2. Add the observers , authorizations and required params to the genesis.json
# 3. Copy the genesis.json to all the nodes .And use it to create a gentx for every node
# 4. Collect all the gentx files in pellcore0 and create the final genesis.json
# 5. Copy the final genesis.json to all the nodes and start the nodes
# 6. Update Config in pellcore0 so that it has the correct persistent peer list
# 7. Start the nodes

# Start of genesis creation . This is done only on pellcore0
if [ $HOSTNAME == "pellcore0" ]
then
  # Misc : Copying the keyring to the client nodes so that they can sign the transactions
  ssh pellclient0 mkdir -p ~/.pellcored/keyring-test/
  scp ~/.pellcored/keyring-test/* pellclient0:~/.pellcored/keyring-test/
  ssh pellclient0 mkdir -p ~/.pellcored/keyring-file/
  scp ~/.pellcored/keyring-file/* pellclient0:~/.pellcored/keyring-file/

# 1. Accumulate all the os_info files from other nodes on zetcacore0 and create a genesis.json
  for NODE in "${NODELIST[@]}"; do
    INDEX=${NODE:0-1}
    ssh pellclient"$INDEX" mkdir -p ~/.pellcored/
    scp "$NODE":~/.pellcored/os_info/os.json ~/.pellcored/os_info/os_z"$INDEX".json
    scp ~/.pellcored/os_info/os_z"$INDEX".json pellclient"$INDEX":~/.pellcored/os.json
  done

  ssh pellclient0 mkdir -p ~/.pellcored/
  scp ~/.pellcored/os_info/os.json pellclient0:/root/.pellcored/os.json

# 2. Add the observers, authorizations, required params and accounts to the genesis.json
  pellcored collect-observer-info
  pellcored add-observer-list /os_info/observer_info.json 100000000000000000000000 1000000000000000000000 100000000000000000000000 --keygen-block 20
  cat $HOME/.pellcored/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="apell"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
  cat $HOME/.pellcored/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="apell"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
  cat $HOME/.pellcored/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="apell"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
  cat $HOME/.pellcored/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="apell"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
  cat $HOME/.pellcored/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="apell"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
  cat $HOME/.pellcored/config/genesis.json | jq '.consensus_params["block"]["max_gas"]="500000000"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
  cat $HOME/.pellcored/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="100s"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
  # TODO:
  cat $HOME/.pellcored/config/genesis.json | jq '.app_state["feemarket"]["params"]["min_gas_price"]="0"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json

 # set admin account
  pellcored add-genesis-account pell18z6s8y83ukj5hu48z2cpy3j2uanhrtd4ehw7xd 100000000000000000000000000apell
  pellcored add-genesis-account pell1n0rn6sne54hv7w2uu93fl48ncyqz97d3szgtee 100000000000000000000000000apell # Funds the localnet_gov_admin account
  cat $HOME/.pellcored/config/genesis.json | jq '.app_state["authority"]["policies"]["items"][0]["address"]="pell18z6s8y83ukj5hu48z2cpy3j2uanhrtd4ehw7xd"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
  cat $HOME/.pellcored/config/genesis.json | jq '.app_state["authority"]["policies"]["items"][1]["address"]="pell18z6s8y83ukj5hu48z2cpy3j2uanhrtd4ehw7xd"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
  cat $HOME/.pellcored/config/genesis.json | jq '.app_state["authority"]["policies"]["items"][2]["address"]="pell18z6s8y83ukj5hu48z2cpy3j2uanhrtd4ehw7xd"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
  cat $HOME/.pellcored/config/genesis.json | jq '.app_state["observer"]["params"]["admin_policy"][0]["address"]="pell18z6s8y83ukj5hu48z2cpy3j2uanhrtd4ehw7xd"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
  cat $HOME/.pellcored/config/genesis.json | jq '.app_state["observer"]["params"]["admin_policy"][1]["address"]="pell18z6s8y83ukj5hu48z2cpy3j2uanhrtd4ehw7xd"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json

# give balance to runner accounts to deploy contracts directly on pEVM
# deployer
  pellcored add-genesis-account pell1uhznv7uzyjq84s3q056suc8pkme85lkv32wqyr 100000000000000000000000000apell
# erc20 tester
  pellcored add-genesis-account pell1datate7xmwm4uk032f9rmcu0cwy7ch7kwngttz 100000000000000000000000000apell
# pell tester
  pellcored add-genesis-account pell1tnp0hvsq4y5mxuhrq9h3jfwulxywpq0aterjsd 100000000000000000000000000apell
# bitcoin tester
  pellcored add-genesis-account pell19q7czqysah6qg0n4y3l2a08gfzqxydlanvxawp 100000000000000000000000000apell
# ethers tester
  pellcored add-genesis-account pell134rakuus43xn63yucgxhn88ywj8ewcv6ltlmpn 100000000000000000000000000apell

# 3. Copy the genesis.json to all the nodes .And use it to create a gentx for every node
  pellcored gentx operator 1000000000000000000000apell --chain-id=$CHAINID --keyring-backend=$KEYRING --gas-prices 20000000000apell
  
  # Copy host gentx to other nodes
  for NODE in "${NODELIST[@]}"; do
    ssh $NODE mkdir -p ~/.pellcored/config/gentx/peer/
    scp ~/.pellcored/config/gentx/* $NODE:~/.pellcored/config/gentx/peer/
  done
  # Create gentx files on other nodes and copy them to host node
  mkdir ~/.pellcored/config/gentx/z2gentx
  for NODE in "${NODELIST[@]}"; do
      ssh $NODE rm -rf ~/.pellcored/genesis.json
      scp ~/.pellcored/config/genesis.json $NODE:~/.pellcored/config/genesis.json
      ssh $NODE pellcored gentx operator 1000000000000000000000apell --chain-id=$CHAINID --keyring-backend=$KEYRING
      scp $NODE:~/.pellcored/config/gentx/*.json ~/.pellcored/config/gentx/
      scp $NODE:~/.pellcored/config/gentx/*.json ~/.pellcored/config/gentx/z2gentx/
      ssh $NODE touch ~/.pellcored/config/init_complete
  done

# 4. Collect all the gentx files in pellcore0 and create the final genesis.json
  pellcored collect-gentxs

  if [ "$OPTION" == "import-data" ] || [ "$OPTION" == "upgrade" ]; then
    echo "Downloading latest state export"
    /root/import-data.sh $NETWORK_TYPE $NETWORK_SNAPSHOT_URL
    echo "Importing data"
    pellcored parse-genesis-file --modify /root/genesis_data/exported-genesis.json
  fi

  pellcored validate-genesis

  # Modify the unbonding time in genesis.json from "172800s" to "100s"
  # TODO: Make this configurable in the future
  sed -i 's/"172800s"/"100s"/g' ~/.pellcored/config/genesis.json

# 5. Copy the final genesis.json to all the nodes
  for NODE in "${NODELIST[@]}"; do
      echo "Copying genesis.json to $NODE"
      ssh $NODE rm -rf ~/.pellcored/genesis.json
      scp ~/.pellcored/config/genesis.json $NODE:~/.pellcored/config/genesis.json
  done
# 6. Update Config in pellcore0 so that it has the correct persistent peer list
   pps=$(cat $HOME/.pellcored/config/gentx/z2gentx/*.json | jq -r '.body.memo' )
   sed -i -e "/^persistent_peers =/s/=.*/= \"$pps\"/" "$HOME"/.pellcored/config/config.toml
fi
# End of genesis creation steps . The steps below are common to all the nodes

# Update persistent peers
if [ $HOSTNAME != "pellcore0" ]
then
  # Misc : Copying the keyring to the client nodes so that they can sign the transactions
  ssh pellclient"$INDEX" mkdir -p ~/.pellcored/keyring-test/
  scp ~/.pellcored/keyring-test/* "pellclient$INDEX":~/.pellcored/keyring-test/
  ssh pellclient"$INDEX" mkdir -p ~/.pellcored/keyring-file/
  scp ~/.pellcored/keyring-file/* "pellclient$INDEX":~/.pellcored/keyring-file/

  pps=$(cat $HOME/.pellcored/config/gentx/peer/*.json | jq -r '.body.memo' )
  sed -i -e "/^persistent_peers =/s/=.*/= \"$pps\"/" "$HOME"/.pellcored/config/config.toml
fi

# 7 Start the nodes
# If upgrade option is passed, use cosmovisor to start the nodes and create a governance proposal for upgrade
if [ "$OPTION" != "upgrade" ]; then
  pellcored start --pruning=nothing --minimum-gas-prices=0.0001apell --json-rpc.api eth,txpool,personal,net,debug,web3,miner --api.enable --home /root/.pellcored --chain-id $CHAINID --log_format=text
else
  # Setup cosmovisor
  mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
  mkdir -p $DAEMON_HOME/cosmovisor/upgrades/"$UpgradeName"/bin

  # Genesis
  cp /usr/local/bin/old_pellcored $DAEMON_HOME/cosmovisor/genesis/bin/pellcored
  cp /usr/local/bin/pellclientd $DAEMON_HOME/cosmovisor/genesis/bin/pellclientd

  #Upgrades
  cp /usr/local/bin/new_pellcored $DAEMON_HOME/cosmovisor/upgrades/$UpgradeName/bin/pellcored

  #Permissions
  chmod +x $DAEMON_HOME/cosmovisor/genesis/bin/pellcored
  chmod +x $DAEMON_HOME/cosmovisor/genesis/bin/pellclientd
  chmod +x $DAEMON_HOME/cosmovisor/upgrades/$UpgradeName/bin/pellcored

  # Start the node using cosmovisor
  cosmovisor run start --pruning=nothing --minimum-gas-prices=0.0001apell --json-rpc.api eth,txpool,personal,net,debug,web3,miner --api.enable --home /root/.pellcored --chain-id $CHAINID >> pellnode.log 2>&1  &
  sleep 20
  cat pellnode.log
  echo

  # Fetch the height of the upgrade, default is 225, if arg3 is passed, use that value
  UPGRADE_HEIGHT=${3:-225}

  # If this is the first node, create a governance proposal for upgrade
  if [ $HOSTNAME = "pellcore0" ]
  then
    /root/.pellcored/cosmovisor/genesis/bin/pellcored tx upgrade software-upgrade $UpgradeName --title $UpgradeName --summary $UpgradeName --upgrade-height $UPGRADE_HEIGHT --upgrade-info '{"binaries":{"os1/arch1":"url1","os2/arch2":"url2"}}' --no-checksum-required --no-validate --deposit="100000000apell" --from operator --fees=0.6pell --gas=60000000 -y
  fi

  # Wait for the proposal to be voted on
  sleep 8
  /root/.pellcored/cosmovisor/genesis/bin/pellcored tx gov vote 1 yes --from operator --fees=2000000000000000apell --chain-id $CHAINID --gas=60000000 --yes 
  sleep 7
  /root/.pellcored/cosmovisor/genesis/bin/pellcored query gov proposals
  /root/.pellcored/cosmovisor/genesis/bin/pellcored query gov proposal 1

  # We use tail -f to keep the container running
  tail -f pellnode.log

fi

