package signer

// import (
// 	"math/big"
// 	"testing"

// 	"github.com/rs/zerolog"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/pell-chain/pellcore/x/xmsg/types"
// )

// func TestGasFromXmsg(t *testing.T) {
// 	logger := zerolog.New(zerolog.NewTestWriter(t))

// 	makeXmsg := func(gasLimit uint64, price, priorityFee string) *types.Xmsg {
// 		xmsg := getXmsg(t)
// 		xmsg.GetOutboundTxParams()[0].OutboundTxGasLimit = gasLimit
// 		xmsg.GetOutboundTxParams()[0].OutboundTxGasPrice = price
// 		xmsg.GetOutboundTxParams()[0].OutboundTxGasPriorityFee = priorityFee

// 		return xmsg
// 	}

// 	for _, tt := range []struct {
// 		name          string
// 		xmsg          *types.Xmsg
// 		errorContains string
// 		assert        func(t *testing.T, g Gas)
// 	}{
// 		{
// 			name: "legacy: gas is too low",
// 			xmsg: makeXmsg(minGasLimit-200, gwei(2).String(), ""),
// 			assert: func(t *testing.T, g Gas) {
// 				assert.True(t, g.isLegacy())
// 				assertGasEquals(t, Gas{
// 					Limit:       minGasLimit,
// 					PriorityFee: gwei(0),
// 					Price:       gwei(2),
// 				}, g)
// 			},
// 		},
// 		{
// 			name: "london: gas is too low",
// 			xmsg: makeXmsg(minGasLimit-200, gwei(2).String(), gwei(1).String()),
// 			assert: func(t *testing.T, g Gas) {
// 				assert.False(t, g.isLegacy())
// 				assertGasEquals(t, Gas{
// 					Limit:       minGasLimit,
// 					Price:       gwei(2),
// 					PriorityFee: gwei(1),
// 				}, g)
// 			},
// 		},
// 		{
// 			name: "pre London gas logic",
// 			xmsg: makeXmsg(minGasLimit+100, gwei(3).String(), ""),
// 			assert: func(t *testing.T, g Gas) {
// 				assert.True(t, g.isLegacy())
// 				assertGasEquals(t, Gas{
// 					Limit:       100_100,
// 					Price:       gwei(3),
// 					PriorityFee: gwei(0),
// 				}, g)
// 			},
// 		},
// 		{
// 			name: "post London gas logic",
// 			xmsg: makeXmsg(minGasLimit+200, gwei(4).String(), gwei(1).String()),
// 			assert: func(t *testing.T, g Gas) {
// 				assert.False(t, g.isLegacy())
// 				assertGasEquals(t, Gas{
// 					Limit:       100_200,
// 					Price:       gwei(4),
// 					PriorityFee: gwei(1),
// 				}, g)
// 			},
// 		},
// 		{
// 			name: "gas is too high, force to the ceiling",
// 			xmsg: makeXmsg(maxGasLimit+200, gwei(4).String(), gwei(1).String()),
// 			assert: func(t *testing.T, g Gas) {
// 				assert.False(t, g.isLegacy())
// 				assertGasEquals(t, Gas{
// 					Limit:       maxGasLimit,
// 					Price:       gwei(4),
// 					PriorityFee: gwei(1),
// 				}, g)
// 			},
// 		},
// 		{
// 			name:          "priority fee is invalid",
// 			xmsg:          makeXmsg(123_000, gwei(4).String(), "oopsie"),
// 			errorContains: "unable to parse priorityFee",
// 		},
// 		{
// 			name:          "priority fee is negative",
// 			xmsg:          makeXmsg(123_000, gwei(4).String(), "-1"),
// 			errorContains: "unable to parse priorityFee: big.Int is negative",
// 		},
// 		{
// 			name:          "gasPrice is less than priorityFee",
// 			xmsg:          makeXmsg(123_000, gwei(4).String(), gwei(5).String()),
// 			errorContains: "gasPrice (4000000000) is less than priorityFee (5000000000)",
// 		},
// 		{
// 			name:          "gasPrice is invalid",
// 			xmsg:          makeXmsg(123_000, "hello", gwei(5).String()),
// 			errorContains: "unable to parse gasPrice",
// 		},
// 	} {
// 		t.Run(tt.name, func(t *testing.T) {
// 			g, err := gasFromXmsg(tt.xmsg, logger)
// 			if tt.errorContains != "" {
// 				assert.ErrorContains(t, err, tt.errorContains)
// 				return
// 			}

// 			assert.NoError(t, err)
// 			assert.NoError(t, g.validate())
// 			tt.assert(t, g)
// 		})
// 	}

// 	t.Run("empty priority fee", func(t *testing.T) {
// 		gas := Gas{
// 			Limit:       123_000,
// 			Price:       gwei(4),
// 			PriorityFee: nil,
// 		}

// 		assert.Error(t, gas.validate())
// 	})
// }

// func assertGasEquals(t *testing.T, expected, actual Gas) {
// 	assert.Equal(t, int64(expected.Limit), int64(actual.Limit), "gas limit")
// 	assert.Equal(t, expected.Price.Int64(), actual.Price.Int64(), "max fee per unit")
// 	assert.Equal(t, expected.PriorityFee.Int64(), actual.PriorityFee.Int64(), "priority fee per unit")
// }

// func gwei(i int64) *big.Int {
// 	const g = 1_000_000_000
// 	return big.NewInt(i * g)
// }
