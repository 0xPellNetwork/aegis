package chains_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
)

func TestReceiveStatusFromString(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    chains.ReceiveStatus
		wantErr bool
	}{
		{
			name:    "created",
			str:     "0",
			want:    chains.ReceiveStatus_CREATED,
			wantErr: false,
		},
		{
			name:    "success",
			str:     "1",
			want:    chains.ReceiveStatus_SUCCESS,
			wantErr: false,
		},
		{
			name:    "failed",
			str:     "2",
			want:    chains.ReceiveStatus_FAILED,
			wantErr: false,
		},
		{
			name:    "wrong status",
			str:     "3",
			want:    chains.ReceiveStatus(0),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := chains.ReceiveStatusFromString(tt.str)
			if tt.wantErr {
				require.Error(t, err)
			} else if got != tt.want {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
