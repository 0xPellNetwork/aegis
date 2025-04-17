package proofs

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/wire"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// NewEthereumHeader returns a new HeaderData containing an Ethereum header
func NewEthereumHeader(header []byte) HeaderData {
	return HeaderData{
		Data: &HeaderData_EthereumHeader{
			EthereumHeader: header,
		},
	}
}

// NewBitcoinHeader returns a new HeaderData containing a Bitcoin header
func NewBitcoinHeader(header []byte) HeaderData {
	return HeaderData{
		Data: &HeaderData_BitcoinHeader{
			BitcoinHeader: header,
		},
	}
}

// ParentHash extracts the parent hash from the header
func (h HeaderData) ParentHash() ([]byte, error) {
	switch data := h.Data.(type) {
	case *HeaderData_EthereumHeader:
		var header ethtypes.Header
		if err := rlp.DecodeBytes(data.EthereumHeader, &header); err != nil {
			return nil, err
		}
		return header.ParentHash.Bytes(), nil
	case *HeaderData_BitcoinHeader:
		var header wire.BlockHeader
		if err := header.Deserialize(bytes.NewReader(data.BitcoinHeader)); err != nil {
			return nil, err
		}
		return header.PrevBlock[:], nil
	default:
		return nil, errors.New("unrecognized header type")
	}
}

func (h HeaderData) ValidateTimestamp(pellTime time.Time) error {
	switch data := h.Data.(type) {
	case *HeaderData_EthereumHeader:
		// No timestamp validation for Ethereum for now
		return nil
	case *HeaderData_BitcoinHeader:
		var header wire.BlockHeader
		if err := header.Deserialize(bytes.NewReader(data.BitcoinHeader)); err != nil {
			return err
		}
		// Below checks are borrowed from btcd/blockchain/validate.go because they are not exported
		//
		// A block timestamp must not have a greater precision than one second.
		// This check is necessary because Go time.Time values support
		// nanosecond precision whereas the consensus rules only apply to
		// seconds and it's much nicer to deal with standard Go time values
		// instead of converting to seconds everywhere.
		if !header.Timestamp.Equal(time.Unix(header.Timestamp.Unix(), 0)) {
			return fmt.Errorf("block timestamp of %v has a higher precision than one second", header.Timestamp)
		}

		// Ensure the block time is not too far in the future.
		maxTimestamp := pellTime.Add(time.Second * blockchain.MaxTimeOffsetSeconds)
		if header.Timestamp.After(maxTimestamp) {
			return fmt.Errorf("block timestamp of %v is too far in the future", header.Timestamp)
		}
		return nil
	default:
		return errors.New("cannot validate timestamp for unrecognized header type")
	}
}

// Validate performs a basic validation of the HeaderData
func (h HeaderData) Validate(blockHash []byte, chainID int64, height int64) error {
	switch data := h.Data.(type) {
	case *HeaderData_EthereumHeader:
		return validateEthereumHeader(data.EthereumHeader, blockHash, height)
	default:
		return errors.New("unrecognized header type")
	}
}

// validateEthereumHeader performs a basic validation of the Ethereum header
func validateEthereumHeader(headerBytes []byte, blockHash []byte, height int64) error {
	// on ethereum the block header is ~538 bytes in RLP encoding
	if len(headerBytes) > 4096 {
		return fmt.Errorf("header too long (%d)", len(headerBytes))
	}

	// RLP encoded block header
	var header ethtypes.Header
	if err := rlp.DecodeBytes(headerBytes, &header); err != nil {
		return fmt.Errorf("cannot decode RLP (%s)", err)
	}
	if err := header.SanityCheck(); err != nil {
		return fmt.Errorf("sanity check failed (%s)", err)
	}
	if !bytes.Equal(blockHash, header.Hash().Bytes()) {
		return fmt.Errorf("block hash mismatch (%s) vs (%s)", hex.EncodeToString(blockHash), header.Hash().Hex())
	}
	if height != header.Number.Int64() {
		return fmt.Errorf("height mismatch (%d) vs (%d)", height, header.Number.Int64())
	}
	return nil
}
