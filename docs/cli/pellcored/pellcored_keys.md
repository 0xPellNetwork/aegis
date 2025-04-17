# keys

Manage your application's keys

### Synopsis

Keyring management commands. These keys may be in any format supported by the
Tendermint crypto library and can be used by light-clients, full nodes, or any other application
that needs to sign with a private key.

The keyring supports the following backends:

    os          Uses the operating system's default credentials store.
    file        Uses encrypted file-based keystore within the app's configuration directory.
                This keyring will request a password each time it is accessed, which may occur
                multiple times in a single command resulting in repeated password prompts.
    kwallet     Uses KDE Wallet Manager as a credentials management application.
    pass        Uses the pass command line utility to store and retrieve keys.
    test        Stores keys insecurely to disk. It does not prompt for a password to be unlocked
                and it should be use only for testing purposes.

kwallet and pass backends depend on external tools. Refer to their respective documentation for more
information:
    KWallet     https://github.com/KDE/kwallet
    pass        https://www.passwordstore.org/

The pass backend requires GnuPG: https://gnupg.org/


### Options

```
  -h, --help                     help for keys
      --home string              The application home directory 
      --keyring-backend string   Select keyring's backend (os|file|test) 
      --keyring-dir string       The client Keyring directory; if omitted, the default 'home' directory will be used
      --output string            Output format (text|json) 
```

### Options inherited from parent commands

```
      --log_format string   The logging format (json|plain) 
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) 
      --trace               print out full stack trace on errors
```

### SEE ALSO

* [pellcored](pellcored.md)	 - Pellcore Daemon (server)
* [pellcored keys ](pellcored_keys_.md)	 - 
* [pellcored keys add](pellcored_keys_add.md)	 - Add an encrypted private key (either newly generated or recovered), encrypt it, and save to [name] file
* [pellcored keys delete](pellcored_keys_delete.md)	 - Delete the given keys
* [pellcored keys export](pellcored_keys_export.md)	 - Export private keys
* [pellcored keys import](pellcored_keys_import.md)	 - Import private keys into the local keybase
* [pellcored keys list](pellcored_keys_list.md)	 - List all keys
* [pellcored keys migrate](pellcored_keys_migrate.md)	 - Migrate keys from amino to proto serialization format
* [pellcored keys mnemonic](pellcored_keys_mnemonic.md)	 - Compute the bip39 mnemonic for some input entropy
* [pellcored keys parse](pellcored_keys_parse.md)	 - Parse address from hex to bech32 and vice versa
* [pellcored keys rename](pellcored_keys_rename.md)	 - Rename an existing key
* [pellcored keys show](pellcored_keys_show.md)	 - Retrieve key information by name or address
* [pellcored keys unsafe-export-eth-key](pellcored_keys_unsafe-export-eth-key.md)	 - **UNSAFE** Export an Ethereum private key
* [pellcored keys unsafe-import-eth-key](pellcored_keys_unsafe-import-eth-key.md)	 - **UNSAFE** Import Ethereum private keys into the local keybase

