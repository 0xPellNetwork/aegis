package types

import (
	"gorm.io/gorm"
)

const (
	// LastBlockNumID is the identifier to access the last block number in the database
	// Using a specific hex value as a unique identifier
	LastBlockNumID uint = 0xBEEF

	// LastTxHashID is the identifier to access the last transaction hash in the database
	// Using a specific hex value as a unique identifier
	LastTxHashID uint = 0xBEF0
)

// LastBlockSQLType is a model for storing the last block number
type LastBlockSQLType struct {
	gorm.Model
	Num              uint64
	LastInboundBlock uint64
}

// LastTransactionSQLType is a model for storing the last transaction hash
type LastTransactionSQLType struct {
	gorm.Model
	Hash string
}

// ToLastBlockSQLType converts a last block number to a LastBlockSQLType
func ToLastBlockSQLType(lastBlock uint64, blockIndex uint64) *LastBlockSQLType {
	return &LastBlockSQLType{
		Model:            gorm.Model{ID: LastBlockNumID},
		Num:              lastBlock,
		LastInboundBlock: blockIndex,
	}
}

// ToLastTxHashSQLType converts a last transaction hash to a LastTransactionSQLType
func ToLastTxHashSQLType(lastTx string) *LastTransactionSQLType {
	return &LastTransactionSQLType{
		Model: gorm.Model{ID: LastTxHashID},
		Hash:  lastTx,
	}
}
