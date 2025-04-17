#!/bin/bash

/usr/sbin/sshd

#retrieve value of ETH TSS Address from localnet
ethTSS_address=$(curl 'http://localhost:1317/pell-chain/relayer/get_tss_address' | jq -r '.eth')

#write value of ETH TSS Address to addresses.txt file
printf "ethTSS:$ethTSS_address\n" > addresses.txt