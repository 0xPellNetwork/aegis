package sample

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"hash/fnv"
	"math/rand"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/cmd/pellcored/config"
	"github.com/0xPellNetwork/aegis/pkg/chains"
)

var ErrSample = errors.New("sample error")

//go:embed genesis.json
var genesisJSON []byte

func newRandFromSeed(s int64) *rand.Rand {
	// #nosec G404 test purpose - weak randomness is not an issue here
	return rand.New(rand.NewSource(s))
}

func newRandFromStringSeed(t *testing.T, s string) *rand.Rand {
	h := fnv.New64a()
	_, err := h.Write([]byte(s))
	require.NoError(t, err)
	return newRandFromSeed(int64(h.Sum64()))
}

// Rand returns a new random number generator
func Rand() *rand.Rand {
	return newRandFromSeed(42)
}

// Validator returns a sample staking validator
func Validator(t testing.TB, r *rand.Rand) stakingtypes.Validator {
	seed := []byte(strconv.Itoa(r.Int()))
	val, err := stakingtypes.NewValidator(
		ValAddress(r).String(),
		ed25519.GenPrivKeyFromSecret(seed).PubKey(),
		stakingtypes.Description{})
	require.NoError(t, err)
	return val
}

func PellIndex(t *testing.T) string {
	msg := Xmsg_pell(t, "foo")
	hash := ethcrypto.Keccak256Hash([]byte(msg.String()))
	return hash.Hex()
}

// Bytes returns a sample byte array
func Bytes() []byte {
	return []byte("sample")
}

// String returns a sample string
func String() string {
	return "sample"
}

// StringRandom returns a sample string with random alphanumeric characters
func StringRandom(r *rand.Rand, length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}

// Coins returns a sample sdk.Coins
func Coins() sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(config.BaseDenom, sdkmath.NewInt(42)))
}

// Uint64InRange returns a sample uint64 in the given ranges
func Uint64InRange(low, high uint64) uint64 {
	r := newRandFromSeed(int64(low))
	return r.Uint64()%(high-low) + low
}

// Int64InRange returns a sample int64 in the given ranges
func Int64InRange(low, high int64) int64 {
	r := newRandFromSeed(low)
	return r.Int63()%(high-low) + low
}

func UintInRange(low, high uint64) sdkmath.Uint {
	u := Uint64InRange(low, high)
	return sdkmath.NewUint(u)
}

func IntInRange(low, high int64) sdkmath.Int {
	i := Int64InRange(low, high)
	return sdkmath.NewInt(i)
}

func AppState(t *testing.T) map[string]json.RawMessage {
	appState, err := genutiltypes.GenesisStateFromAppGenesis(AppGenesis(t))
	require.NoError(t, err)
	return appState
}

func Chain(chainID int64) *chains.Chain {
	return &chains.Chain{
		Id: chainID,
	}
}

func EventIndex() uint64 {
	r := newRandFromSeed(1)
	return r.Uint64()
}

func AppGenesis(t *testing.T) *genutiltypes.AppGenesis {
	reader := bytes.NewReader(genesisJSON)
	genDoc, err := genutiltypes.AppGenesisFromReader(reader)
	require.NoError(t, err)
	return genDoc
}
