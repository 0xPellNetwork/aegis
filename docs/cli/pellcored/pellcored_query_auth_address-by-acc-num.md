# query auth address-by-acc-num

Query for an address by account number

```
pellcored query auth address-by-acc-num [acc-num] [flags]
```

### Examples

```
pellcored q auth address-by-acc-num 1
```

### Options

```
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not TLS the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for address-by-acc-num
      --node string        [host]:[port] to Tendermint RPC interface for this chain 
  -o, --output string      Output format (text|json) 
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

* [pellcored query auth](pellcored_query_auth.md)	 - Querying commands for the auth module

