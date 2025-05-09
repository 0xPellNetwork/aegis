# Pell Tool

Currently, has only one subcommand which finds inbound transactions or deposits that weren't observed on a particular
network. `filterdeposit`

## Configuring 

#### RPC endpoints
Configuring the tool for specific networks will require different reliable endpoints. For example, if you wanted to 
configure an ethereum rpc endpoint, then you will have to find an evm rpc endpoint for eth mainnet and set the field: 
`EthRPCURL`

#### Pell URL
You will need to find an enpoint for pellchain and set the field: `PellURL`

#### Contract Addresses
Depending on the network, connector and custody contract addresses must be set using these fields: `ConnectorAddress`,
`CustodyAddress`

If a configuration file is not provided, a default config will be generated under the name 
`pelltool_config.json`. Below is an example of a configuration file used for mainnet: 

#### Etherscan API Key
In order to make requests to etherscan, an api key will need to be configured.

```
{
 "PellURL": "",
 "BtcExplorerURL": "https://blockstream.info/api/",
 "EthRPCURL": "https://ethereum-rpc.publicnode.com",
 "EtherscanAPIkey": "",
 "ConnectorAddress": "0x000007Cf399229b2f5A4D043F20E90C9C98B7C6a",
 "CustodyAddress": "0x0000030Ec64DF25301d8414eE5a29588C4B0dE10"
}
```

## Running Tool

There are two targets available:

```
filter-missed-btc: install-pelltool
	./tool/filter_missed_deposits/filter_missed_btc.sh

filter-missed-eth: install-pelltool
	./tool/filter_missed_deposits/filter_missed_eth.sh
```

Running the commands can be simply done through the makefile in the node repo:

```
make filter-missed-btc
or ...
make filter-missed-eth
```