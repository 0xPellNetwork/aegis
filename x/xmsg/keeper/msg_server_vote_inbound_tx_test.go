package keeper_test

import (
	"encoding/hex"
	"errors"
	"math/big"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/testutil/sample"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	"github.com/0xPellNetwork/aegis/x/xmsg/keeper"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func setObservers(t *testing.T, k *keeper.Keeper, ctx sdk.Context, zk keepertest.PellKeepers) []string {
	validators, err := k.GetStakingKeeper().GetAllValidators(ctx)
	require.NoError(t, err)

	validatorAddressListFormatted := make([]string, len(validators))
	for i, validator := range validators {
		valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		require.NoError(t, err)
		addressTmp, err := sdk.AccAddressFromHexUnsafe(hex.EncodeToString(valAddr.Bytes()))
		require.NoError(t, err)
		validatorAddressListFormatted[i] = addressTmp.String()
	}

	// Add validator to the observer list for voting
	zk.ObserverKeeper.SetObserverSet(ctx, relayertypes.RelayerSet{
		RelayerList: validatorAddressListFormatted,
	})
	return validatorAddressListFormatted
}

// TODO: Complete the test cases
func TestKeeper_VoteOnObservedInboundTx(t *testing.T) {
	t.Run("successfully vote on evm deposit", func(t *testing.T) {
		k, ctx, sdkk, zk := keepertest.XmsgKeeper(t)

		msgServer := keeper.NewMsgServerImpl(*k)
		validatorList := setObservers(t, k, ctx, zk)
		to, from := int64(1337), int64(186)
		supportedChains := zk.ObserverKeeper.GetSupportedChains(ctx)
		for _, chain := range supportedChains {
			if chains.IsEVMChain(chain.Id) {
				from = chain.Id
			}
			if chains.IsPellChain(chain.Id) {
				to = chain.Id
			}
		}

		zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())

		msg := sample.InboundVote_pell(from, to)

		err := sdkk.EvmKeeper.SetAccount(ctx, ethcommon.HexToAddress(msg.Receiver), statedb.Account{
			Nonce:    0,
			Balance:  big.NewInt(0),
			CodeHash: crypto.Keccak256(nil),
		})
		require.NoError(t, err)

		for _, validatorAddr := range validatorList {
			blockProofMsg := types.NewMsgVoteInboundBlock(
				validatorAddr, uint64(msg.SenderChainId), 0, msg.InBlockHeight, "0x01", []*types.Event{
					{
						Index:     msg.EventIndex,
						TxHash:    msg.InTxHash,
						PellEvent: msg.PellTx,
						Digest:    msg.Digest(),
					},
				},
			)

			_, err := msgServer.VoteOnObservedInboundBlock(ctx, blockProofMsg)
			require.NoError(t, err)
		}

		for _, validatorAddr := range validatorList {
			msg.Signer = validatorAddr

			_, err = msgServer.VoteOnObservedInboundTx(
				ctx,
				&msg,
			)
			require.NoError(t, err)
		}

		ballot, _, _ := zk.ObserverKeeper.FindBallot(
			ctx,
			msg.Digest(),
			zk.ObserverKeeper.GetSupportedChainFromChainID(ctx, msg.SenderChainId),
			relayertypes.ObservationType_IN_BOUND_TX,
		)
		require.Equal(t, ballot.BallotStatus, relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION)
		xmsg, found := k.GetXmsg(ctx, msg.Digest())
		require.True(t, found)
		// TODO: will be aborted because system contract not deployed
		require.Equal(t, types.XmsgStatus_ABORTED, xmsg.XmsgStatus.Status)
		require.Equal(t, xmsg.InboundTxParams.TxFinalizationStatus, types.TxFinalizationStatus_EXECUTED)
	})

	// Test execution order:
	// 1. Block1 (contains msg1,2,3)
	// 2. msg2    pending
	// 3. msg1    execute msg1 -> msg2
	// 4. msg3    execute msg3
	// 5. Block2 (contains msg4,5)
	// 6. msg5    pending
	// 7. msg4    execute msg4 -> msg5
	t.Run("vote two block. Ensure sequential execution", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)

		// MsgServer for the xmsg keeper
		msgServer := keeper.NewMsgServerImpl(*k)

		// Set the chain ids we want to use to be valid
		params := relayertypes.DefaultParams()
		zk.ObserverKeeper.SetParams(
			ctx, params,
		)

		// Convert the validator address into a user address.
		validators, err := k.GetStakingKeeper().GetAllValidators(ctx)
		require.NoError(t, err)

		validatorAddress := validators[0].OperatorAddress
		valAddr, _ := sdk.ValAddressFromBech32(validatorAddress)
		addresstmp, _ := sdk.AccAddressFromHexUnsafe(hex.EncodeToString(valAddr.Bytes()))
		validatorAddr := addresstmp.String()

		// Add validator to the observer list for voting
		zk.ObserverKeeper.SetObserverSet(ctx, relayertypes.RelayerSet{
			RelayerList: []string{validatorAddr},
		})

		// Add tss to the observer keeper
		zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())

		bscTestChainParams := relayertypes.GetDefaultBscTestnetChainParams()
		localTestChainParam := relayertypes.GetDefaultPellPrivnetChainParams()
		bscTestChainParams.IsSupported = true
		localTestChainParam.IsSupported = true

		zk.ObserverKeeper.SetChainParamsList(ctx, relayertypes.ChainParamsList{
			ChainParams: []*relayertypes.ChainParams{bscTestChainParams, localTestChainParam},
		})

		event1 := sample.InboundPellTx_StakerDeposited_pell(sample.Rand())
		event2 := sample.InboundPellTx_StakerDeposited_pell(sample.Rand())
		event3 := sample.InboundPellTx_StakerDeposited_pell(sample.Rand())
		event4 := sample.InboundPellTx_StakerDeposited_pell(sample.Rand())
		event5 := sample.InboundPellTx_StakerDeposited_pell(sample.Rand())

		// Vote on the FIRST message.
		msg1 := &types.MsgVoteOnObservedInboundTx{
			Signer:        validatorAddr,
			Sender:        "0x954598965C2aCdA2885B037561526260764095B8",
			SenderChainId: 97, // ETH
			Receiver:      "0x954598965C2aCdA2885B037561526260764095B8",
			ReceiverChain: 186, // pellchain
			InBlockHeight: 2,
			GasLimit:      1000000000,
			InTxHash:      "0x7a900ef978743f91f57ca47c6d1a1add75df4d3531da17671e9cf149e1aefe0b",
			TxOrigin:      "0x954598965C2aCdA2885B037561526260764095B8",
			EventIndex:    1,
			PellTx:        event1,
		}

		msg2 := &types.MsgVoteOnObservedInboundTx{
			Signer:        validatorAddr,
			Sender:        "0x954598965C2aCdA2885B037561526260764095B8",
			SenderChainId: 97, // ETH
			Receiver:      "0x954598965C2aCdA2885B037561526260764095B8",
			ReceiverChain: 186, // pellchain
			InBlockHeight: 2,
			GasLimit:      1000000000,
			InTxHash:      "0x7a900ef978743f91f57ca47c6d1a1add75df4d3531da17671e9cf149e1aefe0a",
			TxOrigin:      "0x954598965C2aCdA2885B037561526260764095B8",
			EventIndex:    2,
			PellTx:        event2,
		}

		msg3 := &types.MsgVoteOnObservedInboundTx{
			Signer:        validatorAddr,
			Sender:        "0x954598965C2aCdA2885B037561526260764095B8",
			SenderChainId: 97, // ETH
			Receiver:      "0x954598965C2aCdA2885B037561526260764095B8",
			ReceiverChain: 186, // pellchain
			InBlockHeight: 2,
			GasLimit:      1000000000,
			InTxHash:      "0x7a900ef978743f91f57ca47c6d1a1add75df4d3531da17671e9cf149e1aefe0a",
			TxOrigin:      "0x954598965C2aCdA2885B037561526260764095B8",
			EventIndex:    3,
			PellTx:        event3,
		}

		blockProofMsg1 := types.NewMsgVoteInboundBlock(
			validatorAddr, 97, 0, 2, "0x01", []*types.Event{
				{
					Index:     1,
					TxHash:    "0x7a900ef978743f91f57ca47c6d1a1add75df4d3531da17671e9cf149e1aefe0a",
					PellEvent: event1,
					Digest:    msg1.Digest(),
				},
				{
					Index:     2,
					TxHash:    "0x7a900ef978743f91f57ca47c6d1a1add75df4d3531da17671e9cf149e1aefe0b",
					PellEvent: event2,
					Digest:    msg2.Digest(),
				},
				{
					Index:     3,
					TxHash:    "0x7a900ef978743f91f57ca47c6d1a1add75df4d3531da17671e9cf149e1aefe0b",
					PellEvent: event3,
					Digest:    msg3.Digest(),
				},
			},
		)

		msg4 := &types.MsgVoteOnObservedInboundTx{
			Signer:        validatorAddr,
			Sender:        "0x954598965C2aCdA2885B037561526260764095B8",
			SenderChainId: 97, // ETH
			Receiver:      "0x954598965C2aCdA2885B037561526260764095B8",
			ReceiverChain: 186, // pellchain
			InBlockHeight: 5,
			GasLimit:      1000000000,
			InTxHash:      "0x7a900ef978743f91f57ca47c6d1a1add75df4d3531da17671e9cf149e1aefe0z",
			TxOrigin:      "0x954598965C2aCdA2885B037561526260764095B8",
			EventIndex:    4,
			PellTx:        event4,
		}

		msg5 := &types.MsgVoteOnObservedInboundTx{
			Signer:        validatorAddr,
			Sender:        "0x954598965C2aCdA2885B037561526260764095B8",
			SenderChainId: 97, // ETH
			Receiver:      "0x954598965C2aCdA2885B037561526260764095B8",
			ReceiverChain: 186, // pellchain
			InBlockHeight: 5,
			GasLimit:      1000000000,
			InTxHash:      "0x7a900ef978743f91f57ca47c6d1a1add75df4d3531da17671e9cf149e1aefe0z",
			TxOrigin:      "0x954598965C2aCdA2885B037561526260764095B8",
			EventIndex:    5,
			PellTx:        event5,
		}

		blockProofMsg2 := types.NewMsgVoteInboundBlock(
			validatorAddr, 97, 2, 5, "0x02", []*types.Event{
				{
					Index:     4,
					TxHash:    "0x7a900ef978743f91f57ca47c6d1a1add75df4d3531da17671e9cf149e1aefe0z",
					PellEvent: event4,
					Digest:    msg4.Digest(),
				},
				{
					Index:     5,
					TxHash:    "0x7a900ef978743f91f57ca47c6d1a1add75df4d3531da17671e9cf149e1aefe0z",
					PellEvent: event5,
					Digest:    msg5.Digest(),
				},
			},
		)

		_, err = msgServer.VoteOnObservedInboundBlock(ctx, blockProofMsg1)
		require.NoError(t, err)

		_, err = msgServer.VoteOnObservedInboundTx(
			ctx,
			msg2,
		)
		require.NoError(t, err)
		_, found := zk.ObserverKeeper.GetBallot(ctx, msg2.Digest())
		require.True(t, found)

		_, err = msgServer.VoteOnObservedInboundTx(
			ctx,
			msg1,
		)
		require.NoError(t, err)

		_, err = msgServer.VoteOnObservedInboundTx(
			ctx,
			msg3,
		)
		require.NoError(t, err)

		_, err = msgServer.VoteOnObservedInboundBlock(ctx, blockProofMsg2)
		require.NoError(t, err)

		_, err = msgServer.VoteOnObservedInboundTx(
			ctx,
			msg5,
		)
		require.NoError(t, err)

		_, err = msgServer.VoteOnObservedInboundTx(
			ctx,
			msg4,
		)
		require.NoError(t, err)

		// Check that the vote passed
		ballot, found := zk.ObserverKeeper.GetBallot(ctx, blockProofMsg1.Digest())
		require.True(t, found)
		require.Equal(t, ballot.BallotStatus, relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION)

		ballot, found = zk.ObserverKeeper.GetBallot(ctx, blockProofMsg2.Digest())
		require.True(t, found)
		require.Equal(t, ballot.BallotStatus, relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION)

		// Check that the vote passed
		ballot, found = zk.ObserverKeeper.GetBallot(ctx, msg1.Digest())
		require.True(t, found)
		require.Equal(t, ballot.BallotStatus, relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION)

		ballot, found = zk.ObserverKeeper.GetBallot(ctx, msg2.Digest())
		require.True(t, found)
		require.Equal(t, ballot.BallotStatus, relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION)

		ballot, found = zk.ObserverKeeper.GetBallot(ctx, msg3.Digest())
		require.True(t, found)
		require.Equal(t, ballot.BallotStatus, relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION)

		ballot, found = zk.ObserverKeeper.GetBallot(ctx, msg4.Digest())
		require.True(t, found)
		require.Equal(t, ballot.BallotStatus, relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION)

		ballot, found = zk.ObserverKeeper.GetBallot(ctx, msg5.Digest())
		require.True(t, found)
		require.Equal(t, ballot.BallotStatus, relayertypes.BallotStatus_BALLOT_FINALIZED_SUCCESS_OBSERVATION)

	})

	t.Run("should error if vote on inbound ballot fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("VoteOnInboundBallot", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(true, false, errors.New("err"))
		msgServer := keeper.NewMsgServerImpl(*k)
		to, from := int64(1337), int64(186)

		msg := sample.InboundVote_pell(from, to)
		res, err := msgServer.VoteOnObservedInboundTx(
			ctx,
			&msg,
		)
		require.Error(t, err)
		require.NotNil(t, res)
	})

	t.Run("should return if not finalized", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)
		msgServer := keeper.NewMsgServerImpl(*k)
		validatorList := setObservers(t, k, ctx, zk)

		// add one more voter to make it not finalized
		r := rand.New(rand.NewSource(42))
		valAddr := sample.ValAddress(r)
		observerSet := append(validatorList, valAddr.String())
		zk.ObserverKeeper.SetObserverSet(ctx, relayertypes.RelayerSet{
			RelayerList: observerSet,
		})
		to, from := int64(1337), int64(186)
		supportedChains := zk.ObserverKeeper.GetSupportedChains(ctx)
		for _, chain := range supportedChains {
			if chains.IsEVMChain(chain.Id) {
				from = chain.Id
			}
			if chains.IsPellChain(chain.Id) {
				to = chain.Id
			}
		}
		zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())

		msg := sample.InboundVote_pell(from, to)
		for _, validatorAddr := range validatorList {
			msg.Signer = validatorAddr
			_, err := msgServer.VoteOnObservedInboundTx(
				ctx,
				&msg,
			)
			require.NoError(t, err)
		}
		ballot, _, _ := zk.ObserverKeeper.FindBallot(
			ctx,
			msg.Digest(),
			zk.ObserverKeeper.GetSupportedChainFromChainID(ctx, msg.SenderChainId),
			relayertypes.ObservationType_IN_BOUND_TX,
		)
		require.Equal(t, ballot.BallotStatus, relayertypes.BallotStatus_BALLOT_IN_PROGRESS)
		require.Equal(t, ballot.Votes[0], relayertypes.VoteType_SUCCESS_OBSERVATION)
		require.Equal(t, ballot.Votes[1], relayertypes.VoteType_NOT_YET_VOTED)
		_, found := k.GetXmsg(ctx, msg.Digest())
		require.False(t, found)
	})

	t.Run("should err if tss not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UseObserverMock: true,
		})
		observerMock := keepertest.GetXmsgObserverMock(t, k)
		observerMock.On("VoteOnInboundBallot", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(true, false, nil)
		msgServer := keeper.NewMsgServerImpl(*k)
		to, from := int64(1337), int64(186)

		msg := sample.InboundVote_pell(from, to)
		res, err := msgServer.VoteOnObservedInboundTx(
			ctx,
			&msg,
		)
		require.Error(t, err)
		require.Nil(t, res)
	})

	t.Run("should error if block proof not exist", func(t *testing.T) {
		k, ctx, sdkk, zk := keepertest.XmsgKeeper(t)
		msgServer := keeper.NewMsgServerImpl(*k)
		validatorList := setObservers(t, k, ctx, zk)
		to, from := int64(1337), int64(186)
		supportedChains := zk.ObserverKeeper.GetSupportedChains(ctx)
		for _, chain := range supportedChains {
			if chains.IsEVMChain(chain.Id) {
				from = chain.Id
			}
			if chains.IsPellChain(chain.Id) {
				to = chain.Id
			}
		}

		zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())

		msg := sample.InboundVote_pell(from, to)

		err := sdkk.EvmKeeper.SetAccount(ctx, ethcommon.HexToAddress(msg.Receiver), statedb.Account{
			Nonce:    0,
			Balance:  big.NewInt(0),
			CodeHash: crypto.Keccak256(nil),
		})
		require.NoError(t, err)

		for _, validatorAddr := range validatorList {
			msg.Signer = validatorAddr

			_, err = msgServer.VoteOnObservedInboundTx(
				ctx,
				&msg,
			)
			require.Error(t, err)
		}
	})
}

func TestStatus_ChangeStatus(t *testing.T) {
	tt := []struct {
		Name         string
		Status       types.Status
		NonErrStatus types.XmsgStatus
		Msg          string
		IsErr        bool
		ErrStatus    types.XmsgStatus
	}{
		{
			Name: "Transition on finalize Inbound",
			Status: types.Status{
				Status:              types.XmsgStatus_PENDING_INBOUND,
				StatusMessage:       "Getting InTX Votes",
				LastUpdateTimestamp: 0,
			},
			Msg:          "Got super majority and finalized Inbound",
			NonErrStatus: types.XmsgStatus_PENDING_OUTBOUND,
			ErrStatus:    types.XmsgStatus_ABORTED,
			IsErr:        false,
		},
		{
			Name: "Transition on finalize Inbound Fail",
			Status: types.Status{
				Status:              types.XmsgStatus_PENDING_INBOUND,
				StatusMessage:       "Getting InTX Votes",
				LastUpdateTimestamp: 0,
			},
			Msg:          "Got super majority and finalized Inbound",
			NonErrStatus: types.XmsgStatus_OUTBOUND_MINED,
			ErrStatus:    types.XmsgStatus_ABORTED,
			IsErr:        false,
		},
	}
	for _, test := range tt {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			test.Status.ChangeStatus(test.NonErrStatus, test.Msg)
			if test.IsErr {
				require.Equal(t, test.ErrStatus, test.Status.Status)
			} else {
				require.Equal(t, test.NonErrStatus, test.Status.Status)
			}
		})
	}
}

func TestKeeper_SaveInbound(t *testing.T) {
	t.Run("should save the xmsg", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)
		zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())
		receiver := sample.EthAddress()
		senderChain := getValidEthChain()
		xmsg := buildXmsg(t, receiver, *senderChain)
		eventIndex := sample.Uint64InRange(1, 100)
		k.SaveInbound(ctx, xmsg, xmsg.InboundTxParams.InboundTxBlockHeight, eventIndex)
		require.Equal(t, types.TxFinalizationStatus_EXECUTED, xmsg.InboundTxParams.TxFinalizationStatus)
		require.True(t, k.IsFinalizedInbound(ctx, xmsg.GetInboundTxParams().InboundTxHash, xmsg.GetInboundTxParams().SenderChainId, eventIndex))
		_, found := k.GetXmsg(ctx, xmsg.Index)
		require.True(t, found)
	})

	t.Run("should save the xmsg and remove tracker", func(t *testing.T) {
		k, ctx, _, zk := keepertest.XmsgKeeper(t)
		receiver := sample.EthAddress()
		senderChain := getValidEthChain()
		xmsg := buildXmsg(t, receiver, *senderChain)
		hash := sample.Hash()
		xmsg.InboundTxParams.InboundTxHash = hash.String()
		k.SetInTxTracker(ctx, types.InTxTracker{
			ChainId: senderChain.Id,
			TxHash:  hash.String(),
		})
		eventIndex := sample.Uint64InRange(1, 100)
		zk.ObserverKeeper.SetTSS(ctx, sample.Tss_pell())

		k.SaveInbound(ctx, xmsg, xmsg.InboundTxParams.InboundTxBlockHeight, eventIndex)
		require.Equal(t, types.TxFinalizationStatus_EXECUTED, xmsg.InboundTxParams.TxFinalizationStatus)
		require.True(t, k.IsFinalizedInbound(ctx, xmsg.GetInboundTxParams().InboundTxHash, xmsg.GetInboundTxParams().SenderChainId, eventIndex))
		_, found := k.GetXmsg(ctx, xmsg.Index)
		require.True(t, found)
		_, found = k.GetInTxTracker(ctx, senderChain.Id, hash.String())
		require.False(t, found)
	})
}
