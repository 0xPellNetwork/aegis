#!/bin/bash

# This script immediately restarts the pellclientd on pellclient0 and pellclient1 containers in the network

echo restarting pellclients

ssh -o "StrictHostKeyChecking no" "pellclient0" -i ~/.ssh/localtest.pem killall pellclientd
ssh -o "StrictHostKeyChecking no" "pellclient1" -i ~/.ssh/localtest.pem killall pellclientd
ssh -o "StrictHostKeyChecking no" "pellclient0" -i ~/.ssh/localtest.pem "/usr/local/bin/pellclientd start < /root/password.file > $HOME/pellclient.log 2>&1 &"
ssh -o "StrictHostKeyChecking no" "pellclient1" -i ~/.ssh/localtest.pem "/usr/local/bin/pellclientd start < /root/password.file > $HOME/pellclient.log 2>&1 &"

