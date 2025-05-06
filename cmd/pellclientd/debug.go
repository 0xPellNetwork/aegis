package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/onrik/ethrpc"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	evmobserver "github.com/0xPellNetwork/aegis/relayer/chains/evm/observer"
	"github.com/0xPellNetwork/aegis/relayer/config"
	pctx "github.com/0xPellNetwork/aegis/relayer/context"
	"github.com/0xPellNetwork/aegis/relayer/keys"
	"github.com/0xPellNetwork/aegis/relayer/pellcore"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

var debugArgs = debugArguments{}

type debugArguments struct {
	pellCoreHome string
	pellNode     string
	pellChainID  string
}

func init() {
	defaultHomeDir := os.ExpandEnv("$HOME/.pellcored")

	DebugCmd().Flags().StringVar(&debugArgs.pellCoreHome, "core-home", defaultHomeDir, "peer address, e.g. /dns/tss1/tcp/6668/ipfs/16Uiu2HAmACG5DtqmQsHtXg4G2sLS65ttv84e7MrL4kapkjfmhxAp")
	DebugCmd().Flags().StringVar(&debugArgs.pellNode, "node", "46.4.15.110", "public ip address")
	DebugCmd().Flags().StringVar(&debugArgs.pellChainID, "chain-id", "ignite_7001-1", "pre-params file path")

	RootCmd.AddCommand(DebugCmd())
}

func DebugCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-inbound-ballot [inboundHash] [chainID]",
		Short: "provide txHash and chainID to get the ballot status for the txHash",
		RunE:  debugCmd,
	}
}

func debugCmd(_ *cobra.Command, args []string) error {
	cobra.ExactArgs(2)
	cfg, err := config.Load(debugArgs.pellCoreHome)
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}

	txHash := args[0]
	chainID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse chain id")
	}

	// create a new pellcore client
	client, err := pellcore.NewClient(
		zerolog.Nop(),
		&keys.Keys{OperatorAddress: sdk.MustAccAddressFromBech32(sample.AccAddress())},
		debugArgs.pellNode,
		"",
		debugArgs.pellChainID,
		false,
		config.DefaultMaxMsgLen,
	)
	if err != nil {
		return err
	}

	appContext := pctx.NewAppContext(cfg, zerolog.Nop())
	ctx := pctx.WithAppContext(context.Background(), appContext)

	if err := client.UpdateAppContext(ctx, appContext, false, zerolog.Nop()); err != nil {
		return errors.Wrap(err, "failed to update app context")
	}

	chain, exist := chains.GetChainByChainId(chainID)
	if !exist {
		return fmt.Errorf("invalid chain id")
	}

	chainParams, err := client.GetChainParams(context.Background())
	if err != nil {
		return err
	}

	var ballotIdentifier string

	if chains.IsEVMChain(chain.Id) {
		ob := evmobserver.ChainClient{}
		ob.WithPellcoreClient(client)
		var ethRPC *ethrpc.EthRPC
		var client *ethclient.Client
		for chain, evmConfig := range cfg.GetAllEVMConfigs() {
			if chainID == chain {
				ethRPC = ethrpc.NewEthRPC(evmConfig.Endpoint)
				client, err = ethclient.Dial(evmConfig.Endpoint)
				if err != nil {
					return err
				}
				chain, _ := chains.GetChainByChainId(chainID)
				ob.WithEvmClient(client)
				ob.WithEvmJSONRPC(ethRPC)
				ob.WithChain(chain)
			}
		}
		hash := ethcommon.HexToHash(txHash)
		tx, isPending, err := ob.TransactionByHash(txHash)
		if err != nil {
			return fmt.Errorf("tx not found on chain %s , %d", err.Error(), chain.Id)
		}

		if isPending {
			return fmt.Errorf("tx is still pending")
		}

		receipt, err := client.TransactionReceipt(ctx, hash)
		if err != nil {
			return fmt.Errorf("tx receipt not found on chain %s, %d", err.Error(), chain.Id)
		}

		for _, chainParams := range chainParams {
			if chainParams.ChainId == chainID {
				ob.SetChainParams(relayertypes.ChainParams{
					ChainId:                                  chainID,
					ConnectorContractAddress:                 chainParams.ConnectorContractAddress,
					DelegationManagerContractAddress:         chainParams.DelegationManagerContractAddress,
					StrategyManagerContractAddress:           chainParams.StrategyManagerContractAddress,
					OmniOperatorSharesManagerContractAddress: chainParams.OmniOperatorSharesManagerContractAddress,
				})
			}
		}

		msgs, err := ob.GetEvmReactor().CheckAndBuildInboundVoteMsg(tx, receipt, ob.LastBlock())
		if err != nil {
			return errors.Wrapf(err, "error checking and building for intx %s chain %d", tx.Hash, ob.Chain().Id)
		}
		ballotIdentifier, err = ob.PostVoteInboundMsgs(ctx, msgs)
		if err != nil {
			return errors.Wrapf(err, "error voting for intx %s chain %d", tx.Hash, ob.Chain().Id)
		}
	}

	fmt.Println("BallotIdentifier : ", ballotIdentifier)

	ballot, err := client.GetBallot(ctx, ballotIdentifier)
	if err != nil {
		return err
	}

	for _, vote := range ballot.Voters {
		fmt.Printf("%s : %s \n", vote.VoterAddress, vote.VoteType)
	}
	fmt.Println("BallotStatus : ", ballot.BallotStatus)

	return nil
}
