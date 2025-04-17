#!/bin/bash

# This script allows to add keys for operator and hotkey and create the required json structure for os_info

KEYRING_TEST="test"
KEYRING_FILE="file"
HOSTNAME=$(hostname)

set -x
set -e

# Operator key
if [ "$HOSTNAME" = "pellcore0" ]; then
    FIXED_MNEMONIC="pitch omit flag fuel soap artefact sleep hurdle segment hurry then wear plunge talk tragic huge spider open charge father filter behave ski coffee"
    # Note: there must be use eth_secp256k1 algo
    # https://docs.evmos.org/protocol/concepts/keyring
    echo "$FIXED_MNEMONIC" | pellcored keys add operator --recover --algo=eth_secp256k1 --keyring-backend=$KEYRING_TEST
else
    pellcored keys add operator --algo=secp256k1 --keyring-backend=$KEYRING_TEST
fi

operator_address=$(pellcored keys show operator -a --keyring-backend=$KEYRING_TEST)

# Hotkey key depending on the keyring-backend
if [ "$HOTKEY_BACKEND" == "$KEYRING_FILE" ]; then
    printf "%s\n%s\n" "$HOTKEY_PASSWORD" "$HOTKEY_PASSWORD" | pellcored keys add hotkey --algo=secp256k1 --keyring-backend=$KEYRING_FILE
    hotkey_address=$(printf "%s\n%s\n" "$HOTKEY_PASSWORD" "$HOTKEY_PASSWORD" | pellcored keys show hotkey -a --keyring-backend=$KEYRING_FILE)

    # Get hotkey pubkey, the command use keyring-backend in the cosmos config
    pellcored config set client keyring-backend "$KEYRING_FILE"
    pubkey=$(printf "%s\n%s\n" "$HOTKEY_PASSWORD" "$HOTKEY_PASSWORD" | pellcored get-pubkey hotkey | sed -e 's/secp256k1:"\(.*\)"/\1/' |sed 's/ //g' )
    pellcored config set client keyring-backend "$KEYRING_TEST"
else
    pellcored keys add hotkey --algo=secp256k1 --keyring-backend=$KEYRING
    hotkey_address=$(pellcored keys show hotkey -a --keyring-backend=$KEYRING)
    pubkey=$(pellcored get-pubkey hotkey|sed -e 's/secp256k1:"\(.*\)"/\1/' | sed 's/ //g' )
fi

is_observer="y"

echo "operator_address: $operator_address"
echo "hotkey_address: $hotkey_address"
echo "pubkey: $pubkey"
mkdir ~/.pellcored/os_info

# set key in file
jq -n --arg is_observer "$is_observer" --arg operator_address "$operator_address" --arg hotkey_address "$hotkey_address" --arg pubkey "$pubkey" '{"IsObserver":$is_observer,"ObserverAddress":$operator_address,"PellClientGranteeAddress":$hotkey_address,"PellClientGranteePubKey":$pubkey}' > ~/.pellcored/os_info/os.json
