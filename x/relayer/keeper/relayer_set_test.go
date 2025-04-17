package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/pell-chain/pellcore/testutil/keeper"
	"github.com/pell-chain/pellcore/testutil/sample"
)

func TestKeeper_GetObserverSet(t *testing.T) {
	t.Run("get observer set", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		os := sample.ObserverSet_pell(10)
		_, found := k.GetObserverSet(ctx)
		require.False(t, found)
		k.SetObserverSet(ctx, os)
		tfm, found := k.GetObserverSet(ctx)
		require.True(t, found)
		require.Equal(t, os, tfm)
	})
}

func TestKeeper_IsAddressPartOfObserverSet(t *testing.T) {
	t.Run("address is part of observer set", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		os := sample.ObserverSet_pell(10)
		require.False(t, k.IsAddressPartOfObserverSet(ctx, os.RelayerList[0]))
		k.SetObserverSet(ctx, os)
		require.True(t, k.IsAddressPartOfObserverSet(ctx, os.RelayerList[0]))
		require.False(t, k.IsAddressPartOfObserverSet(ctx, sample.AccAddress()))
	})
}

func TestKeeper_AddObserverToSet(t *testing.T) {
	t.Run("add observer to set", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		os := sample.ObserverSet_pell(10)
		k.SetObserverSet(ctx, os)
		newObserver := sample.AccAddress()
		k.AddObserverToSet(ctx, newObserver)
		require.True(t, k.IsAddressPartOfObserverSet(ctx, newObserver))
		require.False(t, k.IsAddressPartOfObserverSet(ctx, sample.AccAddress()))
		osNew, found := k.GetObserverSet(ctx)
		require.True(t, found)
		require.Len(t, osNew.RelayerList, len(os.RelayerList)+1)
	})

	t.Run("add observer to set if set doesn't exist", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		newObserver := sample.AccAddress()
		k.AddObserverToSet(ctx, newObserver)
		require.True(t, k.IsAddressPartOfObserverSet(ctx, newObserver))
		osNew, found := k.GetObserverSet(ctx)
		require.True(t, found)
		require.Len(t, osNew.RelayerList, 1)

		// add same address again, len doesn't change
		k.AddObserverToSet(ctx, newObserver)
		require.True(t, k.IsAddressPartOfObserverSet(ctx, newObserver))
		osNew, found = k.GetObserverSet(ctx)
		require.True(t, found)
		require.Len(t, osNew.RelayerList, 1)
	})
}

func TestKeeper_RemoveObserverFromSet(t *testing.T) {
	t.Run("remove observer from set", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		os := sample.ObserverSet_pell(10)
		k.RemoveObserverFromSet(ctx, os.RelayerList[0])
		k.SetObserverSet(ctx, os)
		k.RemoveObserverFromSet(ctx, os.RelayerList[0])
		require.False(t, k.IsAddressPartOfObserverSet(ctx, os.RelayerList[0]))
		osNew, found := k.GetObserverSet(ctx)
		require.True(t, found)
		require.Len(t, osNew.RelayerList, len(os.RelayerList)-1)
	})
}

func TestKeeper_UpdateObserverAddress(t *testing.T) {
	t.Run("update observer address", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		oldObserverAddress := sample.AccAddress()
		newObserverAddress := sample.AccAddress()
		observerSet := sample.ObserverSet_pell(10)
		observerSet.RelayerList = append(observerSet.RelayerList, oldObserverAddress)
		err := k.UpdateObserverAddress(ctx, oldObserverAddress, newObserverAddress)
		require.Error(t, err)
		k.SetObserverSet(ctx, observerSet)
		err = k.UpdateObserverAddress(ctx, oldObserverAddress, newObserverAddress)
		require.NoError(t, err)
		observerSet, found := k.GetObserverSet(ctx)
		require.True(t, found)
		require.Equal(t, newObserverAddress, observerSet.RelayerList[len(observerSet.RelayerList)-1])
	})
	t.Run("should error if observer address not found", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		oldObserverAddress := sample.AccAddress()
		newObserverAddress := sample.AccAddress()
		observerSet := sample.ObserverSet_pell(10)
		observerSet.RelayerList = append(observerSet.RelayerList, oldObserverAddress)
		k.SetObserverSet(ctx, observerSet)
		err := k.UpdateObserverAddress(ctx, sample.AccAddress(), newObserverAddress)
		require.Error(t, err)
	})
	t.Run("update observer address long observerList", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		oldObserverAddress := sample.AccAddress()
		newObserverAddress := sample.AccAddress()
		observerSet := sample.ObserverSet_pell(10000)
		observerSet.RelayerList = append(observerSet.RelayerList, oldObserverAddress)
		k.SetObserverSet(ctx, observerSet)
		err := k.UpdateObserverAddress(ctx, oldObserverAddress, newObserverAddress)
		require.NoError(t, err)
		observerMappers, found := k.GetObserverSet(ctx)
		require.True(t, found)
		require.Equal(t, newObserverAddress, observerMappers.RelayerList[len(observerMappers.RelayerList)-1])
	})
	t.Run("update observer address short observerList", func(t *testing.T) {
		k, ctx, _, _ := keepertest.RelayerKeeper(t)
		oldObserverAddress := sample.AccAddress()
		newObserverAddress := sample.AccAddress()
		observerSet := sample.ObserverSet_pell(1)
		observerSet.RelayerList = append(observerSet.RelayerList, oldObserverAddress)
		k.SetObserverSet(ctx, observerSet)
		err := k.UpdateObserverAddress(ctx, oldObserverAddress, newObserverAddress)
		require.NoError(t, err)
		observerMappers, found := k.GetObserverSet(ctx)
		require.True(t, found)
		require.Equal(t, newObserverAddress, observerMappers.RelayerList[len(observerMappers.RelayerList)-1])
	})
}
