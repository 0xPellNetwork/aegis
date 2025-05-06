package types

import (
	"fmt"
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

// VerifyInTxBody validates the tx body for a inbound tx
func VerifyInTxBody(
	msg MsgAddToInTxTracker,
	txBytes []byte,
	chainParams relayertypes.ChainParams,
	tss relayertypes.QueryGetTssAddressResponse,
) error {
	// verify message against transaction body
	if chains.IsEVMChain(msg.ChainId) {
		return verifyInTxBodyEVM(msg, txBytes, chainParams, tss)
	}

	// TODO: implement verifyInTxBodyBTC

	return fmt.Errorf("cannot verify inTx body for chain %d", msg.ChainId)
}

// verifyInTxBodyEVM validates the chain id and connector contract address for Pell, ERC20 contract address for ERC20 and TSS address for Gas.
func verifyInTxBodyEVM(
	msg MsgAddToInTxTracker,
	txBytes []byte,
	chainParams relayertypes.ChainParams,
	tss relayertypes.QueryGetTssAddressResponse,
) error {
	var txx ethtypes.Transaction
	err := txx.UnmarshalBinary(txBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal transaction %s", err.Error())
	}
	if txx.Hash().Hex() != msg.TxHash {
		return fmt.Errorf("invalid hash, want tx hash %s, got %s", txx.Hash().Hex(), msg.TxHash)
	}
	if txx.ChainId().Cmp(big.NewInt(msg.ChainId)) != 0 {
		return fmt.Errorf("invalid chain id, want evm chain id %d, got %d", txx.ChainId(), msg.ChainId)
	}

	switch txx.To().Hex() {
	// Inbound ERC20 interacts with strategyManagerContract or delegationManagerContract
	case chainParams.StrategyManagerContractAddress:
	case chainParams.DelegationManagerContractAddress:
	default:
		return fmt.Errorf("tx event is not supported %d:%s", chainParams.ChainId, txx.To().Hex())
	}

	return nil
}

// VerifyOutTxBody verifies the tx body for a outbound tx
func VerifyOutTxBody(msg MsgAddToOutTxTracker, txBytes []byte, tss relayertypes.QueryGetTssAddressResponse) error {
	// verify message against transaction body
	if chains.IsEVMChain(msg.ChainId) {
		return verifyOutTxBodyEVM(msg, txBytes, tss.Eth)
	}
	return fmt.Errorf("cannot verify outTx body for chain %d", msg.ChainId)
}

// verifyOutTxBodyEVM validates the sender address, nonce, chain id and tx hash.
func verifyOutTxBodyEVM(msg MsgAddToOutTxTracker, txBytes []byte, tssEth string) error {
	var txx ethtypes.Transaction
	err := txx.UnmarshalBinary(txBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal transaction %s", err.Error())
	}
	signer := ethtypes.NewLondonSigner(txx.ChainId())
	sender, err := ethtypes.Sender(signer, &txx)
	if err != nil {
		return fmt.Errorf("failed to recover sender %s", err.Error())
	}
	tssAddr := eth.HexToAddress(tssEth)
	if tssAddr == (eth.Address{}) {
		return fmt.Errorf("tss address not found")
	}
	if sender != tssAddr {
		return fmt.Errorf("sender is not tss address %s", sender)
	}
	if txx.ChainId().Cmp(big.NewInt(msg.ChainId)) != 0 {
		return fmt.Errorf("invalid chain id, want evm chain id %d, got %d", txx.ChainId(), msg.ChainId)
	}
	if txx.Nonce() != msg.Nonce {
		return fmt.Errorf("invalid nonce, want nonce %d, got %d", txx.Nonce(), msg.Nonce)
	}
	if txx.Hash().Hex() != msg.TxHash {
		return fmt.Errorf("invalid tx hash, want tx hash %s, got %s", txx.Hash().Hex(), msg.TxHash)
	}
	return nil
}
