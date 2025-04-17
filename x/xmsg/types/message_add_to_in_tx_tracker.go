package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/pell-chain/pellcore/pkg/chains"
	"github.com/pell-chain/pellcore/pkg/coin"
)

const TypeMsgAddToInTxTracker = "AddToInTxTracker"

var _ sdk.Msg = &MsgAddToInTxTracker{}

func NewMsgAddToInTxTracker(creator string, chain int64, coinType coin.CoinType, txHash string) *MsgAddToInTxTracker {
	return &MsgAddToInTxTracker{
		Signer:   creator,
		ChainId:  chain,
		TxHash:   txHash,
		CoinType: coinType,
	}
}

func (msg *MsgAddToInTxTracker) Route() string {
	return RouterKey
}

func (msg *MsgAddToInTxTracker) Type() string {
	return TypeMsgAddToInTxTracker
}

func (msg *MsgAddToInTxTracker) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAddToInTxTracker) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddToInTxTracker) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	chain, exist := chains.GetChainByChainId(msg.ChainId)
	if !exist {
		return errorsmod.Wrapf(ErrInvalidChainID, "chain id (%d)", msg.ChainId)
	}

	if msg.Proof != nil && !chain.SupportMerkleProof() {
		return errorsmod.Wrapf(ErrProofVerificationFail, "chain id %d does not support proof-based trackers", msg.ChainId)
	}
	_, ok := coin.CoinType_value[msg.CoinType.String()]
	if !ok {
		return errorsmod.Wrapf(ErrProofVerificationFail, "coin-type not supported")
	}
	return nil
}
