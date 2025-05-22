# Pell chain Changelog

## Format Specification

Please refer to [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed changelog format guidelines.

## [Unreleased v1.4.0]

### ✨ Features

### 📈 Improvements

- [deps] Bump deps to fix security alert ([#4](https://github.com/0xPellNetwork/aegis/pull/4))
- [deps] update Cosmos SDK to use 0xPellNetwork fork v0.50.10-alpha-1 ([#5](https://github.com/0xPellNetwork/aegis/pull/5))

### 🐛 Bug Fixes

### 📦 Dependencies

## [Released v1.3.0]

### ✨ Features

- [cmd/pellcored] Fix import genesis state from file cmd  ([#260](https://github.com/0xPellNetwork/chain/pull/260))

- [x/xsecurity] feat: add LST token dual staking feature ([#283](https://github.com/0xPellNetwork/chain/pull/283))

### 📈 Improvements

- [client/relayer] Fix inbund event gas limit param use default value ([#258](https://github.com/0xPellNetwork/chain/pull/258)).

- [client/relayer] Set tx.origin in inbound event handler ([#262](https://github.com/0xPellNetwork/chain/pull/262)).

- [client/relayer] Fix dvs registry chain inbound event map struct ordering issue ([#263](https://github.com/0xPellNetwork/chain/pull/263)).

- [x/restaking] feat: add DVS and group data to restaking module genesis exporter ([#284](https://github.com/0xPellNetwork/chain/pull/284)).

### 🐛 Bug Fixes

- [x/restaking] Rename syncModifyStrategyParams to syncModifyPoolParams ([#267](https://github.com/0xPellNetwork/chain/pull/267)).

### 📦 Dependencies

[contract] bump contract version from v0.2.31 to v0.2.34 ([#265](https://github.com/0xPellNetwork/chain/pull/265))

[e2e] Remove pelldvs dependences in e2e test ([#278](https://github.com/0xPellNetwork/chain/pull/278))

## [Released v1.2.0]

### 🚨 State Machine Breaking

### 🔄 API Breaking

### ✨ Features

- [x/xmsg] Implement PELL token bridging functionality from service chain to pell chain ([#237](https://github.com/0xPellNetwork/chain/pull/237)).

### 📈 Improvements

- [CI] Implement e2e testing for upgrade scenarios in CI pipeline ([#203](https://github.com/0xPellNetwork/chain/pull/203)).

- [CI] Add pull request number to changelog check ([#268](https://github.com/0xPellNetwork/chain/pull/268)).

- [refactor] Enhance sorting stability by implementing sort.SliceStable to maintain relative ordering of equal elements ([#232](https://github.com/0xPellNetwork/chain/pull/232)).

- [x/xmsg] Enhance visibility of cross-chain module by exposing chain support parameters and indexing status ([#233](https://github.com/0xPellNetwork/chain/pull/233)).

- [x/restaking] Implement genesis import/export functionality for operator shares ([#236](https://github.com/0xPellNetwork/chain/pull/236)).

- [x/restaking] Add CLI query command for retrieving DVS group information ([#238](https://github.com/0xPellNetwork/chain/pull/238)).

- [x/pevm] Bump contract to v0.2.30, implement reentrant sync group, and introduce upgrade system contract transaction type ([#241](https://github.com/0xPellNetwork/chain/pull/241)).

- [x/xmsg] Add crosschain fee management functionality and implement user fee handling for PEVM cross-chain transactions ([#261](https://github.com/0xPellNetwork/chain/pull/261)).

### 🐛 Bug Fixes

- [x/restaking] Resolve pagination issue in pool sync group functionality ([#226](https://github.com/0xPellNetwork/chain/pull/226)).

- [x/restaking] Address sync group inconsistencies related to operator registration data handling ([#239](https://github.com/0xPellNetwork/chain/pull/239)).

- [x/xmsg] Resolve xmsg migration issues in keygen process and optimize TSS procedure ([#242](https://github.com/0xPellNetwork/chain/pull/242)).

[x/restaking] Store operator registration data in V2 format and add migration in upgrade handler ([#250](https://github.com/0xPellNetwork/chain/pull/250))

[x/xmsg] Fix the issue where the outbound pending nonce tracker is not being deleted ([#248](https://github.com/0xPellNetwork/chain/pull/248))

### 📦 Dependencies

- [IAVL] Bump IAVL to version 1.2.4 to address pruning functionality issues ([#230](https://github.com/0xPellNetwork/chain/pull/230)) ([#ref suggestion](https://github.com/cosmos/cosmos-sdk/discussions/22253)) ([IAVL-1007](https://github.com/cosmos/iavl/pull/1007)).
