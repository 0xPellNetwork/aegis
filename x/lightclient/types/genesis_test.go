package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/pkg/proofs"
	"github.com/pell-chain/pellcore/testutil/sample"
	"github.com/pell-chain/pellcore/x/lightclient/types"
)

func TestGenesisState_Validate(t *testing.T) {
	duplicatedHash := sample.Hash().Bytes()

	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				VerificationFlags: types.VerificationFlags{
					EthTypeChainEnabled: false,
					BtcTypeChainEnabled: true,
				},
				BlockHeaders: []proofs.BlockHeader{
					sample.BlockHeader(sample.Hash().Bytes()),
					sample.BlockHeader(sample.Hash().Bytes()),
					sample.BlockHeader(sample.Hash().Bytes()),
				},
				ChainStates: []types.ChainState{
					sample.ChainState(chains.EthChain().Id),
					sample.ChainState(chains.BscMainnetChain().Id),
				},
			},
			valid: true,
		},
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "duplicate block headers is invalid",
			genState: &types.GenesisState{
				BlockHeaders: []proofs.BlockHeader{
					sample.BlockHeader(sample.Hash().Bytes()),
					sample.BlockHeader(duplicatedHash),
					sample.BlockHeader(duplicatedHash),
				},
			},
			valid: false,
		},
		{
			desc: "duplicate chain state is invalid",
			genState: &types.GenesisState{
				ChainStates: []types.ChainState{
					sample.ChainState(chains.EthChain().Id),
					sample.ChainState(chains.EthChain().Id),
					sample.ChainState(chains.BscMainnetChain().Id),
				},
			},
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
