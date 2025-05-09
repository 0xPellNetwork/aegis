# query staking redelegations-from

Query all outgoing redelegatations from a validator

### Synopsis

Query delegations that are redelegating _from_ a validator.

Example:
$ pellcored query staking redelegations-from pellvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj

```
pellcored query staking redelegations-from [validator-addr] [flags]
```

### Options

```
      --count-total        count total number of records in validator redelegations to query for
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not TLS the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for redelegations-from
      --limit uint         pagination limit of validator redelegations to query for (default 100)
      --node string        [host]:[port] to Tendermint RPC interface for this chain 
      --offset uint        pagination offset of validator redelegations to query for
  -o, --output string      Output format (text|json) 
      --page uint          pagination page of validator redelegations to query for. This sets offset to a multiple of limit (default 1)
      --page-key string    pagination page-key of validator redelegations to query for
      --reverse            results are sorted in descending order
```

### Options inherited from parent commands

```
      --chain-id string     The network chain ID
      --home string         directory for config and data 
      --log_format string   The logging format (json|plain) 
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) 
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [pellcored query staking](pellcored_query_staking.md)	 - Querying commands for the staking module

