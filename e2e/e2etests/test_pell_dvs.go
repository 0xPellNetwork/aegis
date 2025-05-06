package e2etests

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"strconv"
	"time"

	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouterfactory.sol"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/stakeregistryrouter.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/centralscheduler.sol"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/0xPellNetwork/aegis/e2e/runner"
	"github.com/0xPellNetwork/aegis/e2e/utils"
	"github.com/0xPellNetwork/aegis/x/restaking/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func TestDVSCreateRegistryRouter(r *runner.Runner, _ []string) {
	r.UpdateChainParamsLargerGasLimit()

	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	addrBytes, _ := hex.DecodeString(common.HexToAddress(TestDeployerEvmAddr).Hex()[2:])
	deployerPellAddress := sdk.AccAddress(addrBytes)
	err = r.TxServer.SendPellFromAdmin(deployerPellAddress, big.NewInt(200))
	if err != nil {
		panic(err)
	}

	chainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, chainId)
	deployerTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for _, chain := range r.MultiEVM {
		tx, err := r.PellContracts.PellRegistryRouterFactory.CreateRegistryRouter(
			deployerTransact,
			common.HexToAddress(TestDeployerEvmAddr),                   // _initialDVSOwner
			common.HexToAddress(TestDeployerEvmAddr),                   // _dvsChainApprover
			TestOperatorEvmAddr,                                        // _churnApprover
			common.HexToAddress(TestDeployerEvmAddr),                   // _ejector
			[]common.Address{common.HexToAddress(TestDeployerEvmAddr)}, // _pausers
			common.HexToAddress(TestDeployerEvmAddr),                   // _unpauser
			big.NewInt(0),                                              // _initialPausedStatus
		)
		if err != nil {
			panic(err)
		}

		r.Logger.Info("CreateRegistryRouter tx hash: %s", tx.Hash())

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("CreateRegistryRouter failed, tx hash: " + tx.Hash().String())
		}

		registryRouterFactoryABI, err := registryrouterfactory.RegistryRouterFactoryMetaData.GetAbi()
		if err != nil {
			panic(err)
		}

		for _, log := range receipt.Logs {
			if log.Topics[0] == registryRouterFactoryABI.Events["RegistryRouterCreated"].ID {
				// Remove '0x' prefix if present
				logData := log.Data
				if len(logData) >= 2 && logData[0] == '0' && logData[1] == 'x' {
					logData = logData[2:]
				}

				dataBytes, err := hex.DecodeString(hex.EncodeToString(logData))
				if err != nil {
					r.Logger.Error("Failed to decode log data: %v", err)
					continue
				}

				// Extract proxy address from the last 20 bytes
				if len(dataBytes) >= 32 {
					proxyAddress := dataBytes[12:32] // Take last 20 bytes for address
					chain.EvmContracts.RegistryRouterAddr = common.BytesToAddress(proxyAddress)
					r.Logger.Info("Registry router address: %s", chain.EvmContracts.RegistryRouterAddr.String())

					chain.EvmContracts.RegistryRouter, err = registryrouter.NewRegistryRouter(chain.EvmContracts.RegistryRouterAddr, r.PEVMClient)
					if err != nil {
						panic(err)
					}

					chain.EvmContracts.StakeRegistryRouterAddr, err = chain.EvmContracts.RegistryRouter.StakeRegistryRouter(&bind.CallOpts{})
					if err != nil {
						panic(err)
					}

					chain.EvmContracts.StakeRegistryRouter, err = stakeregistryrouter.NewStakeRegistryRouter(chain.EvmContracts.StakeRegistryRouterAddr, r.PEVMClient)
					if err != nil {
						panic(err)
					}
				} else {
					panic(" Data too short to contain address: %d bytes")
				}
				break
			}
		}

		if len(receipt.Logs) == 0 || chain.EvmContracts.RegistryRouterAddr == (common.Address{}) {
			panic("Warning: No logs found in transaction receipt")
		}
	}
}

func TestDVSAddSupportedChain(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	for chainID, chain := range r.MultiEVM {
		deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, big.NewInt(chainID))
		deployerTransact.GasLimit = 2000000
		utils.Assert(err == nil, err)

		salt := generateRandomSalt()
		expiry := big.NewInt(time.Now().Add(100 * time.Hour).Unix())

		digest, err := chain.EvmContracts.RegistryRouter.CalculateAddSupportedDVSApprovalDigestHash(
			&bind.CallOpts{},
			registryrouter.IRegistryRouterDVSInfo{
				ChainId:          big.NewInt(chainID),
				CentralScheduler: chain.EvmContracts.CentralSchedulerAddr,
				StakeManager:     chain.EvmContracts.StakeManagerAddr,
				EjectionManager:  chain.EvmContracts.EjectionManagerAddr,
			}, salt, expiry)
		if err != nil {
			panic(err)
		}

		// Sign the digest
		signature, err := crypto.Sign(digest[:], deployerPrivKey)
		if err != nil {
			panic(err)
		}

		// Adjust v value for Ethereum
		signature[64] += 27

		// Call RegisterChainToPell with the generated signature
		tx, err := chain.EvmContracts.CentralScheduler.RegisterChainToPell(
			deployerTransact,
			chain.EvmContracts.RegistryRouterAddr,
			centralscheduler.ISignatureUtilsSignatureWithSaltAndExpiry{
				Signature: signature,
				Salt:      salt,
				Expiry:    expiry,
			},
		)
		if err != nil {
			panic(err)
		}

		r.Logger.Info("RegisterChainToPell tx hash: %s", tx.Hash().String())

		receipt := utils.MustWaitForTxReceipt(r.Ctx, chain.EVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("RegisterChainToPell failed, tx hash: " + tx.Hash().String())
		}

		// wait for pevm called, this is not cross chain, but also need wait
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)
			chains, err := chain.EvmContracts.RegistryRouter.GetSupportedChain(&bind.CallOpts{})
			if err != nil {
				panic(err)
			}

			if len(chains) == 0 {
				if i == WaitCrossChainTime {
					panic("TestDVSAddSupportedChain failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSCreateGroupOnPell(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, pellChainId)
	deployerTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for chainID, chain := range r.MultiEVM {
		// Set up IRegistryRouterOperatorSetParam parameter
		operatorSetParams := registryrouter.IRegistryRouterOperatorSetParam{
			MaxOperatorCount:        10,  // Set the maximum operator count
			KickBIPsOfOperatorStake: 100, // Set kick BIPs of operator stake
			KickBIPsOfTotalStake:    200, // Set kick BIPs of total stake
		}
		// Set up an array of IRegistryRouterPoolParams
		// poolParamsArray[0] contains 2 pools
		// poolParamsArray[1] contains 3 pools
		poolParamsArray := [][]registryrouter.IStakeRegistryRouterPoolParams{}
		for k := 2; k <= 3; k++ {
			poolParams := []registryrouter.IStakeRegistryRouterPoolParams{}
			for i := 0; i < k; i++ { // 0-2, 0-3
				poolParams = append(poolParams, registryrouter.IStakeRegistryRouterPoolParams{
					ChainId:    big.NewInt(chainID*10 + int64(k)*100 + int64(i)*2), // mock some data, magic number
					Pool:       generateRandomEvmAddress(),                         // Pool contract address
					Multiplier: big.NewInt(10),                                     // Pool multiplier
				})
			}
			r.Logger.Info("CreateGroup poolParams: %v", poolParams)
			poolParamsArray = append(poolParamsArray, poolParams)
		}

		// Set up IRegistryRouterGroupEjectionParams
		quorumEjectionParams := registryrouter.IRegistryRouterGroupEjectionParams{
			RateLimitWindow:       3600, // Rate limit window in seconds
			EjectableStakePercent: 50,   // Ejectable stake percentage
		}
		// Set the minimum stake parameter
		minimumStake := big.NewInt(0) // Minimum stake, e.g., 1 ETH in wei

		// create 11 quorum on pell before enable chain
		// poolParamsArray[0], poolParamsArray[1]
		for i := 0; i <= 1; i++ {
			time.Sleep(time.Second)
			// Call the CreateGroup function
			tx, err := chain.EvmContracts.RegistryRouter.CreateGroup(deployerTransact, operatorSetParams, minimumStake,
				poolParamsArray[i], quorumEjectionParams)
			if err != nil {
				panic(err)
			}

			receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
			if receipt.Status == 0 {
				panic("Create group on pell chain failed, tx hash: " + tx.Hash().String())
			}

			r.Logger.Info("Create group on pell chain success, tx hash: %s", tx.Hash().String())
		}

		// wait for cross chain
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)
			quorumCount, err := chain.EvmContracts.RegistryRouter.GroupCount(&bind.CallOpts{})
			if err != nil {
				panic(err)
			}

			r.Logger.Info("Group count on Pell: %v", quorumCount)

			if quorumCount == 0 {
				if i == WaitCrossChainTime {
					panic("TestDVSCreateGroup failed")
				}
				continue
			}
			break
		}
	}
}

// generateRandomEvmAddress generates a random EVM address
func generateRandomEvmAddress() common.Address {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	return crypto.PubkeyToAddress(*publicKey)
}

func TestDVSSyncGroup(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, pellChainId)
	deployerTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for chainID, chain := range r.MultiEVM {
		groupNumbers := []byte{0, 1}
		tx, err := chain.EvmContracts.RegistryRouter.SyncGroup(deployerTransact, big.NewInt(chainID), groupNumbers)
		if err != nil {
			panic(err)
		}

		r.Logger.Info("SyncGroup tx hash: %s", tx.Hash().String())

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("SyncGroup failed")
		}

		// wait for status changed
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			ret, err := r.QueryDVSGroupSyncStatus(tx.Hash().String())
			if err != nil {
				continue
			}

			isSynced := true
			//r.Logger.Info("Group sync status length %d", len(ret.Xmsg))

			if len(ret.Xmsg) != 4 {
				panic("TestDVSSyncAllGroupData failed")
			}

			for _, v := range ret.Xmsg {
				//r.Logger.Info("Group sync status %s, xmsg index %s", v.XmsgStatus.Status.String(), v.Index)

				// use debug log to print all pool data
				//r.Logger.Info("---------group 0----------")
				//for i := 0; i < 5; i++ {
				//	res, err := chain.EvmContracts.StakeManager.PoolsPerGroup(&bind.CallOpts{}, byte(0), big.NewInt(int64(i)))
				//	r.Logger.Info("SyncGroup debug res: %v, err: %v, index %d", res, err, i)
				//}
				//
				//r.Logger.Info("---------group 1----------")
				//for i := 0; i < 5; i++ {
				//	res, err := chain.EvmContracts.StakeManager.PoolsPerGroup(&bind.CallOpts{}, byte(1), big.NewInt(int64(i)))
				//	r.Logger.Info("SyncGroup debug res: %v, err: %v, index %d", res, err, i)
				//}

				if v.XmsgStatus.Status != xmsgtypes.XmsgStatus_OUTBOUND_MINED {
					isSynced = false
					if v.XmsgStatus.Status == xmsgtypes.XmsgStatus_REVERTED || v.XmsgStatus.Status == xmsgtypes.XmsgStatus_ABORTED {
						panic("TestDVSSyncAllGroupData failed" + v.XmsgStatus.String())
					}

					break
				}
			}

			if !isSynced {
				if i == WaitCrossChainTime {
					panic("TestDVSSyncGroup failed")
				}
				continue
			}
			break
		}

		// wait for status changed
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)
			status := r.QueryDVSSupportedChainStatus(chain.EvmContracts.RegistryRouterAddr.String(), uint64(chainID))
			if status.OutboundState != types.OutboundStatus_OUTBOUND_STATUS_NORMAL {
				if i == WaitCrossChainTime {
					panic("TestDVSAddSupportedChain failed")
				}
			}
			break
		}
	}
}

func TestDVSCreateGroup(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, pellChainId)
	deployerTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for chainID, chain := range r.MultiEVM {
		// Set up IRegistryRouterOperatorSetParam parameter
		operatorSetParams := registryrouter.IRegistryRouterOperatorSetParam{
			MaxOperatorCount:        10,  // Set the maximum operator count
			KickBIPsOfOperatorStake: 100, // Set kick BIPs of operator stake
			KickBIPsOfTotalStake:    200, // Set kick BIPs of total stake
		}
		// Set up an array of IRegistryRouterPoolParams
		poolParams := []registryrouter.IStakeRegistryRouterPoolParams{
			{
				ChainId:    big.NewInt(chainID),             // Chain ID, e.g., 1 for Ethereum mainnet
				Pool:       chain.EvmContracts.StrategyAddr, // Pool contract address
				Multiplier: big.NewInt(10),                  // Pool multiplier
			},
		}
		// Set up IRegistryRouterGroupEjectionParams
		quorumEjectionParams := registryrouter.IRegistryRouterGroupEjectionParams{
			RateLimitWindow:       3600, // Rate limit window in seconds
			EjectableStakePercent: 50,   // Ejectable stake percentage
		}
		// Set the minimum stake parameter
		minimumStake := big.NewInt(0) // Minimum stake, e.g., 1 ETH in wei

		// Call the CreateGroup function
		tx, err := chain.EvmContracts.RegistryRouter.CreateGroup(deployerTransact, operatorSetParams, minimumStake, poolParams, quorumEjectionParams)
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("CreateGroup failed, tx hash: " + tx.Hash().String())
		}

		r.Logger.Info("CreateGroup success, tx hash: %s", tx.Hash().String())

		// wait for cross chain
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)
			quorumCount, err := chain.EvmContracts.CentralScheduler.GroupCount(&bind.CallOpts{})
			if err != nil {
				panic(err)
			}

			r.Logger.Info("Group count on Omni chain: %v", quorumCount)

			if quorumCount == 0 {
				if i == WaitCrossChainTime {
					panic("TestDVSCreateGroup failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSSetOperatorSetParams(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, pellChainId)
	deployerTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for _, chain := range r.MultiEVM {
		operatorSetParams := registryrouter.IRegistryRouterOperatorSetParam{
			MaxOperatorCount:        20,
			KickBIPsOfOperatorStake: 100,
			KickBIPsOfTotalStake:    200,
		}

		tx, err := chain.EvmContracts.RegistryRouter.SetOperatorSetParams(deployerTransact, 0, operatorSetParams)
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("SetOperatorSetParams failed, tx hash: " + tx.Hash().String())
		}

		r.Logger.Info("SetOperatorSetParams success, tx hash: %s", tx.Hash().String())

		// Wait for cross chain message processing
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			params, err := chain.EvmContracts.CentralScheduler.GetOperatorSetParams(&bind.CallOpts{}, 0)
			if err != nil {
				panic(err)
			}

			if params.MaxOperatorCount != 20 {
				if i == WaitCrossChainTime {
					panic("SetOperatorSetParams failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSSetGroupEjectionParams(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, pellChainId)
	deployerTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for _, chain := range r.MultiEVM {
		quorumEjectionParams := registryrouter.IRegistryRouterGroupEjectionParams{
			RateLimitWindow:       3600,
			EjectableStakePercent: 60,
		}

		tx, err := chain.EvmContracts.RegistryRouter.SetGroupEjectionParams(deployerTransact, 0, quorumEjectionParams)
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("TestDVSSetGroupEjectionParams failed, tx hash: " + tx.Hash().String())
		}

		r.Logger.Info("TestDVSSetGroupEjectionParams success, tx hash: %s", tx.Hash().String())

		// Wait for cross chain message processing
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			params, err := chain.EvmContracts.EjectionManager.GroupEjectionParams(&bind.CallOpts{}, 0)
			if err != nil {
				panic(err)
			}

			if params.EjectableStakePercent != 60 {
				if i == WaitCrossChainTime {
					panic("TestDVSSetGroupEjectionParams failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSSetEjectionCooldown(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, pellChainId)
	deployerTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for _, chain := range r.MultiEVM {
		tx, err := chain.EvmContracts.RegistryRouter.SetEjectionCooldown(deployerTransact, big.NewInt(100))
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("TestDVSSetEjectionCooldown failed, tx hash: " + tx.Hash().String())
		}

		r.Logger.Info("TestDVSSetEjectionCooldown success, tx hash: %s", tx.Hash().String())

		// Wait for cross chain message processing
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			ret, err := chain.EvmContracts.CentralScheduler.EjectionCooldown(&bind.CallOpts{})
			if err != nil {
				panic(err)
			}

			r.Logger.Info("eject cooldown: %v", ret)

			if ret.Int64() != 100 {
				if i == WaitCrossChainTime {
					panic("TestDVSSetEjectionCooldown failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSRegisterOperatorBeforeSyncGroup(r *runner.Runner, _ []string) {
	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	operatorPrivKey, err := crypto.HexToECDSA(TestOperatorPrivKey)
	if err != nil {
		panic(err)
	}

	operatorTransact, err := bind.NewKeyedTransactorWithChainID(operatorPrivKey, pellChainId)
	operatorTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for _, chain := range r.MultiEVM {
		groupNumbers := []byte{0} // only have number 0
		socket := "localhost:8545"
		pubkeyParams := generateBLSPubkeyParams(TestOperatorEvmAddr, chain)

		salt := generateRandomSalt()
		expiry := big.NewInt(time.Now().Add(100 * time.Hour).Unix())

		// Create EIP-712 signature
		digest, err := r.PellContracts.PellDvsDirectory.CalculateOperatorDVSRegistrationDigestHash(
			&bind.CallOpts{},
			TestOperatorEvmAddr,
			chain.EvmContracts.RegistryRouterAddr,
			salt,
			expiry,
		)
		if err != nil {
			panic(err)
		}

		// Sign the digest
		dvsSignature, err := crypto.Sign(digest[:], operatorPrivKey)
		if err != nil {
			panic(err)
		}

		// Adjust v value for Ethereum
		dvsSignature[64] += 27
		operatorSignature := registryrouter.ISignatureUtilsSignatureWithSaltAndExpiry{
			Signature: dvsSignature,
			Salt:      salt,
			Expiry:    expiry,
		}

		// Call RegisterOperator
		tx, err := chain.EvmContracts.RegistryRouter.RegisterOperator(
			operatorTransact,
			groupNumbers,
			socket,
			pubkeyParams,
			operatorSignature,
		)
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("RegisterOperator failed, tx hash: " + tx.Hash().String())
		}

		r.Logger.Info("TestDVSRegisterOperatorBeforeSyncGroup success, tx hash: %s", tx.Hash().String())

		// Wait for cross chain message processing
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)
			// don't need to verify cross chain verify because before syncGroup operation
			break
		}
	}
}

func TestDVSRegisterOperator(r *runner.Runner, _ []string) {
	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	operatorPrivKey, err := crypto.HexToECDSA(TestOperatorPrivKey)
	if err != nil {
		panic(err)
	}

	operatorTransact, err := bind.NewKeyedTransactorWithChainID(operatorPrivKey, pellChainId)
	operatorTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for _, chain := range r.MultiEVM {
		groupNumbers := []byte{0} // only have number 0
		socket := "localhost:8545"
		pubkeyParams := generateBLSPubkeyParams(TestOperatorEvmAddr, chain)

		// r.Logger.Info("register operator bls pg1.X: %v", pubkeyParams.PubkeyG1.X)
		// r.Logger.Info("register operator bls pg1.Y: %v", pubkeyParams.PubkeyG1.Y)
		// r.Logger.Info("register operator bls pg2.X: %v", pubkeyParams.PubkeyG2.X)
		// r.Logger.Info("register operator bls pg2.Y: %v", pubkeyParams.PubkeyG2.Y)

		salt := generateRandomSalt()
		expiry := big.NewInt(time.Now().Add(100 * time.Hour).Unix())

		// Create EIP-712 signature
		digest, err := r.PellContracts.PellDvsDirectory.CalculateOperatorDVSRegistrationDigestHash(
			&bind.CallOpts{},
			TestOperatorEvmAddr,
			chain.EvmContracts.RegistryRouterAddr,
			salt,
			expiry,
		)
		if err != nil {
			panic(err)
		}

		// Sign the digest
		dvsSignature, err := crypto.Sign(digest[:], operatorPrivKey)
		if err != nil {
			panic(err)
		}

		// Adjust v value for Ethereum
		dvsSignature[64] += 27
		operatorSignature := registryrouter.ISignatureUtilsSignatureWithSaltAndExpiry{
			Signature: dvsSignature,
			Salt:      salt,
			Expiry:    expiry,
		}

		// Call RegisterOperator
		tx, err := chain.EvmContracts.RegistryRouter.RegisterOperator(
			operatorTransact,
			groupNumbers,
			socket,
			pubkeyParams,
			operatorSignature,
		)
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("RegisterOperator failed, tx hash: " + tx.Hash().String())
		}

		r.Logger.Info("RegisterOperator success, tx hash: %s", tx.Hash().String())

		// Wait for cross chain message processing
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			operator, err := chain.EvmContracts.CentralScheduler.GetOperator(&bind.CallOpts{}, TestOperatorEvmAddr)
			if err != nil {
				panic(err)
			}

			if operator.Status != 1 {
				if i == WaitCrossChainTime {
					panic("TestRegisterOperator failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSAddPools(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	dvsTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, pellChainId)
	dvsTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for chainID, chain := range r.MultiEVM {
		mockChainId := big.NewInt(chainID + 1)
		tx, err := chain.EvmContracts.StakeRegistryRouter.AddPools(
			dvsTransact,
			byte(0),
			[]stakeregistryrouter.IStakeRegistryRouterPoolParams{
				{
					ChainId:    mockChainId,
					Pool:       chain.EvmContracts.StrategyAddr,
					Multiplier: big.NewInt(20),
				},
			},
		)
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("AddPools failed, tx hash: " + tx.Hash().String())
		}

		r.Logger.Info("AddPools success, tx hash: %s", tx.Hash().String())

		// Wait for cross chain message processing
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			_, err1 := chain.EvmContracts.StakeManager.PoolsPerGroup(&bind.CallOpts{}, byte(0), big.NewInt(0))
			_, err2 := chain.EvmContracts.StakeManager.PoolsPerGroup(&bind.CallOpts{}, byte(0), big.NewInt(1))

			// strategy can't be nil
			if err1 != nil || err2 != nil {
				if i == WaitCrossChainTime {
					panic("TestDVSAddPools failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSRemovePools(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	dvsTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, pellChainId)
	dvsTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	groupNumber := 0
	for _, chain := range r.MultiEVM {
		tx, err := chain.EvmContracts.StakeRegistryRouter.RemovePools(dvsTransact, uint8(groupNumber), []*big.Int{big.NewInt(0)})
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("RemovePools failed, tx hash: " + tx.Hash().String())
		}

		r.Logger.Info("RemovePools success, tx hash: %s", tx.Hash().String())

		// Wait for cross chain message processing
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			// useful for debug syncGroup status
			//r.Logger.Info("-------------------")
			//for i := 0; i < 5; i++ {
			//	res, err := chain.EvmContracts.StakeManager.PoolsPerGroup(&bind.CallOpts{}, byte(groupNumber), big.NewInt(int64(i)))
			//	r.Logger.Info("RemovePools res: %v, err: %v, index %d", res, err, i)
			//}

			// already create 2 group before syncGroup
			_, err1 := chain.EvmContracts.StakeManager.PoolsPerGroup(&bind.CallOpts{}, byte(groupNumber), big.NewInt(0))
			_, err2 := chain.EvmContracts.StakeManager.PoolsPerGroup(&bind.CallOpts{}, byte(groupNumber), big.NewInt(1))
			_, err3 := chain.EvmContracts.StakeManager.PoolsPerGroup(&bind.CallOpts{}, byte(groupNumber), big.NewInt(2))

			// expect strategy2 is nil
			if !(err1 == nil && err2 == nil && err3 != nil) {
				if i == WaitCrossChainTime {
					panic("RemovePools failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSModifyPoolParams(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	dvsTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, pellChainId)
	dvsTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for _, chain := range r.MultiEVM {
		poolParamsBefore, err := chain.EvmContracts.StakeManager.PoolParamsByIndex(&bind.CallOpts{}, 0, big.NewInt(0))
		if err != nil {
			panic(err)
		}

		r.Logger.Info("ModifyPoolParams pool params before change: %v", poolParamsBefore)

		newMultoplier := big.NewInt(30)
		tx, err := chain.EvmContracts.StakeRegistryRouter.ModifyPoolParams(
			dvsTransact,
			0,
			[]*big.Int{big.NewInt(0)},
			[]*big.Int{newMultoplier},
		)
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("ModifyPoolParams failed, tx hash: " + tx.Hash().String())
		}

		r.Logger.Info("ModifyPoolParams success, tx hash: %s", tx.Hash().String())

		// Wait for cross chain message processing
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			poolParamsAfter, err := chain.EvmContracts.StakeManager.PoolParamsByIndex(&bind.CallOpts{}, 0, big.NewInt(0))
			if err != nil {
				panic(err)
			}

			r.Logger.Info("ModifyPoolParams pool params after change: %v", poolParamsAfter)

			if poolParamsAfter.Multiplier.Uint64() != newMultoplier.Uint64() {
				if i == WaitCrossChainTime {
					panic("ModifyPoolParams failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSUpdateOperators(r *runner.Runner, _ []string) {
	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	operatorPrivKey, err := crypto.HexToECDSA(TestOperatorPrivKey)
	if err != nil {
		panic(err)
	}

	operatorTransact, err := bind.NewKeyedTransactorWithChainID(operatorPrivKey, pellChainId)
	operatorTransact.GasLimit = 2000000
	if err != nil {
		panic(err)
	}

	for _, chain := range r.MultiEVM {
		operatorId, err := chain.EvmContracts.RegistryRouter.GetOperatorId(&bind.CallOpts{}, TestOperatorEvmAddr)
		if err != nil {
			panic(err)
		}

		beforeUpdate, err := chain.EvmContracts.StakeManager.GetLatestStakeUpdate(&bind.CallOpts{}, operatorId, 0)
		if err != nil {
			panic(err)
		}

		tx, err := chain.EvmContracts.RegistryRouter.UpdateOperators(operatorTransact, []common.Address{TestOperatorEvmAddr})
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("TestDVSUpdateOperators update operators failed, tx hash: " + tx.Hash().String())
		}

		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			update, err := chain.EvmContracts.StakeManager.GetLatestStakeUpdate(&bind.CallOpts{}, operatorId, 0)
			if err != nil {
				panic(err)
			}

			// TODO: how verify
			break
			r.Logger.Info("TestDVSUpdateOperators update block number: %v, before update block number: %v", update.UpdateBlockNumber, beforeUpdate.UpdateBlockNumber)
			if update.UpdateBlockNumber == beforeUpdate.UpdateBlockNumber {
				if i == WaitCrossChainTime {
					panic("TestDVSUpdateOperators failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSUpdateOperatorsForGroup(r *runner.Runner, _ []string) {
	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	operatorPrivKey, err := crypto.HexToECDSA(TestOperatorPrivKey)
	if err != nil {
		panic(err)
	}

	operatorTransact, err := bind.NewKeyedTransactorWithChainID(operatorPrivKey, pellChainId)
	operatorTransact.GasLimit = 2000000
	if err != nil {
		panic(err)
	}

	for _, chain := range r.MultiEVM {
		beforeUpdate, err := chain.EvmContracts.CentralScheduler.GroupUpdateBlockNumber(&bind.CallOpts{}, 0)
		if err != nil {
			panic(err)
		}

		r.Logger.Info("TestDVSUpdateOperatorsForGroup before update: %v", beforeUpdate.Int64())

		tx, err := chain.EvmContracts.RegistryRouter.UpdateOperatorsForGroup(operatorTransact, [][]common.Address{{TestOperatorEvmAddr}}, []byte{0})
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("TestDVSUpdateOperatorsForGroup failed, tx hash: " + tx.Hash().String())
		}

		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			update, err := chain.EvmContracts.CentralScheduler.GroupUpdateBlockNumber(&bind.CallOpts{}, 0)
			if err != nil {
				panic(err)
			}

			// TODO: how verify
			break
			if update.Int64() == beforeUpdate.Int64() {
				if i == WaitCrossChainTime {
					panic("TestDVSUpdateOperatorsForGroup failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSDeregisterOperator(r *runner.Runner, _ []string) {
	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	operatorPrivKey, err := crypto.HexToECDSA(TestOperatorPrivKey)
	if err != nil {
		panic(err)
	}

	operatorTransact, err := bind.NewKeyedTransactorWithChainID(operatorPrivKey, pellChainId)
	operatorTransact.GasLimit = 2000000
	if err != nil {
		panic(err)
	}

	for _, chain := range r.MultiEVM {
		groupNumbers := []byte{0}
		tx, err := chain.EvmContracts.RegistryRouter.DeregisterOperator(operatorTransact, groupNumbers)
		if err != nil {
			panic(err)
		}

		r.Logger.Info("DeregisterOperator tx hash: %s", tx.Hash().String())

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("DeregisterOperator failed")
		}

		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			operator, err := chain.EvmContracts.CentralScheduler.GetOperator(&bind.CallOpts{}, TestOperatorEvmAddr)
			if err != nil {
				panic(err)
			}

			if operator.Status != 2 {
				if i == WaitCrossChainTime {
					panic("DeregisterOperator failed, status: " + strconv.Itoa(int(operator.Status)))
				}
				continue
			}
			break
		}
	}
}

func TestDVSRegisterOperatorWithChurn(r *runner.Runner, _ []string) {
	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	operatorPrivKey, err := crypto.HexToECDSA(TestOperatorPrivKey)
	if err != nil {
		panic(err)
	}

	operatorTransact, err := bind.NewKeyedTransactorWithChainID(operatorPrivKey, pellChainId)
	operatorTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for _, chain := range r.MultiEVM {
		quorumNumbers := []byte{0} // only have number 0
		socket := "localhost:8545"
		pubkeyParams := generateBLSPubkeyParams(TestOperatorEvmAddr, chain)

		salt := generateRandomSalt()
		expiry := big.NewInt(time.Now().Add(100 * time.Hour).Unix())

		// operator sign
		operatorDigest, err := r.PellContracts.PellDvsDirectory.CalculateOperatorDVSRegistrationDigestHash(
			&bind.CallOpts{},
			TestOperatorEvmAddr,
			chain.EvmContracts.RegistryRouterAddr,
			salt,
			expiry,
		)
		if err != nil {
			panic(err)
		}

		operatorSign, err := crypto.Sign(operatorDigest[:], operatorPrivKey)
		if err != nil {
			panic(err)
		}

		operatorSign[64] += 27
		operatorSignature := registryrouter.ISignatureUtilsSignatureWithSaltAndExpiry{
			Signature: operatorSign,
			Salt:      salt,
			Expiry:    expiry,
		}

		// churn appserver
		id, err := chain.EvmContracts.RegistryRouter.GetOperatorId(&bind.CallOpts{}, TestOperatorEvmAddr)
		if err != nil {
			panic(err)
		}

		operatorKickParams := []registryrouter.IRegistryRouterOperatorKickParam{{
			GroupNumber: 0,
			Operator:    TestOperatorEvmAddr,
		}}

		// Create EIP-712 signature
		churnDigest, err := chain.EvmContracts.RegistryRouter.CalculateOperatorChurnApprovalDigestHash(
			&bind.CallOpts{},
			TestOperatorEvmAddr,
			id,
			operatorKickParams,
			salt,
			expiry,
		)
		if err != nil {
			panic(err)
		}

		chrunSign, err := crypto.Sign(churnDigest[:], operatorPrivKey)
		if err != nil {
			panic(err)
		}

		chrunSign[64] += 27
		chrunSignature := registryrouter.ISignatureUtilsSignatureWithSaltAndExpiry{
			Signature: chrunSign,
			Salt:      salt,
			Expiry:    expiry,
		}

		// Call RegisterOperator
		tx, err := chain.EvmContracts.RegistryRouter.RegisterOperatorWithChurn(
			operatorTransact,
			quorumNumbers,
			socket,
			pubkeyParams,
			operatorKickParams,
			chrunSignature,
			operatorSignature,
		)
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("TestDVSRegisterOperatorWithChurn failed, tx hash: " + tx.Hash().String())
		}

		r.Logger.Info("TestDVSRegisterOperatorWithChurn success, tx hash: %s", tx.Hash().String())

		// Wait for cross chain message processing
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			operator, err := chain.EvmContracts.CentralScheduler.GetOperator(&bind.CallOpts{}, TestOperatorEvmAddr)
			if err != nil {
				panic(err)
			}

			if operator.Status != 1 {
				if i == WaitCrossChainTime {
					panic("TestDVSRegisterOperatorWithChurn failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSEjectOperators(r *runner.Runner, _ []string) {
	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, pellChainId)
	deployerTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for _, chain := range r.MultiEVM {
		id, err := chain.EvmContracts.RegistryRouter.GetOperatorId(&bind.CallOpts{}, TestOperatorEvmAddr)
		if err != nil {
			panic(err)
		}

		// Call RegisterOperator
		tx, err := chain.EvmContracts.RegistryRouter.EjectOperators(deployerTransact, [][][32]byte{{id}})
		if err != nil {
			panic(err)
		}

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("TestDVSEjectOperators failed, tx hash: " + tx.Hash().String())
		}

		r.Logger.Info("TestDVSEjectOperators success, tx hash: %s", tx.Hash().String())

		// Wait for cross chain message processing
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			_, err := chain.EvmContracts.EjectionManager.StakeEjectedForGroup(&bind.CallOpts{}, 0, big.NewInt(0))

			if err != nil {
				if i == WaitCrossChainTime {
					panic("TestDVSEjectOperators failed")
				}
				continue
			}
			break
		}
	}
}

func TestDVSSyncGroupFailed(r *runner.Runner, _ []string) {
	deployerPrivKey, err := crypto.HexToECDSA(TestDeployerPrivKey)
	if err != nil {
		panic(err)
	}

	pellChainId, err := r.PEVMClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	deployerTransact, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, pellChainId)
	deployerTransact.GasLimit = 2000000
	utils.Assert(err == nil, err)

	for chainID, chain := range r.MultiEVM {
		groupNumbers := []byte{0}
		tx, err := chain.EvmContracts.RegistryRouter.SyncGroup(deployerTransact, big.NewInt(chainID), groupNumbers)
		if err != nil {
			panic(err)
		}

		r.Logger.Info("TestDVSSyncGroupFailed tx hash: %s", tx.Hash().String())

		receipt := utils.MustWaitForTxReceipt(r.Ctx, r.PEVMClient, tx, r.Logger, r.ReceiptTimeout)
		if receipt.Status == 0 {
			panic("SyncGroup failed")
		}

		// wait for status changed
		for i := 1; i <= WaitCrossChainTime; i++ {
			time.Sleep(time.Second)

			ret, err := r.QueryDVSGroupSyncStatus(tx.Hash().String())
			if err != nil {
				continue
			}

			r.Logger.Info("TestDVSSyncGroupFailed sync status length %d", len(ret.Xmsg))
			for _, v := range ret.Xmsg {
				if v.XmsgStatus.Status != xmsgtypes.XmsgStatus_OUTBOUND_MINED {
					if v.XmsgStatus.Status == xmsgtypes.XmsgStatus_REVERTED {
						panic("TestDVSSyncAllQuorumData failed")
					}

					if v.XmsgStatus.Status == xmsgtypes.XmsgStatus_ABORTED {
						r.Logger.Info("TestDVSSyncGroupFailed sync status %s, xmsg index %s", v.XmsgStatus.Status.String(), v.Index)
					}
					break
				}
			}
			break
		}

	}
}
