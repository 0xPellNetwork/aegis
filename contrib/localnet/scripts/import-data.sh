#!/bin/bash
if [ $# -lt 1 ]
then
  echo "Usage: import-data.sh [network]"
  exit 1
fi

NETWORK=$1
NETWORK_SNAPSHOT_URL=$2
echo "NETWORK: ${NETWORK}"
echo "NETWORK_SNAPSHOT_URL: ${NETWORK_SNAPSHOT_URL}"

rm -rf /root/genesis_data/
mkdir /root/genesis_data/
echo "Download Latest State Export"
LATEST_EXPORT_URL=$(curl http://${NETWORK_SNAPSHOT_URL}/${NETWORK}/state/latest.json | jq -r '.snapshots[0].link')
echo "LATEST EXPORT URL: ${LATEST_EXPORT_URL}"
wget -q ${LATEST_EXPORT_URL} -O /root/genesis_data/exported-genesis.json