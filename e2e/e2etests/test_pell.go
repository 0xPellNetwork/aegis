package e2etests

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/bridge/gatewaypevm.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/omnioperatorsharesmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/pell-chain/pellcore/e2e/runner"
	"github.com/pell-chain/pellcore/e2e/utils"
)

// TODO: support multi chain test case
const WaitCrossChainTime = 240

const (
	// TestStakerPrivKey = "8E18A4583487983A420007E780A3A33EAE3C71CA719AC1FE0396A475E80C542B"

	// TestOperatorPrivKey  = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	// TestOperatorPellAddr = "pell17w0adeg64ky0daxwd2ugyuneellmjgnxnsekc7"
	// TestOperatorEvmAddr  = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"

	TestDeployerPrivKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	TestDeployerEvmAddr = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
)

var (
	TestStakerPrivKey, TestStakerEvmAddr, TestStakerPellAddr       = utils.GenKeypair()
	TestOperatorPrivKey, TestOperatorEvmAddr, TestOperatorPellAddr = utils.GenKeypair()
)

type TxLog struct {
	Address string `json:"address"`
}

func TestUpgradeSystemContract(r *runner.Runner, _ []string) {
	tx, err := r.UpgradeSystemContract()
	utils.Assert(err == nil, err)
	r.Logger.Info("UpgradeSystemContract tx hash: %s", tx.TxHash)

	// The connector must be deployed before deploying the gateway contract.

	tx3, err := r.DeployConnectorContract()
	utils.Assert(err == nil, err)
	r.Logger.Info("DeployConnectorContract tx hash: %s", tx3.TxHash)

	tx2, err := r.DeployGatewayContract()
	utils.Assert(err == nil, err)
	r.Logger.Info("DeployGatewayContract tx hash: %s", tx2.TxHash)

	for _, event := range tx2.Events {
		for _, attr := range event.Attributes {
			if attr.Key == "txLog" {
				var txLog TxLog
				err := json.Unmarshal([]byte(attr.Value), &txLog)
				if err != nil {
					fmt.Println("Error parsing txLog:", err)
					continue
				}
				r.Logger.Info("New gateway contract address: %s", txLog.Address)

				r.UpdateChainParamsGateway()

				tx, err := r.AddAllowXmsgSender(txLog.Address)
				if err != nil {
					panic(err)
				}
				r.Logger.Info("AddAllowXmsgSender tx hash: %s", tx.TxHash)

				r.PellContracts.GatewayAddr = common.HexToAddress(txLog.Address)
				r.PellContracts.Gateway, err = gatewaypevm.NewGatewayPEVM(r.PellContracts.GatewayAddr, r.PEVMClient)
				if err != nil {
					panic(err)
				}
				break
			}
		}
	}

	chainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	for chainID, chain := range r.MultiEVM {
		deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
		if err != nil {
			panic(err)
		}

		deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, big.NewInt(chainID))
		deployerTransact.GasLimit = 2000000
		utils.Assert(err == nil, err)

		tx, err := chain.EvmContracts.GatewayContract.UpdateSourceAddress(deployerTransact, chainId, r.PellContracts.GatewayAddr.Bytes())
		if err != nil {
			panic(err)
		}
		r.Logger.Info("TestUpgradeSystemContract update source address tx hash: %s", tx.Hash().Hex())

		tx2, err := chain.EvmContracts.GatewayContract.UpdateDestinationAddress(deployerTransact, chainId, r.PellContracts.GatewayAddr.Bytes())
		if err != nil {
			panic(err)
		}
		r.Logger.Info("TestUpgradeSystemContract update destination address tx hash: %s", tx2.Hash().Hex())

		receipt := utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		utils.Assert(receipt.Status == 1, "TestUpgradeSystemContract update source address tx failed")

	}

}

func TestDeposit(r *runner.Runner, _ []string) {
	r.Logger.Info("TestPellDeposit processing. staker: %s, operator: %s", TestStakerEvmAddr.String(), TestOperatorEvmAddr.String())

	stakerPrivKey, err := crypto.HexToECDSA(TestStakerPrivKey)
	if err != nil {
		panic(err)
	}

	for chainId, chain := range r.MultiEVM {
		// amount := 10 // 10 * 1e18
		// mint erc20 to staker address
		stakerTransact, err := bind.NewKeyedTransactorWithChainID(stakerPrivKey, big.NewInt(chainId))
		utils.Assert(err == nil, err)
		// mint/transfer to staker address for test
		feeCoinAmt, _ := big.NewInt(0).SetString(chain.E2ETestConfig.DepositStakerFee, 10)
		r.Lock()

		tx, err := r.SendEther(chainId, TestStakerEvmAddr, feeCoinAmt, chain.E2ETestConfig.GasLimit, nil)
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("SendEther failed")
		}

		stTokenAmt, _ := big.NewInt(0).SetString(chain.E2ETestConfig.StTokenAmount, 10)
		tx, err = chain.EvmContracts.STERC20.Transfer(chain.EvmContracts.AdminTransact, TestStakerEvmAddr, stTokenAmt)
		if err != nil {
			panic(err)
		}

		receipt = utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("Transfer failed")
		}
		r.Unlock()

		// amount := 10 // 10 * 1e18
		balance, err := chain.EvmContracts.STERC20.BalanceOf(&bind.CallOpts{}, TestStakerEvmAddr)
		if err != nil {
			panic(err)
		}

		utils.Assert(balance.Cmp(big.NewInt(0)) != 0, fmt.Sprintf("st token(%s) zero balance", chain.EvmContracts.STERC20Addr.String()))

		tx, err = chain.EvmContracts.STERC20.Approve(stakerTransact, chain.EvmContracts.StrategyManagerAddr, balance)
		utils.Assert(err == nil, err)

		receipt = utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("Approve failed")
		}

		tx, err = chain.EvmContracts.StrategyManager.DepositIntoStrategy(stakerTransact, chain.EvmContracts.StrategyAddr, chain.EvmContracts.STERC20Addr, balance)
		utils.Assert(err == nil, err)

		r.Logger.Info("DepositIntoStrategy tx hash: %s", tx.Hash().String())

		receipt = utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("DepositIntoStrategy failed")
		}

		evmShare, err := chain.EvmContracts.StrategyManager.StakerStrategyShares(&bind.CallOpts{}, TestStakerEvmAddr, chain.EvmContracts.StrategyAddr)
		utils.Assert(err == nil, fmt.Sprint(tx.Hash().String(), err))
		utils.Assert(evmShare.Cmp(big.NewInt(0)) != 0, fmt.Sprint("evm StakerStrategyShares", evmShare.String()))
		// retry check
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)
			pellShare, err := r.PellContracts.PellStrategyManager.StakerStrategyShares(&bind.CallOpts{}, big.NewInt(chainId), TestStakerEvmAddr, chain.EvmContracts.StrategyAddr)
			utils.Assert(err == nil, fmt.Sprint(tx.Hash().String(), err, r.PellContracts.PellStrategyManagerAddr.String()))
			if pellShare.Cmp(big.NewInt(0)) == 0 || evmShare.Cmp(pellShare) != 0 {
				if i == WaitCrossChainTime {
					panic("StakerStrategyShares crosschain failed....")
				}
				continue
			}

			break
		}
	}
}

func TestRegisterOperator(r *runner.Runner, _ []string) {
	r.Logger.Info("TestRegisterOperator processing. operator: %s private_address: %s", TestOperatorPrivKey, TestOperatorEvmAddr)

	operatorPrivKey, err := crypto.HexToECDSA(TestOperatorPrivKey)
	if err != nil {
		panic(err)
	}

	err = r.TxServer.SendPellFromAdmin(TestOperatorPellAddr, big.NewInt(200))
	if err != nil {
		panic(err)
	}

	chainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	operatorTransact, err := bind.NewKeyedTransactorWithChainID(operatorPrivKey, chainId)
	operatorTransact.GasLimit = 1000000
	utils.Assert(err == nil, err)

	// check operator registered
	isOperator, err := r.PellContracts.PellDelegationManager.IsOperator(&bind.CallOpts{}, TestOperatorEvmAddr)
	utils.Assert(err == nil, err)
	utils.Assert(!isOperator, "register operator failed")

	metadataUrl := "https://raw.githubusercontent.com/matthew7251/Metadata/main/FourUnit_Metadata.json"
	tx, err := r.PellContracts.PellDelegationManager.RegisterAsOperator(operatorTransact, pelldelegationmanager.IPellDelegationManagerOperatorDetails{
		DelegationApprover: common.Address{},
		StakerOptOutWindow: 0,
	}, metadataUrl)

	r.Logger.Info("RegisterAsOperator tx hash: %s", tx.Hash().String())
	utils.Assert(err == nil, err)
	reciept := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
	utils.Assert(reciept.Status == 1, fmt.Sprint("RegisterAsOperator tx failed", tx.Hash().Hex(), reciept))
	// check operator registered
	isOperator, err = r.PellContracts.PellDelegationManager.IsOperator(&bind.CallOpts{}, TestOperatorEvmAddr)
	utils.Assert(err == nil, err)
	utils.Assert(isOperator, "register operator failed")

	r.Logger.Info("operator address: %s, is operator: %v", TestOperatorEvmAddr.String(), isOperator)

	for _, chain := range r.MultiEVM {
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)
			isOperator, err = chain.EvmContracts.DelegationManager.IsOperator(&bind.CallOpts{}, TestOperatorEvmAddr)
			utils.Assert(err == nil, fmt.Sprint(err, isOperator))
			if !isOperator {
				if i == WaitCrossChainTime {
					panic("RegisterAsOperator xmsg failed")
				}
				continue
			}
			break
		}
	}

	r.Logger.Info("TestRegisterOperator test successful")
}

func TestDelegateTo(r *runner.Runner, _ []string) {
	stakerPrivKey, err := crypto.HexToECDSA(TestStakerPrivKey)
	if err != nil {
		panic(err)
	}

	for chainId, chain := range r.MultiEVM {
		stakerTransact, err := bind.NewKeyedTransactorWithChainID(stakerPrivKey, big.NewInt(chainId))
		if err != nil {
			panic(err)
		}

		stakerTransact.GasLimit = chain.E2ETestConfig.GasLimit
		salt, _ := hex.DecodeString("044852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d")

		tx, err := chain.EvmContracts.DelegationManager.DelegateTo(stakerTransact, TestOperatorEvmAddr, delegationmanager.ISignatureUtilsSignatureWithExpiry{Signature: []byte{}, Expiry: big.NewInt(0)}, [32]byte(salt))
		if err != nil {
			panic(err)
		}

		r.Logger.Info("delegate tx hash: %s", tx.Hash().String())

		receipt := utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		utils.Assert(receipt.Status == 1, "DelegateTo tx failed")

		delegationShares, err := chain.EvmContracts.DelegationManager.GetOperatorShares(&bind.CallOpts{}, TestOperatorEvmAddr, []common.Address{chain.EvmContracts.StrategyAddr})
		if err != nil {
			panic(err)
		}

		utils.Assert(delegationShares[0].Cmp(big.NewInt(0)) != 0, "share is zero")

		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			shares, err := chain.EvmContracts.OmniOperatorSharesManager.GetOperatorShares(&bind.CallOpts{}, TestOperatorEvmAddr, []omnioperatorsharesmanager.IOmniOperatorSharesManagerStrategyWithChain{{ChainId: big.NewInt(chainId), Strategy: chain.EvmContracts.StrategyAddr}})
			if err != nil {
				panic(err)
			}
			// TODO: support multi chain test case
			if delegationShares[0].Cmp(shares[0]) != 0 {
				if i == WaitCrossChainTime {
					panic("TestDelegateTo OmniOperatorSharesManager xmsg failed")
				}
				continue
			}

			break
		}
	}
}

func TestQueueWithdrawals(r *runner.Runner, _ []string) {
	stakerPrivKey, err := crypto.HexToECDSA(TestStakerPrivKey)
	if err != nil {
		panic(err)
	}

	expectedSharesBeforeWithdrawal := big.NewInt(1e18)
	withdrawalSharesP1 := big.NewInt(1e17)
	withdrawalSharesP2 := big.NewInt(2e17)
	expectedSharesAfterWithdrawal := big.NewInt(7e17)

	for chainId, chain := range r.MultiEVM {
		stakerTransact, err := bind.NewKeyedTransactorWithChainID(stakerPrivKey, big.NewInt(chainId))
		if err != nil {
			panic(err)
		}

		stakerTransact.GasLimit = chain.E2ETestConfig.GasLimit
		// before withdrawal
		sharesBeforeWithdrawal, err := chain.EvmContracts.DelegationManager.GetOperatorShares(&bind.CallOpts{}, TestOperatorEvmAddr, []common.Address{chain.EvmContracts.StrategyAddr})
		if err != nil {
			panic(err)
		}

		utils.Assert(sharesBeforeWithdrawal[0].Cmp(expectedSharesBeforeWithdrawal) == 0, "share before withdrawal is unexpected")
		// send queue withdrawal
		withdrawalParams := []delegationmanager.IDelegationManagerQueuedWithdrawalParams{
			{
				Strategies: []common.Address{chain.EvmContracts.StrategyAddr},
				Shares:     []*big.Int{withdrawalSharesP1},
				Withdrawer: TestStakerEvmAddr,
			},
			{
				Strategies: []common.Address{chain.EvmContracts.StrategyAddr},
				Shares:     []*big.Int{withdrawalSharesP2},
				Withdrawer: TestStakerEvmAddr,
			},
		}

		tx, err := chain.EvmContracts.DelegationManager.QueueWithdrawals(stakerTransact, withdrawalParams)
		if err != nil {
			panic(err)
		}

		r.Logger.Info("queueWithdrawals tx hash: %s", tx.Hash().String())
		// check receipt
		receipt := utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		utils.Assert(receipt.Status == 1, "queueWithdrawals tx failed")
		// after withdrawal
		sharesAfterWithdrawal, err := chain.EvmContracts.DelegationManager.GetOperatorShares(&bind.CallOpts{}, TestOperatorEvmAddr, []common.Address{chain.EvmContracts.StrategyAddr})
		if err != nil {
			panic(err)
		}

		utils.Assert(sharesAfterWithdrawal[0].Cmp(expectedSharesAfterWithdrawal) == 0, "share after withdrawal is unexpected")

		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)
			omniShares, err := chain.EvmContracts.OmniOperatorSharesManager.GetOperatorShares(&bind.CallOpts{}, TestOperatorEvmAddr, []omnioperatorsharesmanager.IOmniOperatorSharesManagerStrategyWithChain{{ChainId: big.NewInt(chainId), Strategy: chain.EvmContracts.StrategyAddr}})
			if err != nil {
				panic(err)
			}

			if sharesAfterWithdrawal[0].Cmp(omniShares[0]) != 0 {
				if i == WaitCrossChainTime {
					panic("TestQueueWithdrawals OmniOperatorSharesManager xmsg failed")
				}
				continue
			}
			break
		}
	}
}

func TestUndelegate(r *runner.Runner, _ []string) {
	stakerPrivKey, err := crypto.HexToECDSA(TestStakerPrivKey)
	if err != nil {
		panic(err)
	}

	expectedSharesBeforeUndelegate := big.NewInt(7e17)
	expectedSharesAfterUndelegate := big.NewInt(0)

	for chainId, chain := range r.MultiEVM {
		stakerTransact, err := bind.NewKeyedTransactorWithChainID(stakerPrivKey, big.NewInt(chainId))
		if err != nil {
			panic(err)
		}
		stakerTransact.GasLimit = chain.E2ETestConfig.GasLimit
		// before undelegate
		sharesBeforeUndelegate, err := chain.EvmContracts.DelegationManager.GetOperatorShares(&bind.CallOpts{}, TestOperatorEvmAddr, []common.Address{chain.EvmContracts.StrategyAddr})
		if err != nil {
			panic(err)
		}

		utils.Assert(sharesBeforeUndelegate[0].Cmp(expectedSharesBeforeUndelegate) == 0, "share before undelegate is unexpected")
		// send undelegate
		tx, err := chain.EvmContracts.DelegationManager.Undelegate(stakerTransact, TestStakerEvmAddr)
		if err != nil {
			panic(err)
		}

		r.Logger.Info("undelegate tx hash: %s", tx.Hash().String())
		// check receipt
		reciept := utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		utils.Assert(reciept.Status == 1, "undelegateTo tx failed")
		// check shares
		sharesAfterUndelegate, err := chain.EvmContracts.DelegationManager.GetOperatorShares(&bind.CallOpts{}, TestOperatorEvmAddr, []common.Address{chain.EvmContracts.StrategyAddr})
		if err != nil {
			panic(err)
		}

		utils.Assert(sharesAfterUndelegate[0].Cmp(expectedSharesAfterUndelegate) == 0, "share after undelegate is unexpected")

		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)
			omniShares, err := chain.EvmContracts.OmniOperatorSharesManager.GetOperatorShares(&bind.CallOpts{}, TestOperatorEvmAddr, []omnioperatorsharesmanager.IOmniOperatorSharesManagerStrategyWithChain{{ChainId: big.NewInt(chainId), Strategy: chain.EvmContracts.StrategyAddr}})
			if err != nil {
				panic(err)
			}

			// TODO: support multi chain test case
			if sharesAfterUndelegate[0].Cmp(omniShares[0]) != 0 {
				if i == WaitCrossChainTime {
					panic("TestUndelegate OmniOperatorSharesManager xmsg failed")
				}
				continue
			}
			break
		}
	}
}

func TestPellRechargeToken(r *runner.Runner, _ []string) {
	for _, chain := range r.MultiEVM {
		addresses, err := chain.EvmContracts.GatewayContract.SourceAddresses(&bind.CallOpts{}, big.NewInt(r.PellChainId))
		if err != nil {
			panic(err)
		}

		r.Logger.Info("TestPellRechargeToken query source address %s at %s, remote contract address: %s", common.BytesToAddress(addresses).String(), time.Now().Format(time.RFC3339), chain.EvmContracts.GatewayContractAddr)

		// enable recharge pell token after update source address
		r.UpdateChainParamsEnablePellTokenRecharge()

		// wait for cross chain
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			// query balance of latest tss address
			balance, err := chain.EvmContracts.PellTokenContract.BalanceOf(&bind.CallOpts{}, chain.EvmContracts.GasSwapContractAddr)
			if err != nil {
				panic(err)
			}

			if balance.Cmp(big.NewInt(0)) == 0 {
				if i == WaitCrossChainTime {
					panic("PellTokenRecharge failed")
				}
				continue
			}
			break
		}
	}
}

func TestGasRechargeToken(r *runner.Runner, _ []string) {
	for _, chain := range r.MultiEVM {
		// query balance before recharge
		balanceBefore, err := chain.EVMClient.BalanceAt(context.Background(), r.TSSAddress, nil)
		if err != nil {
			panic(err)
		}

		// enable recharge gas token after update source address
		r.UpdateChainParamsEnableGasTokenRecharge()

		// wait for cross chain
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			// query balance of latest tss address
			balanceAfter, err := chain.EVMClient.BalanceAt(context.Background(), r.TSSAddress, nil)
			if err != nil {
				panic(err)
			}

			if balanceAfter.Cmp(balanceBefore) <= 0 {
				if i == WaitCrossChainTime {
					panic("GasTokenRecharge failed")
				}
				continue
			}
			break
		}
	}
}

func TestBridgePellInbound(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	pellChainID, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	for chainId, chain := range r.MultiEVM {
		deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, big.NewInt(chainId))
		if err != nil {
			panic(err)
		}
		deployerTransact.GasLimit = chain.E2ETestConfig.GasLimit
		receiver := common.HexToAddress(TestDeployerEvmAddr)

		balanceBefore, err := r.PEVMClient.BalanceAt(context.Background(), receiver, nil)
		if err != nil {
			panic(err)
		}

		r.Logger.Info("Balance before bridge inbound: address=%s, balance=%s", receiver.Hex(), balanceBefore)

		destinationChainId := pellChainID
		pellValue := big.NewInt(1000000000000000000)
		destinationGasLimit := big.NewInt(1000000)
		tx, err := chain.EvmContracts.GatewayContract.Bridge(deployerTransact, destinationChainId, receiver.Bytes(), pellValue, destinationGasLimit)
		if err != nil {
			panic(err)
		}
		r.Logger.Info("Bridge inbound tx hash: %s", tx.Hash().Hex())

		// check receipt
		reciept := utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		utils.Assert(reciept.Status == 1, "TestBridgePellInbound tx failed")

		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			balanceAfter, err := r.PEVMClient.BalanceAt(context.Background(), receiver, nil)
			if err != nil {
				panic(err)
			}

			r.Logger.Info("Balance after bridge inbound: address=%s, balance=%s", receiver.Hex(), balanceAfter)

			if balanceAfter.Cmp(balanceBefore) <= 0 {
				if i == WaitCrossChainTime {
					panic("TestBridgePellInbound failed")
				}
				continue
			}
			break
		}
	}
}

func TestBridgePellOutbound(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	chainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, chainId)
	deployerTransact.GasLimit = 1000000
	utils.Assert(err == nil, err)

	for chainID, chain := range r.MultiEVM {
		receiver := common.HexToAddress(TestDeployerEvmAddr)

		balanceBefore, err := chain.EvmContracts.PellTokenContract.BalanceOf(&bind.CallOpts{}, receiver)
		if err != nil {
			panic(err)
		}

		r.Logger.Info("Balance before bridge outbound: address=%s, balance=%s", receiver.Hex(), balanceBefore)

		destinationGasLimit := big.NewInt(10000000)
		deployerTransact.Value = big.NewInt(100000000000000000)
		tx, err := r.PellContracts.Gateway.BridgePell(deployerTransact, big.NewInt(chainID), receiver.Bytes(), destinationGasLimit)
		if err != nil {
			panic(err)
		}
		r.Logger.Info("Bridge outbound tx hash: %s", tx.Hash().Hex())

		// check receipt
		reciept := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		utils.Assert(reciept.Status == 1, "TestBridgePellInbound BridgePell tx failed")

		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			balanceAfter, err := chain.EvmContracts.PellTokenContract.BalanceOf(&bind.CallOpts{}, receiver)
			if err != nil {
				panic(err)
			}

			r.Logger.Info("Balance after bridge outbound: address=%s, balance=%s", receiver.Hex(), balanceAfter)

			if balanceAfter.Cmp(balanceBefore) <= 0 {
				if i == WaitCrossChainTime {
					panic("TestBridgePellOutbound failed")
				}
				continue
			}
			break
		}
	}
}
