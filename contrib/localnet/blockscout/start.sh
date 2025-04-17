#!/bin/sh
docker compose -f geth.yml down --volumes
rm -rf services/blockscout-db-data
rm -rf services/logs
rm -rf services/stats-db-data
rm -rf services/stats-db-data

docker compose -f geth.yml up -d
