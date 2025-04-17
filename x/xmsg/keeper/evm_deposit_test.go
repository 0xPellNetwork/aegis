package keeper_test

import (
	"errors"
	"math/big"
	"testing"

	"cosmossdk.io/math"
	ethcommon "github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
	pevmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

func TestMsgServer_HandleEVMEvents(t *testing.T) {
	t.Run("can process stakerdeposited calling pevm method", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock: true,
		})

		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		sender := sample.EthAddress()
		senderChainId := getValidEthChainID()
		receiver := sample.EthAddress()
		staker := sample.EthAddress()
		token := sample.EthAddress()
		strategy := sample.EthAddress()
		shares := big.NewInt(42)

		pevmMock.On("GetPellStrategyManagerProxyContractAddress", ctx).Return(sample.EthAddress(), nil)

		// expect CallsyncDepositStateOnPellStrategyManager to be called
		pevmMock.On("CallSyncDepositStateOnPellStrategyManager",
			ctx,
			sender.Bytes(),
			senderChainId,
			staker,
			strategy,
			shares).Return(&evmtypes.MsgEthereumTxResponse{}, false, nil)

		// call HandleEVMEvents
		xmsg := sample.Xmsg_pell(t, "foo")
		xmsg.GetCurrentOutTxParam().Receiver = receiver.String()
		xmsg.GetInboundTxParams().SenderChainId = senderChainId
		xmsg.GetInboundTxParams().Sender = sender.String()
		xmsg.GetInboundTxParams().InboundPellTx = &types.InboundPellEvent{
			PellData: &types.InboundPellEvent_StakerDeposited{
				StakerDeposited: &types.StakerDeposited{
					Staker:   staker.String(),
					Token:    token.String(),
					Strategy: strategy.String(),
					Shares:   math.NewUintFromBigInt(shares),
				},
			},
		}
		reverted, err := k.HandleEVMEvents(
			ctx,
			xmsg,
		)
		require.NoError(t, err)
		require.False(t, reverted)
		pevmMock.AssertExpectations(t)
	})

	t.Run("should error on processing stakerdeposited calling pevm method for contract call if process logs fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock: true,
		})

		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		sender := sample.EthAddress()
		senderChainId := getValidEthChainID()
		receiver := sample.EthAddress()
		staker := sample.EthAddress()
		token := sample.EthAddress()
		strategy := sample.EthAddress()
		shares := big.NewInt(42)
		errMsg := errors.New("stakerdelegated failed")

		pevmMock.On("GetPellStrategyManagerProxyContractAddress", ctx).Return(sample.EthAddress(), nil)

		// expect CallsyncDepositStateOnPellStrategyManager to be called
		pevmMock.On("CallSyncDepositStateOnPellStrategyManager",
			ctx,
			sender.Bytes(),
			senderChainId,
			staker,
			strategy,
			shares).Return(&evmtypes.MsgEthereumTxResponse{
			Logs: []*evmtypes.Log{
				{
					Address:     receiver.Hex(),
					Topics:      []string{},
					Data:        []byte{},
					BlockNumber: uint64(ctx.BlockHeight()),
					TxHash:      sample.Hash().Hex(),
					TxIndex:     1,
					BlockHash:   sample.Hash().Hex(),
					Index:       1,
				},
			},
		}, true, nil)

		// failed on processing logs
		pevmMock.On("GetPellConnectorContractAddress", mock.Anything).Return(ethcommon.Address{}, errMsg)

		// call HandleEVMDeposit
		xmsg := sample.Xmsg_pell(t, "foo")
		xmsg.GetCurrentOutTxParam().Receiver = receiver.String()
		xmsg.GetInboundTxParams().Sender = sender.String()
		xmsg.GetInboundTxParams().SenderChainId = senderChainId
		xmsg.GetInboundTxParams().InboundPellTx = &types.InboundPellEvent{
			PellData: &types.InboundPellEvent_StakerDeposited{
				StakerDeposited: &types.StakerDeposited{
					Staker:   staker.String(),
					Token:    token.String(),
					Strategy: strategy.String(),
					Shares:   math.NewUintFromBigInt(shares),
				},
			},
		}

		reverted, err := k.HandleEVMEvents(
			ctx,
			xmsg,
		)
		require.Error(t, err)
		require.False(t, reverted)
		pevmMock.AssertExpectations(t)
	})

	t.Run("should error on processing stakerdeposited calling pevm method for contract call if process logs doesn't fails", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock: true,
		})

		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		sender := sample.EthAddress()
		senderChainId := getValidEthChainID()
		receiver := sample.EthAddress()
		staker := sample.EthAddress()
		token := sample.EthAddress()
		strategy := sample.EthAddress()
		shares := big.NewInt(42)

		pevmMock.On("GetPellStrategyManagerProxyContractAddress", ctx).Return(sample.EthAddress(), nil)

		// expect CallsyncDepositStateOnPellStrategyManager to be called
		pevmMock.On("CallSyncDepositStateOnPellStrategyManager",
			ctx,
			sender.Bytes(),
			senderChainId,
			staker,
			strategy,
			shares).Return(&evmtypes.MsgEthereumTxResponse{
			Logs: []*evmtypes.Log{
				{
					Address:     receiver.Hex(),
					Topics:      []string{},
					Data:        []byte{},
					BlockNumber: uint64(ctx.BlockHeight()),
					TxHash:      sample.Hash().Hex(),
					TxIndex:     1,
					BlockHash:   sample.Hash().Hex(),
					Index:       1,
				},
			},
		}, true, nil)

		// success on processing logs
		pevmMock.On("GetPellConnectorContractAddress", mock.Anything).Return(sample.EthAddress(), nil)

		// call HandleEVMDeposit
		xmsg := sample.Xmsg_pell(t, "foo")
		xmsg.GetCurrentOutTxParam().Receiver = receiver.String()
		xmsg.GetInboundTxParams().Sender = sender.String()
		xmsg.GetInboundTxParams().SenderChainId = senderChainId
		xmsg.GetInboundTxParams().InboundPellTx = &types.InboundPellEvent{
			PellData: &types.InboundPellEvent_StakerDeposited{
				StakerDeposited: &types.StakerDeposited{
					Staker:   staker.String(),
					Token:    token.String(),
					Strategy: strategy.String(),
					Shares:   math.NewUintFromBigInt(shares),
				},
			},
		}

		reverted, err := k.HandleEVMEvents(
			ctx,
			xmsg,
		)
		require.NoError(t, err)
		require.False(t, reverted)
		pevmMock.AssertExpectations(t)
	})

	t.Run("should error if invalid sender", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock: true,
		})

		sender := "invalid"
		senderChainId := int64(987)
		receiver := sample.EthAddress()
		staker := sample.EthAddress()
		token := sample.EthAddress()
		strategy := sample.EthAddress()
		shares := big.NewInt(42)

		// call HandleEVMEvents
		xmsg := sample.Xmsg_pell(t, "foo")
		xmsg.GetCurrentOutTxParam().Receiver = receiver.String()
		xmsg.GetInboundTxParams().SenderChainId = senderChainId
		xmsg.InboundTxParams.Sender = sender
		xmsg.InboundTxParams.InboundPellTx = &types.InboundPellEvent{
			PellData: &types.InboundPellEvent_StakerDeposited{
				StakerDeposited: &types.StakerDeposited{
					Staker:   staker.String(),
					Token:    token.String(),
					Strategy: strategy.String(),
					Shares:   math.NewUintFromBigInt(shares),
				},
			},
		}
		reverted, err := k.HandleEVMEvents(
			ctx,
			xmsg,
		)
		require.Error(t, err)
		require.False(t, reverted)
	})

	t.Run("should return error with non-reverted if stakerdelegated events fails with tx non-failed", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock: true,
		})

		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		sender := sample.EthAddress()
		senderChain := getValidEthChainID()
		receiver := sample.EthAddress()
		staker := sample.EthAddress()
		operator := sample.EthAddress()
		errMsg := errors.New("stakerdelegated failed")

		pevmMock.On("GetPellDelegationManagerProxyContractAddress", ctx).Return(sample.EthAddress(), nil)

		// expect CallsyncDelegatedStateOnPellDelegationManager to be called
		pevmMock.On("CallSyncDelegatedStateOnPellDelegationManager",
			ctx,            // types.Context
			sender.Bytes(), // Chain ID as []uint8
			senderChain,    // Height as int64
			staker,         // Staker
			operator,       // Operator
		).Return(&evmtypes.MsgEthereumTxResponse{}, false, errMsg)

		// call HandleEVMEvents
		xmsg := sample.Xmsg_pell(t, "foo")
		xmsg.GetCurrentOutTxParam().Receiver = receiver.String()
		xmsg.GetInboundTxParams().Sender = sender.String()
		xmsg.GetInboundTxParams().SenderChainId = senderChain
		xmsg.InboundTxParams.InboundPellTx = &types.InboundPellEvent{
			PellData: &types.InboundPellEvent_StakerDelegated{
				StakerDelegated: &types.StakerDelegated{
					Staker:   staker.String(),
					Operator: operator.String(),
				},
			},
		}

		reverted, err := k.HandleEVMEvents(
			ctx,
			xmsg,
		)
		require.ErrorIs(t, err, errMsg)
		require.False(t, reverted)
		pevmMock.AssertExpectations(t)
	})

	t.Run("should return error with reverted if stakerdelegated fails with tx failed", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock: true,
		})

		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		sender := sample.EthAddress()
		senderChain := getValidEthChainID()
		receiver := sample.EthAddress()
		staker := sample.EthAddress()
		operator := sample.EthAddress()
		errMsg := errors.New("stakerdelegated failed")

		pevmMock.On("GetPellDelegationManagerProxyContractAddress", ctx).Return(sample.EthAddress(), nil)

		// expect CallsyncDelegatedStateOnPellDelegationManager to be called
		pevmMock.On("CallSyncDelegatedStateOnPellDelegationManager",
			ctx,            // types.Context
			sender.Bytes(), // Chain ID as []uint8
			senderChain,    // Height as int64
			staker,         // Staker
			operator,       // Operator
		).Return(&evmtypes.MsgEthereumTxResponse{VmError: "reverted"}, false, errMsg)

		// call HandleEVMEvents
		xmsg := sample.Xmsg_pell(t, "foo")
		xmsg.GetCurrentOutTxParam().Receiver = receiver.String()
		xmsg.GetInboundTxParams().Sender = sender.String()
		xmsg.GetInboundTxParams().SenderChainId = senderChain
		xmsg.GetInboundTxParams().InboundPellTx = &types.InboundPellEvent{
			PellData: &types.InboundPellEvent_StakerDelegated{
				StakerDelegated: &types.StakerDelegated{
					Staker:   staker.String(),
					Operator: operator.String(),
				},
			},
		}
		reverted, err := k.HandleEVMEvents(
			ctx,
			xmsg,
		)
		require.ErrorIs(t, err, errMsg)
		require.True(t, reverted)
		pevmMock.AssertExpectations(t)
	})

	t.Run("should return error with reverted if stakerdelegated fails with calling a non-contract address", func(t *testing.T) {
		k, ctx, _, _ := keepertest.XmsgKeeperWithMocks(t, keepertest.XmsgMockOptions{
			UsePevmMock: true,
		})

		pevmMock := keepertest.GetXmsgPevmMock(t, k)
		sender := sample.EthAddress()
		senderChain := getValidEthChainID()
		receiver := sample.EthAddress()
		staker := sample.EthAddress()
		operator := sample.EthAddress()

		pevmMock.On("GetPellDelegationManagerProxyContractAddress", ctx).Return(sample.EthAddress(), nil)

		// expect CallsyncDelegatedStateOnPellDelegationManager to be called
		pevmMock.On("CallSyncDelegatedStateOnPellDelegationManager",
			ctx,            // types.Context
			sender.Bytes(), // Chain ID as []uint8
			senderChain,    // Height as int64
			staker,         // Staker
			operator,       // Operator
		).Return(&evmtypes.MsgEthereumTxResponse{}, false, pevmtypes.ErrCallNonContract)

		// call HandleEVMEvents
		xmsg := sample.Xmsg_pell(t, "foo")
		xmsg.GetCurrentOutTxParam().Receiver = receiver.String()
		xmsg.GetInboundTxParams().Sender = sender.String()
		xmsg.GetInboundTxParams().SenderChainId = senderChain
		xmsg.InboundTxParams.InboundPellTx = &types.InboundPellEvent{
			PellData: &types.InboundPellEvent_StakerDelegated{
				StakerDelegated: &types.StakerDelegated{
					Staker:   staker.String(),
					Operator: operator.String(),
				},
			},
		}
		reverted, err := k.HandleEVMEvents(
			ctx,
			xmsg,
		)
		require.ErrorIs(t, err, pevmtypes.ErrCallNonContract)
		require.True(t, reverted)
		pevmMock.AssertExpectations(t)
	})

	// TODO: add test cases for testing logs process
}
