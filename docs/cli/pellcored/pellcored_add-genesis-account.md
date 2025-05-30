# add-genesis-account

Add a genesis account to genesis.json

### Synopsis

Add a genesis account to genesis.json. The provided account must specify
the account address or key name and a list of initial coins. If a key name is given,
the address will be looked up in the local Keybase. The list of initial tokens must
contain valid denominations. Accounts may optionally be supplied with vesting parameters.


```
pellcored add-genesis-account [address_or_key_name] [coin][,[coin]] [flags]
```

### Options

```
      --grpc-addr string         the gRPC endpoint to use for this chain
      --grpc-insecure            allow gRPC over insecure channels, if not TLS the server must use TLS
      --height int               Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help                     help for add-genesis-account
      --home string              The application home directory 
      --keyring-backend string   Select keyring's backend (os|file|kwallet|pass|test) 
      --node string              [host]:[port] to Tendermint RPC interface for this chain 
  -o, --output string            Output format (text|json) 
      --vesting-amount string    amount of coins for vesting accounts
      --vesting-end-time int     schedule end time (unix epoch) for vesting accounts
      --vesting-start-time int   schedule start time (unix epoch) for vesting accounts
```

### Options inherited from parent commands

```
      --log_format string   The logging format (json|plain) 
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) 
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [pellcored](pellcored.md)	 - Pellcore Daemon (server)

