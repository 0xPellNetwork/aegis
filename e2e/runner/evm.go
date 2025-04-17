package runner

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/0xPellNetwork/aegis/e2e/contracts/erc20"
	"github.com/0xPellNetwork/aegis/e2e/utils"
)

// WaitForTxReceiptOnEvm waits for a tx receipt on EVM
func (runner *Runner) WaitForTxReceiptOnEvm(chainId int64, tx *ethtypes.Transaction) {
	runner.Lock()
	defer runner.Unlock()

	receipt := utils.MustWaitForTxReceipt(runner.Ctx, runner.MultiEVM[chainId].EVMClient, tx, runner.Logger, runner.ReceiptTimeout)
	if receipt.Status != 1 {
		panic("tx failed")
	}
}

// SendERC20OnEvm sends ERC20 to an address on EVM
// this allows the ERC20 contract deployer to funds other accounts on EVM
// amountERC20 is a multiple of 1e18
func (runner *Runner) SendERC20OnEvm(erc20 *erc20.ERC20, adminTransact *bind.TransactOpts, address ethcommon.Address, amountERC20 int64) *ethtypes.Transaction {
	// the deployer might be sending ERC20 in different goroutines
	runner.Lock()
	defer runner.Unlock()

	amount := big.NewInt(0).Mul(big.NewInt(1e18), big.NewInt(amountERC20))

	// transfer
	tx, err := erc20.Transfer(adminTransact, address, amount)
	if err != nil {
		panic(err)
	}

	return tx
}

// SendEther sends ethers to the TSS on EVM
func (runner *Runner) SendEther(chainId int64, to ethcommon.Address, value *big.Int, gasLimit uint64, data []byte) (*ethtypes.Transaction, error) {
	chain := runner.MultiEVM[chainId]

	nonce, err := chain.EVMClient.PendingNonceAt(runner.Ctx, runner.DeployerAddress)
	if err != nil {
		return nil, err
	}

	gasPrice, err := chain.EVMClient.SuggestGasPrice(runner.Ctx)
	if err != nil {
		return nil, err
	}

	runner.Logger.Print("evm client SuggestGasPrice: %s", gasPrice.String())

	tx := ethtypes.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)

	deployerPrivkey, err := crypto.HexToECDSA(runner.DeployerPrivateKey)
	if err != nil {
		return nil, err
	}

	signedTx, err := ethtypes.SignTx(tx, ethtypes.NewEIP155Signer(big.NewInt(chainId)), deployerPrivkey)
	if err != nil {
		return nil, err
	}

	err = chain.EVMClient.SendTransaction(runner.Ctx, signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}
