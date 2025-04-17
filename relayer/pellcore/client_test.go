package pellcore

import (
	"context"
	"net"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.nhat.io/grpcmock"
	"go.nhat.io/grpcmock/planner"

	"github.com/pell-chain/pellcore/cmd/pellcored/config"
	"github.com/pell-chain/pellcore/relayer/keys"
	keyinterfaces "github.com/pell-chain/pellcore/relayer/keys/interfaces"
	"github.com/pell-chain/pellcore/relayer/testutils/stub"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

const skipMethod = "skip"

// setupMockServer setup mock pellcore GRPC server
func setupMockServer(
	t *testing.T,
	serviceFunc any, method string, input any, expectedOutput any,
	extra ...grpcmock.ServerOption,
) *grpcmock.Server {
	listener, err := net.Listen("tcp", "127.0.0.1:9090")
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

func withDummyServer(pellBlockHeight int64) []grpcmock.ServerOption {
	return []grpcmock.ServerOption{
		grpcmock.RegisterService(xmsgtypes.RegisterQueryServer),
		grpcmock.RegisterService(xmsgtypes.RegisterMsgServer),
		grpcmock.RegisterService(feemarkettypes.RegisterQueryServer),
		grpcmock.RegisterService(authtypes.RegisterQueryServer),
		// grpcmock.RegisterService(abci.RegisterABCIApplicationServer),
		func(s *grpcmock.Server) {
			// Block Height
			s.ExpectUnary("/xmsg.Query/LastPellHeight").
				UnlimitedTimes().
				Return(xmsgtypes.QueryLastPellHeightResponse{Height: pellBlockHeight})

			// London Base Fee
			s.ExpectUnary("/ethermint.feemarket.v1.Query/Params").
				UnlimitedTimes().
				Return(feemarkettypes.QueryParamsResponse{
					Params: feemarkettypes.Params{BaseFee: sdkmath.NewInt(100)},
				})
		},
	}
}

type clientTestConfig struct {
	keys keyinterfaces.ObserverKeys
	opts []Opt
}

type clientTestOpt func(*clientTestConfig)

func withObserverKeys(keys keyinterfaces.ObserverKeys) clientTestOpt {
	return func(cfg *clientTestConfig) { cfg.keys = keys }
}

func withDefaultObserverKeys() clientTestOpt {
	var (
		key     = stub.TestKeyringPair
		address = types.AccAddress(key.PubKey().Address().Bytes())
		keyRing = stub.NewKeyring()
	)

	return withObserverKeys(keys.NewKeysWithKeybase(keyRing, address, testSigner, ""))
}

func setupPellcoreClient(t *testing.T, opts ...clientTestOpt) *PellCoreBridge {
	const (
		chainIP = "127.0.0.1"
		signer  = testSigner
		chainID = "pellchain_186-1"
	)

	var cfg clientTestConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.keys == nil {
		cfg.keys = &keys.Keys{}
	}

	c, err := NewClient(
		zerolog.Nop(),
		cfg.keys,
		chainIP, signer,
		chainID,
		false,
		13,
	)

	require.NoError(t, err)

	return c
}

// Need to test after refactor
func TestPellcore_GetGenesisSupply(t *testing.T) {
}

func TestPellcore_GetPellHotKeyBalance(t *testing.T) {
	ctx := context.Background()

	expectedOutput := banktypes.QueryBalanceResponse{
		Balance: &types.Coin{
			Denom:  config.BaseDenom,
			Amount: sdkmath.NewInt(55646484),
		},
	}
	input := banktypes.QueryBalanceRequest{
		Address: types.AccAddress(stub.TestKeyringPair.PubKey().Address().Bytes()).String(),
		Denom:   config.BaseDenom,
	}
	method := "/cosmos.bank.v1beta1.Query/Balance"
	setupMockServer(t, banktypes.RegisterQueryServer, method, input, expectedOutput)

	client := setupPellcoreClient(t, withDefaultObserverKeys())

	// should be able to get balance of signer
	client.keys = keys.NewKeysWithKeybase(stub.NewKeyring(), types.AccAddress{}, "bob", "")
	resp, err := client.GetPellHotKeyBalance(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedOutput.Balance.Amount, resp)

	// should return error on empty signer
	// client.keys = keys.NewKeysWithKeybase(stub.NewKeyring(), types.AccAddress{}, "", "")
	// resp, err = client.GetPellHotKeyBalance(ctx)
	// require.Error(t, err)
	// require.Equal(t, types.ZeroInt(), resp)
}
