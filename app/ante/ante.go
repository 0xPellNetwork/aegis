// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE
package ante

import (
	"fmt"
	"runtime/debug"

	errorsmod "cosmossdk.io/errors"
	tmlog "cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	relayertypes "github.com/pell-chain/pellcore/x/relayer/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

func ValidateHandlerOptions(options HandlerOptions) error {
	if options.AccountKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "bank keeper is required for AnteHandler")
	}

	if options.IBCKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "IBCKeeper is required for AnteHandler")
	}

	if options.WasmConfig == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "WasmConfig is required for AnteHandler")
	}

	if options.WasmKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "WasmKeeper is required for AnteHandler")
	}

	if options.TXCounterStoreService == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "TXCounterStoreService is required for AnteHandler")
	}

	if options.CircuitKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "circuit keeper is required for AnteHandler")
	}

	if options.SignModeHandler == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.FeeMarketKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "fee market keeper is required for AnteHandler")
	}
	if options.EvmKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "evm keeper is required for AnteHandler")
	}
	if options.RelayerKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "relayer keeper is required for AnteHandler")
	}

	return nil
}

// NewAnteHandler returns an ante handler responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if err := ValidateHandlerOptions(options); err != nil {
		return nil, err
	}

	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler

		defer Recover(ctx.Logger(), &err)

		txWithExtensions, ok := tx.(authante.HasExtensionOptionsTx)
		if ok {
			opts := txWithExtensions.GetExtensionOptions()
			if len(opts) > 0 {
				switch typeURL := opts[0].GetTypeUrl(); typeURL {
				case "/ethermint.evm.v1.ExtensionOptionsEthereumTx":
					// handle as *evmtypes.MsgEthereumTx
					anteHandler = newEthAnteHandler(options)
				case "/ethermint.types.v1.ExtensionOptionsWeb3Tx":
					// Deprecated: Handle as normal Cosmos SDK tx, except signature is checked for Legacy EIP712 representation
					anteHandler = NewLegacyCosmosAnteHandlerEip712(options)
				case "/ethermint.types.v1.ExtensionOptionDynamicFeeTx":
					// cosmos-sdk tx with dynamic fee extension
					anteHandler = newCosmosAnteHandler(options)
				default:
					return ctx, errorsmod.Wrapf(
						errortypes.ErrUnknownExtensionOptions,
						"rejecting tx with unsupported extension option: %s", typeURL,
					)
				}

				return anteHandler(ctx, tx, sim)
			}
		}

		// handle as totally normal Cosmos SDK tx
		switch tx.(type) {
		case sdk.Tx:
			// default: handle as normal Cosmos SDK tx
			anteHandler = newCosmosAnteHandler(options)

			// if tx is a system tx, and singer is authorized, use system tx handler

			isAuthorized := func(creator string) bool {
				return options.RelayerKeeper.IsNonTombstonedObserver(ctx, creator)
			}

			if IsSystemTx(tx, isAuthorized) {
				anteHandler = newCosmosAnteHandlerForSystemTx(options)
			}

			// if tx is MsgCreatorValidator, use the newCosmosAnteHandlerForSystemTx handler to
			// exempt gas fee requirement in genesis because it's not possible to pay gas fee in genesis
			if len(tx.GetMsgs()) == 1 {
				if _, ok := tx.GetMsgs()[0].(*stakingtypes.MsgCreateValidator); ok && ctx.BlockHeight() == 0 {
					anteHandler = newCosmosAnteHandlerForSystemTx(options)
				}
			}

		default:
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid transaction type: %T", tx)
		}

		return anteHandler(ctx, tx, sim)
	}, nil
}

func Recover(logger tmlog.Logger, err *error) {
	if r := recover(); r != nil {
		if err != nil {
			// #nosec G703 err is checked non-nil above
			*err = errorsmod.Wrapf(errortypes.ErrPanic, "%v", r)
		}

		if e, ok := r.(error); ok {
			logger.Error(
				"ante handler panicked",
				"error", e,
				"stack trace", string(debug.Stack()),
			)
		} else {
			logger.Error(
				"ante handler panicked",
				"recover", fmt.Sprintf("%v", r),
			)
		}
	}
}

// IsSystemTx determines whether tx is a system tx that's signed by an authorized signer
// system tx are special types of txs (see in the switch below), or such txs wrapped inside a MsgExec
// the parameter isAuthorizedSigner is a caller specified function that determines whether the signer of
// the tx is authorized.
func IsSystemTx(tx sdk.Tx, isAuthorizedSigner func(string) bool) bool {
	// the following determines whether the tx is a system tx which will uses different handler
	// System txs are always single Msg txs, optionally wrapped by one level of MsgExec
	if len(tx.GetMsgs()) != 1 { // this is not a system tx
		return false
	}
	msg := tx.GetMsgs()[0]

	// if wrapped inside a MsgExec, unwrap it and reveal the innerMsg.
	var innerMsg sdk.Msg
	innerMsg = msg
	if mm, ok := msg.(*authz.MsgExec); ok { // authz tx; look inside it
		msgs, err := mm.GetMessages()
		if err == nil && len(msgs) == 1 {
			innerMsg = msgs[0]
		}
	}

	signer := ""
	// TODO: get signer from the innerMsg maybe better
	switch msg := innerMsg.(type) {
	case *xmsgtypes.MsgVoteGasPrice:
		signer = msg.Signer
	case *xmsgtypes.MsgVoteOnObservedInboundTx:
		signer = msg.Signer
	case *xmsgtypes.MsgVoteOnObservedOutboundTx:
		signer = msg.Signer
	case *xmsgtypes.MsgAddToOutTxTracker:
		signer = msg.Signer
	case *xmsgtypes.MsgVoteInboundBlock:
		signer = msg.Signer
	case *relayertypes.MsgVoteBlockHeader:
		signer = msg.Signer
	case *relayertypes.MsgVoteTSS:
		signer = msg.Signer
	case *relayertypes.MsgAddBlameVote:
		signer = msg.Signer
	case *xmsgtypes.MsgVoteOnGasRecharge:
		signer = msg.Signer
	case *xmsgtypes.MsgVoteOnPellRecharge:
		signer = msg.Signer
	case *xmsgtypes.MsgInboundTxMaintenance: // TODO: remove this after the next upgrade
		signer = msg.Signer
	default:
		return false
	}

	return isAuthorizedSigner(signer)
}
