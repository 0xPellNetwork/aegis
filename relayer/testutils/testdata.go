package testutils

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/btcsuite/btcd/btcjson"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/onrik/ethrpc"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/coin"
	"github.com/0xPellNetwork/aegis/relayer/config"
	testxmsg "github.com/0xPellNetwork/aegis/relayer/testdata/xmsg"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

const (
	TestDataPathEVM          = "testdata/evm"
	TestDataPathBTC          = "testdata/btc"
	TestDataPathXmsg         = "testdata/xmsg"
	RestrictedEVMAddressTest = "0x8a81Ba8eCF2c418CAe624be726F505332DF119C6"
	RestrictedBtcAddressTest = "bcrt1qzp4gt6fc7zkds09kfzaf9ln9c5rvrzxmy6qmpp"
)

// cloneXmsg returns a deep copy of the xmsg
func cloneXmsg(t *testing.T, xmsg *xmsgtypes.Xmsg) *xmsgtypes.Xmsg {
	data, err := xmsg.Marshal()
	require.NoError(t, err)
	cloned := &xmsgtypes.Xmsg{}
	err = cloned.Unmarshal(data)
	require.NoError(t, err)
	return cloned
}

// SaveObjectToJSONFile saves an object to a file in JSON format
func SaveObjectToJSONFile(obj interface{}, filename string) error {
	file, err := os.Create(filepath.Clean(filename))
	if err != nil {
		return err
	}
	defer file.Close()

	// write the struct to the file
	encoder := json.NewEncoder(file)
	return encoder.Encode(obj)
}

// LoadObjectFromJSONFile loads an object from a file in JSON format
func LoadObjectFromJSONFile(t *testing.T, obj interface{}, filename string) {
	file, err := os.Open(filepath.Clean(filename))
	require.NoError(t, err)
	defer file.Close()

	// read the struct from the file
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&obj)
	require.NoError(t, err)
}

func ComplianceConfigTest() config.ComplianceConfig {
	return config.ComplianceConfig{
		RestrictedAddresses: []string{RestrictedEVMAddressTest, RestrictedBtcAddressTest},
	}
}

// SaveTrimedEVMBlockTrimTxInput trims tx input data from a block and saves it to a file
func SaveEVMBlockTrimTxInput(block *ethrpc.Block, filename string) error {
	for i := range block.Transactions {
		block.Transactions[i].Input = "0x"
	}
	return SaveObjectToJSONFile(block, filename)
}

// SaveTrimedBTCBlockTrimTx trims tx data from a block and saves it to a file
func SaveBTCBlockTrimTx(blockVb *btcjson.GetBlockVerboseTxResult, filename string) error {
	for i := range blockVb.Tx {
		// reserve one coinbase tx and one non-coinbase tx
		if i >= 2 {
			blockVb.Tx[i].Hex = ""
			blockVb.Tx[i].Vin = nil
			blockVb.Tx[i].Vout = nil
		}
	}
	return SaveObjectToJSONFile(blockVb, filename)
}

// LoadXmsgByIntx loads archived xmsg by intx
func LoadXmsgByIntx(
	t *testing.T,
	chainID int64,
	coinType coin.CoinType,
	intxHash string,
) *xmsgtypes.Xmsg {
	// nameXmsg := path.Join("../", TestDataPathXmsg, FileNameXmsgByIntx(chainID, intxHash, coinType))

	// xmsg := &xmsgtypes.Xmsg{}
	// LoadObjectFromJSONFile(t, &xmsg, nameXmsg)
	// return xmsg

	// get xmsg
	xmsg, found := testxmsg.XmsgByIntxMap[chainID][coinType][intxHash]
	require.True(t, found)

	// clone xmsg for each individual test
	cloned := cloneXmsg(t, xmsg)
	return cloned
}

// LoadXmsgByNonce loads archived xmsg by nonce
func LoadXmsgByNonce(
	t *testing.T,
	chainID int64,
	nonce uint64,
) *xmsgtypes.Xmsg {
	// nameXmsg := path.Join("../", TestDataPathXmsg, FileNameXmsgByNonce(chainID, nonce))

	// xmsg := &xmsgtypes.Xmsg{}
	// LoadObjectFromJSONFile(t, &xmsg, nameXmsg)

	// get xmsg
	xmsg, found := testxmsg.XmsgByNonceMap[chainID][nonce]
	require.True(t, found)

	// clone xmsg for each individual test
	cloned := cloneXmsg(t, xmsg)
	return cloned
}

// LoadEVMBlock loads archived evm block from file
func LoadEVMBlock(t *testing.T, dir string, chainID int64, blockNumber uint64, trimmed bool) *ethrpc.Block {
	name := path.Join(dir, TestDataPathEVM, FileNameEVMBlock(chainID, blockNumber, trimmed))
	block := &ethrpc.Block{}
	LoadObjectFromJSONFile(t, block, name)
	return block
}

// LoadBTCIntxRawResult loads archived Bitcoin intx raw result from file
func LoadBTCIntxRawResult(t *testing.T, dir string, chainID int64, txHash string, donation bool) *btcjson.TxRawResult {
	name := path.Join(dir, TestDataPathBTC, FileNameBTCIntx(chainID, txHash, donation))
	rawResult := &btcjson.TxRawResult{}
	LoadObjectFromJSONFile(t, rawResult, name)
	return rawResult
}

// LoadBTCTxRawResultNXmsg loads archived Bitcoin outtx raw result and corresponding xmsg
func LoadBTCTxRawResultNXmsg(t *testing.T, dir string, chainID int64, nonce uint64) (*btcjson.TxRawResult, *xmsgtypes.Xmsg) {
	//nameTx := FileNameBTCOuttx(chainID, nonce)
	nameTx := path.Join(dir, TestDataPathBTC, FileNameBTCOuttx(chainID, nonce))
	rawResult := &btcjson.TxRawResult{}
	LoadObjectFromJSONFile(t, rawResult, nameTx)

	xmsg := LoadXmsgByNonce(t, chainID, nonce)
	return rawResult, xmsg
}

// LoadEVMIntx loads archived intx from file
func LoadEVMIntx(
	t *testing.T,
	dir string,
	chainID int64,
	intxHash string,
	coinType coin.CoinType) *ethrpc.Transaction {
	nameTx := path.Join(dir, TestDataPathEVM, FileNameEVMIntx(chainID, intxHash, coinType, false))

	tx := &ethrpc.Transaction{}
	LoadObjectFromJSONFile(t, &tx, nameTx)
	return tx
}

// LoadEVMIntxReceipt loads archived intx receipt from file
func LoadEVMIntxReceipt(
	t *testing.T,
	dir string,
	chainID int64,
	intxHash string,
	coinType coin.CoinType) *ethtypes.Receipt {
	nameReceipt := path.Join(dir, TestDataPathEVM, FileNameEVMIntxReceipt(chainID, intxHash, coinType, false))

	receipt := &ethtypes.Receipt{}
	LoadObjectFromJSONFile(t, &receipt, nameReceipt)
	return receipt
}

// LoadEVMIntxNReceipt loads archived intx and receipt from file
func LoadEVMIntxNReceipt(
	t *testing.T,
	dir string,
	chainID int64,
	intxHash string,
	coinType coin.CoinType) (*ethrpc.Transaction, *ethtypes.Receipt) {
	// load archived intx and receipt
	tx := LoadEVMIntx(t, dir, chainID, intxHash, coinType)
	receipt := LoadEVMIntxReceipt(t, dir, chainID, intxHash, coinType)

	return tx, receipt
}

// LoadEVMIntxDonation loads archived donation intx from file
func LoadEVMIntxDonation(
	t *testing.T,
	dir string,
	chainID int64,
	intxHash string,
	coinType coin.CoinType) *ethrpc.Transaction {
	nameTx := path.Join(dir, TestDataPathEVM, FileNameEVMIntx(chainID, intxHash, coinType, true))

	tx := &ethrpc.Transaction{}
	LoadObjectFromJSONFile(t, &tx, nameTx)
	return tx
}

// LoadEVMIntxReceiptDonation loads archived donation intx receipt from file
func LoadEVMIntxReceiptDonation(
	t *testing.T,
	dir string,
	chainID int64,
	intxHash string,
	coinType coin.CoinType) *ethtypes.Receipt {
	nameReceipt := path.Join(dir, TestDataPathEVM, FileNameEVMIntxReceipt(chainID, intxHash, coinType, true))

	receipt := &ethtypes.Receipt{}
	LoadObjectFromJSONFile(t, &receipt, nameReceipt)
	return receipt
}

// LoadEVMIntxNReceiptDonation loads archived donation intx and receipt from file
func LoadEVMIntxNReceiptDonation(
	t *testing.T,
	dir string,
	chainID int64,
	intxHash string,
	coinType coin.CoinType) (*ethrpc.Transaction, *ethtypes.Receipt) {
	// load archived donation intx and receipt
	tx := LoadEVMIntxDonation(t, dir, chainID, intxHash, coinType)
	receipt := LoadEVMIntxReceiptDonation(t, dir, chainID, intxHash, coinType)

	return tx, receipt
}

// LoadTxNReceiptNXmsg loads archived intx, receipt and corresponding xmsg from file
func LoadEVMIntxNReceiptNXmsg(
	t *testing.T,
	dir string,
	chainID int64,
	intxHash string,
	coinType coin.CoinType) (*ethrpc.Transaction, *ethtypes.Receipt, *xmsgtypes.Xmsg) {
	// load archived intx, receipt and xmsg
	tx := LoadEVMIntx(t, dir, chainID, intxHash, coinType)
	receipt := LoadEVMIntxReceipt(t, dir, chainID, intxHash, coinType)
	xmsg := LoadXmsgByIntx(t, chainID, coinType, intxHash)

	return tx, receipt, xmsg
}

// LoadEVMOuttx loads archived evm outtx from file
func LoadEVMOuttx(
	t *testing.T,
	dir string,
	chainID int64,
	txHash string,
	coinType coin.CoinType) *ethtypes.Transaction {
	nameTx := path.Join(dir, TestDataPathEVM, FileNameEVMOuttx(chainID, txHash, coinType))

	tx := &ethtypes.Transaction{}
	LoadObjectFromJSONFile(t, &tx, nameTx)
	return tx
}

// LoadEVMOuttxReceipt loads archived evm outtx receipt from file
func LoadEVMOuttxReceipt(
	t *testing.T,
	dir string,
	chainID int64,
	txHash string,
	coinType coin.CoinType,
	eventName string) *ethtypes.Receipt {
	nameReceipt := path.Join(dir, TestDataPathEVM, FileNameEVMOuttxReceipt(chainID, txHash, coinType, eventName))

	receipt := &ethtypes.Receipt{}
	LoadObjectFromJSONFile(t, &receipt, nameReceipt)
	return receipt
}

// LoadEVMOuttxNReceipt loads archived evm outtx and receipt from file
func LoadEVMOuttxNReceipt(
	t *testing.T,
	dir string,
	chainID int64,
	txHash string,
	coinType coin.CoinType) (*ethtypes.Transaction, *ethtypes.Receipt) {
	// load archived evm outtx and receipt
	tx := LoadEVMOuttx(t, dir, chainID, txHash, coinType)
	receipt := LoadEVMOuttxReceipt(t, dir, chainID, txHash, coinType, "")

	return tx, receipt
}

// LoadEVMOuttxNReceiptNEvent loads archived xmsg, outtx and receipt from file
func LoadEVMXmsgNOuttxNReceipt(
	t *testing.T,
	dir string,
	chainID int64,
	nonce uint64,
	eventName string) (*xmsgtypes.Xmsg, *ethtypes.Transaction, *ethtypes.Receipt) {
	xmsg := LoadXmsgByNonce(t, chainID, nonce)
	coinType := coin.CoinType_PELL
	txHash := xmsg.GetCurrentOutTxParam().OutboundTxHash
	outtx := LoadEVMOuttx(t, dir, chainID, txHash, coinType)
	receipt := LoadEVMOuttxReceipt(t, dir, chainID, txHash, coinType, eventName)
	return xmsg, outtx, receipt
}
