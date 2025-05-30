# keys show

Retrieve key information by name or address

### Synopsis

Display keys details. If multiple names or addresses are provided,
then an ephemeral multisig key will be created under the name "multi"
consisting of all the keys provided by name and multisig threshold.

```
pellcored keys show [name_or_address [name_or_address...]] [flags]
```

### Options

```
  -a, --address                  Output the address only (overrides --output)
      --bech string              The Bech32 prefix encoding for a key (acc|val|cons) 
  -d, --device                   Output the address in a ledger device
  -h, --help                     help for show
      --multisig-threshold int   K out of N required signatures (default 1)
  -p, --pubkey                   Output the public key only (overrides --output)
```

### Options inherited from parent commands

```
      --home string              The application home directory 
      --keyring-backend string   Select keyring's backend (os|file|test) 
      --keyring-dir string       The client Keyring directory; if omitted, the default 'home' directory will be used
      --log_format string        The logging format (json|plain) 
      --log_level string         The logging level (trace|debug|info|warn|error|fatal|panic) 
      --output string            Output format (text|json) 
      --trace                    print out full stack trace on errors
```

### SEE ALSO

* [pellcored keys](pellcored_keys.md)	 - Manage your application's keys

