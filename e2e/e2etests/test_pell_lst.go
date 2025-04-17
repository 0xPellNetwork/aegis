package e2etests

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/pelldelegationmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/service_evm/omnioperatorsharesmanager.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/staking_evm/core/v3/delegationmanager.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/0xPellNetwork/aegis/e2e/runner"
	"github.com/0xPellNetwork/aegis/e2e/utils"
	"github.com/0xPellNetwork/aegis/x/restaking/types"
	xsecuritytypes "github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

const (
	pellcore0EVMAddr  = "0x0D42281f7a604C313931BfE1E75e58aCC42541E3"
	pellcore0EVMPriv  = "4e89c6bf87f85fea856d33b57880bee42fc2004678c71318f952925d96bc1201"
	pellcore0PellAddr = "pell1p4pzs8m6vpxrzwf3hls7whjc4nzz2s0rcf3xvy"
)

var pellcore0EVMAddrHash = ethcommon.HexToAddress(pellcore0EVMAddr)

var LSTTestStakerPrivKey, LSTTestStakerEvmAddr, LSTTestStakerPellAddr = utils.GenKeypair()

// TestLSTSetVotingPowerRatio tests the LST voting power ratio setting
// First it updates the LST staking enabled status, then sets the voting power ratio to 50%.
func TestLSTSetVotingPowerRatio(r *runner.Runner, _ []string) {
	// Update the LST staking enabled status
	stakingEnabled, err := r.LSTUpdateLSTStakingEnabled(true)
	utils.Assert(err == nil, err)
	r.Logger.Info("TestLSTSetVotingPowerRatio update LST staking enabled tx hash: %s", stakingEnabled.TxHash)

	enabled, err := r.LSTQueryLSTStakingEnabled()
	utils.Assert(err == nil, err)
	utils.Assert(enabled.LstStakingEnabled, "LST staking enabled should be true")

	// Set up the voting power ratio
	numerator := uint64(50)
	denominator := uint64(100)
	ratio, err := r.LSTUpdateVotingPowerRatio(numerator, denominator)
	utils.Assert(err == nil, err)
	r.Logger.Info("TestLSTSetVotingPowerRatio tx hash: %s", ratio.TxHash)

	powerRatio, err := r.LSTQueryVotingPowerRatio()
	utils.Assert(err == nil, err)
	utils.Assert(powerRatio.Numerator == 50, "voting power Numerator should be 50")
	utils.Assert(powerRatio.Denominator == 100, "voting power Denominator should be 50")
}

// TestLSTCreateRegistryRouter tests the creation of a registry router
func TestLSTCreateRegistryRouter(r *runner.Runner, _ []string) {
	tx, err := r.LSTCreateRegistryRouter(pellcore0EVMAddr, pellcore0EVMAddr, pellcore0EVMAddr, pellcore0EVMAddr, pellcore0EVMAddr, 0)
	utils.Assert(err == nil, err)
	r.Logger.Info("TestLSTCreateRegistryRouter tx hash: %s", tx.TxHash)

	address, err := r.LSTQueryRegistryRouter()
	utils.Assert(err == nil, err)
	r.Logger.Info("TestLSTCreateRegistryRouter query registry router address: %s", address)

	r.PellContracts.PellRegistryRouterAddr = ethcommon.HexToAddress(address.RegistryRouterAddress)
	r.PellContracts.PellRegistryRouter, err = registryrouter.NewRegistryRouter(r.PellContracts.PellRegistryRouterAddr, r.PEVMClient)
	utils.Assert(err == nil, err)
}

func TestLSTCreateGroup(r *runner.Runner, _ []string) {
	// Set up OperatorSetParam parameter
	operatorSetParams := &types.OperatorSetParam{
		MaxOperatorCount:        10,  // Maximum operator count
		KickBipsOfOperatorStake: 100, // Kick BIPs of operator stake
		KickBipsOfTotalStake:    200, // Kick BIPs of total stake
	}

	// Set up an array of PoolParams
	poolParams := []*types.PoolParams{}
	for chainID, chain := range r.MultiEVM {
		poolParams = append(poolParams, &types.PoolParams{
			ChainId:    uint64(chainID), // Chain ID
			Pool:       chain.EvmContracts.StrategyAddr.String(),
			Multiplier: 10, // Pool multiplier
		})
	}
	r.Logger.Info("CreateGroup poolParams: %v", poolParams)

	// Set up GroupEjectionParams
	ejectionParams := &types.GroupEjectionParam{
		RateLimitWindow:       3600, // Rate limit window in seconds
		EjectableStakePercent: 50,   // Ejectable stake percentage
	}

	// Set the minimum stake parameter
	minimumStake := 0

	tx, err := r.LSTCreateGroup(operatorSetParams, poolParams, ejectionParams, int64(minimumStake))
	utils.Assert(err == nil, err)
	r.Logger.Info("TestLSTCreateGroup tx hash: %s", tx.TxHash)

	info, err := r.LSTQueryGroupInfo()
	utils.Assert(err == nil, err)
	r.Logger.Info("TestLSTCreateGroup query group info: %s", info)
}

// TestLSTRegisterOperatorToDelegationManager tests the registration of an operator to the delegation manager
func TestLSTRegisterOperatorToDelegationManager(r *runner.Runner, _ []string) {
	r.Logger.Info("TestLSTRegisterOperatorToDelegationManager processing. operator: %s private_address: %s", pellcore0EVMAddr, pellcore0EVMPriv)

	err := r.TxServer.SendPellFromAdmin(sdk.AccAddress(pellcore0PellAddr), big.NewInt(200))
	utils.Assert(err == nil, err)

	chainId, err := r.PEVMClient.ChainID(context.Background())
	utils.Assert(err == nil, err)

	operatorPrivKey, err := crypto.HexToECDSA(pellcore0EVMPriv)
	utils.Assert(err == nil, err)

	// check operator registered
	isOperator, err := r.PellContracts.PellDelegationManager.IsOperator(&bind.CallOpts{}, pellcore0EVMAddrHash)
	utils.Assert(err == nil, err)
	utils.Assert(!isOperator, "TestLSTRegisterOperatorToDelegationManager register operator failed")
	r.Logger.Info("TestLSTRegisterOperatorToDelegationManager isOperator: %v", isOperator)

	operatorTransact, err := bind.NewKeyedTransactorWithChainID(operatorPrivKey, chainId)
	operatorTransact.GasLimit = 1000000
	utils.Assert(err == nil, err)

	metadataUrl := "https://raw.githubusercontent.com/matthew7251/Metadata/main/FourUnit_Metadata.json"
	tx, err := r.PellContracts.PellDelegationManager.RegisterAsOperator(operatorTransact, pelldelegationmanager.IPellDelegationManagerOperatorDetails{
		DelegationApprover: ethcommon.Address{},
		StakerOptOutWindow: 0,
	}, metadataUrl)
	utils.Assert(err == nil, err)
	r.Logger.Info("TestLSTRegisterOperatorToDelegationManager tx hash: %s", tx.Hash().String())

	reciept := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
	utils.Assert(reciept.Status == 1, fmt.Sprint("TestLSTRegisterOperatorToDelegationManager tx failed", tx.Hash().Hex(), reciept))

	// check operator registered
	isOperator, err = r.PellContracts.PellDelegationManager.IsOperator(&bind.CallOpts{}, pellcore0EVMAddrHash)
	utils.Assert(err == nil, err)
	utils.Assert(isOperator, "TestRegisterOperator register operator failed")

	r.Logger.Info("TestRegisterOperator operator address: %s, is operator: %v", pellcore0EVMAddrHash.String(), isOperator)
}

// TestLSTRegisterOperator tests the registration of an operator
func TestLSTRegisterOperator(r *runner.Runner, _ []string) {
	g1HashedMsgToSign, err := r.PellContracts.PellRegistryRouter.PubkeyRegistrationMessageHash(&bind.CallOpts{}, pellcore0EVMAddrHash)
	utils.Assert(err == nil, err)

	socket := "localhost:8545"
	pubkeyParams := generateBLSPubkeyParamsBySign(g1HashedMsgToSign)

	salt := generateRandomSalt()
	expiry := big.NewInt(time.Now().Add(100 * time.Hour).Unix())

	// Create EIP-712 signature
	digest, err := r.PellContracts.PellDvsDirectory.CalculateOperatorDVSRegistrationDigestHash(
		&bind.CallOpts{},
		pellcore0EVMAddrHash,
		r.PellContracts.PellRegistryRouterAddr,
		salt,
		expiry,
	)
	if err != nil {
		panic(err)
	}

	// Sign the digest
	operatorPrivKey, err := crypto.HexToECDSA(pellcore0EVMPriv)
	if err != nil {
		panic(err)
	}
	dvsSignature, err := crypto.Sign(digest[:], operatorPrivKey)
	if err != nil {
		panic(err)
	}

	// Adjust v value for Ethereum
	dvsSignature[64] += 27

	operatorSignature := &types.SignatureWithSaltAndExpiry{
		Signature: dvsSignature,
		Salt:      salt[:],
		Expiry:    expiry.Uint64(),
	}

	param := &xsecuritytypes.RegisterOperatorParam{
		Socket:       socket,
		PubkeyParams: ConvertPubkeyRegistrationParamsFromEventToStore(pubkeyParams),
		Signature:    operatorSignature,
	}

	tx, err := r.LSTRegisterOperator(param, pellcore0EVMAddr)
	utils.Assert(err == nil, err)
	r.Logger.Info("TestLSTRRegisterOperator tx hash: %s", tx.TxHash)
}

// TestLSTDeposit tests the deposit of LST tokens
// Staker deposits LST tokens into the strategy manager
func TestLSTDeposit(r *runner.Runner, _ []string) {
	r.Logger.Info("TestLSTDeposit processing. staker: %s", LSTTestStakerEvmAddr.String())

	stakerPrivKey, err := crypto.HexToECDSA(LSTTestStakerPrivKey)
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

		tx, err := r.SendEther(chainId, LSTTestStakerEvmAddr, feeCoinAmt, chain.E2ETestConfig.GasLimit, nil)
		utils.Assert(err == nil, err)
		r.Logger.Info("TestLSTDeposit tx hash: %s", tx.Hash().String())

		receipt := utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		utils.Assert(receipt.Status != 0, "SendEther failed")

		stTokenAmt, _ := big.NewInt(0).SetString(chain.E2ETestConfig.StTokenAmount, 10)
		tx, err = chain.EvmContracts.STERC20.Transfer(chain.EvmContracts.AdminTransact, LSTTestStakerEvmAddr, stTokenAmt)
		utils.Assert(err == nil, err)
		r.Logger.Info("TestLSTDeposit transfer tx hash: %s", tx.Hash().String())

		receipt = utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		utils.Assert(receipt.Status != 0, "Transfer failed")

		// amount := 10 // 10 * 1e18
		balance, err := chain.EvmContracts.STERC20.BalanceOf(&bind.CallOpts{}, LSTTestStakerEvmAddr)
		utils.Assert(err == nil, err)

		r.Logger.Info("TestLSTDeposit balance: %s", balance.String())
		utils.Assert(balance.Cmp(feeCoinAmt) == 0, fmt.Sprintf("st token(%s) zero balance", chain.EvmContracts.STERC20Addr.String()))

		tx, err = chain.EvmContracts.STERC20.Approve(stakerTransact, chain.EvmContracts.StrategyManagerAddr, balance)
		utils.Assert(err == nil, err)

		receipt = utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		utils.Assert(receipt.Status != 0, "Approve failed")

		tx, err = chain.EvmContracts.StrategyManager.DepositIntoStrategy(stakerTransact, chain.EvmContracts.StrategyAddr, chain.EvmContracts.STERC20Addr, balance)
		utils.Assert(err == nil, err)
		r.Logger.Info("DepositIntoStrategy tx hash: %s", tx.Hash().String())

		receipt = utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		utils.Assert(receipt.Status != 0, fmt.Sprint("DepositIntoStrategy tx failed", tx.Hash().Hex(), receipt))

		evmShare, err := chain.EvmContracts.StrategyManager.StakerStrategyShares(&bind.CallOpts{}, LSTTestStakerEvmAddr, chain.EvmContracts.StrategyAddr)
		utils.Assert(err == nil, fmt.Sprint(tx.Hash().String(), err))

		r.Logger.Info("TestLSTDeposit evmShare: %v", evmShare)
		utils.Assert(evmShare.Cmp(feeCoinAmt) == 0, fmt.Sprint("evm StakerStrategyShares", evmShare.String()))

		// retry check
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)
			pellShare, err := r.PellContracts.PellStrategyManager.StakerStrategyShares(&bind.CallOpts{}, big.NewInt(chainId), LSTTestStakerEvmAddr, chain.EvmContracts.StrategyAddr)
			utils.Assert(err == nil, fmt.Sprint(tx.Hash().String(), err, r.PellContracts.PellStrategyManagerAddr.String()))

			if pellShare.Cmp(big.NewInt(0)) == 0 || evmShare.Cmp(pellShare) != 0 {
				if i == WaitCrossChainTime {
					panic("StakerStrategyShares cross chain failed....")
				}
				continue
			}
			break
		}
	}
}

// TestLSTDelegate tests the delegation of LST tokens
// Staker delegates LST tokens to an operator
func TestLSTDelegate(r *runner.Runner, _ []string) {
	stakerPrivKey, err := crypto.HexToECDSA(LSTTestStakerPrivKey)
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

		tx, err := chain.EvmContracts.DelegationManager.DelegateTo(stakerTransact, pellcore0EVMAddrHash, delegationmanager.ISignatureUtilsSignatureWithExpiry{Signature: []byte{}, Expiry: big.NewInt(0)}, [32]byte(salt))
		if err != nil {
			panic(err)
		}

		r.Logger.Info("TestLSTDelegate delegate tx hash: %s", tx.Hash().String())

		receipt := utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		utils.Assert(receipt.Status == 1, "TestLSTDelegate tx failed")

		delegationShares, err := chain.EvmContracts.DelegationManager.GetOperatorShares(&bind.CallOpts{}, pellcore0EVMAddrHash, []ethcommon.Address{chain.EvmContracts.StrategyAddr})
		if err != nil {
			panic(err)
		}

		utils.Assert(delegationShares[0].Cmp(big.NewInt(0)) != 0, "TestLSTDelegate share is zero")
		r.Logger.Info("TestLSTDelegate delegate shares: %s", delegationShares[0].String())

		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			shares, err := chain.EvmContracts.OmniOperatorSharesManager.GetOperatorShares(&bind.CallOpts{}, pellcore0EVMAddrHash, []omnioperatorsharesmanager.IOmniOperatorSharesManagerStrategyWithChain{{ChainId: big.NewInt(chainId), Strategy: chain.EvmContracts.StrategyAddr}})
			if err != nil {
				panic(err)
			}

			// TODO: support multi chain test case
			if delegationShares[0].Cmp(shares[0]) != 0 {
				if i == WaitCrossChainTime {
					panic("TestLSTDelegate OmniOperatorSharesManager xmsg failed")
				}
				continue
			}

			break
		}
	}

	// verify the LST token voting power works
	// wait for the epoch change the voting power
	for i := 1; i <= WaitCrossChainTime; i++ {
		time.Sleep(time.Second)

		validators, err := r.LSTQueryValidator()
		if err != nil {
			panic(err)
		}

		// If both validators are found, break the loop
		if validators[0].VotingPower == "2000" && validators[1].VotingPower == "1000" {
			r.Logger.Info("TestLSTDelegate: found validator(1000/2000) in %d seconds", i)
			break
		}

		// If the timeout is reached, panic
		if i == WaitCrossChainTime {
			panic("TestLSTDelegate: waiting for validator failed")
		}
	}

}

// TestLSTOperatePools tests the addition and removal of pools
// This case also test the update of the operator set parameters
func TestLSTOperatePools(r *runner.Runner, _ []string) {
	poolParams := []*types.PoolParams{}
	for chainID := range r.MultiEVM {
		poolParams = append(poolParams, &types.PoolParams{
			ChainId:    uint64(chainID), // Chain ID
			Pool:       generateRandomEvmAddress().String(),
			Multiplier: 10, // Pool multiplier
		})
	}
	r.Logger.Info("TestLSTAddPools poolParams: %v", poolParams)

	tx, err := r.LSTAddPools(poolParams)
	utils.Assert(err == nil, err)
	r.Logger.Info("TestLSTAddPools tx hash: %s", tx.TxHash)

	for i := 1; i <= WaitCrossChainTime; i++ {
		time.Sleep(time.Second)

		info, err := r.LSTQueryGroupInfo()
		utils.Assert(err == nil, err)

		if len(info.PoolParams) == 2 {
			r.Logger.Info("TestLSTAddPools query group info: %v", info)
			break
		}

		if i == WaitCrossChainTime {
			panic("TestLSTAddPools waiting group info failed")
		}
	}

	tx, err = r.LSTRemovePools(poolParams)
	utils.Assert(err == nil, err)
	r.Logger.Info("TestLSTRemovePools tx hash: %s", tx.TxHash)

	for i := 1; i <= WaitCrossChainTime; i++ {
		time.Sleep(time.Second)

		info, err := r.LSTQueryGroupInfo()
		utils.Assert(err == nil, err)

		if len(info.PoolParams) == 1 {
			r.Logger.Info("TestLSTRemovePools query group info: %v", info)
			break
		}

		if i == WaitCrossChainTime {
			panic("TestLSTRemovePools waiting group info failed")
		}
	}

	operatorSetParams := &types.OperatorSetParam{
		MaxOperatorCount:        11,  // Maximum operator count
		KickBipsOfOperatorStake: 100, // Kick BIPs of operator stake
		KickBipsOfTotalStake:    200, // Kick BIPs of total stake
	}
	tx, err = r.LSTSetGroupParam(operatorSetParams)
	utils.Assert(err == nil, err)
	r.Logger.Info("TestLSTSetGroupParam tx hash: %s", tx.TxHash)

	for i := 1; i <= WaitCrossChainTime; i++ {
		time.Sleep(time.Second)

		info, err := r.LSTQueryGroupInfo()
		utils.Assert(err == nil, err)

		if info.OperatorSetParam.MaxOperatorCount == 11 {
			r.Logger.Info("TestLSTSetGroupParam query group info: %v", info)
			break
		}

		if i == WaitCrossChainTime {
			panic("TestLSTSetGroupParam waiting group info failed")
		}
	}
}
