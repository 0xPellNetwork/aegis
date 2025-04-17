package pellcore

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/pell-chain/pellcore/pkg/chains"
	clientauthz "github.com/pell-chain/pellcore/relayer/authz"
	clientcommon "github.com/pell-chain/pellcore/relayer/common"
	"github.com/pell-chain/pellcore/x/xmsg/types"
)

// GetInBoundVoteMessage returns a new MsgVoteOnObservedInboundTx
func GetInBoundVoteMessage(
	sender string,
	senderChain int64,
	txOrigin string,
	receiver string,
	receiverChain int64,
	inTxHash string,
	inBlockHeight uint64,
	gasLimit uint64,
	signerAddress string,
	eventIndex uint,
	pellData types.InboundPellEvent,
) *types.MsgVoteOnObservedInboundTx {
	msg := types.NewMsgVoteOnObservedInboundTx(
		signerAddress,
		sender,
		senderChain,
		txOrigin,
		receiver,
		receiverChain,
		inTxHash,
		inBlockHeight,
		gasLimit,
		eventIndex,
		pellData,
	)
	return msg
}

// GasPriceMultiplier returns the gas price multiplier for the given chain
func GasPriceMultiplier(chainID int64) float64 {
	if chains.IsEVMChain(chainID) {
		return clientcommon.EVMOutboundGasPriceMultiplier
	}

	return clientcommon.DefaultGasPriceMultiplier
}

func WrapMessageWithAuthz(msg sdk.Msg) (sdk.Msg, clientauthz.Signer, error) {
	msgURL := sdk.MsgTypeURL(msg)

	authzSigner := clientauthz.GetSigner(msgURL)
	authzMessage := authz.NewMsgExec(authzSigner.GranteeAddress, []sdk.Msg{msg})
	return &authzMessage, authzSigner, nil
}
