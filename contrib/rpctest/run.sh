#!/usr/bin/env bash

killall pellcored
pellcored start --trace \
--minimum-gas-prices=0.0001apell \
--json-rpc.api eth,txpool,personal,net,debug,web3,miner \
--api.enable \

#--home ~/.pellcored \
#--p2p.laddr 0.0.0.0:27655  \
#--grpc.address 0.0.0.0:9096 \
#--grpc-web.address 0.0.0.0:9093 \
#--address tcp://0.0.0.0:27659 \
#--rpc.laddr tcp://127.0.0.1:26657 \
#--pruning custom \
#--pruning-keep-recent 1 \
#--pruning-keep-every 10  \
#--pruning-interval 10 \
#--state-sync.snapshot-interval 10 \
#--state-sync.snapshot-keep-recent 1
