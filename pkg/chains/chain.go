package chains

import (
	"fmt"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

type SigninAlgo string

// Chains represent a slice of Chain
type Chains []Chain

// IsEqual compare two chain to see whether they represent the same chain
func (chain Chain) IsEqual(c Chain) bool {
	return chain.Id == c.Id
}

func (chain Chain) IsPellChain() bool {
	return chain.Network == NetWork_PELL
}

func (chain Chain) IsExternalChain() bool {
	return chain.Network != NetWork_PELL
}

// EncodeAddress bytes representations of address
// on EVM chain, it is 20Bytes
// on Bitcoin chain, it is P2WPKH address, []byte(bech32 encoded string)
func (chain Chain) EncodeAddress(b []byte) (string, error) {
	if IsEVMChain(chain.Id) {
		addr := ethcommon.BytesToAddress(b)
		if addr == (ethcommon.Address{}) {
			return "", fmt.Errorf("invalid EVM address")
		}
		return addr.Hex(), nil
	}

	return "", fmt.Errorf("chain (%d) not supported", chain.Id)
}

// DecodeAddress decode the address string to bytes
func (chain Chain) DecodeAddress(addr string) ([]byte, error) {
	return DecodeAddressFromChainID(chain.Id, addr)
}

// DecodeAddressFromChainID decode the address string to bytes
func DecodeAddressFromChainID(chainID int64, addr string) ([]byte, error) {
	if IsEVMChain(chainID) {
		return ethcommon.HexToAddress(addr).Bytes(), nil
	}

	return nil, fmt.Errorf("chain (%d) not supported", chainID)
}

func IsPellChain(chainID int64) bool {
	_, exist := FindChain(func(c Chain) bool {
		return c.Network == NetWork_PELL && c.Id == chainID
	})

	return exist
}

// IsEVMChain returns true if the chain is an EVM chain
func IsEVMChain(chainID int64) bool {
	_, isEvmChain := FindChain(func(c Chain) bool {
		return c.VmType == VMType_EVM && c.Id == chainID
	})

	return isEvmChain
}

// IsHeaderSupportedEvmChain returns true if the chain is an EVM chain supporting block header-based verification
// TODO: put this information directly in chain object
func IsHeaderSupportedEvmChain(chainID int64) bool {
	return chainID == 5 || // Goerli
		chainID == 11155111 || // Sepolia
		chainID == 97 || // BSC testnet
		chainID == 1337 || // eth privnet
		chainID == 1 || // eth mainnet
		chainID == 56 // bsc mainnet
}

// SupportMerkleProof returns true if the chain supports block header-based verification
func (chain Chain) SupportMerkleProof() bool {
	return IsEVMChain(chain.Id)
}

// // IsEthereumChain returns true if the chain is an Ethereum chain
// // TODO: put this information directly in chain object
func IsEthereumChain(chainID int64) bool {
	_, exist := FindChain(func(c Chain) bool {
		return c.Network == NetWork_ETH
	})

	return exist
}

// IsEmpty is to determinate whether the chain is empty
func (chain Chain) IsEmpty() bool {
	return strings.TrimSpace(chain.String()) == ""
}

// Has check whether chain c is in the list
func (chains Chains) Has(c Chain) bool {
	for _, ch := range chains {
		if ch.IsEqual(c) {
			return true
		}
	}
	return false
}

// Distinct return a distinct set of chains, no duplicates
func (chains Chains) Distinct() Chains {
	var newChains Chains
	for _, chain := range chains {
		if !newChains.Has(chain) {
			newChains = append(newChains, chain)
		}
	}
	return newChains
}

func (chains Chains) Strings() []string {
	str := make([]string, len(chains))
	for i, c := range chains {
		str[i] = c.String()
	}
	return str
}

// InChainList checks whether the chain is in the chain list
func (chain Chain) InChainList(chainList []*Chain) bool {
	return ChainIDInChainList(chain.Id, chainList)
}

// ChainIDInChainList checks whether the chainID is in the chain list
func ChainIDInChainList(chainID int64, chainList []*Chain) bool {
	for _, c := range chainList {
		if chainID == c.Id {
			return true
		}
	}
	return false
}
