package filterdeposit

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/strategymanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nanmu42/etherscan-api"
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/cmd/pelltool/config"
	"github.com/pell-chain/pellcore/pkg/constant"
	"github.com/pell-chain/pellcore/relayer/chains/evm"
)

const (
	EvmMaxRangeFlag   = "evm-max-range"
	EvmStartBlockFlag = "evm-start-block"
)

func NewEvmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eth",
		Short: "Filter inbound eth deposits",
		RunE:  FilterEVMTransactions,
	}

	cmd.Flags().Uint64(EvmMaxRangeFlag, 1000, "number of blocks to scan per iteration")
	cmd.Flags().Uint64(EvmStartBlockFlag, 19463725, "block height to start scanning from")

	return cmd
}

// FilterEVMTransactions is a command that queries an EVM explorer and Contracts for inbound transactions that qualify
// for cross chain transactions.
func FilterEVMTransactions(cmd *cobra.Command, _ []string) error {
	// Get flags
	configFile, err := cmd.Flags().GetString(config.FlagConfig)
	if err != nil {
		return err
	}
	startBlock, err := cmd.Flags().GetUint64(EvmStartBlockFlag)
	if err != nil {
		return err
	}
	blockRange, err := cmd.Flags().GetUint64(EvmMaxRangeFlag)
	if err != nil {
		return err
	}
	btcChainID, err := cmd.Flags().GetString(BTCChainIDFlag)
	if err != nil {
		return err
	}
	// Scan for deposits
	cfg, err := config.GetConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
	res, err := GetTssAddress(cfg, btcChainID)
	if err != nil {
		return err
	}
	list, err := GetEthHashList(cfg, res.Eth, startBlock, blockRange)
	if err != nil {
		return err
	}
	_, err = CheckForXmsg(list, cfg)
	return err
}

// GetEthHashList is a helper function querying total inbound txns by segments of blocks in ranges defined by the config
func GetEthHashList(cfg *config.Config, tssAddress string, startBlock uint64, blockRange uint64) ([]Deposit, error) {
	client, err := ethclient.Dial(cfg.EthRPCURL)
	if err != nil {
		return []Deposit{}, err
	}
	fmt.Println("Connection successful")

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return []Deposit{}, err
	}
	latestBlock := header.Number.Uint64()
	fmt.Println("latest Block: ", latestBlock)

	endBlock := startBlock + blockRange
	deposits := make([]Deposit, 0)
	segment := 0
	for startBlock < latestBlock {
		fmt.Printf("adding segment: %d, startblock: %d\n", segment, startBlock)
		segmentRes, err := GetHashListSegment(client, startBlock, endBlock, tssAddress, cfg)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		deposits = append(deposits, segmentRes...)
		startBlock = endBlock
		endBlock = endBlock + blockRange
		if endBlock > latestBlock {
			endBlock = latestBlock
		}
		segment++
	}
	return deposits, nil
}

// GetHashListSegment queries and filters deposits for a given range
func GetHashListSegment(
	client *ethclient.Client,
	startBlock uint64,
	endBlock uint64,
	tssAddress string,
	cfg *config.Config) ([]Deposit, error) {

	deposits := make([]Deposit, 0)

	strategyManagerAddress := common.HexToAddress(cfg.StrategyManagerAddress)
	strategyManagerContract, err := strategymanager.NewStrategyManager(strategyManagerAddress, client)
	if err != nil {
		return deposits, err
	}
	delegationManagerAddress := common.HexToAddress(cfg.DelegationManagerAddress)
	delegationManagerContract, err := delegationmanager.NewDelegationManager(delegationManagerAddress, client)
	if err != nil {
		return deposits, err
	}

	strategyManagerIter, err := strategyManagerContract.FilterDeposit(&bind.FilterOpts{
		Start:   startBlock,
		End:     &endBlock,
		Context: context.TODO(),
	})
	if err != nil {
		return deposits, err
	}
	delegationManagerIter, err := delegationManagerContract.FilterStakerDelegated(&bind.FilterOpts{
		Start:   startBlock,
		End:     &endBlock,
		Context: context.TODO(),
	}, []common.Address{}, []common.Address{})
	if err != nil {
		return deposits, err
	}

	// Get Staker Deposited events
	for strategyManagerIter.Next() {
		// sanity check tx event
		err := CheckEvmTxLog(&strategyManagerIter.Event.Raw, strategyManagerAddress, "", evm.TopicsPellStakerDeposited)
		if err == nil {
			deposits = append(deposits, Deposit{
				TxID: strategyManagerIter.Event.Raw.TxHash.Hex(),
			})
		}
	}
	// Get Staker Delegated events
	for delegationManagerIter.Next() {
		// sanity check tx event
		err := CheckEvmTxLog(&delegationManagerIter.Event.Raw, delegationManagerAddress, "", evm.TopicsPellStakerDelegated)
		if err == nil {
			deposits = append(deposits, Deposit{
				TxID: delegationManagerIter.Event.Raw.TxHash.Hex(),
			})
		}
	}

	return deposits, nil
}

// getTSSDeposits more specifically queries and filters deposits based on direct transfers the TSS address.
func getTSSDeposits(tssAddress string, startBlock uint64, endBlock uint64, apiKey string) ([]Deposit, error) {
	client := etherscan.New(etherscan.Mainnet, apiKey)
	deposits := make([]Deposit, 0)

	// #nosec G701 these block numbers need to be *int for this particular client package
	startInt := int(startBlock)
	// #nosec G701
	endInt := int(endBlock)
	txns, err := client.NormalTxByAddress(tssAddress, &startInt, &endInt, 0, 0, true)
	if err != nil {
		return deposits, err
	}

	fmt.Println("getTSSDeposits - Num of transactions: ", len(txns))

	for _, tx := range txns {
		if tx.To == tssAddress {
			if strings.Compare(tx.Input, constant.DonationMessage) == 0 {
				continue // skip donation tx
			}
			if tx.TxReceiptStatus != "1" {
				continue
			}
			//fmt.Println("getTSSDeposits - adding Deposit")
			deposits = append(deposits, Deposit{
				TxID:   tx.Hash,
				Amount: tx.Value.Int().Uint64(),
			})
		}
	}

	return deposits, nil
}

// CheckEvmTxLog is a helper function used to validate receipts, logic is taken from pellclient.
func CheckEvmTxLog(vLog *ethtypes.Log, wantAddress common.Address, wantHash string, wantTopics int) error {
	if vLog.Removed {
		return fmt.Errorf("log is removed, chain reorg?")
	}
	if vLog.Address != wantAddress {
		return fmt.Errorf("log emitter address mismatch: want %s got %s", wantAddress.Hex(), vLog.Address.Hex())
	}
	if vLog.TxHash.Hex() == "" {
		return fmt.Errorf("log tx hash is empty: %d %s", vLog.BlockNumber, vLog.TxHash.Hex())
	}
	if wantHash != "" && vLog.TxHash.Hex() != wantHash {
		return fmt.Errorf("log tx hash mismatch: want %s got %s", wantHash, vLog.TxHash.Hex())
	}
	if len(vLog.Topics) != wantTopics {
		return fmt.Errorf("number of topics mismatch: want %d got %d", wantTopics, len(vLog.Topics))
	}
	return nil
}
