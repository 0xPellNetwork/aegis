package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/thorchain/tss/go-tss/blame"

	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func TestConvertNodes(t *testing.T) {
	tests := []struct {
		name     string
		input    []blame.Node
		expected []*types.Node
	}{
		{
			name:     "TestEmptyInput",
			input:    []blame.Node{},
			expected: nil,
		},
		{
			name:     "TestNilInput",
			input:    nil,
			expected: nil,
		},
		{
			name: "TestSingleInput",
			input: []blame.Node{
				{Pubkey: "pubkey1", BlameSignature: []byte("signature1"), BlameData: []byte("data1")},
			},
			expected: []*types.Node{
				{PubKey: "pubkey1", BlameSignature: []byte("signature1"), BlameData: []byte("data1")},
			},
		},
		{
			name: "TestMultipleInputs",
			input: []blame.Node{
				{Pubkey: "pubkey1", BlameSignature: []byte("signature1"), BlameData: []byte("data1")},
				{Pubkey: "pubkey2", BlameSignature: []byte("signature2"), BlameData: []byte("data2")},
			},
			expected: []*types.Node{
				{PubKey: "pubkey1", BlameSignature: []byte("signature1"), BlameData: []byte("data1")},
				{PubKey: "pubkey2", BlameSignature: []byte("signature2"), BlameData: []byte("data2")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := types.ConvertNodes(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
