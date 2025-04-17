package runner

import (
	"context"
	"sync"
	"time"

	tmservice "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/bridge/gatewaypevm.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/connector/pellconnector.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/dvsdirectory.sol"
	pellDelegationManager "github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	pellStrategyManager "github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pellstrategymanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouterfactory.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/bridge/gatewayevm.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/omnioperatorsharesmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/swap/gasswapevm.sol"
	tssManager "github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/tssmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/strategymanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/centralscheduler.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/ejectionmanager.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/operatorstakemanager.sol"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"google.golang.org/grpc"

	"github.com/pell-chain/pellcore/e2e/config"
	"github.com/pell-chain/pellcore/e2e/contracts/erc20"
	"github.com/pell-chain/pellcore/e2e/txserver"
	"github.com/pell-chain/pellcore/pkg/chains"
	lightclienttypes "github.com/pell-chain/pellcore/x/lightclient/types"
	pevmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	relayertypes "github.com/pell-chain/pellcore/x/relayer/types"
	restakingtypes "github.com/pell-chain/pellcore/x/restaking/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
	xsecuritytypes "github.com/pell-chain/pellcore/x/xsecurity/types"
)

const (
	ReceiptTimeout = 20 * time.Second
	LOG_NAME       = "pell-e2e"
)

type Runner struct {
	PellChainId int64

	DeployerAddress       ethcommon.Address
	DeployerPrivateKey    string
	TSSAddress            ethcommon.Address
	FungibleAdminMnemonic string

	// pell
	PEVMClient        *ethclient.Client
	PellClients       PellClients
	TxServer          txserver.PellTxServer
	TxServerPellCore0 txserver.PellTxServer
	PellContracts     PellContracts
	PEVMAuth          *bind.TransactOpts
	// multi evm
	MultiEVM map[int64]*EVMChain

	Logger *Logger

	Ctx            context.Context
	CtxCancel      context.CancelFunc
	mutex          *sync.Mutex
	ReceiptTimeout time.Duration
}

type EVMChain struct {
	EVMClient     *ethclient.Client
	EVMAuth       *bind.TransactOpts
	EvmContracts  EvmContracts
	E2ETestConfig config.E2ETestConfig
}

type PellClients struct {
	XmsgClient        xmsgtypes.QueryClient
	FungibleClient    pevmtypes.QueryClient
	AuthClient        authtypes.QueryClient
	BankClient        banktypes.QueryClient
	RelayerClient     relayertypes.QueryClient
	StakingClient     stakingtypes.QueryClient
	LightclientClient lightclienttypes.QueryClient
	RestakingClient   restakingtypes.QueryClient
	XSecurityClient   xsecuritytypes.QueryClient
	RPCClientURL      string
}

// pevm contracts
type PellContracts struct {
	PellDelegationManager         *pellDelegationManager.PellDelegationManager
	PellDelegationManagerAddr     ethcommon.Address
	PellStrategyManager           *pellStrategyManager.PellStrategyManager
	PellStrategyManagerAddr       ethcommon.Address
	Gateway                       *gatewaypevm.GatewayPEVM
	GatewayAddr                   ethcommon.Address
	PellRegistryRouterFactory     *registryrouterfactory.RegistryRouterFactory
	PellRegistryRouterFactoryAddr ethcommon.Address
	PellDvsDirectory              *dvsdirectory.DVSDirectory
	PellDvsDirectoryAddr          ethcommon.Address
	PellRegistryRouter            *registryrouter.RegistryRouter
	PellRegistryRouterAddr        ethcommon.Address
}

// evm contracts
type EvmContracts struct {
	AdminTransact *bind.TransactOpts
	// evm related contract
	STERC20Addr                   ethcommon.Address
	STERC20                       *erc20.ERC20
	StrategyManagerAddr           ethcommon.Address
	StrategyManager               *strategymanager.StrategyManager
	StrategyAddr                  ethcommon.Address
	DelegationManagerAddr         ethcommon.Address
	DelegationManager             *delegationmanager.DelegationManager
	OmniOperatorSharesManagerAddr ethcommon.Address
	OmniOperatorSharesManager     *omnioperatorsharesmanager.OmniOperatorSharesManager
	TssManagerAddr                ethcommon.Address
	TssManager                    *tssManager.TSSManager
	EvmPellConnectorAddr          ethcommon.Address
	EvmPellConnector              *pellconnector.PellConnector
	PellTokenContractAddr         ethcommon.Address
	PellTokenContract             *erc20.ERC20
	GatewayContractAddr           ethcommon.Address
	GatewayContract               *gatewayevm.GatewayEVM
	GasSwapContractAddr           ethcommon.Address
	GasSwapContract               *gasswapevm.GasSwapEVM
	CentralSchedulerAddr          ethcommon.Address
	CentralScheduler              *centralscheduler.CentralScheduler
	StakeManagerAddr              ethcommon.Address
	StakeManager                  *operatorstakemanager.OperatorStakeManager
	EjectionManagerAddr           ethcommon.Address
	EjectionManager               *ejectionmanager.EjectionManager
	RegistryRouterAddr            ethcommon.Address
	RegistryRouter                *registryrouter.RegistryRouter
	StakeRegistryRouterAddr       ethcommon.Address
	StakeRegistryRouter           *stakeregistryrouter.StakeRegistryRouter
}

func NewFromConfig(conf *config.PellConfig) *Runner {
	r := new(Runner)
	r.FungibleAdminMnemonic = conf.Accounts.FungibleAdminMnemonic
	r.ReceiptTimeout = ReceiptTimeout
	r.Logger = NewLogger(true, color.BgHiBlack, LOG_NAME)
	r.mutex = new(sync.Mutex)
	r.DeployerPrivateKey = conf.Accounts.DeployerPrivKey
	r.DeployerAddress = ethcommon.HexToAddress(conf.Accounts.DeployerAddress)

	r.Ctx = context.Background()

	if err := r.fillClients(conf); err != nil {
		panic(err)
	}

	if err := r.fillTxServer(conf.Rpcs.PellCoreRpc, r.FungibleAdminMnemonic, conf.PellChainId); err != nil {
		panic(err)
	}

	if err := r.fillTssAddr(); err != nil {
		panic(err)
	}

	if err := r.fillEvmContracts(conf); err != nil {
		panic(err)
	}

	if err := r.fillPellContracts(conf); err != nil {
		panic(err)
	}

	return r
}

func (r *Runner) Lock() {
	r.mutex.Lock()
}

func (r *Runner) Unlock() {
	r.mutex.Unlock()
}

func (r *Runner) fillEvmContracts(conf *config.PellConfig) error {
	var err error

	r.MultiEVM = make(map[int64]*EVMChain)

	for _, chain := range conf.MultiChain {
		evm := new(EVMChain)
		evm.EVMClient, err = ethclient.Dial(chain.Rpc)
		if err != nil {
			return err
		}

		evm.EvmContracts.EvmPellConnectorAddr = ethcommon.HexToAddress(chain.Connector)
		evm.EvmContracts.EvmPellConnector, err = pellconnector.NewPellConnector(evm.EvmContracts.EvmPellConnectorAddr, evm.EVMClient)
		if err != nil {
			return err
		}

		evm.EvmContracts.STERC20Addr = ethcommon.HexToAddress(chain.StERC20Addr)
		evm.EvmContracts.STERC20, err = erc20.NewERC20(evm.EvmContracts.STERC20Addr, evm.EVMClient)
		if err != nil {
			return err
		}

		evm.EvmContracts.StrategyManagerAddr = ethcommon.HexToAddress(chain.StrategyManagerAddr)
		evm.EvmContracts.StrategyManager, err = strategymanager.NewStrategyManager(evm.EvmContracts.StrategyManagerAddr, evm.EVMClient)
		if err != nil {
			return err
		}

		evm.EvmContracts.StrategyAddr = ethcommon.HexToAddress(chain.StrategyAddr)

		evm.EvmContracts.DelegationManagerAddr = ethcommon.HexToAddress(chain.DelegationManagerAddr)
		evm.EvmContracts.DelegationManager, err = delegationmanager.NewDelegationManager(evm.EvmContracts.DelegationManagerAddr, evm.EVMClient)
		if err != nil {
			return nil
		}

		evm.EvmContracts.OmniOperatorSharesManagerAddr = ethcommon.HexToAddress(chain.OmniOperatorSharesManagerAddr)
		evm.EvmContracts.OmniOperatorSharesManager, err = omnioperatorsharesmanager.NewOmniOperatorSharesManager(evm.EvmContracts.OmniOperatorSharesManagerAddr, evm.EVMClient)
		if err != nil {
			return err
		}

		evm.EvmContracts.TssManagerAddr = ethcommon.HexToAddress(chain.TssManagerAddr)
		evm.EvmContracts.TssManager, err = tssManager.NewTSSManager(evm.EvmContracts.TssManagerAddr, evm.EVMClient)
		if err != nil {
			return err
		}

		evm.EvmContracts.PellTokenContractAddr = ethcommon.HexToAddress(chain.PellTokenAddr)
		evm.EvmContracts.PellTokenContract, err = erc20.NewERC20(evm.EvmContracts.PellTokenContractAddr, evm.EVMClient)
		if err != nil {
			return err
		}

		evm.EvmContracts.GatewayContractAddr = ethcommon.HexToAddress(chain.GatewayEVMAddr)
		evm.EvmContracts.GatewayContract, err = gatewayevm.NewGatewayEVM(evm.EvmContracts.GatewayContractAddr, evm.EVMClient)
		if err != nil {
			return err
		}

		evm.EvmContracts.GasSwapContractAddr = ethcommon.HexToAddress(chain.GasSwapEVMAddr)
		evm.EvmContracts.GasSwapContract, err = gasswapevm.NewGasSwapEVM(evm.EvmContracts.GasSwapContractAddr, evm.EVMClient)
		if err != nil {
			return err
		}

		evm.EvmContracts.CentralSchedulerAddr = ethcommon.HexToAddress(chain.CentralSchedulerAddr)
		evm.EvmContracts.CentralScheduler, err = centralscheduler.NewCentralScheduler(evm.EvmContracts.CentralSchedulerAddr, evm.EVMClient)
		if err != nil {
			return err
		}

		evm.EvmContracts.StakeManagerAddr = ethcommon.HexToAddress(chain.OperatorStakeManagerAddr)
		evm.EvmContracts.StakeManager, err = operatorstakemanager.NewOperatorStakeManager(evm.EvmContracts.StakeManagerAddr, evm.EVMClient)
		if err != nil {
			return err
		}

		evm.EvmContracts.EjectionManagerAddr = ethcommon.HexToAddress(chain.EjectionManagerAddr)
		evm.EvmContracts.EjectionManager, err = ejectionmanager.NewEjectionManager(evm.EvmContracts.EjectionManagerAddr, evm.EVMClient)
		if err != nil {
			return err
		}

		chainId, err := evm.EVMClient.ChainID(context.Background())
		if err != nil {
			return err
		}

		deployerPrivkey, err := crypto.HexToECDSA(conf.Accounts.DeployerPrivKey)
		if err != nil {
			return err
		}

		evm.EvmContracts.AdminTransact, err = bind.NewKeyedTransactorWithChainID(deployerPrivkey, chainId)
		if err != nil {
			return err
		}

		evm.E2ETestConfig = chain.E2ETestConfig

		r.MultiEVM[chainId.Int64()] = evm
	}

	return nil
}

func (r *Runner) fillPellContracts(conf *config.PellConfig) error {
	var err error
	r.PellContracts.PellDelegationManagerAddr = ethcommon.HexToAddress(conf.PEVM.PellDelegationManager)
	r.PellContracts.PellDelegationManager, err = pellDelegationManager.NewPellDelegationManager(r.PellContracts.PellDelegationManagerAddr, r.PEVMClient)
	if err != nil {
		return err
	}

	r.PellContracts.PellStrategyManagerAddr = ethcommon.HexToAddress(conf.PEVM.PellStrategyManager)
	r.PellContracts.PellStrategyManager, err = pellStrategyManager.NewPellStrategyManager(r.PellContracts.PellStrategyManagerAddr, r.PEVMClient)
	if err != nil {
		return err
	}

	r.PellContracts.GatewayAddr = ethcommon.HexToAddress(conf.PEVM.PellGatewayAddr)
	r.PellContracts.Gateway, err = gatewaypevm.NewGatewayPEVM(r.PellContracts.GatewayAddr, r.PEVMClient)
	if err != nil {
		return err
	}

	r.PellContracts.PellRegistryRouterFactoryAddr = ethcommon.HexToAddress(conf.PEVM.PellRegistryRouterFactory)
	r.PellContracts.PellRegistryRouterFactory, err = registryrouterfactory.NewRegistryRouterFactory(r.PellContracts.PellRegistryRouterFactoryAddr, r.PEVMClient)
	if err != nil {
		return err
	}

	r.PellContracts.PellDvsDirectoryAddr = ethcommon.HexToAddress(conf.PEVM.PellDvsDirectory)
	r.PellContracts.PellDvsDirectory, err = dvsdirectory.NewDVSDirectory(r.PellContracts.PellDvsDirectoryAddr, r.PEVMClient)

	return nil
}

func (r *Runner) fillClients(conf *config.PellConfig) error {
	ctx := context.Background()
	var err error

	r.PEVMClient, err = ethclient.Dial(conf.Rpcs.PellEvm)
	if err != nil {
		return err
	}

	pevmChainId, err := r.PEVMClient.ChainID(ctx)
	if err != nil {
		return err
	}

	deployerPrivkey, err := crypto.HexToECDSA(conf.Accounts.DeployerPrivKey)
	if err != nil {
		return err
	}

	r.PEVMAuth, err = bind.NewKeyedTransactorWithChainID(deployerPrivkey, pevmChainId)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	grpcConn, err := grpc.DialContext(ctx, conf.Rpcs.PellCoreGrpc, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}

	r.PellClients.XmsgClient = xmsgtypes.NewQueryClient(grpcConn)
	r.PellClients.FungibleClient = pevmtypes.NewQueryClient(grpcConn)
	r.PellClients.AuthClient = authtypes.NewQueryClient(grpcConn)
	r.PellClients.BankClient = banktypes.NewQueryClient(grpcConn)
	r.PellClients.RelayerClient = relayertypes.NewQueryClient(grpcConn)
	r.PellClients.StakingClient = stakingtypes.NewQueryClient(grpcConn)
	r.PellClients.LightclientClient = lightclienttypes.NewQueryClient(grpcConn)
	r.PellClients.RestakingClient = restakingtypes.NewQueryClient(grpcConn)
	r.PellClients.XSecurityClient = xsecuritytypes.NewQueryClient(grpcConn)

	r.PellClients.RPCClientURL = conf.Rpcs.PellCoreRpc

	// This client might be used later, keep it for now
	tmClient := tmservice.NewServiceClient(grpcConn)
	nodeInfo, err := tmClient.GetNodeInfo(context.Background(), &tmservice.GetNodeInfoRequest{})
	if err != nil {
		return err
	}

	r.PellChainId, err = chains.CosmosToEthChainID(nodeInfo.DefaultNodeInfo.Network)
	if err != nil {
		return err
	}

	return nil
}
