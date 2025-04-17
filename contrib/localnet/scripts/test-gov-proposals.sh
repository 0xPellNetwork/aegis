#!/bin/bash

# This script is used to create a variety of proposals for testing purposes
# It creates proposals with different deposit amounts, voting periods, and content
# It also creates proposals with different voting options and votes on them
# It is intended to be run after the network has been started and the pellcored client is running
# It is intended to be run from the root of the pellcored repository
# It is intended to be run with the following command
# docker exec -it pellcore0 bash
# #/root/test-gov-proposals.sh

SCRIPT_DIR=$(dirname "$0")
cd "$SCRIPT_DIR" || exit

WALLET_NAME=operator

# Create a few short lived proposals for variety of testing
pellcored tx gov submit-proposal proposals/proposal_for_failure.json --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12
pellcored tx gov vote 1 VOTE_OPTION_NO --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12

pellcored tx gov submit-proposal proposals/proposal_for_success.json --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12
pellcored tx gov vote 2 VOTE_OPTION_YES --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12

pellcored tx gov submit-proposal proposals/v100.0.0_proposal.json --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12
pellcored tx gov vote 3 VOTE_OPTION_YES --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12

# Increase the length of the voting period to 1 week
pellcored tx gov submit-proposal proposals/proposal_voting_period.json --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12
pellcored tx gov vote 4 VOTE_OPTION_YES --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12

# Create a few long lived proposals for variety of testing

echo "Sleeping for 3 minutes to allow the voting period to end and the voting period will be increased to 1 week on future proposals"
sleep 180

pellcored tx gov submit-proposal proposals/proposal_voting_period.json --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12
pellcored tx gov vote 5 VOTE_OPTION_YES --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12

pellcored tx gov submit-proposal proposals/v100.0.0_proposal.json --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12
pellcored tx gov vote 6 VOTE_OPTION_YES --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12

pellcored tx gov submit-proposal proposals/proposal_for_deposit.json --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12
pellcored tx gov vote 7 VOTE_OPTION_YES --from $WALLET_NAME --keyring-backend test --chain-id ignite_186-1 --fees 2000000000000000apell --yes && sleep 12
