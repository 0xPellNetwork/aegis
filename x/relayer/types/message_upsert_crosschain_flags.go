package types

import (
	"errors"

	cosmoserrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
)

const (
	TypeMsgUpsertCrosschainFlags = "update_crosschain_flags"
)

var _ sdk.Msg = &MsgUpsertCrosschainFlags{}

func NewMsgUpsertCrosschainFlags(creator string, isInboundEnabled, isOutboundEnabled bool) *MsgUpsertCrosschainFlags {
	return &MsgUpsertCrosschainFlags{
		Signer:            creator,
		IsInboundEnabled:  isInboundEnabled,
		IsOutboundEnabled: isOutboundEnabled,
	}
}

func (msg *MsgUpsertCrosschainFlags) Route() string {
	return RouterKey
}

func (msg *MsgUpsertCrosschainFlags) Type() string {
	return TypeMsgUpsertCrosschainFlags
}

func (msg *MsgUpsertCrosschainFlags) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpsertCrosschainFlags) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpsertCrosschainFlags) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return cosmoserrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.GasPriceIncreaseFlags != nil {
		if err := msg.GasPriceIncreaseFlags.Validate(); err != nil {
			return cosmoserrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}
	}

	return nil
}

func (gpf GasPriceIncreaseFlags) Validate() error {
	if gpf.EpochLength <= 0 {
		return errors.New("epoch length must be positive")
	}
	if gpf.RetryInterval <= 0 {
		return errors.New("retry interval must be positive")
	}
	return nil
}

// GetRequiredPolicyType returns the required policy type for the message to execute the message
// Group emergency should only be able to stop or disable functionalities in case of emergency
// this concerns disabling inbound and outbound txs or block header verification
// every other action requires group admin
// TODO: add separate message for each group
func (msg *MsgUpsertCrosschainFlags) GetRequiredPolicyType() authoritytypes.PolicyType {
	if msg.IsInboundEnabled || msg.IsOutboundEnabled {
		return authoritytypes.PolicyType_GROUP_OPERATIONAL
	}
	if msg.GasPriceIncreaseFlags != nil {
		return authoritytypes.PolicyType_GROUP_OPERATIONAL
	}
	if msg.BlockHeaderVerificationFlags != nil && (msg.BlockHeaderVerificationFlags.IsEthTypeChainEnabled || msg.BlockHeaderVerificationFlags.IsBtcTypeChainEnabled) {
		return authoritytypes.PolicyType_GROUP_OPERATIONAL

	}
	return authoritytypes.PolicyType_GROUP_EMERGENCY
}
