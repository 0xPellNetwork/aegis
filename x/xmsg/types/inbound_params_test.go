package types_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/testutil/sample"
)

func TestInboundTxParams_Validate(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	inTxParams := sample.InboundTxParamsValidChainID_pell(r)
	inTxParams.Sender = ""
	require.ErrorContains(t, inTxParams.Validate(), "sender cannot be empty")
	inTxParams = sample.InboundTxParamsValidChainID_pell(r)
	inTxParams.SenderChainId = 1000
	require.ErrorContains(t, inTxParams.Validate(), "invalid sender chain id 1000")

	inTxParams = sample.InboundTxParamsValidChainID_pell(r)
	inTxParams.InboundTxHash = sample.Hash().String()
	inTxParams.InboundTxBallotIndex = sample.PellIndex(t)
	require.NoError(t, inTxParams.Validate())
}
