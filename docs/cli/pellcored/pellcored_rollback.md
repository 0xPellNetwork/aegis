# rollback

rollback cosmos-sdk and tendermint state by one height

### Synopsis


A state rollback is performed to recover from an incorrect application state transition,
when Tendermint has persisted an incorrect app hash and is thus unable to make
progress. Rollback overwrites a state at height n with the state at height n - 1.
The application also rolls back to height n - 1. No blocks are removed, so upon
restarting Tendermint the transactions in block n will be re-executed against the
application.


```
pellcored rollback [flags]
```

### Options

```
  -h, --help          help for rollback
      --home string   The application home directory 
```

### Options inherited from parent commands

```
      --log_format string   The logging format (json|plain) 
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) 
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [pellcored](pellcored.md)	 - Pellcore Daemon (server)

