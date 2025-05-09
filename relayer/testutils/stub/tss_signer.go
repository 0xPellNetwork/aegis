package stub

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/relayer/testutils"
)

var TestPrivateKey *ecdsa.PrivateKey

var _ interfaces.TSSSigner = (*TSS)(nil)

func init() {
	var err error
	TestPrivateKey, err = crypto.GenerateKey()
	if err != nil {
		fmt.Println(err.Error())
	}
}

// TSS is a mock of TSS signer for testing
type TSS struct {
	chain      chains.Chain
	evmAddress string
}

func NewMockTSS(chain chains.Chain, evmAddress string) *TSS {
	return &TSS{
		chain:      chain,
		evmAddress: evmAddress,
	}
}

func NewTSSMainnet() *TSS {
	return NewMockTSS(chains.BscTestnetChain(), testutils.TSSAddressEVMMainnet)
}

func NewTSSIgnite3() *TSS {
	return NewMockTSS(chains.BscTestnetChain(), testutils.TSSAddressEVMIgnite3)
}

// Sign uses test key unrelated to any tss key in production
func (s *TSS) Sign(ctx context.Context, data []byte, _ uint64, _ uint64, _ *chains.Chain, _ string) ([65]byte, error) {
	signature, err := crypto.Sign(data, TestPrivateKey)
	if err != nil {
		return [65]byte{}, err
	}
	var sigbyte [65]byte
	_ = copy(sigbyte[:], signature[:65])

	return sigbyte, nil
}

// Pubkey uses the hardcoded private test key to generate the public key in bytes
func (s *TSS) Pubkey() []byte {
	publicKey := TestPrivateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("error casting public key to ECDSA")
	}
	return crypto.FromECDSAPub(publicKeyECDSA)
}

func (s *TSS) EVMAddress() ethcommon.Address {
	return ethcommon.HexToAddress(s.evmAddress)
}

func (s *TSS) PubKeyCompressedBytes() []byte {
	return []byte{}
}
