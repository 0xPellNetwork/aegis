package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	sdkmath "cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/pell-chain/pellcore/e2e/txserver"
	"github.com/pell-chain/pellcore/e2e/utils"
	pevmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	relayertypes "github.com/pell-chain/pellcore/x/relayer/types"
	restakingtypes "github.com/pell-chain/pellcore/x/restaking/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
	xsecuritytypes "github.com/pell-chain/pellcore/x/xsecurity/types"
)

// Waiting for the network to be ready
const MAX_WAITING_PREPARE_TIME = 90

// update multi chain param on pell
func (r *Runner) updateChainParamsWith(updateFunc func(*relayertypes.ChainParams, *EvmContracts)) {
	// update chain params
	for _, v := range relayertypes.GetDefaultChainParams().ChainParams {
		evmChain, ok := r.MultiEVM[v.ChainId]
		if !ok {
			continue
		}

		v.GatewayEvmContractAddress = evmChain.EvmContracts.GatewayContractAddr.String()
		v.GasSwapContractAddress = evmChain.EvmContracts.GasSwapContractAddr.String()
		v.OmniOperatorSharesManagerContractAddress = evmChain.EvmContracts.OmniOperatorSharesManagerAddr.Hex()

		updateFunc(v, &evmChain.EvmContracts)

		msg := relayertypes.NewMsgUpsertChainParams(r.TxServer.GetAccountAddress(0), v)
		pellTx, err := r.TxServer.BroadcastTx(utils.FungibleAdminName, msg)

		utils.Assert(err == nil, err)
		utils.Assert(pellTx.Code == 0, fmt.Sprintf("NewMsgUpdateAndSyncChainParams failed. chain_id: %d tx_hash: %s", v.ChainId, pellTx.TxHash))

		outboundStateMsg := restakingtypes.MsgUpsertOutboundState{
			Signer: r.TxServer.GetAccountAddress(0),
			OutboundState: &restakingtypes.EpochOutboundState{
				ChainId:        uint64(v.ChainId),
				OutboundStatus: restakingtypes.OutboundStatus_OUTBOUND_STATUS_NORMAL,
				EpochNumber:    0,
			},
		}

		pellTx, err = r.TxServer.BroadcastTx(utils.FungibleAdminName, &outboundStateMsg)

		utils.Assert(err == nil, err)
		utils.Assert(pellTx.Code == 0, fmt.Sprintf("MsgUpsertOutboundState failed. chain_id: %d tx_hash: %s", v.ChainId, pellTx.TxHash))
	}
}

func (r *Runner) UpdateChainParamsEnablePellTokenRecharge() {
	r.updateChainParamsWith(func(v *relayertypes.ChainParams, evm *EvmContracts) {
		v.OmniOperatorSharesManagerContractAddress = evm.OmniOperatorSharesManagerAddr.Hex()
		v.PellTokenRechargeEnabled = true
		v.IsSupported = true
	})
}

func (r *Runner) UpdateChainParamsEnableGasTokenRecharge() {
	r.updateChainParamsWith(func(v *relayertypes.ChainParams, evm *EvmContracts) {
		v.OmniOperatorSharesManagerContractAddress = evm.OmniOperatorSharesManagerAddr.Hex()
		v.GasTokenRechargeEnabled = true
		v.IsSupported = true
	})
}

func (r *Runner) UpdateChainParamsLargerGasLimit() {
	r.updateChainParamsWith(func(v *relayertypes.ChainParams, evm *EvmContracts) {
		v.OmniOperatorSharesManagerContractAddress = evm.OmniOperatorSharesManagerAddr.Hex()
		v.IsSupported = true
		v.GasLimit = 20000000
	})
}

func (r *Runner) fillTxServer(pellCoreRpc, adminMnemonic, chain_id string) error {
	var err error
	r.TxServer, err = txserver.NewTxServer(pellCoreRpc, []string{utils.FungibleAdminName}, []string{adminMnemonic}, chain_id)
	if err != nil {
		return err
	}

	r.TxServerPellCore0, err = txserver.NewTxServer(pellCoreRpc, []string{utils.Pellcore0Name}, []string{utils.Pellcore0Mnemonic}, chain_id)
	return err
}

// fill tss address from pellcore observer. to be compatible with the localnet, retry here
func (r *Runner) fillTssAddr() error {
	r.Logger.Print("⚙️ setting up TSS address")
	var err error
	for i := 0; i <= MAX_WAITING_PREPARE_TIME; i++ {
		time.Sleep(time.Second)

		if i == MAX_WAITING_PREPARE_TIME {
			panic("network prepare timeout. get tss address failed")
		}

		res := &relayertypes.QueryGetTssAddressResponse{}
		res, err = r.PellClients.RelayerClient.GetTssAddress(context.Background(), &relayertypes.QueryGetTssAddressRequest{})
		if err != nil || res.Eth == "" {
			continue
		}

		r.TSSAddress = ethcommon.HexToAddress(res.Eth)
		break
	}

	return nil
}

// WaitForTxReceiptOnPEVM waits for a tx receipt on EVM
func (runner *Runner) WaitForTxReceiptOnPEVM(tx *ethtypes.Transaction) {
	runner.Lock()
	defer runner.Unlock()

	receipt := utils.MustWaitForTxReceipt(runner.Ctx, runner.PEVMClient, tx, runner.Logger, runner.ReceiptTimeout)
	if receipt.Status != 1 {
		panic("tx failed")
	}
}

func (r *Runner) QueryDVSSupportedChainStatus(registryRouterAddress string, chaindID uint64) *restakingtypes.QueryDVSSupportedChainStatusResponse {
	status, err := r.PellClients.RestakingClient.QueryDVSSupportedChainStatus(context.Background(), &restakingtypes.QueryDVSSupportedChainStatusRequest{
		RegistryRouterAddress: registryRouterAddress,
		ChainId:               chaindID,
	})

	utils.Assert(err == nil, err)
	return status
}

func (r *Runner) QueryDVSGroupSyncStatus(txHash string) (*restakingtypes.QueryDVSGroupSyncStatusResponse, error) {
	status, err := r.PellClients.RestakingClient.QueryDVSGroupSyncStatus(context.Background(), &restakingtypes.QueryDVSGroupSyncStatusRequest{
		TxHash: txHash,
	})
	if err != nil {
		return nil, err
	}

	return status, nil
}

func (r *Runner) UpgradeSystemContract() (*sdktypes.TxResponse, error) {
	msg := pevmtypes.MsgUpgradeSystemContracts{
		Signer: r.TxServer.GetAccountAddress(0),
	}

	return r.TxServer.BroadcastTx(utils.FungibleAdminName, &msg)
}

func (r *Runner) DeployGatewayContract() (*sdktypes.TxResponse, error) {
	msg := pevmtypes.MsgDeployGatewayContract{
		Signer: r.TxServer.GetAccountAddress(0),
	}

	return r.TxServer.BroadcastTx(utils.FungibleAdminName, &msg)
}

func (r *Runner) UpdateChainParamsGateway() {
	r.updateChainParamsWith(func(v *relayertypes.ChainParams, evm *EvmContracts) {
		v.IsSupported = true
	})
}

func (r *Runner) AddAllowXmsgSender(addr string) (*sdktypes.TxResponse, error) {
	msg := xmsgtypes.MsgAddAllowedXmsgSender{
		Signer: r.TxServer.GetAccountAddress(0),
		Builders: []string{
			addr,
		},
	}

	return r.TxServer.BroadcastTx(utils.FungibleAdminName, &msg)
}

func (r *Runner) DeployConnectorContract() (*sdktypes.TxResponse, error) {
	msg := pevmtypes.MsgDeployConnectorContract{
		Signer: r.TxServer.GetAccountAddress(0),
	}

	return r.TxServer.BroadcastTx(utils.FungibleAdminName, &msg)
}

// -------------- lst token staking --------------

// LSTUpdateLSTStakingEnabled send a tx to update LST staking enabled
func (r *Runner) LSTUpdateLSTStakingEnabled(enable bool) (*sdktypes.TxResponse, error) {
	msg := xsecuritytypes.MsgUpdateLSTStakingEnabled{
		Signer:  r.TxServer.GetAccountAddress(0),
		Enabled: enable,
	}

	return r.TxServer.BroadcastTx(utils.FungibleAdminName, &msg)
}

// LSTQueryLSTStakingEnabled query LST staking enabled
func (r *Runner) LSTQueryLSTStakingEnabled() (*xsecuritytypes.QueryLSTStakingEnabledResponse, error) {
	return r.PellClients.XSecurityClient.QueryLSTStakingEnabled(
		context.Background(),
		&xsecuritytypes.QueryLSTStakingEnabledRequest{},
	)
}

// LSTUpdateVotingPowerRatio send a tx to update voting power ratio
func (r *Runner) LSTUpdateVotingPowerRatio(numerator, denominator uint64) (*sdktypes.TxResponse, error) {
	msg := xsecuritytypes.MsgUpdateVotingPowerRatio{
		Signer:      r.TxServer.GetAccountAddress(0),
		Numerator:   sdkmath.NewInt(int64(numerator)),
		Denominator: sdkmath.NewInt(int64(denominator)),
	}

	return r.TxServer.BroadcastTx(utils.FungibleAdminName, &msg)
}

// LSTQueryVotingPowerRatio query voting power ratio
func (r *Runner) LSTQueryVotingPowerRatio() (*xsecuritytypes.QueryVotingPowerRatioResponse, error) {
	return r.PellClients.XSecurityClient.QueryVotingPowerRatio(
		context.Background(),
		&xsecuritytypes.QueryVotingPowerRatioRequest{},
	)
}

// LSTCreateRegistryRouter send a tx to create registry router
func (r *Runner) LSTCreateRegistryRouter(chainApprover, churnApprover, ejector, pauser, unpauser string, initialPausedStatus int64) (*sdktypes.TxResponse, error) {
	msg := xsecuritytypes.MsgCreateRegistryRouter{
		Signer:              r.TxServer.GetAccountAddress(0),
		ChainApprover:       chainApprover,
		ChurnApprover:       churnApprover,
		Ejector:             ejector,
		Pauser:              pauser,
		Unpauser:            unpauser,
		InitialPausedStatus: initialPausedStatus,
	}

	return r.TxServer.BroadcastTx(utils.FungibleAdminName, &msg)
}

// LSTQueryRegistryRouter query registry router address
func (r *Runner) LSTQueryRegistryRouter() (*xsecuritytypes.QueryRegistryRouterAddressResponse, error) {
	return r.PellClients.XSecurityClient.QueryRegistryRouterAddress(context.Background(), &xsecuritytypes.QueryRegistryRouterAddressRequest{})
}

// LSTCreateGroup send a tx to create a DVS group
func (r *Runner) LSTCreateGroup(operatorSetParam *restakingtypes.OperatorSetParam, poolParams []*restakingtypes.PoolParams,
	groupEjectionParam *restakingtypes.GroupEjectionParam, minStake int64) (*sdktypes.TxResponse, error) {

	msg := xsecuritytypes.MsgCreateGroup{
		Signer:              r.TxServer.GetAccountAddress(0),
		OperatorSetParams:   operatorSetParam,
		PoolParams:          poolParams,
		GroupEjectionParams: groupEjectionParam,
		MinStake:            sdkmath.NewInt(minStake),
	}

	return r.TxServer.BroadcastTx(utils.FungibleAdminName, &msg)
}

// LSTQueryGroupInfo query group info
func (r *Runner) LSTQueryGroupInfo() (*xsecuritytypes.QueryGroupInfoResponse, error) {
	return r.PellClients.XSecurityClient.QueryGroupInfo(context.Background(), &xsecuritytypes.QueryGroupInfoRequest{})
}

// LSTRegisterOperator send a tx to register an operator
func (r *Runner) LSTRegisterOperator(param *xsecuritytypes.RegisterOperatorParam, operatorAddress string) (*sdktypes.TxResponse, error) {
	msg := xsecuritytypes.MsgRegisterOperator{
		Signer:                r.TxServerPellCore0.GetAccountAddress(0),
		OperatorAddress:       operatorAddress,
		RegisterOperatorParam: param,
	}
	return r.TxServerPellCore0.BroadcastTx(utils.Pellcore0Name, &msg)
}

type Validator struct {
	Address     string `json:"address"`
	PubKey      any    `json:"pub_key"`
	VotingPower string `json:"voting_power"`
}

type Result struct {
	Validators []Validator `json:"validators"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  Result `json:"result"`
}

// LSTQueryValidator query validator info from staking module
func (r *Runner) LSTQueryValidator() ([]Validator, error) {
	url := r.PellClients.RPCClientURL + "/validators"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch validators: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	var data Response
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return data.Result.Validators, nil
}

// LSTAddPools send a tx to add pools
func (r *Runner) LSTAddPools(poolParams []*restakingtypes.PoolParams) (*sdktypes.TxResponse, error) {
	msg := xsecuritytypes.MsgAddPools{
		Signer: r.TxServer.GetAccountAddress(0),
		Pools:  poolParams,
	}

	return r.TxServer.BroadcastTx(utils.FungibleAdminName, &msg)
}

// LSTRemovePools send a tx to remove pools
func (r *Runner) LSTRemovePools(poolParams []*restakingtypes.PoolParams) (*sdktypes.TxResponse, error) {
	msg := xsecuritytypes.MsgRemovePools{
		Signer: r.TxServer.GetAccountAddress(0),
		Pools:  poolParams,
	}

	return r.TxServer.BroadcastTx(utils.FungibleAdminName, &msg)
}

// LSTSetGroupParam send a tx to set group param
func (r *Runner) LSTSetGroupParam(param *restakingtypes.OperatorSetParam) (*sdktypes.TxResponse, error) {
	msg := xsecuritytypes.MsgSetGroupParam{
		Signer:            r.TxServer.GetAccountAddress(0),
		OperatorSetParams: param,
	}

	return r.TxServer.BroadcastTx(utils.FungibleAdminName, &msg)
}
