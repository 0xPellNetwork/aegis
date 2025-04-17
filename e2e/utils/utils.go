package utils

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ScriptPKToAddress is a hex string for P2WPKH script
func ScriptPKToAddress(scriptPKHex string, params *chaincfg.Params) string {
	pkh, err := hex.DecodeString(scriptPKHex[4:])
	if err == nil {
		addr, err := btcutil.NewAddressWitnessPubKeyHash(pkh, params)
		if err == nil {
			return addr.EncodeAddress()
		}
	}
	return ""
}

type infoLogger interface {
	Info(message string, args ...interface{})
}

type NoopLogger struct{}

func (nl NoopLogger) Info(_ string, _ ...interface{}) {}

func Assert(conidtion bool, msg ...any) {
	if !conidtion {
		panic(fmt.Sprint(msg...))
	}
}

// gen new account. private_key, eth_address, pell_address
func GenKeypair() (string, ethcommon.Address, sdk.AccAddress) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()

	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	addrBytes, _ := hex.DecodeString(address.Hex()[2:])
	pellAddress := sdk.AccAddress(addrBytes)

	privkeyBytes := crypto.FromECDSA(privateKey)

	return hex.EncodeToString(privkeyBytes), address, pellAddress
}

// IsEthAddress checks if a string is a valid Ethereum address, with or without the '0x' prefix.
func IsEthAddress(address string) bool {
	re := regexp.MustCompile(`^(0x)?[0-9a-fA-F]{40}$`)
	return re.MatchString(address)
}
