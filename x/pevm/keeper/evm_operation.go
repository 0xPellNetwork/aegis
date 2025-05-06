package keeper

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	cosmoserrors "cosmossdk.io/errors"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/server/config"
	"github.com/0xPellNetwork/aegis/x/pevm/types"
	restakingtypes "github.com/0xPellNetwork/aegis/x/restaking/types"
	xmsgType "github.com/0xPellNetwork/aegis/x/xmsg/types"
	xsecuritytypes "github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

func (k Keeper) CallMethodOnSystemContract(ctx sdk.Context, systemContractAddr common.Address, method string, args ...interface{}) (*evmtypes.MsgEthereumTxResponse, error) {
	return k.CallEVM(
		ctx,
		*systemContractMetaDataABI,
		types.ModuleAddressEVM,
		systemContractAddr,
		types.BigIntZero,
		nil,
		true,
		false,
		method,
		args...,
	)
}

func (k Keeper) CallMethodOnGateway(ctx sdk.Context, gatewayAddress common.Address, method string, args ...interface{}) (*evmtypes.MsgEthereumTxResponse, error) {
	return k.CallEVM(
		ctx,
		*gatewayPEVMMetaDataABI,
		types.ModuleAddressEVM,
		gatewayAddress,
		types.BigIntZero,
		nil,
		true,
		false,
		method,
		args...,
	)
}

func (k Keeper) CallMethodOnContractByProxyAdmin(
	ctx sdk.Context,
	proxyAdmin common.Address,
	proxy, impl common.Address,
	implMetaData *bind.MetaData,
	method string,
	args ...interface{},
) (*evmtypes.MsgEthereumTxResponse, error) {
	implABI, err := implMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	data, err := implABI.Pack(method, args...)
	if err != nil {
		return nil, cosmoserrors.Wrap(
			types.ErrABIPack,
			cosmoserrors.Wrap(err, "failed to create transaction data").Error(),
		)
	}
	return k.call_UpgradeAndCallOnProxyAdmin(ctx, proxyAdmin, proxy, impl, data)
}

func (k Keeper) call_UpgradeAndCallOnProxyAdmin(
	ctx sdk.Context,
	proxyAdmin common.Address,
	proxy, impl common.Address,
	data []byte,
) (*evmtypes.MsgEthereumTxResponse, error) {
	return k.CallEVM(
		ctx,
		*proxyAdminMetaDataABI,
		types.ModuleAddressEVM,
		proxyAdmin,
		types.BigIntZero,
		nil,
		true,
		false,
		"upgradeAndCall",
		proxy,
		impl,
		data,
	)
}

// Call_syncDepositStateOnPellStrategyManager  call contract function in a single tx
// callable from pevm module
// Returns directly results from CallEVM
func (k Keeper) callSyncDepositStateOnPellStrategyManager(
	ctx sdk.Context,
	chainID *big.Int,
	staker common.Address,
	strategy common.Address,
	shares *big.Int,
) (*evmtypes.MsgEthereumTxResponse, error) {
	to, err := k.GetPellStrategyManagerProxyContractAddress(ctx)
	if err != nil {
		return nil, cosmoserrors.Wrapf(types.ErrContractNotFound, "GetPellStrategyManagerContractAddress address not found")
	}

	return k.CallEVM(
		ctx,
		*pellStrategyManagerMetaDataABI,
		types.ModuleAddressEVM,
		to,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"syncDepositState",
		chainID,
		staker,
		strategy,
		shares,
	)
}

func (k Keeper) callSyncDelegatedStateOnPellDelegationManager(
	ctx sdk.Context,
	chainID *big.Int,
	staker common.Address,
	operator common.Address,
) (*evmtypes.MsgEthereumTxResponse, error) {
	to, err := k.GetPellDelegationManagerProxyContractAddress(ctx)
	if err != nil {
		return nil, cosmoserrors.Wrapf(types.ErrContractNotFound, "GetPellDelegationManagerContractAddress address not found")
	}

	return k.CallEVM(
		ctx,
		*delegationManagerMetaDataABI,
		types.ModuleAddressEVM,
		to,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"syncDelegateState",
		chainID,
		staker,
		operator,
	)
}

func (k Keeper) callSyncWithdrawalStateOnPellDelegationManager(
	ctx sdk.Context,
	chainID *big.Int,
	staker common.Address,
	withdrawalParam *xmsgType.WithdrawalQueued,
) (*evmtypes.MsgEthereumTxResponse, error) {
	to, err := k.GetPellDelegationManagerProxyContractAddress(ctx)
	if err != nil {
		return nil, cosmoserrors.Wrapf(types.ErrContractNotFound, "GetPellDelegationManagerContractAddress address not found")
	}

	param, err := QueuedWithdrawalsToWithdrawalParams(withdrawalParam)
	if err != nil {
		return nil, err
	}

	return k.CallEVM(
		ctx,
		*delegationManagerMetaDataABI,
		types.ModuleAddressEVM,
		to,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"syncWithdrawalState",
		chainID,
		staker,
		common.HexToAddress(withdrawalParam.Withdrawal.DelegatedTo),
		param,
	)
}

func QueuedWithdrawalsToWithdrawalParams(queuedWithdrawals *xmsgType.WithdrawalQueued) (struct {
	Strategies []common.Address `json:"strategies" binding:"required"`
	Shares     []*big.Int       `json:"shares" binding:"required"`
}, error) {
	if queuedWithdrawals == nil || queuedWithdrawals.Withdrawal == nil {
		return struct {
			Strategies []common.Address `json:"strategies" binding:"required"`
			Shares     []*big.Int       `json:"shares" binding:"required"`
		}{}, nil
	}

	strategies := make([]common.Address, len(queuedWithdrawals.Withdrawal.Strategies))
	for j, strategy := range queuedWithdrawals.Withdrawal.Strategies {
		strategies[j] = common.HexToAddress(strategy)
	}

	shares := make([]*big.Int, len(queuedWithdrawals.Withdrawal.Shares))
	for j, shareStr := range queuedWithdrawals.Withdrawal.Shares {
		share, ok := new(big.Int).SetString(shareStr, 10)
		if !ok {
			return struct {
				Strategies []common.Address `json:"strategies" binding:"required"`
				Shares     []*big.Int       `json:"shares" binding:"required"`
			}{}, errors.New("invalid share value: " + shareStr)
		}
		shares[j] = share
	}

	return struct {
		Strategies []common.Address `json:"strategies" binding:"required"`
		Shares     []*big.Int       `json:"shares" binding:"required"`
	}{
		Strategies: strategies,
		Shares:     shares,
	}, nil
}

func (k Keeper) callSyncUndelegatedStateOnPellDelegationManager(
	ctx sdk.Context,
	chainID *big.Int,
	staker common.Address,
) (*evmtypes.MsgEthereumTxResponse, error) {
	to, err := k.GetPellDelegationManagerProxyContractAddress(ctx)
	if err != nil {
		return nil, cosmoserrors.Wrapf(types.ErrContractNotFound, "GetPellDelegationManagerContractAddress address not found")
	}

	return k.CallEVM(
		ctx,
		*delegationManagerMetaDataABI,
		types.ModuleAddressEVM,
		to,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"syncUndelegateState",
		chainID,
		staker,
	)
}

// CallBridgePellOnPellGatewayPEVM calls the contract
func (k Keeper) callBridgePellOnPellGateway(
	ctx sdk.Context,
	destinationChainId *big.Int,
	receiver common.Address,
	amount *big.Int,
) (*evmtypes.MsgEthereumTxResponse, error) {
	pellGatewayAddr, err := k.GetPellGatewayEVMContractAddress(ctx)
	if err != nil {
		return nil, err
	}

	return k.CallEVM(
		ctx,
		*gatewayPEVMMetaDataABI,
		types.ModuleAddressEVM,
		pellGatewayAddr,
		amount,
		types.PEVMGasLimit,
		true,
		false,
		"bridgePell",
		destinationChainId,
		receiver.Bytes(),
	)
}

// CallUpdateDestinationAddressOnPellGatewayPEVM calls the contract
func (k Keeper) callUpdateDestinationAddressOnPellGateway(
	ctx sdk.Context,
	chainId *big.Int,
	destinationAddress common.Address,
) (*evmtypes.MsgEthereumTxResponse, error) {
	pellGatewayAddr, err := k.GetPellGatewayEVMContractAddress(ctx)
	if err != nil {
		return nil, err
	}

	return k.CallEVM(
		ctx,
		*gatewayPEVMMetaDataABI,
		types.ModuleAddressEVM,
		pellGatewayAddr,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"updateDestinationAddress",
		chainId,
		destinationAddress[:],
	)
}

func (k Keeper) callUpdateSourceAddressOnPellGateway(
	ctx sdk.Context,
	chainId *big.Int,
	sourceAddress common.Address,
) (*evmtypes.MsgEthereumTxResponse, error) {
	pellGatewayAddr, err := k.GetPellGatewayEVMContractAddress(ctx)
	if err != nil {
		return nil, err
	}

	return k.CallEVM(
		ctx,
		*gatewayPEVMMetaDataABI,
		types.ModuleAddressEVM,
		pellGatewayAddr,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"updateSourceAddress",
		chainId,
		sourceAddress[:],
	)
}

// CallSwapOnPellGasSwapEVM calls the contract
func (k Keeper) callSwapOnPellGasSwap(
	ctx sdk.Context,
	destinationChainId *big.Int,
	amountIn *big.Int,
	receiver common.Address,
) (*evmtypes.MsgEthereumTxResponse, error) {
	contractAddr, err := k.GetGasSwapPEVMContractAddress(ctx)
	if err != nil {
		return nil, err
	}

	return k.CallEVM(
		ctx,
		*gasSwapPEVMMetaDataABI,
		types.ModuleAddressEVM,
		contractAddr,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"swap",
		destinationChainId,
		amountIn,
		big.NewInt(0), // TODO: amountOutMin
		big.NewInt(0), // TODO: amountOutMin
		receiver.Bytes(),
	)
}

// callUpdateDestinationAddressOnGasSwapPEVM calls the contract
func (k Keeper) callUpdateDestinationAddressOnGasSwapPEVM(
	ctx sdk.Context,
	chainId *big.Int,
	destinationAddress common.Address,
) (*evmtypes.MsgEthereumTxResponse, error) {
	contractAddr, err := k.GetGasSwapPEVMContractAddress(ctx)
	if err != nil {
		return nil, err
	}

	return k.CallEVM(
		ctx,
		*gasSwapPEVMMetaDataABI,
		types.ModuleAddressEVM,
		contractAddr,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"updateDestinationAddress",
		chainId,
		destinationAddress[:],
	)
}

// DeployContract deploys a new contract in the PEVM
func (k Keeper) DeployContract(ctx sdk.Context, metadata *bind.MetaData, ctorArguments ...interface{}) (common.Address, error) {
	contractABI, err := metadata.GetAbi()
	if err != nil {
		return common.Address{}, cosmoserrors.Wrapf(types.ErrABIGet, "failed to get  ABI: %s", err.Error())
	}
	ctorArgs, err := contractABI.Pack(
		"",               // function--empty string for constructor
		ctorArguments..., // feeToSetter
	)
	if err != nil {
		return common.Address{}, cosmoserrors.Wrapf(types.ErrABIGet, "failed to abi.Pack ctor arguments: %s", err.Error())
	}

	if len(metadata.Bin) <= 2 {
		return common.Address{}, cosmoserrors.Wrapf(types.ErrABIGet, "metadata Bin field too short")
	}

	bin, err := hex.DecodeString(metadata.Bin[2:])
	if err != nil {
		return common.Address{}, cosmoserrors.Wrapf(types.ErrABIPack, "error decoding %s hex bytecode string: %s", metadata.Bin[2:], err.Error())
	}

	data := make([]byte, len(bin)+len(ctorArgs))
	copy(data[:len(bin)], bin)
	copy(data[len(bin):], ctorArgs)

	nonce, err := k.authKeeper.GetSequence(ctx, types.ModuleAddress.Bytes())
	if err != nil {
		return common.Address{}, err
	}

	contractAddr := crypto.CreateAddress(types.ModuleAddressEVM, nonce)
	_, err = k.CallEVMWithData(ctx, types.ModuleAddressEVM, nil, data, true, false, types.BigIntZero, nil)
	if err != nil {
		return common.Address{}, cosmoserrors.Wrapf(err, "failed to deploy contract")
	}

	return contractAddr, nil
}

// CallEVM performs a smart contract method call using given args
// returns (msg,err) the EVM execution result if there is any, even if error is non-nil due to contract reverts
// Furthermore, err!=nil && msg!=nil && msg.Failed() means the contract call reverted.
func (k Keeper) CallEVM(
	ctx sdk.Context,
	abi abi.ABI,
	from, contract common.Address,
	value, gasLimit *big.Int,
	commit bool,
	noEthereumTxEvent bool,
	method string,
	args ...interface{},
) (*evmtypes.MsgEthereumTxResponse, error) {
	data, err := abi.Pack(method, args...)
	if err != nil {
		return nil, cosmoserrors.Wrap(
			types.ErrABIPack,
			cosmoserrors.Wrap(err, "failed to create transaction data").Error(),
		)
	}

	k.Logger(ctx).Debug("calling EVM", "from", from, "contract", contract, "value", value, "method", method)
	resp, err := k.CallEVMWithData(ctx, from, &contract, data, commit, noEthereumTxEvent, value, gasLimit)
	if err != nil {
		errMes := fmt.Sprintf("contract call failed: method '%s', contract '%s', args: %v", method, contract.Hex(), args)

		// if it is a revert error then add the revert reason to the error message
		revertErr, ok := err.(*evmtypes.RevertError)
		if ok {
			errMes = fmt.Sprintf("%s, reason: %v", errMes, revertErr.ErrorData())
		}
		return resp, cosmoserrors.Wrapf(err, errMes)
	}
	return resp, nil
}

// CallEVMWithData performs a smart contract method call using contract data
// value is the amount of wei to send; gaslimit is the custom gas limit, if nil EstimateGas is used
// to bisect the correct gas limit (this may sometimes result in insufficient gas limit; not sure why)
//
// returns (msg,err) the EVM execution result if there is any, even if error is non-nil due to contract reverts
// Furthermore, err!=nil && msg!=nil && msg.Failed() means the contract call reverted; in which case
// msg.Ret gives the RET code if contract revert with REVERT opcode with parameters.
func (k Keeper) CallEVMWithData(
	ctx sdk.Context,
	from common.Address,
	contract *common.Address,
	data []byte,
	commit bool,
	noEthereumTxEvent bool,
	value *big.Int,
	gasLimit *big.Int,
) (*evmtypes.MsgEthereumTxResponse, error) {
	nonce, err := k.authKeeper.GetSequence(ctx, from.Bytes())
	if err != nil {
		return nil, err
	}
	gasCap := config.DefaultGasCap
	if commit && gasLimit == nil {
		args, err := json.Marshal(evmtypes.TransactionArgs{
			From: &from,
			To:   contract,
			Data: (*hexutil.Bytes)(&data),
		})
		if err != nil {
			return nil, cosmoserrors.Wrapf(sdkerrors.ErrJSONMarshal, "failed to marshal tx args: %s", err.Error())
		}

		gasRes, err := k.evmKeeper.EstimateGas(sdk.WrapSDKContext(ctx), &evmtypes.EthCallRequest{
			Args:   args,
			GasCap: config.DefaultGasCap,
		})
		if err != nil {
			return nil, err
		}
		gasCap = gasRes.Gas
		k.Logger(ctx).Info("call evm", "EstimateGas", gasCap)
	}
	if gasLimit != nil {
		gasCap = gasLimit.Uint64()
	}

	msg := ethtypes.NewMessage(
		from,
		contract,
		nonce,
		value,         // amount
		gasCap,        // gasLimit
		big.NewInt(0), // gasFeeCap
		big.NewInt(0), // gasTipCap
		big.NewInt(0), // gasPrice
		data,
		ethtypes.AccessList{}, // AccessList
		!commit,               // isFake
	)
	k.evmKeeper.WithChainID(ctx) //FIXME:  set chainID for signer; should not need to do this; but seems necessary. Why?
	k.Logger(ctx).Debug("call evm", "gasCap", gasCap, "chainid", k.evmKeeper.ChainID(), "ctx.chainid", ctx.ChainID())
	res, err := k.evmKeeper.ApplyMessage(ctx, msg, evmtypes.NewNoOpTracer(), commit)
	if err != nil {
		return nil, err
	}

	if res.Failed() {
		return res, cosmoserrors.Wrap(evmtypes.ErrVMExecution, fmt.Sprintf("%s: ret 0x%x", res.VmError, res.Ret))
	}

	// Emit events and log for the transaction if it is committed
	if commit {
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			return nil, cosmoserrors.Wrap(err, "failed to encode msg")
		}
		ethTxHash := common.BytesToHash(crypto.Keccak256(msgBytes)) // NOTE(pwu): this is a fake txhash
		attrs := []sdk.Attribute{}
		if len(ctx.TxBytes()) > 0 {
			// add event for tendermint transaction hash format
			hash := tmbytes.HexBytes(tmtypes.Tx(ctx.TxBytes()).Hash())
			ethTxHash = common.BytesToHash(hash) // NOTE(pwu): use cosmos tx hash as eth tx hash if available
			attrs = append(attrs, sdk.NewAttribute(evmtypes.AttributeKeyTxHash, hash.String()))
		}
		attrs = append(attrs, []sdk.Attribute{
			sdk.NewAttribute(sdk.AttributeKeyAmount, value.String()),
			// add event for ethereum transaction hash format; NOTE(pwu): this is a fake txhash
			sdk.NewAttribute(evmtypes.AttributeKeyEthereumTxHash, ethTxHash.String()),
			// add event for index of valid ethereum tx; NOTE(pwu): fake txindex
			sdk.NewAttribute(evmtypes.AttributeKeyTxIndex, strconv.FormatUint(8888, 10)),
			// add event for eth tx gas used, we can't get it from cosmos tx result when it contains multiple eth tx msgs.
			sdk.NewAttribute(evmtypes.AttributeKeyTxGasUsed, strconv.FormatUint(res.GasUsed, 10)),
		}...)

		// recipient: contract address
		if contract != nil {
			attrs = append(attrs, sdk.NewAttribute(evmtypes.AttributeKeyRecipient, contract.Hex()))
		}
		if res.Failed() {
			attrs = append(attrs, sdk.NewAttribute(evmtypes.AttributeKeyEthereumTxFailed, res.VmError))
		}

		txLogAttrs := make([]sdk.Attribute, len(res.Logs))
		for i, log := range res.Logs {
			log.TxHash = ethTxHash.String()
			value, err := json.Marshal(log)
			if err != nil {
				return nil, cosmoserrors.Wrap(err, "failed to encode log")
			}
			txLogAttrs[i] = sdk.NewAttribute(evmtypes.AttributeKeyTxLog, string(value))
		}

		if !noEthereumTxEvent {
			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					evmtypes.EventTypeEthereumTx,
					attrs...,
				),
				sdk.NewEvent(
					evmtypes.EventTypeTxLog,
					txLogAttrs...,
				),
				sdk.NewEvent(
					sdk.EventTypeMessage,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(sdk.AttributeKeySender, from.Hex()),
					sdk.NewAttribute(evmtypes.AttributeKeyTxType, "88"), // type 88: synthetic Eth tx
				),
			})
		}

		logs := evmtypes.LogsToEthereum(res.Logs)
		var bloomReceipt ethtypes.Bloom
		if len(logs) > 0 {
			bloom := k.evmKeeper.GetBlockBloomTransient(ctx)
			bloom.Or(bloom, big.NewInt(0).SetBytes(ethtypes.LogsBloom(logs)))
			bloomReceipt = ethtypes.BytesToBloom(bloom.Bytes())
			k.evmKeeper.SetBlockBloomTransient(ctx, bloomReceipt.Big())
			k.evmKeeper.SetLogSizeTransient(ctx, (k.evmKeeper.GetLogSizeTransient(ctx))+uint64(len(logs)))
		}
	}

	return res, nil
}

func (k Keeper) callAddSupportedChainOnRegistryRouter(
	ctx sdk.Context,
	params *xmsgType.RegisterChainDVSToPell,
) (*evmtypes.MsgEthereumTxResponse, error) {
	dvsInfo := registryrouter.IRegistryRouterDVSInfo{
		ChainId:          new(big.Int).SetUint64(params.ChainId),
		CentralScheduler: common.HexToAddress(params.CentralScheduler),
		EjectionManager:  common.HexToAddress(params.EjectionManager),
		StakeManager:     common.HexToAddress(params.StakeManager),
	}
	signature := registryrouter.ISignatureUtilsSignatureWithSaltAndExpiry{
		Signature: params.DvsChainApproverSignature.Signature,
		Salt:      [32]byte(params.DvsChainApproverSignature.Salt),
		Expiry:    new(big.Int).SetUint64(params.DvsChainApproverSignature.Expiry),
	}
	return k.CallEVM(
		ctx,
		*registryRouterMetaDataABI,
		types.ModuleAddressEVM,
		common.HexToAddress(params.RegistryRouterOnPell),
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"addSupportedChain",
		dvsInfo,
		signature,
	)
}

func (k Keeper) callBeaconUpgrade(
	ctx sdk.Context,
	to, upgradeTo common.Address,
) (*evmtypes.MsgEthereumTxResponse, error) {
	return k.CallEVM(
		ctx,
		*upgradeableBeaconMetaDataABI,
		types.ModuleAddressEVM,
		to,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"upgradeTo",
		upgradeTo,
	)
}

func (k Keeper) callProcessPellSent(
	ctx sdk.Context,
	pellSent *xmsgType.PellSent,
	xmsgIndex string,
) (*evmtypes.MsgEthereumTxResponse, error) {
	if pellSent == nil {
		return nil, fmt.Errorf("no pell data to be sent")
	}

	callContractAddr, err := k.GetPellConnectorContractAddress(ctx)
	if err != nil {
		return nil, err
	}

	paramType, err := types.PellSentParamTypeFromString(pellSent.PellParams)
	if err != nil {
		return nil, fmt.Errorf("parse pellSent.PellParams error")
	}

	methodName, err := paramType.MethodName()
	if err != nil {
		return nil, fmt.Errorf("get method name error")
	}

	//  function onReceive(
	//    bytes calldata pellTxSenderAddress,
	//    uint256 sourceChainId,
	//    address destinationAddress,
	//    uint256 pellValue,
	//    bytes calldata message,
	//    bytes32 internalSendHash
	//  )
	// sender not pellSent's sender param
	//pellTxSenderAddress := common.HexToAddress(pellSent.Sender).Bytes()
	// sender is pell connector
	pellTxSenderAddress := callContractAddr.Bytes()
	sourceChainIdInt, err := chains.CosmosToEthChainID(ctx.ChainID())
	if err != nil {
		return nil, fmt.Errorf("get source chain id error")
	}

	sourceChainId := big.NewInt(sourceChainIdInt)
	destinationAddress := common.HexToAddress(pellSent.Receiver)
	pellValue := pellSent.PellValue.BigInt()
	message, err := base64.StdEncoding.DecodeString(pellSent.Message)
	if err != nil {
		return nil, fmt.Errorf("decode pellSent.Message %s error", pellSent.Message)
	}

	xmsgIndexSlice, err := hex.DecodeString(xmsgIndex[2:])
	if err != nil || len(xmsgIndexSlice) != 32 {
		return nil, fmt.Errorf("unable to decode xmsg index %s", xmsgIndex)
	}

	var internalSendHash [32]byte
	copy(internalSendHash[:32], xmsgIndexSlice[:32])

	k.Logger(ctx).Info("pevm call pell connector by pellSent event", "from", types.ModuleAddressEVM,
		"to", callContractAddr, "value", pellSent.PellValue.BigInt(), "destinationAddress", destinationAddress,
		"xmsgIndex", xmsgIndex)

	return k.CallEVM(
		ctx,
		*pellConnectorMetaDataABI,
		types.ModuleAddressEVM,
		callContractAddr,
		pellSent.PellValue.BigInt(),
		types.PEVMGasLimit,
		true,
		false,
		methodName,
		pellTxSenderAddress,
		sourceChainId,
		destinationAddress,
		pellValue,
		message,
		internalSendHash,
	)
}

// ---------------- LST Token staking ----------------

// callRegistryRouterFactory calls the contract
func (k Keeper) callRegistryRouterFactory(
	ctx sdk.Context,
	dvsChainApprover, churnApprover, ejector, pauser, unpauser common.Address,
	initialPausedStatus uint,
) (*evmtypes.MsgEthereumTxResponse, error) {
	contractAddr, err := k.GetRegistryRouterFactoryContractAddress(ctx)
	if err != nil {
		return nil, err
	}

	//function createRegistryRouter(
	//	address _initialDVSOwner,
	//	address _dvsChainApprover,
	//	address _churnApprover,
	//	address _ejector,
	//	address[] memory _pausers,
	//	address _unpauser,
	//	uint256 _initialPausedStatus
	//) external returns (address, address) {
	pausers := []common.Address{pauser}
	initialPausedStatusBigInt := new(big.Int).SetUint64(uint64(initialPausedStatus))

	return k.CallEVM(
		ctx,
		*registryRouterFactoryMetaDataABI,
		types.ModuleAddressEVM,
		contractAddr,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"createRegistryRouter",
		types.ModuleAddressEVM,
		dvsChainApprover,
		churnApprover,
		ejector,
		pausers,
		unpauser,
		initialPausedStatusBigInt,
	)
}

// callRegistryRouterToCreateGroup calls the registry router contract and creates a group
func (k Keeper) callRegistryRouterToCreateGroup(
	ctx sdk.Context,
	registryRouterAddress common.Address,
	operatorSetParams restakingtypes.OperatorSetParam,
	minimumStake int64,
	poolParams []restakingtypes.PoolParams,
	groupEjectionParams restakingtypes.GroupEjectionParam,
) (*evmtypes.MsgEthereumTxResponse, error) {

	//function createGroup(
	//	OperatorSetParam memory operatorSetParams,
	//	uint96 minimumStake,
	//	IStakeRegistryRouter.PoolParams[] memory poolParams,
	//	GroupEjectionParams memory _groupEjectionParams
	//) external virtual onlyOwner {
	operatorSetParamsEVM := registryrouter.IRegistryRouterOperatorSetParam{
		MaxOperatorCount:        operatorSetParams.MaxOperatorCount,
		KickBIPsOfOperatorStake: uint16(operatorSetParams.KickBipsOfOperatorStake),
		KickBIPsOfTotalStake:    uint16(operatorSetParams.KickBipsOfTotalStake),
	}
	minimumStakeEVM := big.NewInt(minimumStake)
	poolParamsEVM := make([]registryrouter.IStakeRegistryRouterPoolParams, len(poolParams))
	for i, poolParam := range poolParams {
		poolParamsEVM[i] = registryrouter.IStakeRegistryRouterPoolParams{
			ChainId:    big.NewInt(int64(poolParam.ChainId)),
			Pool:       common.HexToAddress(poolParam.Pool),
			Multiplier: big.NewInt(int64(poolParam.Multiplier)),
		}
	}
	groupEjectionParamsEVM := registryrouter.IRegistryRouterGroupEjectionParams{
		RateLimitWindow:       groupEjectionParams.RateLimitWindow,
		EjectableStakePercent: uint16(groupEjectionParams.EjectableStakePercent),
	}

	return k.CallEVM(
		ctx,
		*registryRouterMetaDataABI,
		types.ModuleAddressEVM,
		registryRouterAddress,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"createGroup",
		operatorSetParamsEVM,
		minimumStakeEVM,
		poolParamsEVM,
		groupEjectionParamsEVM,
	)
}

// callRegistryRouterToRegisterOperator calls the registry router contract and registers an operator
func (k Keeper) callRegistryRouterToRegisterOperator(
	ctx sdk.Context,
	registryRouterAddress common.Address,
	param xsecuritytypes.RegisterOperatorParam,
	operatorAddress common.Address,
	groupNumbers uint64,
) (*evmtypes.MsgEthereumTxResponse, error) {

	//function registerOperator(
	//	bytes calldata groupNumbers,
	//	string calldata socket,
	//	PubkeyRegistrationParams calldata params,
	//	SignatureWithSaltAndExpiry memory operatorSignature
	//) external onlyWhenNotPaused(PAUSED_REGISTER_OPERATOR) {
	groupNumbersBytes := []byte{byte(groupNumbers)}

	pubKey := ConvertPubkeyRegistrationParamsFromStore(param.PubkeyParams)

	sign := registryrouter.ISignatureUtilsSignatureWithSaltAndExpiry{
		Signature: param.Signature.Signature,
		Salt:      [32]byte(param.Signature.Salt),
		Expiry:    big.NewInt(int64(param.Signature.Expiry)),
	}

	return k.CallEVM(
		ctx,
		*registryRouterMetaDataABI,
		operatorAddress,
		registryRouterAddress,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"registerOperator",
		groupNumbersBytes,
		param.Socket,
		pubKey,
		sign,
	)
}

// ConvertPubkeyRegistrationParamsFromStore converts the pubkey registration params from store to the format used by the registry router
func ConvertPubkeyRegistrationParamsFromStore(params *xsecuritytypes.PubkeyRegistrationParams) registryrouter.IRegistryRouterPubkeyRegistrationParams {
	return registryrouter.IRegistryRouterPubkeyRegistrationParams{
		PubkeyRegistrationSignature: registryrouter.BN254G1Point{
			X: params.PubkeyRegistrationSignature.X.BigInt(),
			Y: params.PubkeyRegistrationSignature.Y.BigInt(),
		},
		PubkeyG1: registryrouter.BN254G1Point{
			X: params.PubkeyG1.X.BigInt(),
			Y: params.PubkeyG1.Y.BigInt(),
		},
		PubkeyG2: registryrouter.BN254G2Point{
			X: [2]*big.Int{
				params.PubkeyG2.X[0].BigInt(),
				params.PubkeyG2.X[1].BigInt(),
			},
			Y: [2]*big.Int{
				params.PubkeyG2.Y[0].BigInt(),
				params.PubkeyG2.Y[1].BigInt(),
			},
		},
	}
}

// callRegistryRouterToAddPools calls the registry router contract and adds pools
func (k Keeper) callStakeRegistryRouterToAddPools(
	ctx sdk.Context,
	stakeRegistryRouterAddress common.Address,
	groupNumbers uint64,
	poolParams []*restakingtypes.PoolParams,
) (*evmtypes.MsgEthereumTxResponse, error) {

	// function addPools(uint8 groupNumber, PoolParams[] memory _poolParams) public onlyRegitryRouterOwner {
	groupNumbersUint8 := uint8(groupNumbers)

	poolParamsEVM := make([]registryrouter.IStakeRegistryRouterPoolParams, len(poolParams))
	for i, poolParam := range poolParams {
		poolParamsEVM[i] = registryrouter.IStakeRegistryRouterPoolParams{
			ChainId:    big.NewInt(int64(poolParam.ChainId)),
			Pool:       common.HexToAddress(poolParam.Pool),
			Multiplier: big.NewInt(int64(poolParam.Multiplier)),
		}
	}

	return k.CallEVM(
		ctx,
		*stakeRegistryRouterMetaDataABI,
		types.ModuleAddressEVM,
		stakeRegistryRouterAddress,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"addPools",
		groupNumbersUint8,
		poolParamsEVM,
	)
}

// callStakeRegistryRouterToRemovePools calls the registry router contract and removes pools
func (k Keeper) callStakeRegistryRouterToRemovePools(
	ctx sdk.Context,
	stakeRegistryRouterAddress common.Address,
	groupNumbers uint64,
	indicesToRemove []uint,
) (*evmtypes.MsgEthereumTxResponse, error) {

	//   function removePools(uint8 groupNumber, uint256[] memory indicesToRemove) public onlyRegitryRouterOwner {
	groupNumbersUint8 := uint8(groupNumbers)

	indicesToRemoveEVM := make([]*big.Int, len(indicesToRemove))
	for i, index := range indicesToRemove {
		indicesToRemoveEVM[i] = big.NewInt(int64(index))
	}

	return k.CallEVM(
		ctx,
		*stakeRegistryRouterMetaDataABI,
		types.ModuleAddressEVM,
		stakeRegistryRouterAddress,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"removePools",
		groupNumbersUint8,
		indicesToRemoveEVM,
	)
}

// callRegistryRouterToSetOperatorSetParams calls the registry router contract and sets operator set params
func (k Keeper) callRegistryRouterToSetOperatorSetParams(
	ctx sdk.Context,
	registryRouterAddress common.Address,
	groupNumbers uint64,
	operatorSetParams *restakingtypes.OperatorSetParam,
) (*evmtypes.MsgEthereumTxResponse, error) {

	//   function removePools(uint8 groupNumber, uint256[] memory indicesToRemove) public onlyRegitryRouterOwner {
	groupNumbersUint8 := uint8(groupNumbers)
	operatorSetParamsEVM := registryrouter.IRegistryRouterOperatorSetParam{
		MaxOperatorCount:        operatorSetParams.MaxOperatorCount,
		KickBIPsOfOperatorStake: uint16(operatorSetParams.KickBipsOfOperatorStake),
		KickBIPsOfTotalStake:    uint16(operatorSetParams.KickBipsOfTotalStake),
	}

	return k.CallEVM(
		ctx,
		*registryRouterMetaDataABI,
		types.ModuleAddressEVM,
		registryRouterAddress,
		types.BigIntZero,
		types.PEVMGasLimit,
		true,
		false,
		"setOperatorSetParams",
		groupNumbersUint8,
		operatorSetParamsEVM,
	)
}
