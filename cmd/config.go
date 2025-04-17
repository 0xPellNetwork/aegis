package cmd

const (
	Bech32PrefixAccAddr         = "pell"
	Bech32PrefixAccPub          = "pellpub"
	Bech32PrefixValAddr         = "pellv"
	Bech32PrefixValPub          = "pellvpub"
	Bech32PrefixConsAddr        = "pellc"
	Bech32PrefixConsPub         = "pellcpub"
	DenomRegex                  = `[a-zA-Z][a-zA-Z0-9:\\/\\\-\\_\\.]{2,127}`
	PellChainHDPath      string = `m/44'/60'/0'/0/0`
)
