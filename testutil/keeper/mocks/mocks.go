package mocks

import (
	emissionstypes "github.com/0xPellNetwork/aegis/x/emissions/types"
	lightclienttypes "github.com/0xPellNetwork/aegis/x/lightclient/types"
	pevmtypes "github.com/0xPellNetwork/aegis/x/pevm/types"
	relayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
	restakingtypes "github.com/0xPellNetwork/aegis/x/restaking/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
	xsecuritytypes "github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

/**
 * Emissions Mocks
 */

//go:generate mockery --name EmissionAccountKeeper --filename account.go --case underscore --output ./emissions
type EmissionAccountKeeper interface {
	emissionstypes.AccountKeeper
}

//go:generate mockery --name EmissionBankKeeper --filename bank.go --case underscore --output ./emissions
type EmissionBankKeeper interface {
	emissionstypes.BankKeeper
}

//go:generate mockery --name EmissionStakingKeeper --filename staking.go --case underscore --output ./emissions
type EmissionStakingKeeper interface {
	emissionstypes.StakingKeeper
}

//go:generate mockery --name EmissionRelayerKeeper --filename relayer.go --case underscore --output ./emissions
type EmissionRelayerKeeper interface {
	emissionstypes.RelayerKeeper
}

//go:generate mockery --name EmissionParamStore --filename param_store.go --case underscore --output ./emissions
type EmissionParamStore interface {
	emissionstypes.ParamStore
}

/**
 * Relayer Mocks
 */

//go:generate mockery --name RelayerStakingKeeper --filename staking.go --case underscore --output ./relayer
type RelayerStakingKeeper interface {
	relayertypes.StakingKeeper
}

//go:generate mockery --name RelayerSlashingKeeper --filename slashing.go --case underscore --output ./relayer
type RelayerSlashingKeeper interface {
	relayertypes.SlashingKeeper
}

//go:generate mockery --name RelayerAuthorityKeeper --filename authority.go --case underscore --output ./relayer
type RelayerAuthorityKeeper interface {
	relayertypes.AuthorityKeeper
}

//go:generate mockery --name RelayerLightclientKeeper --filename lightclient.go --case underscore --output ./relayer
type RelayerLightclientKeeper interface {
	relayertypes.LightclientKeeper
}

//go:generate mockery --name RelayerPevmKeeper --filename pevm.go --case underscore --output ./relayer
type RelayerPevmKeeper interface {
	relayertypes.PevmKeeper
}

/**
 * Lightclient Mocks
 */

//go:generate mockery --name LightclientAuthorityKeeper --filename authority.go --case underscore --output ./lightclient
type LightclientAuthorityKeeper interface {
	lightclienttypes.AuthorityKeeper
}

/**
 * Xmsg Mocks
 */

//go:generate mockery --name XmsgAccountKeeper --filename account.go --case underscore --output ./xmsg
type XmsgAccountKeeper interface {
	xmsgtypes.AccountKeeper
}

//go:generate mockery --name XmsgBankKeeper --filename bank.go --case underscore --output ./xmsg
type XmsgBankKeeper interface {
	xmsgtypes.BankKeeper
}

//go:generate mockery --name XmsgStakingKeeper --filename staking.go --case underscore --output ./xmsg
type XmsgStakingKeeper interface {
	xmsgtypes.StakingKeeper
}

//go:generate mockery --name XmsgRelayerKeeper --filename relayer.go --case underscore --output ./xmsg
type XmsgRelayerKeeper interface {
	xmsgtypes.RelayerKeeper
}

//go:generate mockery --name XmsgPevmKeeper --filename pevm.go --case underscore --output ./xmsg
type XmsgPevmKeeper interface {
	xmsgtypes.PevmKeeper
}

//go:generate mockery --name XmsgAuthorityKeeper --filename authority.go --case underscore --output ./xmsg
type XmsgAuthorityKeeper interface {
	xmsgtypes.AuthorityKeeper
}

//go:generate mockery --name XmsgLightclientKeeper --filename lightclient.go --case underscore --output ./xmsg
type XmsgLightclientKeeper interface {
	xmsgtypes.LightclientKeeper
}

/**
 * Pevm Mocks
 */

//go:generate mockery --name PevmAccountKeeper --filename account.go --case underscore --output ./pevm
type PevmAccountKeeper interface {
	pevmtypes.AccountKeeper
}

//go:generate mockery --name PevmBankKeeper --filename bank.go --case underscore --output ./pevm
type PevmBankKeeper interface {
	pevmtypes.BankKeeper
}

//go:generate mockery --name PevmRelayerKeeper --filename relayer.go --case underscore --output ./pevm
type PevmRelayerKeeper interface {
	pevmtypes.RelayerKeeper
}

//go:generate mockery --name PevmEVMKeeper --filename evm.go --case underscore --output ./pevm
type PevmEVMKeeper interface {
	pevmtypes.EVMKeeper
}

//go:generate mockery --name PevmAuthorityKeeper --filename authority.go --case underscore --output ./pevm
type PevmAuthorityKeeper interface {
	pevmtypes.AuthorityKeeper
}

/**
 * Restaking Mocks
 */

//go:generate mockery --name RestakingAccountKeeper --filename account.go --case underscore --output ./restaking
type RestakingAccountKeeper interface {
	restakingtypes.AccountKeeper
}

//go:generate mockery --name RestakingEVMKeeper --filename evm.go --case underscore --output ./restaking
type RestakingEVMKeeper interface {
	restakingtypes.EVMKeeper
}

//go:generate mockery --name RestakingBankKeeper --filename bank.go --case underscore --output ./restaking
type RestakingBankKeeper interface {
	restakingtypes.BankKeeper
}

//go:generate mockery --name RestakingRelayerKeeper --filename relayer.go --case underscore --output ./restaking
type RestakingRelayerKeeper interface {
	restakingtypes.RelayerKeeper
}

//go:generate mockery --name RestakingAuthorityKeeper --filename authority.go --case underscore --output ./restaking
type RestakingAuthorityKeeper interface {
	restakingtypes.AuthorityKeeper
}

//go:generate mockery --name RestakingPevmKeeper --filename pevm.go --case underscore --output ./restaking
type RestakingPevmKeeper interface {
	restakingtypes.PevmKeeper
}

/**
 * XSecurity Mocks
 */

//go:generate mockery --name XSecurityStakingKeeper --filename staking.go --case underscore --output ./xsecurity
type XSecurityStakingKeeper interface {
	xsecuritytypes.StakingKeeper
}

//go:generate mockery --name XSecuritySlashingKeeper --filename slashing.go --case underscore --output ./xsecurity
type XSecuritySlashingKeeper interface {
	xsecuritytypes.SlashingKeeper
}

//go:generate mockery --name XSecurityStakingHooks --filename staking_hooks.go --case underscore --output ./xsecurity
type XSecurityStakingHooks interface {
	xsecuritytypes.StakingHooks
}

//go:generate mockery --name XSecurityAuthorityKeeper --filename authority.go --case underscore --output ./xsecurity
type XSecurityAuthorityKeeper interface {
	xsecuritytypes.AuthorityKeeper
}

//go:generate mockery --name XSecurityRelayerKeeper --filename relayer.go --case underscore --output ./xsecurity
type XSecurityRelayerKeeper interface {
	xsecuritytypes.RelayerKeeper
}

//go:generate mockery --name XSecurityLightclientKeeper --filename lightclient.go --case underscore --output ./xsecurity
type XSecurityLightclientKeeper interface {
	xsecuritytypes.LightclientKeeper
}

//go:generate mockery --name XSecurityPevmKeeper --filename pevm.go --case underscore --output ./xsecurity
type XSecurityPevmKeeper interface {
	xsecuritytypes.PevmKeeper
}

//go:generate mockery --name XSecurityRestakingKeeper --filename restaking.go --case underscore --output ./xsecurity
type XSecurityRestakingKeeper interface {
	xsecuritytypes.RestakingKeeper
}
