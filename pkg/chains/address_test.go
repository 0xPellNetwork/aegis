package chains

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	. "gopkg.in/check.v1"
)

func TestPackage(t *testing.T) { TestingT(t) }

func TestAddress(t *testing.T) {
	addr := NewAddress("bnb1lejrrtta9cgr49fuh7ktu3sddhe0ff7wenlpn6")
	require.EqualValuesf(t, NoAddress, addr, "address string should be empty")
	require.True(t, addr.IsEmpty())

	addr = NewAddress("bogus")
	require.EqualValuesf(t, NoAddress, addr, "address string should be empty")
	require.True(t, addr.IsEmpty())

	addr = NewAddress("0x90f2b1ae50e6018230e90a33f98c7844a0ab635a")
	require.EqualValuesf(t, "0x90f2b1ae50e6018230e90a33f98c7844a0ab635a", addr.String(), "address string should be equal")
	require.False(t, addr.IsEmpty())

	addr2 := NewAddress("0x95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5")
	require.EqualValuesf(t, "0x95222290DD7278Aa3Ddd389Cc1E1d165CC4BAfe5", addr2.String(), "address string should be equal")
	require.False(t, addr.IsEmpty())

	require.False(t, addr.Equals(addr2))
	require.True(t, addr.Equals(addr))
}

func TestConvertRecoverToError(t *testing.T) {
	t.Run("recover with string", func(t *testing.T) {
		err := ConvertRecoverToError("error occurred")
		require.Error(t, err)
		require.Equal(t, "error occurred", err.Error())
	})

	t.Run("recover with error", func(t *testing.T) {
		originalErr := errors.New("original error")
		err := ConvertRecoverToError(originalErr)
		require.Error(t, err)
		require.Equal(t, originalErr, err)
	})

	t.Run("recover with non-string and non-error", func(t *testing.T) {
		err := ConvertRecoverToError(12345)
		require.Error(t, err)
		require.Equal(t, "12345", err.Error())
	})

	t.Run("recover with nil", func(t *testing.T) {
		err := ConvertRecoverToError(nil)
		require.Error(t, err)
		require.Equal(t, "<nil>", err.Error())
	})
}
