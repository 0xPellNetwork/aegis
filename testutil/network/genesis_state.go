package network

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/require"

	cmdcfg "github.com/0xPellNetwork/aegis/cmd/pellcored/config"
	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/testutil/nullify"
	authoritytypes "github.com/0xPellNetwork/aegis/x/authority/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func SetupPellGenesisState(t *testing.T, genesisState map[string]json.RawMessage, codec codec.Codec, observerList []string, setupChainNonces bool) {
	// Cross-chain genesis state
	var xmsgGenesis types.GenesisState
	require.NoError(t, codec.UnmarshalJSON(genesisState[types.ModuleName], &xmsgGenesis))
	nodeAccountList := make([]*relayertypes.NodeAccount, len(observerList))
	for i, operator := range observerList {
		nodeAccountList[i] = &relayertypes.NodeAccount{
			Operator:   operator,
			NodeStatus: relayertypes.NodeStatus_ACTIVE,
		}
	}

	require.NoError(t, xmsgGenesis.Validate())
	xmsgGenesisBz, err := codec.MarshalJSON(&xmsgGenesis)
	require.NoError(t, err)

	// EVM genesis state
	var evmGenesisState evmtypes.GenesisState
	require.NoError(t, codec.UnmarshalJSON(genesisState[evmtypes.ModuleName], &evmGenesisState))
	evmGenesisState.Params.EvmDenom = cmdcfg.BaseDenom
	require.NoError(t, evmGenesisState.Validate())
	evmGenesisBz, err := codec.MarshalJSON(&evmGenesisState)
	require.NoError(t, err)

	// Staking genesis state
	var stakingGenesisState stakingtypes.GenesisState
	require.NoError(t, codec.UnmarshalJSON(genesisState[stakingtypes.ModuleName], &stakingGenesisState))
	stakingGenesisState.Params.BondDenom = cmdcfg.BaseDenom
	stakingGenesisStateBz, err := codec.MarshalJSON(&stakingGenesisState)
	require.NoError(t, err)

	// Observer genesis state
	var observerGenesis relayertypes.GenesisState
	require.NoError(t, codec.UnmarshalJSON(genesisState[relayertypes.ModuleName], &observerGenesis))
	observerSet := relayertypes.RelayerSet{
		RelayerList: observerList,
	}

	privnetChainList := chains.PrivnetChainList()
	if setupChainNonces {
		chainNonceList := make([]relayertypes.ChainNonces, len(privnetChainList))
		for i, chain := range privnetChainList {
			chainNonceList[i] = relayertypes.ChainNonces{
				Index:   chain.ChainName(),
				ChainId: chain.Id,
				Nonce:   0,
			}
		}
		observerGenesis.ChainNonces = chainNonceList
	}

	observerGenesis.Observers = observerSet
	observerGenesis.NodeAccountList = nodeAccountList
	observerGenesis.Keygen = &relayertypes.Keygen{
		Status:         relayertypes.KeygenStatus_PENDING,
		GranteePubkeys: observerList,
		BlockNumber:    5,
	}
	observerGenesis.CrosschainFlags = &relayertypes.CrosschainFlags{
		IsInboundEnabled:  true,
		IsOutboundEnabled: true,
	}
	require.NoError(t, observerGenesis.Validate())
	observerGenesisBz, err := codec.MarshalJSON(&observerGenesis)
	require.NoError(t, err)

	// authority genesis state
	var authorityGenesis authoritytypes.GenesisState
	require.NoError(t, codec.UnmarshalJSON(genesisState[authoritytypes.ModuleName], &authorityGenesis))
	policies := authoritytypes.Policies{
		Items: []*authoritytypes.Policy{
			{
				PolicyType: authoritytypes.PolicyType_GROUP_EMERGENCY,
				Address:    "pell1xmcw3ckru3mgqy8ztk3hy8tm9pqxhpaq3q0tgm",
			},
			{
				PolicyType: authoritytypes.PolicyType_GROUP_ADMIN,
				Address:    "pell1xmcw3ckru3mgqy8ztk3hy8tm9pqxhpaq3q0tgm",
			},
			{
				PolicyType: authoritytypes.PolicyType_GROUP_OPERATIONAL,
				Address:    "pell1xmcw3ckru3mgqy8ztk3hy8tm9pqxhpaq3q0tgm",
			},
		},
	}
	authorityGenesis.Policies = policies
	require.NoError(t, authorityGenesis.Validate())
	authorityGenesisBz, err := codec.MarshalJSON(&authorityGenesis)
	require.NoError(t, err)

	genesisState[types.ModuleName] = xmsgGenesisBz
	genesisState[stakingtypes.ModuleName] = stakingGenesisStateBz
	genesisState[relayertypes.ModuleName] = observerGenesisBz
	genesisState[evmtypes.ModuleName] = evmGenesisBz
	genesisState[authoritytypes.ModuleName] = authorityGenesisBz
}

func AddObserverData(t *testing.T, n int, genesisState map[string]json.RawMessage, codec codec.Codec, ballots []*relayertypes.Ballot) *relayertypes.GenesisState {
	state := relayertypes.GenesisState{}
	require.NoError(t, codec.UnmarshalJSON(genesisState[relayertypes.ModuleName], &state))

	// set chain params with chains all enabled
	state.ChainParamsList = relayertypes.GetDefaultChainParams()
	for i := range state.ChainParamsList.ChainParams {
		state.ChainParamsList.ChainParams[i].IsSupported = true
	}

	// set params
	if len(ballots) > 0 {
		state.Ballots = ballots
	}
	state.Params.BallotMaturityBlocks = 3
	state.Keygen = &relayertypes.Keygen{BlockNumber: 10, GranteePubkeys: []string{}}

	// set tss
	tss := relayertypes.TSS{
		TssPubkey:           "tssPubkey",
		TssParticipantList:  []string{"tssParticipantList"},
		OperatorAddressList: []string{"operatorAddressList"},
		FinalizedPellHeight: 1,
		KeygenPellHeight:    1,
	}
	pendingNonces := make([]relayertypes.PendingNonces, len(chains.ChainsList()))
	for i, chain := range chains.ChainsList() {
		pendingNonces[i] = relayertypes.PendingNonces{
			ChainId:   chain.Id,
			NonceLow:  0,
			NonceHigh: 0,
			Tss:       tss.TssPubkey,
		}
	}
	state.Tss = &tss
	state.TssHistory = []relayertypes.TSS{tss}
	state.PendingNonces = pendingNonces

	// set crosschain flags
	crosschainFlags := &relayertypes.CrosschainFlags{
		IsInboundEnabled:             true,
		IsOutboundEnabled:            true,
		GasPriceIncreaseFlags:        &relayertypes.DefaultGasPriceIncreaseFlags,
		BlockHeaderVerificationFlags: &relayertypes.DefaultBlockHeaderVerificationFlags,
	}
	nullify.Fill(&crosschainFlags)
	state.CrosschainFlags = crosschainFlags

	for i := 0; i < n; i++ {
		state.ChainNonces = append(state.ChainNonces, relayertypes.ChainNonces{Signer: "ANY", Index: strconv.Itoa(i), Signers: []string{}})
	}

	// check genesis state validity
	require.NoError(t, state.Validate())

	// marshal genesis state
	buf, err := codec.MarshalJSON(&state)
	require.NoError(t, err)
	genesisState[relayertypes.ModuleName] = buf
	return &state
}

func AddXmsgData(t *testing.T, n int, genesisState map[string]json.RawMessage, codec codec.Codec) *types.GenesisState {
	state := types.GenesisState{}
	require.NoError(t, codec.UnmarshalJSON(genesisState[types.ModuleName], &state))
	// TODO : Fix add EVM balance to deploy contracts
	for i := 0; i < n; i++ {
		state.Xmsgs = append(state.Xmsgs, &types.Xmsg{
			Signer: "ANY",
			Index:  strconv.Itoa(i),
			XmsgStatus: &types.Status{
				Status:              types.XmsgStatus_PENDING_INBOUND,
				StatusMessage:       "",
				LastUpdateTimestamp: 0,
			},
			InboundTxParams:  &types.InboundTxParams{InboundTxHash: fmt.Sprintf("Hash-%d", i)},
			OutboundTxParams: []*types.OutboundTxParams{},
		})
	}

	for i := 0; i < n; i++ {
		state.GasPriceList = append(state.GasPriceList, &types.GasPrice{Signer: "ANY", ChainId: int64(i), Index: strconv.Itoa(i), Prices: []uint64{}, BlockNums: []uint64{}, Signers: []string{}})
	}
	for i := 0; i < n; i++ {
		state.LastBlockHeightList = append(state.LastBlockHeightList, &types.LastBlockHeight{Signer: "ANY", Index: strconv.Itoa(i)})
	}

	for i := 0; i < n; i++ {
		inTxTracker := types.InTxTracker{
			ChainId: 5,
			TxHash:  fmt.Sprintf("txHash-%d", i),
		}
		nullify.Fill(&inTxTracker)
		state.InTxTrackerList = append(state.InTxTrackerList, inTxTracker)
	}

	for i := 0; i < n; i++ {
		inTxHashToXmsg := types.InTxHashToXmsg{
			InTxHash: strconv.Itoa(i),
		}
		nullify.Fill(&inTxHashToXmsg)
		state.InTxHashToXmsgList = append(state.InTxHashToXmsgList, inTxHashToXmsg)
	}
	for i := 0; i < n; i++ {
		outTxTracker := types.OutTxTracker{
			Index:   fmt.Sprintf("%d-%d", i, i),
			ChainId: int64(i),
			Nonce:   uint64(i),
		}
		nullify.Fill(&outTxTracker)
		state.OutTxTrackerList = append(state.OutTxTrackerList, outTxTracker)
	}

	require.NoError(t, state.Validate())

	buf, err := codec.MarshalJSON(&state)
	require.NoError(t, err)
	genesisState[types.ModuleName] = buf
	return &state
}
