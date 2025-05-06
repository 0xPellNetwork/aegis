package types

import (
	"fmt"

	"cosmossdk.io/errors"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/0xPellNetwork/aegis/pkg/chains"
)

// ValidatePellIndex validates the pell index
func ValidatePellIndex(index string) error {
	if len(index) != PellIndexLength {
		return errors.Wrap(ErrInvalidIndexValue, fmt.Sprintf("invalid index length %d", len(index)))
	}
	return nil
}

// ValidateHashForChain validates the hash for the chain
func ValidateHashForChain(hash string, chainID int64) error {
	if chains.IsEthereumChain(chainID) || chains.IsPellChain(chainID) {
		_, err := hexutil.Decode(hash)
		if err != nil {
			return fmt.Errorf("hash must be a valid ethereum hash %s", hash)
		}
		return nil
	}

	return fmt.Errorf("invalid chain id %d", chainID)
}

// ValidateAddressForChain validates the address for the chain
func ValidateAddressForChain(address string, chainID int64) error {
	// we do not validate the address for pell chain as the address field can be btc or eth address
	if chains.IsPellChain(chainID) {
		return nil
	}
	if chains.IsEthereumChain(chainID) {
		if !ethcommon.IsHexAddress(address) {
			return fmt.Errorf("invalid address %s , chain %d", address, chainID)
		}
		return nil
	}

	return fmt.Errorf("invalid chain id %d", chainID)
}
