#!/usr/bin/env bash

CHAINID="localnet_101-1"
KEYRING="test"
export DAEMON_HOME=$HOME/.pellcored
export DAEMON_NAME=pellcored

### chain init script for development purposes only ###
rm -rf ~/.pellcored
kill -9 $(lsof -ti:26657)
pellcored config keyring-backend $KEYRING --home ~/.pellcored
pellcored config chain-id $CHAINID --home ~/.pellcored
echo "race draft rival universe maid cheese steel logic crowd fork comic easy truth drift tomorrow eye buddy head time cash swing swift midnight borrow" | pellcored keys add pell --algo=secp256k1 --recover --keyring-backend=test
echo "hand inmate canvas head lunar naive increase recycle dog ecology inhale december wide bubble hockey dice worth gravity ketchup feed balance parent secret orchard" | pellcored keys add mario --algo secp256k1 --recover --keyring-backend=test
echo "lounge supply patch festival retire duck foster decline theme horror decline poverty behind clever harsh layer primary syrup depart fantasy session fossil dismiss east" | pellcored keys add pelleth --recover --keyring-backend=test

pellcored init test --chain-id=$CHAINID

#Set config to use apell
cat $HOME/.pellcored/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="apell"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
cat $HOME/.pellcored/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="apell"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
cat $HOME/.pellcored/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="apell"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
cat $HOME/.pellcored/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="apell"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
cat $HOME/.pellcored/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="apell"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json
cat $HOME/.pellcored/config/genesis.json | jq '.consensus_params["block"]["max_gas"]="10000000"' > $HOME/.pellcored/config/tmp_genesis.json && mv $HOME/.pellcored/config/tmp_genesis.json $HOME/.pellcored/config/genesis.json






pellcored add-genesis-account $(pellcored keys show pell -a --keyring-backend=test) 500000000000000000000000000000000000000apell --keyring-backend=test
pellcored add-genesis-account $(pellcored keys show mario -a --keyring-backend=test) 50000000000000000000000000000000000000apell --keyring-backend=test
pellcored add-genesis-account $(pellcored keys show pelleth -a --keyring-backend=test) 500000000000000000000000000000000apell --keyring-backend=test


ADDR1=$(pellcored keys show pell -a --keyring-backend=test)
observer+=$ADDR1
observer+=","
ADDR2=$(pellcored keys show mario -a --keyring-backend=test)
observer+=$ADDR2
observer+=","


observer_list=$(echo $observer | rev | cut -c2- | rev)

echo $observer_list



pellcored add-observer 1337 "$observer_list"
pellcored add-observer 101 "$observer_list"




pellcored gentx pell 50000000000000000000000000apell --chain-id=localnet_101-1 --keyring-backend=test

contents="$(jq '.app_state.gov.voting_params.voting_period = "10s"' $DAEMON_HOME/config/genesis.json)" && \
echo "${contents}" > $DAEMON_HOME/config/genesis.json

echo "Collecting genesis txs..."
pellcored collect-gentxs

echo "Validating genesis file..."
pellcored validate-genesis
#
#export DUMMY_PRICE=yes
#export DISABLE_TSS_KEYGEN=yes
#export GOERLI_ENDPOINT=https://goerli.infura.io/v3/faf5188f178a4a86b3a63ce9f624eb1b
