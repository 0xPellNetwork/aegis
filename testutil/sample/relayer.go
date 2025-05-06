package sample

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/pkg/cosmos"
	pellcrypto "github.com/0xPellNetwork/aegis/pkg/crypto"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func Ballot_pell(t *testing.T, index string) *types.Ballot {
	r := newRandFromStringSeed(t, index)

	return &types.Ballot{
		Index:                index,
		BallotIdentifier:     StringRandom(r, 32),
		VoterList:            []string{AccAddress(), AccAddress()},
		Votes:                []types.VoteType{types.VoteType_FAILURE_OBSERVATION, types.VoteType_SUCCESS_OBSERVATION},
		ObservationType:      types.ObservationType_EMPTY_OBSERVER_TYPE,
		BallotThreshold:      sdkmath.LegacyNewDec(1),
		BallotStatus:         types.BallotStatus_BALLOT_IN_PROGRESS,
		BallotCreationHeight: r.Int63(),
	}
}

func ObserverSet_pell(n int) types.RelayerSet {
	observerList := make([]string, n)
	for i := 0; i < n; i++ {
		observerList[i] = AccAddress()
	}

	return types.RelayerSet{
		RelayerList: observerList,
	}
}

func NodeAccount_pell() *types.NodeAccount {
	return &types.NodeAccount{
		Operator:       AccAddress(),
		GranteeAddress: AccAddress(),
		GranteePubkey:  PubKeySet(),
		NodeStatus:     types.NodeStatus_ACTIVE,
	}
}

func CrosschainFlags_pell() *types.CrosschainFlags {
	return &types.CrosschainFlags{
		IsInboundEnabled:  true,
		IsOutboundEnabled: true,
	}
}

func Keygen_pell(t *testing.T) *types.Keygen {
	pubKey := ed25519.GenPrivKey().PubKey().String()
	r := newRandFromStringSeed(t, pubKey)

	return &types.Keygen{
		Status:         types.KeygenStatus_SUCCESS,
		GranteePubkeys: []string{pubKey},
		BlockNumber:    r.Int63(),
	}
}

func LastObserverCount_pell(lastChangeHeight int64) *types.LastRelayerCount {
	r := newRandFromSeed(lastChangeHeight)

	return &types.LastRelayerCount{
		Count:            r.Uint64(),
		LastChangeHeight: lastChangeHeight,
	}
}

func ChainParams_pell(chainID int64) *types.ChainParams {
	r := newRandFromSeed(chainID)

	fiftyPercent, err := sdkmath.LegacyNewDecFromStr("0.5")
	if err != nil {
		return nil
	}

	return &types.ChainParams{
		ChainId:           chainID,
		ConfirmationCount: r.Uint64(),

		GasPriceTicker:                           Uint64InRange(1, 300),
		InTxTicker:                               Uint64InRange(1, 300),
		OutTxTicker:                              Uint64InRange(1, 300),
		StrategyManagerContractAddress:           EthAddress().String(),
		ConnectorContractAddress:                 EthAddress().String(),
		DelegationManagerContractAddress:         EthAddress().String(),
		OmniOperatorSharesManagerContractAddress: EthAddress().String(),
		OutboundTxScheduleInterval:               Int64InRange(1, 100),
		OutboundTxScheduleLookahead:              Int64InRange(1, 500),
		BallotThreshold:                          fiftyPercent,
		MinObserverDelegation:                    sdkmath.LegacyNewDec(r.Int63()),
		IsSupported:                              false,
	}
}

func ChainParamsSupported_pell(chainID int64) *types.ChainParams {
	cp := ChainParams_pell(chainID)
	cp.IsSupported = true
	return cp
}

func ChainParamsList_pell() (cpl types.ChainParamsList) {
	for _, chain := range chains.PrivnetChainList() {
		cpl.ChainParams = append(cpl.ChainParams, ChainParams_pell(chain.Id))
	}

	return
}

func Tss_pell() types.TSS {
	_, pubKey, _ := testdata.KeyTestPubAddr()
	spk, err := cosmos.Bech32ifyPubKey(cosmos.Bech32PubKeyTypeAccPub, pubKey)
	if err != nil {
		panic(err)
	}
	pk, err := pellcrypto.NewPubKey(spk)
	if err != nil {
		panic(err)
	}
	pubkeyString := pk.String()
	return types.TSS{
		TssPubkey:           pubkeyString,
		FinalizedPellHeight: 1000,
		KeygenPellHeight:    1000,
	}
}

func TssList_pell(n int) (tssList []types.TSS) {
	for i := 0; i < n; i++ {
		tss := Tss_pell()
		tss.FinalizedPellHeight = tss.FinalizedPellHeight + int64(i)
		tss.KeygenPellHeight = tss.KeygenPellHeight + int64(i)
		tssList = append(tssList, tss)
	}
	return
}

func TssFundsMigrator_pell(chainID int64) types.TssFundMigratorInfo {
	return types.TssFundMigratorInfo{
		ChainId:            chainID,
		MigrationXmsgIndex: "sampleIndex",
	}
}

func BlameRecord_pell(t *testing.T, index string) types.Blame {
	r := newRandFromStringSeed(t, index)
	return types.Blame{
		Index:         fmt.Sprintf("%d-%s", r.Int63(), index),
		FailureReason: "sample failure reason",
		Nodes:         nil,
	}
}

func BlameRecordsList_pell(t *testing.T, n int) []types.Blame {
	blameList := make([]types.Blame, n)
	for i := 0; i < n; i++ {
		blameList[i] = BlameRecord_pell(t, fmt.Sprint(i))
	}
	return blameList
}

func ChainNonces_pell(t *testing.T, index string) types.ChainNonces {
	r := newRandFromStringSeed(t, index)
	return types.ChainNonces{
		Signer:          AccAddress(),
		Index:           index,
		ChainId:         r.Int63(),
		Nonce:           r.Uint64(),
		Signers:         []string{AccAddress(), AccAddress()},
		FinalizedHeight: r.Uint64(),
	}
}

func ChainNoncesList_pell(t *testing.T, n int) []types.ChainNonces {
	chainNoncesList := make([]types.ChainNonces, n)
	for i := 0; i < n; i++ {
		chainNoncesList[i] = ChainNonces_pell(t, fmt.Sprint(i))
	}
	return chainNoncesList
}

func PendingNoncesList_pell(t *testing.T, index string, count int) []types.PendingNonces {
	r := newRandFromStringSeed(t, index)
	nonceLow := r.Int63()
	list := make([]types.PendingNonces, count)
	for i := 0; i < count; i++ {
		list[i] = types.PendingNonces{
			ChainId:   int64(i),
			NonceLow:  nonceLow,
			NonceHigh: nonceLow + r.Int63(),
			Tss:       StringRandom(r, 32),
		}
	}
	return list
}

func NonceToXmsgList_pell(t *testing.T, index string, count int) []types.NonceToXmsg {
	r := newRandFromStringSeed(t, index)
	list := make([]types.NonceToXmsg, count)
	for i := 0; i < count; i++ {
		list[i] = types.NonceToXmsg{
			ChainId:   int64(i),
			Nonce:     r.Int63(),
			XmsgIndex: StringRandom(r, 32),
		}
	}
	return list
}

func LegacyObserverMapper_pell(t *testing.T, index string, observerList []string) *types.RelayerMapper {
	r := newRandFromStringSeed(t, index)

	return &types.RelayerMapper{
		Index:        index,
		RelayerChain: Chain(r.Int63()),
		RelayerList:  observerList,
	}
}

func LegacyObserverMapperList_pell(t *testing.T, n int, index string) []*types.RelayerMapper {
	r := newRandFromStringSeed(t, index)
	observerList := []string{AccAddress(), AccAddress()}
	observerMapperList := make([]*types.RelayerMapper, n)
	for i := 0; i < n; i++ {
		observerMapperList[i] = LegacyObserverMapper_pell(t, fmt.Sprintf("%d-%s", r.Int63(), index), observerList)
	}
	return observerMapperList
}

func BallotList_pell(n int, observerSet []string) []types.Ballot {
	r := newRandFromSeed(int64(n))
	ballotList := make([]types.Ballot, n)

	for i := 0; i < n; i++ {
		identifier := crypto.Keccak256Hash([]byte(fmt.Sprintf("%d-%d-%d", r.Int63(), r.Int63(), r.Int63())))
		ballotList[i] = types.Ballot{
			Index:                identifier.Hex(),
			BallotIdentifier:     identifier.Hex(),
			VoterList:            observerSet,
			Votes:                VotesSuccessOnly_pell(len(observerSet)),
			ObservationType:      types.ObservationType_IN_BOUND_TX,
			BallotThreshold:      sdkmath.LegacyOneDec(),
			BallotStatus:         types.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION,
			BallotCreationHeight: 0,
		}
	}
	return ballotList
}

func VotesSuccessOnly_pell(voteCount int) []types.VoteType {
	votes := make([]types.VoteType, voteCount)
	for i := 0; i < voteCount; i++ {
		votes[i] = types.VoteType_SUCCESS_OBSERVATION
	}
	return votes
}

func NonceToXmsg_pell(t *testing.T, seed string) types.NonceToXmsg {
	r := newRandFromStringSeed(t, seed)
	return types.NonceToXmsg{
		ChainId:   r.Int63(),
		Nonce:     r.Int63(),
		XmsgIndex: StringRandom(r, 64),
		Tss:       Tss_pell().TssPubkey,
	}
}
