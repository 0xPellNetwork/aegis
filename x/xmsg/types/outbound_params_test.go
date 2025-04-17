package types_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/testutil/sample"
)

func TestOutboundTxParams_Validate(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	outTxParams := sample.OutboundTxParamsValidChainID_pell(r)
	outTxParams.Receiver = ""
	require.ErrorContains(t, outTxParams.Validate(), "receiver cannot be empty")

	outTxParams = sample.OutboundTxParamsValidChainID_pell(r)
	outTxParams.ReceiverChainId = 1000
	require.ErrorContains(t, outTxParams.Validate(), "invalid receiver chain id 1000")

	outTxParams = sample.OutboundTxParamsValidChainID_pell(r)
	outTxParams.OutboundTxBallotIndex = sample.PellIndex(t)
	outTxParams.OutboundTxHash = sample.Hash().String()
	require.NoError(t, outTxParams.Validate())

	// Disabled checks
	// TODO: Improve the checks, move the validation call to a new place and reenable
	//outTxParams = sample.OutboundParamsValidChainID(r)
	//outTxParams.Receiver = "0x123"
	//require.ErrorContains(t, outTxParams.Validate(), "invalid address 0x123")
	//outTxParams = sample.OutboundParamsValidChainID(r)
	//outTxParams.BallotIndex = "12"
	//require.ErrorContains(t, outTxParams.Validate(), "invalid index length 2")
}

func TestOutboundTxParams_GetGasPrice(t *testing.T) {
	// #nosec G404 - random seed is not used for security purposes
	r := rand.New(rand.NewSource(42))
	outTxParams := sample.OutboundTxParams_pell(r)

	outTxParams.OutboundTxGasPrice = "42"
	gasPrice, err := outTxParams.GetGasPrice()
	require.NoError(t, err)
	require.EqualValues(t, uint64(42), gasPrice)

	outTxParams.OutboundTxGasPrice = "invalid"
	_, err = outTxParams.GetGasPrice()
	require.Error(t, err)
}
