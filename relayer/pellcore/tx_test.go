package pellcore

import (
	"bytes"
	"context"
	"encoding/hex"
	"net"
	"os"
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"gitlab.com/thorchain/tss/go-tss/blame"
	"go.nhat.io/grpcmock"
	"go.nhat.io/grpcmock/planner"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/proofs"
	"github.com/0xPellNetwork/aegis/relayer/config"
	corecontext "github.com/0xPellNetwork/aegis/relayer/context"
	"github.com/0xPellNetwork/aegis/relayer/keys"
	"github.com/0xPellNetwork/aegis/relayer/testutils/stub"
	lightclienttypes "github.com/0xPellNetwork/aegis/x/lightclient/types"
	observertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

const (
	testSigner   = "jack"
	sampleHash   = "fa51db4412144f1130669f2bae8cb44aadbd8d85958dbffcb0fe236878097e1a"
	ethBlockHash = "1a17bcc359e84ba8ae03b17ec425f97022cd11c3e279f6bdf7a96fcffa12b366"
)

func Test_GasPriceMultiplier(t *testing.T) {
	tt := []struct {
		name       string
		chainID    int64
		multiplier float64
		fail       bool
	}{
		{
			name:       "get Ethereum multiplier",
			chainID:    1,
			multiplier: 1.2,
			fail:       false,
		},
		{
			name:       "get Goerli multiplier",
			chainID:    5,
			multiplier: 1.2,
			fail:       false,
		},
		{
			name:       "get BSC multiplier",
			chainID:    56,
			multiplier: 1.2,
			fail:       false,
		},
		{
			name:       "get BSC Testnet multiplier",
			chainID:    97,
			multiplier: 1.2,
			fail:       false,
		},
		{
			name:       "get Polygon multiplier",
			chainID:    137,
			multiplier: 1.2,
			fail:       false,
		},
		{
			name:       "get Mumbai Testnet multiplier",
			chainID:    80001,
			multiplier: 1.2,
			fail:       false,
		},
		{
			name:       "get Bitcoin multiplier",
			chainID:    8332,
			multiplier: 1.0,
			fail:       false,
		},
		{
			name:       "get Bitcoin Testnet multiplier",
			chainID:    18332,
			multiplier: 1.0,
			fail:       false,
		},
		{
			name:       "get unknown chain gas price multiplier",
			chainID:    1234,
			multiplier: 1.0,
			fail:       true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			multiplier := GasPriceMultiplier(tc.chainID)
			require.Equal(t, tc.multiplier, multiplier)
		})
	}
}

func getHeaderData(t *testing.T) proofs.HeaderData {
	var header ethtypes.Header
	file, err := os.Open("../../pkg/testdata/eth_header_18495266.json")
	require.NoError(t, err)
	defer file.Close()
	headerBytes := make([]byte, 4096)
	n, err := file.Read(headerBytes)
	require.NoError(t, err)
	err = header.UnmarshalJSON(headerBytes[:n])
	require.NoError(t, err)
	var buffer bytes.Buffer
	err = header.EncodeRLP(&buffer)
	require.NoError(t, err)
	return proofs.NewEthereumHeader(buffer.Bytes())
}

func TestPellCoreBridge_PostGasPrice(t *testing.T) {
	ctx := context.Background()

	extraGRPC := withDummyServer(100)
	setupMockServer(t, observertypes.RegisterQueryServer, skipMethod, nil, nil, extraGRPC...)

	client := setupPellcoreClient(t,
		withDefaultObserverKeys(),
	)

	t.Run("post gas price success", func(t *testing.T) {
		hash, err := client.PostGasPrice(ctx, chains.BscMainnetChain(), 1000000, "0", 1234)
		// TODO: fix it
		t.Log(hash, err)
		//require.NoError(t, err)
		//require.Equal(t, sampleHash, hash)
	})
}

// TODO: fix it
//func TestPellCoreBridge_AddTxHashToOutTxTracker(t *testing.T) {
//	ctx := context.Background()
//
//	const nonce = 123
//	chainID := chains.BscMainnetChain().Id
//
//	method := "/pellchain.pellcore.xmsg.Query/OutTxTracker"
//	input := &xmsgtypes.QueryGetOutTxTrackerRequest{
//		ChainId: chains.BscMainnetChain().Id,
//		Nonce:   nonce,
//	}
//	output := &xmsgtypes.QueryOutTxTrackerResponse{
//		OutTxTracker: xmsgtypes.OutTxTracker{
//			Index:     "456",
//			ChainId:   chainID,
//			Nonce:     nonce,
//			HashLists: nil,
//		},
//	}
//
//	extraGRPC := withDummyServer(100)
//	setupMockServer(t, observertypes.RegisterQueryServer, method, input, output, extraGRPC...)
//
//	client := setupPellcoreClient(t,
//		withDefaultObserverKeys(),
//	)
//
//	t.Run("add tx hash success", func(t *testing.T) {
//		hash, err := client.PostAddTxHashToOutTxTracker(ctx, chainID, nonce, "", nil, "", 456)
//		assert.NoError(t, err)
//		assert.Equal(t, sampleHash, hash)
//	})
//
//	t.Run("add tx hash fail", func(t *testing.T) {
//		hash, err := client.PostAddTxHashToOutTxTracker(ctx, chainID, nonce, "", nil, "", 456)
//		assert.Error(t, err)
//		assert.Empty(t, hash)
//	})
//}

func TestPellCoreBridge_SetTSS(t *testing.T) {
	ctx := context.Background()

	extraGRPC := withDummyServer(100)
	setupMockServer(t, xmsgtypes.RegisterMsgServer, skipMethod, nil, nil, extraGRPC...)

	client := setupPellcoreClient(t,
		withDefaultObserverKeys(),
	)

	t.Run("set tss success", func(t *testing.T) {
		hash, err := client.PostVoteTSS(
			ctx,
			"pellpub1addwnpepqtadxdyt037h86z60nl98t6zk56mw5zpnm79tsmvspln3hgt5phdc79kvfc",
			9987,
			chains.ReceiveStatus_SUCCESS,
		)
		// TODO: 26657 not setup
		t.Log(hash, err)
		//require.NoError(t, err)
		//require.Equal(t, sampleHash, hash)
	})
}

func TestPellCoreBridge_UpdateAppContext(t *testing.T) {
	ctx := context.Background()

	//Setup server for multiple grpc calls
	listener, err := net.Listen("tcp", "127.0.0.1:9090")
	require.NoError(t, err)

	server := grpcmock.MockUnstartedServer(
		grpcmock.RegisterService(xmsgtypes.RegisterQueryServer),
		grpcmock.RegisterService(upgradetypes.RegisterQueryServer),
		grpcmock.RegisterService(observertypes.RegisterQueryServer),
		grpcmock.RegisterService(lightclienttypes.RegisterQueryServer),
		grpcmock.WithPlanner(planner.FirstMatch()),
		grpcmock.WithListener(listener),
		func(s *grpcmock.Server) {
			method := "/xmsg.Query/LastPellHeight"
			s.ExpectUnary(method).
				UnlimitedTimes().
				WithPayload(xmsgtypes.QueryLastPellHeightRequest{}).
				Return(xmsgtypes.QueryLastPellHeightResponse{Height: 12345})

			method = "/cosmos.upgrade.v1beta1.Query/CurrentPlan"
			s.ExpectUnary(method).
				UnlimitedTimes().
				WithPayload(upgradetypes.QueryCurrentPlanRequest{}).
				Return(upgradetypes.QueryCurrentPlanResponse{
					Plan: &upgradetypes.Plan{
						Name:   "big upgrade",
						Height: 100,
					},
				})

			method = "/relayer.Query/GetChainParams"
			s.ExpectUnary(method).
				UnlimitedTimes().
				WithPayload(observertypes.QueryGetChainParamsRequest{}).
				Return(observertypes.QueryGetChainParamsResponse{ChainParams: &observertypes.ChainParamsList{
					ChainParams: []*observertypes.ChainParams{
						{
							ChainId: 86,
						},
					},
				}})

			method = "/relayer.Query/SupportedChains"
			s.ExpectUnary(method).
				UnlimitedTimes().
				WithPayload(observertypes.QuerySupportedChains{}).
				Return(observertypes.QuerySupportedChainsResponse{
					Chains: []*chains.Chain{
						{
							Id: chains.BscMainnetChain().Id,
						},
						{
							Id: chains.EthChain().Id,
						},
					},
				})

			method = "/relayer.Query/Keygen"
			s.ExpectUnary(method).
				UnlimitedTimes().
				WithPayload(observertypes.QueryGetKeygenRequest{}).
				Return(observertypes.QueryKeygenResponse{
					Keygen: &observertypes.Keygen{
						Status:         observertypes.KeygenStatus_SUCCESS,
						GranteePubkeys: nil,
						BlockNumber:    5646,
					}})

			method = "/relayer.Query/TSS"
			s.ExpectUnary(method).
				UnlimitedTimes().
				WithPayload(observertypes.QueryGetTSSRequest{}).
				Return(observertypes.QueryTSSResponse{
					Tss: observertypes.TSS{
						TssPubkey:           "pellpub1addwnpepqtadxdyt037h86z60nl98t6zk56mw5zpnm79tsmvspln3hgt5phdc79kvfc",
						TssParticipantList:  nil,
						OperatorAddressList: nil,
						FinalizedPellHeight: 1000,
						KeygenPellHeight:    900,
					},
				})

			method = "/relayer.Query/CrosschainFlags"
			s.ExpectUnary(method).
				UnlimitedTimes().
				WithPayload(observertypes.QueryGetCrosschainFlagsRequest{}).
				Return(observertypes.QueryCrosschainFlagsResponse{CrosschainFlags: observertypes.CrosschainFlags{
					IsInboundEnabled:             true,
					IsOutboundEnabled:            false,
					GasPriceIncreaseFlags:        nil,
					BlockHeaderVerificationFlags: nil,
				}})

			method = "/lightclient.Query/VerificationFlags"
			s.ExpectUnary(method).
				UnlimitedTimes().
				WithPayload(lightclienttypes.QueryVerificationFlagsRequest{}).
				Return(lightclienttypes.QueryVerificationFlagsResponse{VerificationFlags: lightclienttypes.VerificationFlags{
					EthTypeChainEnabled: true,
					BtcTypeChainEnabled: false,
				}})

		},
	)(t)

	server.Serve()
	defer server.Close()

	address := sdktypes.AccAddress(stub.TestKeyringPair.PubKey().Address().Bytes())
	client := setupPellcoreClient(t,
		withObserverKeys(keys.NewKeysWithKeybase(stub.NewKeyring(), address, testSigner, "")),
	)

	t.Run("pellcore update success", func(t *testing.T) {
		cfg := config.NewConfig()
		appContext := corecontext.NewAppContext(cfg, zerolog.Nop())
		err := client.UpdateAppContext(ctx, appContext, false, zerolog.New(zerolog.NewTestWriter(t)))
		require.NoError(t, err)
	})
}

func TestPellCoreBridge_PostBlameData(t *testing.T) {
	ctx := context.Background()

	extraGRPC := withDummyServer(100)
	setupMockServer(t, observertypes.RegisterQueryServer, skipMethod, nil, nil, extraGRPC...)

	client := setupPellcoreClient(t,
		withDefaultObserverKeys(),
	)

	t.Run("post blame data success", func(t *testing.T) {
		hash, err := client.PostBlameData(
			ctx,
			&blame.Blame{
				FailReason: "",
				IsUnicast:  false,
				BlameNodes: nil,
			},
			chains.BscMainnetChain().Id,
			"102394876-bsc",
		)
		// TODO: 26657 not setup
		t.Log(hash, err)
		//assert.NoError(t, err)
		//assert.Equal(t, sampleHash, hash)
	})
}

func TestPellCoreBridge_PostVoteBlockHeader(t *testing.T) {
	ctx := context.Background()

	extraGRPC := withDummyServer(100)
	setupMockServer(t, observertypes.RegisterQueryServer, skipMethod, nil, nil, extraGRPC...)

	client := setupPellcoreClient(t,
		withDefaultObserverKeys(),
	)

	blockHash, err := hex.DecodeString(ethBlockHash)
	require.NoError(t, err)

	t.Run("post add block header success", func(t *testing.T) {
		hash, err := client.PostVoteBlockHeader(
			ctx,
			chains.EthChain().Id,
			blockHash,
			18495266,
			getHeaderData(t),
		)
		// TODO: 26657 not setup
		t.Log(hash, err)
		//require.NoError(t, err)
		//require.Equal(t, sampleHash, hash)
	})
}

func TestPellCoreBridge_PostVoteInbound(t *testing.T) {
	ctx := context.Background()

	address := sdktypes.AccAddress(stub.TestKeyringPair.PubKey().Address().Bytes())

	expectedOutput := observertypes.QueryHasVotedResponse{HasVoted: false}
	input := observertypes.QueryHasVotedRequest{
		BallotIdentifier: "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
		VoterAddress:     address.String(),
	}
	method := "/relayer.Query/HasVoted"

	extraGRPC := withDummyServer(100)
	setupMockServer(t, observertypes.RegisterQueryServer, method, input, expectedOutput, extraGRPC...)

	client := setupPellcoreClient(t,
		withDefaultObserverKeys(),
	)

	t.Run("post inbound vote already voted", func(t *testing.T) {
		hash, _, err := client.PostVoteInboundEvents(ctx, 100, 200, []*xmsgtypes.MsgVoteOnObservedInboundTx{
			{
				Signer: address.String(),
			},
		})
		// TODO: 26657 not setup
		t.Log(hash, err)
		//require.NoError(t, err)
		//require.Equal(t, sampleHash, hash)
	})
}

func TestPellcore_GetInboundVoteMessage(t *testing.T) {
	address := sdktypes.AccAddress(stub.TestKeyringPair.PubKey().Address().Bytes())
	t.Run("get inbound vote message", func(t *testing.T) {
		msg := GetInBoundVoteMessage(
			address.String(),
			chains.EthChain().Id,
			"",
			address.String(),
			chains.PellChainMainnet().Id,
			"",
			12345,
			1000,
			address.String(),
			0,
			xmsgtypes.InboundPellEvent{
				PellData: &xmsgtypes.InboundPellEvent_StakerDelegated{
					StakerDelegated: &xmsgtypes.StakerDelegated{
						Staker:   "",
						Operator: "",
					},
				},
			})
		require.Equal(t, address.String(), msg.Signer)
	})
}

// TODO: will be timeout here
//func TestPellCoreBridge_MonitorVoteInboundTxResult(t *testing.T) {
//	ctx := context.Background()
//
//	address := sdktypes.AccAddress(stub.TestKeyringPair.PubKey().Address().Bytes())
//	client := setupPellcoreClient(t,
//		withObserverKeys(keys.NewKeysWithKeybase(stub.NewKeyring(), address, testSigner, "")),
//	)
//
//	t.Run("monitor inbound vote", func(t *testing.T) {
//		err := client.MonitorVoteInboundTxResult(ctx, sampleHash, 1000, &xmsgtypes.MsgVoteOnObservedInboundTx{
//			Signer: address.String(),
//		})
//
//		require.NoError(t, err)
//	})
//}

func TestPellCoreBridge_PostVoteOutbound(t *testing.T) {
	const (
		blockHeight = 1234
		accountNum  = 10
		accountSeq  = 10
	)

	// TODO: fix it
	//ctx := context.Background()
	//
	//address := sdktypes.AccAddress(stub.TestKeyringPair.PubKey().Address().Bytes())
	//
	//expectedOutput := observertypes.QueryHasVotedResponse{HasVoted: false}
	//input := observertypes.QueryHasVotedRequest{
	//	BallotIdentifier: "0xd2200e02c15c44031b8b52d79a6237081661ca9c81ff98c09b75185641d5dfbd",
	//	VoterAddress:     address.String(),
	//}
	//method := "/relayer.Query/HasVoted"
	//
	//extraGRPC := withDummyServer(blockHeight)
	//
	//server := setupMockServer(t, observertypes.RegisterQueryServer, method, input, expectedOutput, extraGRPC...)
	//require.NotNil(t, server)

	//client := setupPellcoreClient(t,
	//	withDefaultObserverKeys(),
	//)

	//hash, ballot, err := client.PostVoteOutbound(
	//	ctx,
	//	sampleHash,
	//	sampleHash,
	//	1234,
	//	1000,
	//	big.NewInt(100),
	//	1200,
	//	chains.ReceiveStatus_SUCCESS,
	//	"",
	//	chains.EthChain(),
	//	10001,
	//)

	// TODO: fix it
	//t.Log(hash, ballot, err)
	//assert.NoError(t, err)
	//assert.Equal(t, sampleHash, hash)
	//assert.Equal(t, "0xd2200e02c15c44031b8b52d79a6237081661ca9c81ff98c09b75185641d5dfbd", ballot)
}

func TestPellCoreBridge_MonitorVoteOutboundTxResult(t *testing.T) {
	//ctx := context.Background()
	//
	//address := sdktypes.AccAddress(stub.TestKeyringPair.PubKey().Address().Bytes())
	//client := setupPellcoreClient(t,
	//	withObserverKeys(keys.NewKeysWithKeybase(stub.NewKeyring(), address, testSigner, "")),
	//)

	//t.Run("monitor outbound vote", func(t *testing.T) {
	//	msg := &xmsgtypes.MsgVoteOnObservedOutboundTx{Signer: address.String()}
	//
	//	err := client.MonitorVoteOutboundTxResult(ctx, sampleHash, 1000, msg)
	//	assert.NoError(t, err)
	//})
}
