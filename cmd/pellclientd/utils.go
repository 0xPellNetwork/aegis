package main

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"

	"github.com/pell-chain/pellcore/relayer/authz"
	"github.com/pell-chain/pellcore/relayer/config"
	"github.com/pell-chain/pellcore/relayer/keys"
	"github.com/pell-chain/pellcore/relayer/pellcore"
)

func CreateAuthzSigner(granter string, grantee sdk.AccAddress) {
	authz.SetupAuthZSignerList(granter, grantee)
}

func CreatePellcoreClient(cfg config.Config, hotkeyPassword string, logger zerolog.Logger) (*pellcore.PellCoreBridge, error) {
	hotKey := cfg.AuthzHotkey
	if cfg.HsmMode {
		hotKey = cfg.HsmHotKey
	}

	chainIP := cfg.PellCoreURL

	kb, _, err := keys.GetKeyringKeybase(cfg, hotkeyPassword)
	if err != nil {
		return nil, err
	}

	granterAddreess, err := sdk.AccAddressFromBech32(cfg.AuthzGranter)
	if err != nil {
		return nil, err
	}

	k := keys.NewKeysWithKeybase(kb, granterAddreess, cfg.AuthzHotkey, hotkeyPassword)

	bridge, err := pellcore.NewClient(logger, k, chainIP, hotKey, cfg.ChainID, cfg.HsmMode, cfg.PellTxMsgLength)
	if err != nil {
		return nil, err
	}

	return bridge, nil
}
