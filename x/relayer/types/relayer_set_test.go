package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/relayer/types"
)

func TestObserverSet(t *testing.T) {
	observerSet := sample.ObserverSet_pell(4)

	require.Equal(t, int(4), observerSet.Len())
	require.Equal(t, uint64(4), observerSet.LenUint())
	err := observerSet.Validate()
	require.NoError(t, err)

	observerSet.RelayerList[0] = "invalid"
	err = observerSet.Validate()
	require.Error(t, err)
}

func TestCheckReceiveStatus(t *testing.T) {
	err := types.CheckReceiveStatus(chains.ReceiveStatus_SUCCESS)
	require.NoError(t, err)
	err = types.CheckReceiveStatus(chains.ReceiveStatus_FAILED)
	require.NoError(t, err)
	err = types.CheckReceiveStatus(chains.ReceiveStatus_CREATED)
	require.Error(t, err)
}
