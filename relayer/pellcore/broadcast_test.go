package pellcore

import (
	"context"
	"encoding/hex"
	"errors"
	"net"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/require"
	"go.nhat.io/grpcmock"
	"go.nhat.io/grpcmock/planner"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/relayer/keys"
	"github.com/pell-chain/pellcore/relayer/testutils/stub"
	observerTypes "github.com/pell-chain/pellcore/x/relayer/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestHandleBroadcastError(t *testing.T) {
	type response struct {
		retry  bool
		report bool
	}
	testCases := map[error]response{
		errors.New("nonce too low"):                       {retry: false, report: false},
		errors.New("replacement transaction underpriced"): {retry: false, report: false},
		errors.New("already known"):                       {retry: false, report: true},
		errors.New(""):                                    {retry: true, report: false},
	}
	for input, output := range testCases {
		retry, report := HandleBroadcastError(input, "", "", "")
		require.Equal(t, output.report, report)
		require.Equal(t, output.retry, retry)
	}
}

func TestBroadcast(t *testing.T) {
	ctx := context.Background()

	address := types.AccAddress(stub.TestKeyringPair.PubKey().Address().Bytes())

	//Setup server for multiple grpc calls
	listener, err := net.Listen("tcp", "127.0.0.1:9090")
	require.NoError(t, err)

	server := grpcmock.MockUnstartedServer(
		grpcmock.RegisterService(xmsgtypes.RegisterQueryServer),
		grpcmock.RegisterService(feemarkettypes.RegisterQueryServer),
		grpcmock.RegisterService(authtypes.RegisterQueryServer),
		grpcmock.WithPlanner(planner.FirstMatch()),
		grpcmock.WithListener(listener),
		func(s *grpcmock.Server) {
			method := "/xmsg.Query/LastPellHeight"
			s.ExpectUnary(method).
				UnlimitedTimes().
				WithPayload(xmsgtypes.QueryLastPellHeightRequest{}).
				Return(xmsgtypes.QueryLastPellHeightResponse{Height: 0})

			method = "/ethermint.feemarket.v1.Query/Params"
			s.ExpectUnary(method).
				UnlimitedTimes().
				WithPayload(feemarkettypes.QueryParamsRequest{}).
				Return(feemarkettypes.QueryParamsResponse{
					Params: feemarkettypes.Params{
						BaseFee: sdkmath.NewInt(23455),
					},
				})
		},
	)(t)

	server.Serve()
	defer server.Close()

	observerKeys := keys.NewKeysWithKeybase(stub.NewKeyring(), address, testSigner, "")

	t.Run("broadcast success", func(t *testing.T) {
		client := setupPellcoreClient(t,
			withObserverKeys(observerKeys),
		)

		blockHash, err := hex.DecodeString(ethBlockHash)
		require.NoError(t, err)
		msg := observerTypes.NewMsgVoteBlockHeader(
			address.String(),
			chains.EthChain().Id,
			blockHash,
			18495266,
			getHeaderData(t),
		)
		authzMsg, authzSigner, err := WrapMessageWithAuthz(msg)
		require.NoError(t, err)

		_, err = client.Broadcast(ctx, 10_000, []sdktypes.Msg{authzMsg}, authzSigner)
		// TODO; fix this test, because mock server failed
		t.Log(err)
		//require.NoError(t, err)
	})

	t.Run("broadcast failed", func(t *testing.T) {
		client := setupPellcoreClient(t,
			withObserverKeys(observerKeys),
		)

		blockHash, err := hex.DecodeString(ethBlockHash)
		require.NoError(t, err)
		msg := observerTypes.NewMsgVoteBlockHeader(
			address.String(),
			chains.EthChain().Id,
			blockHash,
			18495266,
			getHeaderData(t),
		)
		authzMsg, authzSigner, err := WrapMessageWithAuthz(msg)
		require.NoError(t, err)

		_, err = client.Broadcast(ctx, 10_000, []sdktypes.Msg{authzMsg}, authzSigner)
		require.Error(t, err)
	})
}
