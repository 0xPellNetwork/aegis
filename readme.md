# Pell Chain

PellChain is an L1 blockchain compatible with both CosmWasm and EVM virtual machines, enabling omnichain, generic
smart contracts and messaging between any blockchain.

## Prerequisites

- [Go](https://go.dev/dl/#go1.23.2) 1.23.2
- [Docker](https://docs.docker.com/install/) and
  [Docker Compose](https://docs.docker.com/compose/install/) (optional, for
  running tests locally)
- [buf](https://buf.build/) (optional, for processing protocol buffer files)
- [jq](https://stedolan.github.io/jq/download/) (optional, for running scripts)

## Components of PellChain

PellChain is built with [Cosmos SDK](https://github.com/cosmos/cosmos-sdk), a
modular framework for building blockchain and
[Ethermint](https://github.com/evmos/ethermint), a module that implements
EVM-compatibility.

- [pell-node](https://github.com/0xPellNetwork/chain) (this repository)
  contains the source code for the PellChain node (`pellcored`) and the
  PellChain relayer (`pellclientd`).
- [protocol-contracts](https://github.com/0xPellNetwork/contracts)
  contains the source code for the Solidity smart contracts that implement the
  core functionality of PellChain.

## Building the pellcored/pellclientd binaries

For the Ignite testnet, clone this repository, checkout the latest release tag, and type the following command to build the binaries:

```
make install
```

to build.

This command will install the `pellcored` and `pellclientd` binaries in your
`$GOPATH/bin` directory.

Verify that the version of the binaries match the release tag.  

```
pellcored version
pellclientd version
```

## Making changes to the source code

After making changes to any of the protocol buffer files, run the following
command to generate the Go files:

```
make proto
```

This command will use `buf` to generate the Go files from the protocol buffer
files and move them to the correct directories inside `x/`. It will also
generate an OpenAPI spec.

### Generate documentation

To generate the documentation, run the following command:

```
make specs
```

This command will run a script to update the modules' documentation. The script
uses static code analysis to read the protocol buffer files and identify all
Cosmos SDK messages. It then searches the source code for the corresponding
message handler functions and retrieves the documentation for those functions.
Finally, it creates a `messages.md` file for each module, which contains the
documentation for all the messages in that module.

## Running tests

To check that the source code is working as expected, refer to the manual on how
to [run the E2E test](./LOCAL_TESTING.md).

## Community

[Twitter](https://twitter.com/pellblockchain) |
[Discord](https://discord.com/invite/pellchain) |
[Telegram](https://t.me/pellchainofficial) | [Website](https://pellchain.com)
