package keys

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	ckeys "github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"

	"github.com/0xPellNetwork/aegis/cmd"
	"github.com/0xPellNetwork/aegis/pkg/cosmos"
	pellcrypto "github.com/0xPellNetwork/aegis/pkg/crypto"
	"github.com/0xPellNetwork/aegis/relayer/config"
)

var (
	// ErrBech32ifyPubKey is an error when Bech32ifyPubKey fails
	ErrBech32ifyPubKey = errors.New("Bech32ifyPubKey fail in main")

	// ErrNewPubKey is an error when NewPubKey fails
	ErrNewPubKey = errors.New("NewPubKey error from string")
)

// Keys manages all the keys used by pell client
type Keys struct {
	signerName      string
	kb              ckeys.Keyring
	OperatorAddress sdk.AccAddress
	hotkeyPassword  string
}

// NewKeysWithKeybase create a new instance of Keys
func NewKeysWithKeybase(kb ckeys.Keyring, granterAddress sdk.AccAddress, granteeName string, hotkeyPassword string) *Keys {
	return &Keys{
		signerName:      granteeName,
		kb:              kb,
		OperatorAddress: granterAddress,
		hotkeyPassword:  hotkeyPassword,
	}
}

func GetGranteeKeyName(signerName string) string {
	return signerName
}

// GetKeyringKeybase return keyring and key info
func GetKeyringKeybase(cfg config.Config, hotkeyPassword string) (ckeys.Keyring, string, error) {
	granteeName := cfg.AuthzHotkey
	chainHomeFolder := cfg.PellCoreHome
	logger := log.Logger.With().Str("module", "GetKeyringKeybase").Logger()
	if len(granteeName) == 0 {
		return nil, "", fmt.Errorf("signer name is empty")
	}

	// read password from env if using keyring backend file
	buf := bytes.NewBufferString("")
	if cfg.KeyringBackend == config.KeyringBackendFile {
		buf.WriteString(hotkeyPassword)
		buf.WriteByte('\n') // the library used by keyring is using ReadLine , which expect a new line
		buf.WriteString(hotkeyPassword)
		buf.WriteByte('\n')
	}

	kb, err := getKeybase(chainHomeFolder, buf, cfg.KeyringBackend)
	if err != nil {
		return nil, "", fmt.Errorf("fail to get keybase,err:%w", err)
	}

	oldStdIn := os.Stdin
	defer func() {
		os.Stdin = oldStdIn
	}()
	os.Stdin = nil

	logger.Debug().Msgf("Checking for Hotkey Key: %s \nFolder %s\nBackend %s", granteeName, chainHomeFolder, kb.Backend())
	rc, err := kb.Key(granteeName)
	if err != nil {
		return nil, "", fmt.Errorf("key not in backend %s present with name (%s): %w", kb.Backend(), granteeName, err)
	}

	pubkeyBech32, err := pellcrypto.GetPubkeyBech32FromRecord(rc)
	if err != nil {
		return nil, "", fmt.Errorf("fail to get pubkey from record,err:%w", err)
	}

	return kb, pubkeyBech32, nil
}

// GetSignerInfo return signer info
func (k *Keys) GetSignerInfo() *ckeys.Record {
	signer := GetGranteeKeyName(k.signerName)
	info, err := k.kb.Key(signer)
	if err != nil {
		return nil
	}
	return info
}

func (k *Keys) GetOperatorAddress() sdk.AccAddress {
	return k.OperatorAddress
}

func (k *Keys) GetAddress() (sdk.AccAddress, error) {
	signer := GetGranteeKeyName(k.signerName)
	info, err := k.kb.Key(signer)
	if err != nil {
		return nil, err
	}
	addr, err := info.GetAddress()
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// GetPrivateKey return the private key
func (k *Keys) GetPrivateKey(password string) (cryptotypes.PrivKey, error) {
	signer := GetGranteeKeyName(k.signerName)
	privKeyArmor, err := k.kb.ExportPrivKeyArmor(signer, password)
	if err != nil {
		return nil, err
	}
	priKey, _, err := crypto.UnarmorDecryptPrivKey(privKeyArmor, password)
	if err != nil {
		return nil, fmt.Errorf("fail to unarmor private key: %w", err)
	}
	return priKey, nil
}

// GetKeybase return the keybase
func (k *Keys) GetKeybase() ckeys.Keyring {
	return k.kb
}

func (k *Keys) GetPubKeySet(password string) (pellcrypto.PubKeySet, error) {
	pubkeySet := pellcrypto.PubKeySet{
		Secp256k1: "",
		Ed25519:   "",
	}

	pK, err := k.GetPrivateKey(password)
	if err != nil {
		return pubkeySet, err
	}

	s, err := cosmos.Bech32ifyPubKey(cosmos.Bech32PubKeyTypeAccPub, pK.PubKey())
	if err != nil {
		return pubkeySet, ErrBech32ifyPubKey
	}
	pubkey, err := pellcrypto.NewPubKey(s)
	if err != nil {
		return pubkeySet, ErrNewPubKey
	}
	pubkeySet.Secp256k1 = pubkey
	return pubkeySet, nil
}

// GetHotkeyPassword returns the password to be used
// returns empty if no password is needed
func (k *Keys) GetHotkeyPassword() string {
	if k.GetKeybase().Backend() == ckeys.BackendFile {
		return k.hotkeyPassword
	}
	return ""
}

// getKeybase will create an instance of Keybase
func getKeybase(pellCoreHome string, reader io.Reader, keyringBackend config.KeyringBackend) (ckeys.Keyring, error) {
	cliDir := pellCoreHome
	if len(pellCoreHome) == 0 {
		return nil, fmt.Errorf("pellCoreHome is empty")
	}
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	// create a new keybase based on the selected backend
	backend := ckeys.BackendTest
	if keyringBackend == config.KeyringBackendFile {
		backend = ckeys.BackendFile
	}

	return ckeys.New(sdk.KeyringServiceName(), backend, cliDir, reader, cdc)
}

func SetupConfigForTest() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(cmd.Bech32PrefixAccAddr, cmd.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(cmd.Bech32PrefixValAddr, cmd.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(cmd.Bech32PrefixConsAddr, cmd.Bech32PrefixConsPub)
	//config.SetCoinType(cmd.MetaChainCoinType)
	config.SetFullFundraiserPath(cmd.PellChainHDPath)
	sdk.SetCoinDenomRegex(func() string {
		return cmd.DenomRegex
	})
}
