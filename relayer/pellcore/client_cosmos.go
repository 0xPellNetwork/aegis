package pellcore

import (
	"context"
	"encoding/json"
	"fmt"

	sdkmath "cosmossdk.io/math"
	tmhttp "github.com/cometbft/cometbft/rpc/client/http"
	comebfttypes "github.com/cometbft/cometbft/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"

	"github.com/pell-chain/pellcore/cmd/pellcored/config"
)

// GetGenesisSupply returns the genesis supply.
// NOTE that this method is brittle as it uses STATEFUL connection
func (c *PellCoreBridge) GetGenesisSupply(ctx context.Context) (sdkmath.Int, error) {
	tmURL := fmt.Sprintf("http://%s", c.config.ChainRPC)

	s, err := tmhttp.New(tmURL, "/websocket")
	if err != nil {
		return sdkmath.ZeroInt(), errors.Wrap(err, "failed to create tm client")
	}

	// nolint:errcheck
	defer s.Stop()

	res, err := s.Genesis(ctx)
	if err != nil {
		return sdkmath.ZeroInt(), errors.Wrap(err, "failed to get genesis")
	}

	bankState, err := parseBankGenesisState(res.Genesis)
	if err != nil {
		return sdkmath.ZeroInt(), err
	}

	return bankState.Supply.AmountOf(config.BaseDenom), nil
}

// GetPellHotKeyBalance returns the pell hot key balance
func (c *PellCoreBridge) GetPellHotKeyBalance(ctx context.Context) (sdkmath.Int, error) {
	address, err := c.keys.GetAddress()
	if err != nil {
		return sdkmath.ZeroInt(), errors.Wrap(err, "failed to get address")
	}

	in := &banktypes.QueryBalanceRequest{
		Address: address.String(),
		Denom:   config.BaseDenom,
	}

	resp, err := c.Clients.Bank.Balance(ctx, in)
	if err != nil {
		return sdkmath.ZeroInt(), errors.Wrap(err, "failed to get pell hot key balance")
	}

	return resp.Balance.Amount, nil
}

func parseBankGenesisState(genDoc *comebfttypes.GenesisDoc) (*banktypes.GenesisState, error) {
	var appState map[string]json.RawMessage
	err := json.Unmarshal(genDoc.AppState, &appState)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal app state: %w", err)
	}

	var bankGenesis banktypes.GenesisState
	if err := json.Unmarshal(appState[banktypes.ModuleName], &bankGenesis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bank genesis state: %w", err)
	}

	return &bankGenesis, nil
}
