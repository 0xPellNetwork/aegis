package rpc

import (
	"context"
	"net"
	"testing"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/require"
	"go.nhat.io/grpcmock"
	"go.nhat.io/grpcmock/planner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/0xPellNetwork/aegis/cmd/pellcored/config"
	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/relayer/chains/interfaces"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	lightclienttypes "github.com/0xPellNetwork/aegis/x/lightclient/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

const skipMethod = "skip"

const gRPCListenPath = "127.0.0.1:47392"

// setupMockServer setup mock pellcore GRPC server
func setupMockServer(
	t *testing.T,
	serviceFunc any,
	method string,
	input any,
	expectedOutput any,
	extra ...grpcmock.ServerOption,
) *grpcmock.Server {
	listener, err := net.Listen("tcp", gRPCListenPath)
	require.NoError(t, err)

	opts := []grpcmock.ServerOption{
		grpcmock.RegisterService(serviceFunc),
		grpcmock.WithPlanner(planner.FirstMatch()),
		grpcmock.WithListener(listener),
	}

	opts = append(opts, extra...)

	if method != skipMethod {
		opts = append(opts, func(s *grpcmock.Server) {
			s.ExpectUnary(method).
				UnlimitedTimes().
				WithPayload(input).
				Return(expectedOutput)
		})
	}

	server := grpcmock.MockUnstartedServer(opts...)(t)

	server.Serve()

	t.Cleanup(func() {
		require.NoError(t, server.Close())
	})

	return server
}

func setupPellcoreClients(t *testing.T) Clients {
	c, err := NewGRPCClients(gRPCListenPath, grpc.WithTransportCredentials(insecure.NewCredentials()))

	require.NoError(t, err)

	return c
}

func TestPellCoreBridge_GetBallot(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryBallotByIdentifierResponse{
		BallotIdentifier: "123",
		Voters:           nil,
		ObservationType:  0,
		BallotStatus:     0,
	}
	input := relayertypes.QueryBallotByIdentifierRequest{BallotIdentifier: "123"}
	method := "/relayer.Query/BallotByIdentifier"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetBallotByID(ctx, "123")
	require.NoError(t, err)
	require.Equal(t, expectedOutput, *resp)
}

func TestPellCoreBridge_GetCrosschainFlags(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryCrosschainFlagsResponse{CrosschainFlags: relayertypes.CrosschainFlags{
		IsInboundEnabled:             true,
		IsOutboundEnabled:            false,
		GasPriceIncreaseFlags:        nil,
		BlockHeaderVerificationFlags: nil,
	}}
	input := relayertypes.QueryGetCrosschainFlagsRequest{}
	method := "/relayer.Query/CrosschainFlags"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetCrosschainFlags(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.CrosschainFlags, resp)
}

func TestPellcore_GetRateLimiterFlags(t *testing.T) {
	ctx := context.Background()

	// create sample flags
	rateLimiterFlags := sample.RateLimiterFlags()
	expectedOutput := xmsgtypes.QueryRateLimiterFlagsResponse{
		RateLimiterFlags: rateLimiterFlags,
	}

	// setup mock server
	input := xmsgtypes.QueryRateLimiterFlagsRequest{}
	method := "/xmsg.Query/RateLimiterFlags"
	setupMockServer(t, xmsgtypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	// query
	resp, err := client.GetRateLimiterFlags(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.RateLimiterFlags, resp)
}

func TestPellCoreBridge_GetVerificationFlags(t *testing.T) {
	ctx := context.Background()

	expectedOutput := lightclienttypes.QueryVerificationFlagsResponse{VerificationFlags: lightclienttypes.VerificationFlags{
		EthTypeChainEnabled: true,
		BtcTypeChainEnabled: false,
	}}
	input := lightclienttypes.QueryVerificationFlagsRequest{}
	method := "/lightclient.Query/VerificationFlags"
	setupMockServer(t, lightclienttypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetVerificationFlags(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.VerificationFlags, resp)
}

func TestPellCoreBridge_GetChainParamsForChainID(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryGetChainParamsForChainResponse{ChainParams: &relayertypes.ChainParams{
		ChainId:               123,
		BallotThreshold:       sdkmath.LegacyZeroDec(),
		MinObserverDelegation: sdkmath.LegacyZeroDec(),
	}}
	input := relayertypes.QueryGetChainParamsForChainRequest{ChainId: 123}
	method := "/relayer.Query/GetChainParamsForChain"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetChainParamsForChainID(ctx, 123)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.ChainParams, resp)
}

func TestPellCoreBridge_GetChainParams(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryGetChainParamsResponse{ChainParams: &relayertypes.ChainParamsList{
		ChainParams: []*relayertypes.ChainParams{
			{
				ChainId:               123,
				BallotThreshold:       sdkmath.LegacyZeroDec(),
				MinObserverDelegation: sdkmath.LegacyZeroDec(),
			},
		},
	}}
	input := relayertypes.QueryGetChainParamsRequest{}
	method := "/relayer.Query/GetChainParams"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetChainParams(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.ChainParams.ChainParams, resp)
}

func TestPellCoreBridge_GetUpgradePlan(t *testing.T) {
	ctx := context.Background()

	expectedOutput := upgradetypes.QueryCurrentPlanResponse{
		Plan: &upgradetypes.Plan{
			Name:   "big upgrade",
			Height: 100,
		},
	}
	input := upgradetypes.QueryCurrentPlanRequest{}
	method := "/cosmos.upgrade.v1beta1.Query/CurrentPlan"
	setupMockServer(t, upgradetypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetUpgradePlan(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.Plan, resp)
}

func TestPellCoreBridge_GetAllXmsg(t *testing.T) {
	ctx := context.Background()

	expectedOutput := xmsgtypes.QueryXmsgAllResponse{
		Xmsgs: []*xmsgtypes.Xmsg{
			{
				Index: "cross-chain4456",
			},
		},
		Pagination: nil,
	}
	input := xmsgtypes.QueryAllXmsgRequest{}
	method := "/xmsg.Query/XmsgAll"
	setupMockServer(t, xmsgtypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetAllXmsg(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.Xmsgs, resp)
}

func TestPellCoreBridge_GetXmsgByHash(t *testing.T) {
	ctx := context.Background()

	expectedOutput := xmsgtypes.QueryXmsgResponse{Xmsg: &xmsgtypes.Xmsg{
		Index: "9c8d02b6956b9c78ecb6090a8160faaa48e7aecfd0026fcdf533721d861436a3",
	}}
	input := xmsgtypes.QueryGetXmsgRequest{Index: "9c8d02b6956b9c78ecb6090a8160faaa48e7aecfd0026fcdf533721d861436a3"}
	method := "/xmsg.Query/Xmsg"
	setupMockServer(t, xmsgtypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetXmsgByHash(ctx, "9c8d02b6956b9c78ecb6090a8160faaa48e7aecfd0026fcdf533721d861436a3")
	require.NoError(t, err)
	require.Equal(t, expectedOutput.Xmsg, resp)
}

func TestPellCoreBridge_GetXmsgByNonce(t *testing.T) {
	ctx := context.Background()

	expectedOutput := xmsgtypes.QueryXmsgByNonceResponse{Xmsg: &xmsgtypes.Xmsg{
		Index: "9c8d02b6956b9c78ecb6090a8160faaa48e7aecfd0026fcdf533721d861436a3",
	}}
	input := xmsgtypes.QueryGetXmsgByNonceRequest{
		ChainId: 86,
		Nonce:   55,
	}
	method := "/xmsg.Query/XmsgByNonce"
	setupMockServer(t, xmsgtypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetXmsgByNonce(ctx, 86, 55)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.Xmsg, resp)
}

func TestPellCoreBridge_GetRelayerList(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryObserverSetResponse{
		Observers: []string{
			"pell1p6qlkxfss34lqwggak6nknygaaaxswvtjruevc",
			"pell1e8z7yggs7zd48mc7qhncg285awzfxckutdasus",
			"pell1pu5xy7wnwt7ukvt4yvvkldshhh0lhq6q6rhhxp",
		},
	}
	input := relayertypes.QueryObserverSet{}
	method := "/relayer.Query/ObserverSet"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetObserverList(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.Observers, resp)
}

func TestPellcore_GetRateLimiterInput(t *testing.T) {
	ctx := context.Background()

	expectedOutput := &xmsgtypes.QueryRateLimiterInputResponse{
		Height:                  10,
		XmsgsMissed:             []*xmsgtypes.Xmsg{sample.Xmsg_pell(t, "1-1")},
		XmsgsPending:            []*xmsgtypes.Xmsg{sample.Xmsg_pell(t, "1-2")},
		TotalPending:            1,
		LowestPendingXmsgHeight: 2,
	}
	input := xmsgtypes.QueryRateLimiterInputRequest{Window: 10}
	method := "/xmsg.Query/RateLimiterInput"
	setupMockServer(t, xmsgtypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetRateLimiterInput(ctx, 10)
	require.NoError(t, err)
	require.Equal(t, expectedOutput, resp)
}

func TestPellCoreBridge_ListPendingXmsg(t *testing.T) {
	ctx := context.Background()

	expectedOutput := xmsgtypes.QueryListPendingXmsgResponse{
		Xmsg: []*xmsgtypes.Xmsg{
			{
				Index: "cross-chain4456",
			},
		},
		TotalPending: 1,
	}
	input := xmsgtypes.QueryListPendingXmsgRequest{ChainId: 86}
	method := "/xmsg.Query/ListPendingXmsg"
	setupMockServer(t, xmsgtypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, totalPending, err := client.ListPendingXmsg(ctx, 86)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.Xmsg, resp)
	require.Equal(t, expectedOutput.TotalPending, totalPending)
}

// Need to test after refactor
func TestPellCoreBridge_GetGenesisSupply(t *testing.T) {
}

func TestPellCoreBridge_GetPellTokenSupplyOnNode(t *testing.T) {
	ctx := context.Background()

	expectedOutput := banktypes.QuerySupplyOfResponse{
		Amount: types.Coin{
			Denom:  config.BaseDenom,
			Amount: sdkmath.NewInt(329438),
		}}
	input := banktypes.QuerySupplyOfRequest{Denom: config.BaseDenom}
	method := "/cosmos.bank.v1beta1.Query/SupplyOf"
	setupMockServer(t, banktypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetPellTokenSupplyOnNode(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.GetAmount().Amount, resp)
}

func TestPellcore_GetBlockHeight(t *testing.T) {
	ctx := context.Background()

	method := "/xmsg.Query/LastPellHeight"
	input := &xmsgtypes.QueryLastPellHeightRequest{}
	output := &xmsgtypes.QueryLastPellHeightResponse{Height: 12345}

	setupMockServer(t, xmsgtypes.RegisterQueryServer, method, input, output)

	client := setupPellcoreClients(t)

	t.Run("last block height", func(t *testing.T) {
		height, err := client.GetBlockHeight(ctx)
		require.NoError(t, err)
		require.Equal(t, int64(12345), height)
	})
}

func TestPellcore_GetLatestPellBlock(t *testing.T) {
	ctx := context.Background()

	expectedOutput := cmtservice.GetLatestBlockResponse{
		SdkBlock: &cmtservice.Block{
			Header:     cmtservice.Header{},
			Data:       tmtypes.Data{},
			Evidence:   tmtypes.EvidenceList{},
			LastCommit: nil,
		},
	}
	input := cmtservice.GetLatestBlockRequest{}
	method := "/cosmos.base.tendermint.v1beta1.Service/GetLatestBlock"
	setupMockServer(t, cmtservice.RegisterServiceServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetLatestPellBlock(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.SdkBlock, resp)
}

func TestPellCoreBridge_GetNodeInfo(t *testing.T) {
	ctx := context.Background()

	expectedOutput := cmtservice.GetNodeInfoResponse{
		DefaultNodeInfo:    nil,
		ApplicationVersion: &cmtservice.VersionInfo{},
	}
	input := cmtservice.GetNodeInfoRequest{}
	method := "/cosmos.base.tendermint.v1beta1.Service/GetNodeInfo"
	setupMockServer(t, cmtservice.RegisterServiceServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetNodeInfo(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput, *resp)
}

func TestPellCoreBridge_GetBaseGasPrice(t *testing.T) {
	ctx := context.Background()

	expectedOutput := feemarkettypes.QueryParamsResponse{
		Params: feemarkettypes.Params{
			BaseFee: sdkmath.NewInt(23455),
		},
	}
	input := feemarkettypes.QueryParamsRequest{}
	method := "/ethermint.feemarket.v1.Query/Params"
	setupMockServer(t, feemarkettypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetFeemarketParams(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.Params.BaseFee.Int64(), resp.BaseFee.Int64())
}

func TestPellCoreBridge_GetNonceByChain(t *testing.T) {
	ctx := context.Background()

	chain := chains.BscMainnetChain()
	expectedOutput := relayertypes.QueryChainNoncesResponse{
		ChainNonces: relayertypes.ChainNonces{
			Signer:          "",
			Index:           "",
			ChainId:         chain.Id,
			Nonce:           8446,
			Signers:         nil,
			FinalizedHeight: 0,
		},
	}
	input := relayertypes.QueryGetChainNoncesRequest{Index: chain.ChainName()}
	method := "/relayer.Query/ChainNonces"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetNonceByChain(ctx, chain)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.ChainNonces, resp)
}

func TestPellCoreBridge_GetAllNodeAccounts(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryNodeAccountAllResponse{
		NodeAccount: []*relayertypes.NodeAccount{
			{
				Operator:       "pell1p6qlkxfss34lqwggak6nknygaaaxswvtjruevc",
				GranteeAddress: "pell1e8z7yggs7zd48mc7qhncg285awzfxckutdasus",
				GranteePubkey:  nil,
				NodeStatus:     0,
			},
		},
	}
	input := relayertypes.QueryAllNodeAccountRequest{}
	method := "/relayer.Query/NodeAccountAll"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetAllNodeAccounts(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.NodeAccount, resp)
}

func TestPellCoreBridge_GetKeyGen(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryKeygenResponse{
		Keygen: &relayertypes.Keygen{
			Status:         relayertypes.KeygenStatus_SUCCESS,
			GranteePubkeys: nil,
			BlockNumber:    5646,
		}}
	input := relayertypes.QueryGetKeygenRequest{}
	method := "/relayer.Query/Keygen"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetKeyGen(ctx)
	require.NoError(t, err)
	require.Equal(t, *expectedOutput.Keygen, resp)
}

func TestPellCoreBridge_GetBallotByID(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryBallotByIdentifierResponse{
		BallotIdentifier: "ballot1235",
	}
	input := relayertypes.QueryBallotByIdentifierRequest{BallotIdentifier: "ballot1235"}
	method := "/relayer.Query/BallotByIdentifier"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetBallot(ctx, "ballot1235")
	require.NoError(t, err)
	require.Equal(t, expectedOutput, *resp)
}

func TestPellCoreBridge_GetInboundTrackersForChain(t *testing.T) {
	ctx := context.Background()

	chainID := chains.BscMainnetChain().Id
	expectedOutput := xmsgtypes.QueryInTxTrackerAllByChainResponse{
		InTxTrackers: []xmsgtypes.InTxTracker{
			{
				ChainId: chainID,
				TxHash:  "DC76A6DCCC3AA62E89E69042ADC44557C50D59E4D3210C37D78DC8AE49B3B27F",
			},
		},
	}
	input := xmsgtypes.QueryAllInTxTrackerByChainRequest{ChainId: chainID}
	method := "/xmsg.Query/InTxTrackerAllByChain"
	setupMockServer(t, xmsgtypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetInboundTrackersForChain(ctx, chainID)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.InTxTrackers, resp)
}

func TestPellCoreBridge_GetCurrentTss(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryTSSResponse{
		Tss: relayertypes.TSS{
			TssPubkey:           "pellpub1addwnpepqtadxdyt037h86z60nl98t6zk56mw5zpnm79tsmvspln3hgt5phdc79kvfc",
			TssParticipantList:  nil,
			OperatorAddressList: nil,
			FinalizedPellHeight: 1000,
			KeygenPellHeight:    900,
		},
	}
	input := relayertypes.QueryGetTSSRequest{}
	method := "/relayer.Query/TSS"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetCurrentTss(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.Tss, resp)
}

func TestPellCoreBridge_GetEthTssAddress(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryGetTssAddressResponse{
		Eth: "0x70e967acfcc17c3941e87562161406d41676fd83",
	}
	input := relayertypes.QueryGetTssAddressRequest{}
	method := "/relayer.Query/GetTssAddress"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetEVMTSSAddress(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.Eth, resp)
}

func TestPellCoreBridge_GetTssHistory(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryTssHistoryResponse{
		TssList: []relayertypes.TSS{
			{
				TssPubkey:           "pellpub1addwnpepqtadxdyt037h86z60nl98t6zk56mw5zpnm79tsmvspln3hgt5phdc79kvfc",
				TssParticipantList:  nil,
				OperatorAddressList: nil,
				FinalizedPellHeight: 46546,
				KeygenPellHeight:    6897,
			},
		},
	}
	input := relayertypes.QueryTssHistoryRequest{}
	method := "/relayer.Query/TssHistory"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetTssHistory(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.TssList, resp)
}

func TestPellCoreBridge_GetOutTxTracker(t *testing.T) {
	ctx := context.Background()

	chain := chains.BscMainnetChain()
	expectedOutput := xmsgtypes.QueryOutTxTrackerResponse{
		OutTxTracker: xmsgtypes.OutTxTracker{
			Index:     "tracker12345",
			ChainId:   chain.Id,
			Nonce:     456,
			HashLists: nil,
		},
	}
	input := xmsgtypes.QueryGetOutTxTrackerRequest{
		ChainId: chain.Id,
		Nonce:   456,
	}
	method := "/xmsg.Query/OutTxTracker"
	setupMockServer(t, xmsgtypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetOutTxTracker(ctx, chain, 456)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.OutTxTracker, *resp)
}

func TestPellCoreBridge_GetAllOutTxTrackerByChain(t *testing.T) {
	ctx := context.Background()

	chain := chains.BscMainnetChain()
	expectedOutput := xmsgtypes.QueryOutTxTrackerAllByChainResponse{
		OutTxTrackers: []xmsgtypes.OutTxTracker{
			{
				Index:     "tracker23456",
				ChainId:   chain.Id,
				Nonce:     123456,
				HashLists: nil,
			},
		},
	}
	input := xmsgtypes.QueryAllOutTxTrackerByChainRequest{
		Chain: chain.Id,
		Pagination: &query.PageRequest{
			Key:        nil,
			Offset:     0,
			Limit:      2000,
			CountTotal: false,
			Reverse:    false,
		},
	}
	method := "/xmsg.Query/OutTxTrackerAllByChain"
	setupMockServer(t, xmsgtypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetAllOutTxTrackerByChain(ctx, chain.Id, interfaces.Ascending)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.OutTxTrackers, resp)

	resp, err = client.GetAllOutTxTrackerByChain(ctx, chain.Id, interfaces.Descending)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.OutTxTrackers, resp)
}

func TestPellCoreBridge_GetPendingNoncesByChain(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryPendingNoncesByChainResponse{
		PendingNonces: relayertypes.PendingNonces{
			NonceLow:  0,
			NonceHigh: 0,
			ChainId:   chains.EthChain().Id,
			Tss:       "",
		},
	}
	input := relayertypes.QueryPendingNoncesByChainRequest{ChainId: chains.EthChain().Id}
	method := "/relayer.Query/PendingNoncesByChain"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetPendingNoncesByChain(ctx, chains.EthChain().Id)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.PendingNonces, resp)
}

func TestPellCoreBridge_GetBlockHeaderChainState(t *testing.T) {
	ctx := context.Background()

	chainID := chains.BscMainnetChain().Id
	expectedOutput := lightclienttypes.QueryChainStateResponse{ChainState: &lightclienttypes.ChainState{
		ChainId:         chainID,
		LatestHeight:    5566654,
		EarliestHeight:  4454445,
		LatestBlockHash: nil,
	}}
	input := lightclienttypes.QueryGetChainStateRequest{ChainId: chainID}
	method := "/lightclient.Query/ChainState"
	setupMockServer(t, lightclienttypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetBlockHeaderChainState(ctx, chainID)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.ChainState, resp.ChainState)
}

func TestPellCoreBridge_GetSupportedChains(t *testing.T) {
	ctx := context.Background()

	bscMainnet := chains.BscMainnetChain()
	ethMainnet := chains.EthChain()
	expectedOutput := relayertypes.QuerySupportedChainsResponse{
		Chains: []*chains.Chain{
			&bscMainnet, &ethMainnet,
		},
	}
	input := relayertypes.QuerySupportedChains{}
	method := "/relayer.Query/SupportedChains"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetSupportedChains(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.Chains, resp)
}

func TestPellCoreBridge_GetPendingNonces(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryPendingNoncesAllResponse{
		PendingNonces: []relayertypes.PendingNonces{
			{
				NonceLow:  225,
				NonceHigh: 226,
				ChainId:   8332,
				Tss:       "bc1qm24wp577nk8aacckv8np465z3dvmu7ry45el6y",
			},
		},
	}
	input := relayertypes.QueryAllPendingNoncesRequest{}
	method := "/relayer.Query/PendingNoncesAll"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.GetPendingNonces(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput, *resp)
}

func TestPellCoreBridge_Prove(t *testing.T) {
	ctx := context.Background()

	chainId := chains.BscMainnetChain().Id
	txHash := "9c8d02b6956b9c78ecb6090a8160faaa48e7aecfd0026fcdf533721d861436a3"
	blockHash := "0000000000000000000172c9a64f86f208b867a84dc7a0b7c75be51e750ed8eb"
	txIndex := 555
	expectedOutput := lightclienttypes.QueryProveResponse{
		Valid: true,
	}
	input := lightclienttypes.QueryProveRequest{
		ChainId:   chainId,
		TxHash:    txHash,
		Proof:     nil,
		BlockHash: blockHash,
		TxIndex:   int64(txIndex),
	}
	method := "/lightclient.Query/Prove"
	setupMockServer(t, lightclienttypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.Prove(ctx, blockHash, txHash, int64(txIndex), nil, chainId)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.Valid, resp)
}

func TestPellCoreBridge_HasVoted(t *testing.T) {
	ctx := context.Background()

	expectedOutput := relayertypes.QueryHasVotedResponse{HasVoted: true}
	input := relayertypes.QueryHasVotedRequest{
		BallotIdentifier: "123456asdf",
		VoterAddress:     "pell1pu5xy7wnwt7ukvt4yvvkldshhh0lhq6q6rhhxp",
	}
	method := "/relayer.Query/HasVoted"
	setupMockServer(t, relayertypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClients(t)

	resp, err := client.HasVoted(ctx, "123456asdf", "pell1pu5xy7wnwt7ukvt4yvvkldshhh0lhq6q6rhhxp")
	require.NoError(t, err)
	require.Equal(t, expectedOutput.HasVoted, resp)
}
