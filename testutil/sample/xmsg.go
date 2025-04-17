package sample

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/pkg/chains"
	"github.com/0xPellNetwork/aegis/x/xmsg/types"
)

func RateLimiterFlags() types.RateLimiterFlags {
	r := Rand()

	return types.RateLimiterFlags{
		Enabled: true,
		Window:  r.Int63(),
		Rate:    sdkmath.NewUint(r.Uint64()),
	}
}

func InboundPellTx_StakerDeposited_pell(r *rand.Rand) *types.InboundPellEvent {
	pellTx := &types.InboundPellEvent{
		PellData: &types.InboundPellEvent_StakerDeposited{
			StakerDeposited: &types.StakerDeposited{
				Staker:   EthAddress().String(),
				Token:    EthAddress().String(),
				Strategy: EthAddress().String(),
				Shares:   math.NewUint(r.Uint64()),
			},
		},
	}

	return pellTx
}

func InboundPellTx_StakerDelegated_pell(r *rand.Rand) *types.InboundPellEvent {
	pellTx := &types.InboundPellEvent{
		PellData: &types.InboundPellEvent_StakerDelegated{
			StakerDelegated: &types.StakerDelegated{
				Staker:   EthAddress().String(),
				Operator: EthAddress().String(),
			},
		},
	}

	return pellTx
}

func InboundPellTx_PellSend_pell(r *rand.Rand) *types.InboundPellEvent {
	pellTx := &types.InboundPellEvent{
		PellData: &types.InboundPellEvent_PellSent{
			PellSent: &types.PellSent{
				TxOrigin:        EthAddress().String(),
				Sender:          EthAddress().String(),
				ReceiverChainId: r.Int63(),
				Receiver:        EthAddress().String(),
				Message:         StringRandom(r, 32),
				PellParams:      StringRandom(r, 32),
			},
		},
	}

	return pellTx
}

func RateLimiterFlags_pell() types.RateLimiterFlags {
	r := Rand()

	return types.RateLimiterFlags{
		Enabled: true,
		Window:  r.Int63(),
		Rate:    sdkmath.NewUint(r.Uint64()),
	}
}

func OutTxTracker_pell(t *testing.T, index string) types.OutTxTracker {
	r := newRandFromStringSeed(t, index)

	return types.OutTxTracker{
		Index:   index,
		ChainId: r.Int63(),
		Nonce:   r.Uint64(),
	}
}

func InTxTracker_pell(t *testing.T, index string) types.InTxTracker {
	r := newRandFromStringSeed(t, index)

	return types.InTxTracker{
		ChainId: r.Int63(),
		TxHash:  Hash().Hex(),
	}
}

func GasPrice_pell(t *testing.T, index string) *types.GasPrice {
	r := newRandFromStringSeed(t, index)

	return &types.GasPrice{
		Signer:      AccAddress(),
		Index:       index,
		ChainId:     r.Int63(),
		Signers:     []string{AccAddress(), AccAddress()},
		BlockNums:   []uint64{r.Uint64(), r.Uint64()},
		Prices:      []uint64{r.Uint64(), r.Uint64()},
		MedianIndex: 0,
	}
}

func InboundTxParams_pell(r *rand.Rand) *types.InboundTxParams {
	return &types.InboundTxParams{
		Sender:                       EthAddress().String(),
		SenderChainId:                r.Int63(),
		TxOrigin:                     EthAddress().String(),
		InboundTxHash:                StringRandom(r, 32),
		InboundTxBlockHeight:         r.Uint64(),
		InboundTxBallotIndex:         StringRandom(r, 32),
		InboundTxFinalizedPellHeight: r.Uint64(),
		InboundPellTx:                InboundPellTx_StakerDelegated_pell(r),
	}
}

func InboundTxParamsValidChainID_pell(r *rand.Rand) *types.InboundTxParams {
	return &types.InboundTxParams{
		Sender:                       EthAddress().String(),
		SenderChainId:                chains.GoerliChain().Id,
		TxOrigin:                     EthAddress().String(),
		InboundTxHash:                StringRandom(r, 32),
		InboundTxBlockHeight:         r.Uint64(),
		InboundTxBallotIndex:         StringRandom(r, 32),
		InboundTxFinalizedPellHeight: r.Uint64(),
		InboundPellTx:                InboundPellTx_StakerDeposited_pell(r),
	}
}

func OutboundTxParams_pell(r *rand.Rand) *types.OutboundTxParams {
	return &types.OutboundTxParams{
		Receiver:                    EthAddress().String(),
		ReceiverChainId:             r.Int63(),
		OutboundTxTssNonce:          r.Uint64(),
		OutboundTxGasLimit:          r.Uint64(),
		OutboundTxGasPrice:          math.NewUint(uint64(r.Int63())).String(),
		OutboundTxHash:              StringRandom(r, 32),
		OutboundTxBallotIndex:       StringRandom(r, 32),
		OutboundTxExternalHeight:    r.Uint64(),
		OutboundTxGasUsed:           r.Uint64(),
		OutboundTxEffectiveGasPrice: math.NewInt(r.Int63()),
	}
}

func OutboundTxParamsValidChainID_pell(r *rand.Rand) *types.OutboundTxParams {
	return &types.OutboundTxParams{
		Receiver:                    EthAddress().String(),
		ReceiverChainId:             chains.GoerliChain().Id,
		OutboundTxTssNonce:          r.Uint64(),
		OutboundTxGasLimit:          r.Uint64(),
		OutboundTxGasPrice:          math.NewUint(uint64(r.Int63())).String(),
		OutboundTxHash:              StringRandom(r, 32),
		OutboundTxBallotIndex:       StringRandom(r, 32),
		OutboundTxExternalHeight:    r.Uint64(),
		OutboundTxGasUsed:           r.Uint64(),
		OutboundTxEffectiveGasPrice: math.NewInt(r.Int63()),
	}
}

func Status_pell(t *testing.T, index string) *types.Status {
	r := newRandFromStringSeed(t, index)

	return &types.Status{
		Status:              types.XmsgStatus(r.Intn(100)),
		StatusMessage:       String(),
		LastUpdateTimestamp: r.Int63(),
	}
}

func GetXmsgIndicesFromString_pell(index string) string {
	return crypto.Keccak256Hash([]byte(index)).String()
}

func Xmsg_pell(t *testing.T, index string) *types.Xmsg {
	r := newRandFromStringSeed(t, index)

	return &types.Xmsg{
		Signer:           AccAddress(),
		Index:            GetXmsgIndicesFromString_pell(index),
		XmsgStatus:       Status_pell(t, index),
		InboundTxParams:  InboundTxParams_pell(r),
		OutboundTxParams: []*types.OutboundTxParams{OutboundTxParams_pell(r), OutboundTxParams_pell(r)},
	}
}

func LastBlockHeight_pell(t *testing.T, index string) *types.LastBlockHeight {
	r := newRandFromStringSeed(t, index)

	return &types.LastBlockHeight{
		Signer:            AccAddress(),
		Index:             index,
		Chain:             StringRandom(r, 32),
		LastSendHeight:    r.Uint64(),
		LastReceiveHeight: r.Uint64(),
	}
}

func InTxHashToXmsg_pell(t *testing.T, inTxHash string) types.InTxHashToXmsg {
	r := newRandFromStringSeed(t, inTxHash)

	return types.InTxHashToXmsg{
		InTxHash:    inTxHash,
		XmsgIndices: []string{StringRandom(r, 32), StringRandom(r, 32)},
	}
}

func InboundVote_pell(from, to int64) types.MsgVoteOnObservedInboundTx {
	r := Rand()

	return types.MsgVoteOnObservedInboundTx{
		Signer:        "",
		Sender:        EthAddress().String(),
		SenderChainId: Chain(from).Id,
		Receiver:      EthAddress().String(),
		ReceiverChain: Chain(to).Id,
		InBlockHeight: Uint64InRange(1, 10000),
		GasLimit:      1000000000,
		InTxHash:      Hash().String(),
		TxOrigin:      EthAddress().String(),
		EventIndex:    EventIndex(),
		PellTx:        InboundPellTx_StakerDeposited_pell(r),
	}
}

// CustomXmsgsInBlockRange create 1 cctx per block in block range [lowBlock, highBlock] (inclusive)
func CustomXmsgsInBlockRange(
	t *testing.T,
	lowBlock uint64,
	highBlock uint64,
	senderChainID int64,
	receiverChainID int64,
	status types.XmsgStatus,
) (cctxs []*types.Xmsg) {
	// create 1 cctx per block
	for i := lowBlock; i <= highBlock; i++ {
		nonce := i - 1
		cctx := Xmsg_pell(t, fmt.Sprintf("%d-%d", receiverChainID, nonce))
		cctx.XmsgStatus.Status = status
		cctx.InboundTxParams.SenderChainId = senderChainID
		cctx.InboundTxParams.InboundTxBlockHeight = i
		cctx.GetCurrentOutTxParam().ReceiverChainId = receiverChainID
		cctx.GetCurrentOutTxParam().OutboundTxTssNonce = nonce
		cctxs = append(cctxs, cctx)
	}
	return cctxs
}

func GetValidPellSent_pell(t *testing.T) (receipt ethtypes.Receipt) {
	block := "{\"root\":\"0x\",\"status\":\"0x1\",\"cumulativeGasUsed\":\"0xd75f4f\",\"logsBloom\":\"0x00000000000000000000000000000000800800000000000000000000100000000000002000000100000000000000000000000000000000000000000000240000000000000000000000000008000000000800000000440000000000008080000000000000000000000000000000000000000000000000040000000010000000000000000000000000000000000000000200000001000000000000000040000000020000000000000000000000008200000000000000000000000000000000000000000002000000000000008000000000000000000000000000080002000041000010000000000000000000000000000000000000000000400000000000000000\",\"logs\":[{\"address\":\"0x5f0b1a82749cb4e2278ec87f8bf6b618dc71a8bf\",\"topics\":[\"0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c\",\"0x000000000000000000000000f0a3f93ed1b126142e61423f9546bf1323ff82df\"],\"data\":\"0x000000000000000000000000000000000000000000000003cb71f51fc5580000\",\"blockNumber\":\"0x1bedc8\",\"transactionHash\":\"0x19d8a67a05998f1cb19fe731b96d817d5b186b62c9430c51679664959c952ef0\",\"transactionIndex\":\"0x5f\",\"blockHash\":\"0x198fdd1f4bc6b910db978602cb15bdb2bcc6fd960e9324e9b9675dc062133794\",\"logIndex\":\"0x13b\",\"removed\":false},{\"address\":\"0x5f0b1a82749cb4e2278ec87f8bf6b618dc71a8bf\",\"topics\":[\"0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925\",\"0x000000000000000000000000f0a3f93ed1b126142e61423f9546bf1323ff82df\",\"0x000000000000000000000000239e96c8f17c85c30100ac26f635ea15f23e9c67\"],\"data\":\"0x000000000000000000000000000000000000000000000003cb71f51fc5580000\",\"blockNumber\":\"0x1bedc8\",\"transactionHash\":\"0x19d8a67a05998f1cb19fe731b96d817d5b186b62c9430c51679664959c952ef0\",\"transactionIndex\":\"0x5f\",\"blockHash\":\"0x198fdd1f4bc6b910db978602cb15bdb2bcc6fd960e9324e9b9675dc062133794\",\"logIndex\":\"0x13c\",\"removed\":false},{\"address\":\"0x5f0b1a82749cb4e2278ec87f8bf6b618dc71a8bf\",\"topics\":[\"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef\",\"0x000000000000000000000000f0a3f93ed1b126142e61423f9546bf1323ff82df\",\"0x000000000000000000000000239e96c8f17c85c30100ac26f635ea15f23e9c67\"],\"data\":\"0x000000000000000000000000000000000000000000000003cb71f51fc5580000\",\"blockNumber\":\"0x1bedc8\",\"transactionHash\":\"0x19d8a67a05998f1cb19fe731b96d817d5b186b62c9430c51679664959c952ef0\",\"transactionIndex\":\"0x5f\",\"blockHash\":\"0x198fdd1f4bc6b910db978602cb15bdb2bcc6fd960e9324e9b9675dc062133794\",\"logIndex\":\"0x13d\",\"removed\":false},{\"address\":\"0x5f0b1a82749cb4e2278ec87f8bf6b618dc71a8bf\",\"topics\":[\"0x7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65\",\"0x000000000000000000000000239e96c8f17c85c30100ac26f635ea15f23e9c67\"],\"data\":\"0x000000000000000000000000000000000000000000000003cb71f51fc5580000\",\"blockNumber\":\"0x1bedc8\",\"transactionHash\":\"0x19d8a67a05998f1cb19fe731b96d817d5b186b62c9430c51679664959c952ef0\",\"transactionIndex\":\"0x5f\",\"blockHash\":\"0x198fdd1f4bc6b910db978602cb15bdb2bcc6fd960e9324e9b9675dc062133794\",\"logIndex\":\"0x13e\",\"removed\":false},{\"address\":\"0x239e96c8f17c85c30100ac26f635ea15f23e9c67\",\"topics\":[\"0x7ec1c94701e09b1652f3e1d307e60c4b9ebf99aff8c2079fd1d8c585e031c4e4\",\"0x000000000000000000000000f0a3f93ed1b126142e61423f9546bf1323ff82df\",\"0x0000000000000000000000000000000000000000000000000000000000000001\"],\"data\":\"0x00000000000000000000000060983881bdf302dcfa96603a58274d15d596620900000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000003cb71f51fc558000000000000000000000000000000000000000000000000000000000000000186a000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000001460983881bdf302dcfa96603a58274d15d59662090000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000\",\"blockNumber\":\"0x1bedc8\",\"transactionHash\":\"0x19d8a67a05998f1cb19fe731b96d817d5b186b62c9430c51679664959c952ef0\",\"transactionIndex\":\"0x5f\",\"blockHash\":\"0x198fdd1f4bc6b910db978602cb15bdb2bcc6fd960e9324e9b9675dc062133794\",\"logIndex\":\"0x13f\",\"removed\":false}],\"transactionHash\":\"0x19d8a67a05998f1cb19fe731b96d817d5b186b62c9430c51679664959c952ef0\",\"contractAddress\":\"0x0000000000000000000000000000000000000000\",\"gasUsed\":\"0x2406d\",\"blockHash\":\"0x198fdd1f4bc6b910db978602cb15bdb2bcc6fd960e9324e9b9675dc062133794\",\"blockNumber\":\"0x1bedc8\",\"transactionIndex\":\"0x5f\"}\n"
	err := json.Unmarshal([]byte(block), &receipt)
	require.NoError(t, err)
	return
}
