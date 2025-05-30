package sample

import (
	"crypto/ecdsa"
	"math/big"
	"math/rand"
	"strconv"
	"testing"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/cosmos"
	"github.com/0xPellNetwork/aegis/pkg/crypto"
)

func PubKeySet() *crypto.PubKeySet {
	pubKeySet := crypto.PubKeySet{
		Secp256k1: crypto.PubKey(secp256k1.GenPrivKey().PubKey().Bytes()),
		Ed25519:   crypto.PubKey(ed25519.GenPrivKey().PubKey().String()),
	}
	return &pubKeySet
}

// PubKeyString returns a sample public key string
func PubKeyString() string {
	priKey := ed25519.GenPrivKey()
	s, err := cosmos.Bech32ifyPubKey(cosmos.Bech32PubKeyTypeAccPub, priKey.PubKey())
	if err != nil {
		panic(err)
	}
	pubkey, err := crypto.NewPubKey(s)
	if err != nil {
		panic(err)
	}
	return pubkey.String()
}

// PrivKeyAddressPair returns a private key, address pair
func PrivKeyAddressPair() (*ed25519.PrivKey, sdk.AccAddress) {
	privKey := ed25519.GenPrivKey()
	addr := privKey.PubKey().Address()

	return privKey, sdk.AccAddress(addr)
}

// EthAddress returns a sample ethereum address
func EthAddress() ethcommon.Address {
	return ethcommon.BytesToAddress(sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address()).Bytes())
}

// Hash returns a sample hash
func Hash() ethcommon.Hash {
	return EthAddress().Hash()
}

// PubKey returns a sample account PubKey
func PubKey(r *rand.Rand) cryptotypes.PubKey {
	seed := []byte(strconv.Itoa(r.Int()))
	return ed25519.GenPrivKeyFromSecret(seed).PubKey()
}

// Bech32AccAddress returns a sample account address
func Bech32AccAddress() sdk.AccAddress {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.AccAddress(addr)
}

// AccAddress returns a sample account address in string
func AccAddress() string {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.AccAddress(addr).String()
}

// ValAddress returns a sample validator operator address
func ValAddress(r *rand.Rand) sdk.ValAddress {
	return sdk.ValAddress(PubKey(r).Address())
}

// EthTx returns a sample ethereum transaction with the associated tx data bytes
func EthTx(t *testing.T, chainID int64, to ethcommon.Address, nonce uint64) (*ethtypes.Transaction, []byte) {
	tx := ethtypes.NewTx(&ethtypes.DynamicFeeTx{
		ChainID:   big.NewInt(chainID),
		Nonce:     nonce,
		GasTipCap: nil,
		GasFeeCap: nil,
		Gas:       21000,
		To:        &to,
		Value:     big.NewInt(5),
		Data:      nil,
	})

	txBytes, err := tx.MarshalBinary()
	require.NoError(t, err)

	return tx, txBytes
}

// EthTxSigned returns a sample signed ethereum transaction with the address of the sender
func EthTxSigned(
	t *testing.T,
	chainID int64,
	to ethcommon.Address,
	nonce uint64,
) (*ethtypes.Transaction, []byte, ethcommon.Address) {
	tx, _ := EthTx(t, chainID, to, nonce)

	// generate a private key and get address
	privateKey, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	fromAddress := ethcrypto.PubkeyToAddress(*publicKeyECDSA)
	require.True(t, ok)

	// sign the transaction
	signer := ethtypes.NewLondonSigner(tx.ChainId())
	signedTx, err := ethtypes.SignTx(tx, signer, privateKey)
	require.NoError(t, err)

	txBytes, err := signedTx.MarshalBinary()
	require.NoError(t, err)

	return signedTx, txBytes, fromAddress
}
