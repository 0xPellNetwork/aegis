package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	// #nosec G101 - used for testing only
	PellEthPriv           = "9D00E4D7A8A14384E01CD90B83745BCA847A66AD8797A9904A200C28C2648E64"
	SystemContractAddress = "0x91d18e54DAf4F677cB28167158d6dd21F6aB3921"
)

type Request struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type Response struct {
	Result json.RawMessage `json:"result"`
	Error  *Error          `json:"error"`
	ID     int             `json:"id"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <blocknum>\n", os.Args[0])
		os.Exit(1)
	}
	fmt.Printf("Start testing the pEVM ETH JSON-RPC for all txs...\n")
	fmt.Printf("Test1: simple gas voter tx\n")

	bn, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		panic(err)
	}
	if bn < 0 {
		panic("Block number must be non-negative")
	}
	// #nosec G701 check as positive
	bnUint64 := uint64(bn)

	if false {
		// USE RAW JSON-RPC INSTEAD
		pevmClient, err := ethclient.Dial("http://localhost:8545")
		if err != nil {
			panic(err)
		}

		block, err := pevmClient.BlockByNumber(context.Background(), big.NewInt(bn))
		if err != nil {
			panic(err)
		}

		fmt.Printf("Block number: %d, num of txs %d (should be 1)\n", block.Number(), len(block.Transactions()))
	}

	client := &EthClient{
		Endpoint:   "http://localhost:8545",
		HTTPClient: &http.Client{},
	}
	resp := client.EthGetBlockByNumber(bnUint64, false)
	var jsonObject map[string]interface{}
	if resp.Error != nil {
		fmt.Printf("Error: %s (code %d)\n", resp.Error.Message, resp.Error.Code)
		panic(resp.Error.Message)
	}
	err = json.Unmarshal(resp.Result, &jsonObject)
	if err != nil {
		panic(err)
	}

	txs, ok := jsonObject["transactions"].([]interface{})
	if !ok || len(txs) != 1 {
		panic("Wrong number of txs")
	}
	txhash, ok := txs[0].(string)
	if !ok {
		panic("Wrong tx type")
	}
	fmt.Printf("Tx hash: %s\n", txhash)
	tx := client.EthGetTransactionReceipt(txhash)
	if tx.Error != nil {
		fmt.Printf("Error: %s (code %d)\n", tx.Error.Message, tx.Error.Code)
		panic(tx.Error.Message)
	}

	// tx receipt can be queried by ethclient queries.
	pevmClient, err := ethclient.Dial(client.Endpoint)
	if err != nil {
		panic(err)
	}
	receipt, err := pevmClient.TransactionReceipt(context.Background(), ethcommon.HexToHash(txhash))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Receipt status: %+v\n", receipt.Status)

	// HeaderByHash works; BlockByHash does not work;
	// main offending RPC is the transaction type; we have custom type id 56
	// which is not recognized by the go-ethereum client.
	blockHeader, err := pevmClient.HeaderByNumber(context.Background(), big.NewInt(bn))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Block header TxHash: %+v\n", blockHeader.TxHash)

	// TODO
}

type EthClient struct {
	Endpoint   string
	HTTPClient *http.Client
}

func (c *EthClient) EthGetBlockByNumber(blockNum uint64, verbose bool) *Response {
	client := c.HTTPClient
	hexBlockNum := fmt.Sprintf("0x%x", blockNum)
	req := &Request{
		Jsonrpc: "2.0",
		Method:  "eth_getBlockByNumber",
		Params: []interface{}{
			hexBlockNum,
			verbose,
		},
		ID: 1,
	}

	// Encode the request to JSON
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		panic(err)
	}
	// Create a new HTTP request
	httpReq, err := http.NewRequest("POST", c.Endpoint, buf)
	if err != nil {
		panic(err)
	}
	// Set the content type header
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	resp, err := client.Do(httpReq)
	if err != nil {
		panic(err)
	}
	// #nosec G107 - defer close
	defer resp.Body.Close()
	// Decode the response from JSON
	var rpcResp Response
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		panic(err)
	}

	return &rpcResp
}

func (c *EthClient) EthGetTransactionReceipt(txhash string) *Response {
	client := c.HTTPClient
	req := &Request{
		Jsonrpc: "2.0",
		Method:  "eth_getTransactionReceipt",
		Params: []interface{}{
			txhash,
		},
		ID: 1,
	}

	// Encode the request to JSON
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		panic(err)
	}
	// Create a new HTTP request
	httpReq, err := http.NewRequest("POST", c.Endpoint, buf)
	if err != nil {
		panic(err)
	}
	// Set the content type header
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	resp, err := client.Do(httpReq)
	if err != nil {
		panic(err)
	}
	// #nosec G107 - defer close
	defer resp.Body.Close()
	// Decode the response from JSON
	var rpcResp Response
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		panic(err)
	}

	return &rpcResp
}

func (c *EthClient) EthGetLogs() {
	//client := c.HTTPClient
	//req := &Request{
	//	Jsonrpc: "2.0",
	//	Method:  "eth_getTransactionReceipt",
	//	Params: []interface{}{
	//		txhash,
	//	},
	//	ID: 1,
	//}
	//
	//// Encode the request to JSON
	//buf := &bytes.Buffer{}
	//err := json.NewEncoder(buf).Encode(req)
	//if err != nil {
	//	panic(err)
	//}
	//// Create a new HTTP request
	//httpReq, err := http.NewRequest("POST", c.Endpoint, buf)
	//if err != nil {
	//	panic(err)
	//}
	//// Set the content type header
	//httpReq.Header.Set("Content-Type", "application/json")
	//
	//// Send the HTTP request
	//resp, err := client.Do(httpReq)
	//if err != nil {
	//	panic(err)
	//}
	//// #nosec G107 - defer close
	//defer resp.Body.Close()
	//// Decode the response from JSON
	//var rpcResp Response
	//err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	//if err != nil {
	//	panic(err)
	//}
	//
	//return &rpcResp
}

func MustWaitForReceipt(ctx context.Context, client *ethclient.Client, txhash ethcommon.Hash) *types.Receipt {
	for {
		select {
		case <-ctx.Done():
			panic("timeout waiting for transaction receipt")
		default:
			receipt, err := client.TransactionReceipt(context.Background(), txhash)
			if err == nil {
				return receipt
			}
			time.Sleep(1 * time.Second)
		}
	}
}
