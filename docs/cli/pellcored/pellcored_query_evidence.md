# query evidence

Query for evidence by hash or for all (paginated) submitted evidence

### Synopsis

Query for specific submitted evidence by hash or query for all (paginated) evidence:

Example:
$ pellcored query evidence DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660
$ pellcored query evidence --page=2 --limit=50

```
pellcored query evidence [flags]
```

### Options

```
      --count-total        count total number of records in evidence to query for
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not TLS the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for evidence
      --limit uint         pagination limit of evidence to query for (default 100)
      --node string        [host]:[port] to Tendermint RPC interface for this chain 
      --offset uint        pagination offset of evidence to query for
  -o, --output string      Output format (text|json) 
      --page uint          pagination page of evidence to query for. This sets offset to a multiple of limit (default 1)
      --page-key string    pagination page-key of evidence to query for
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

* [pellcored query](pellcored_query.md)	 - Querying subcommands

