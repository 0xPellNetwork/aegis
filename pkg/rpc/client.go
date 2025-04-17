package rpc

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	rpcclient "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"google.golang.org/grpc"

	etherminttypes "github.com/pell-chain/pellcore/rpc/types"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	lightclienttypes "github.com/pell-chain/pellcore/x/lightclient/types"
	pevmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	relayertypes "github.com/pell-chain/pellcore/x/relayer/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

// Clients contains RPC client interfaces to interact with PellCore
//
// Clients also has some high level wrappers for the clients
type Clients struct {
	// Cosmos SDK clients

	// Auth is a github.com/cosmos/cosmos-sdk/x/auth/types QueryClient
	Auth authtypes.QueryClient
	// Bank is a github.com/cosmos/cosmos-sdk/x/bank/types QueryClient
	Bank banktypes.QueryClient
	// Bank is a github.com/cosmos/cosmos-sdk/x/staking/types QueryClient
	Staking stakingtypes.QueryClient
	// Upgrade is a cosmossdk.io/x/upgrade/types QueryClient
	Upgrade upgradetypes.QueryClient

	// PellCore specific clients

	// Authority is a github.com/pell-chain/pellcore/x/authority/types QueryClient
	Authority authoritytypes.QueryClient
	// Xmsg is a github.com/pell-chain/pellcore/x/xmsg/types QueryClient
	Xmsg xmsgtypes.QueryClient
	// Pevm is a github.com/pell-chain/pellcore/x/pevm/types QueryClient
	Pevm pevmtypes.QueryClient
	// Relayer is a github.com/pell-chain/pellcore/x/relayer/types QueryClient
	Relayer relayertypes.QueryClient
	// Lightclient is a github.com/pell-chain/pellcore/x/lightclient/types QueryClient
	Lightclient lightclienttypes.QueryClient

	// Ethermint specific clients

	// Ethermint is a github.com/pell-chain/pellcore/rpc/types QueryClient
	Ethermint *etherminttypes.QueryClient
	// EthermintFeeMarket is a github.com/pell-chain/ethermint/x/feemarket/types QueryClient
	EthermintFeeMarket feemarkettypes.QueryClient

	// Tendermint specific clients

	// Tendermint is a github.com/cosmos/cosmos-sdk/client/grpc/cmtservice QueryClient
	Tendermint cmtservice.ServiceClient
}

func newClients(ctx client.Context) (Clients, error) {
	return Clients{
		// Cosmos SDK clients
		Auth:      authtypes.NewQueryClient(ctx),
		Bank:      banktypes.NewQueryClient(ctx),
		Staking:   stakingtypes.NewQueryClient(ctx),
		Upgrade:   upgradetypes.NewQueryClient(ctx),
		Authority: authoritytypes.NewQueryClient(ctx),

		// PellCore specific clients
		Xmsg:        xmsgtypes.NewQueryClient(ctx),
		Pevm:        pevmtypes.NewQueryClient(ctx),
		Relayer:     relayertypes.NewQueryClient(ctx),
		Lightclient: lightclienttypes.NewQueryClient(ctx),

		// Ethermint specific clients
		Ethermint:          etherminttypes.NewQueryClient(ctx),
		EthermintFeeMarket: feemarkettypes.NewQueryClient(ctx),

		// Tendermint specific clients
		Tendermint: cmtservice.NewServiceClient(ctx),
	}, nil
}

// NewCometBFTClients creates a Clients which uses cometbft abci_query as the transport
func NewCometBFTClients(url string) (Clients, error) {
	cometRPCClient, err := rpcclient.New(url, "/websocket")
	if err != nil {
		return Clients{}, fmt.Errorf("create cometbft rpc client: %w", err)
	}
	clientCtx := client.Context{}.WithClient(cometRPCClient)

	return newClients(clientCtx)
}

// NewGRPCClient creates a Clients which uses gRPC as the transport
func NewGRPCClients(url string, opts ...grpc.DialOption) (Clients, error) {
	grpcConn, err := grpc.Dial(url, opts...)
	if err != nil {
		return Clients{}, err
	}
	clientCtx := client.Context{}.WithGRPCClient(grpcConn)
	return newClients(clientCtx)
}
