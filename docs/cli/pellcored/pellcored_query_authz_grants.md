# query authz grants

query grants for a granter-grantee pair and optionally a msg-type-url

### Synopsis

Query authorization grants for a granter-grantee pair. If msg-type-url
is set, it will select grants only for that msg type.
Examples:
$ pellcored query authz grants cosmos1skj.. cosmos1skjwj..
$ pellcored query authz grants cosmos1skjw.. cosmos1skjwj.. /cosmos.bank.v1beta1.MsgSend

```
pellcored query authz grants [granter-addr] [grantee-addr] [msg-type-url]? [flags]
```

### Options

```
      --count-total        count total number of records in grants to query for
      --grpc-addr string   the gRPC endpoint to use for this chain
      --grpc-insecure      allow gRPC over insecure channels, if not TLS the server must use TLS
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for grants
      --limit uint         pagination limit of grants to query for (default 100)
      --node string        [host]:[port] to Tendermint RPC interface for this chain 
      --offset uint        pagination offset of grants to query for
  -o, --output string      Output format (text|json) 
      --page uint          pagination page of grants to query for. This sets offset to a multiple of limit (default 1)
      --page-key string    pagination page-key of grants to query for
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

* [pellcored query authz](pellcored_query_authz.md)	 - Querying commands for the authz module

