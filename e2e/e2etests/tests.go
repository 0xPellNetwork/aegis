package e2etests

import "github.com/0xPellNetwork/aegis/e2e/runner"

const (
	// pell contract
	TEST_UPGRADE_SYSTEM_CONTRACT                = "pell_upgrade_system_contract"
	TEST_DEPOSIT_FOR_INBOUND                    = "pell_xmsg_deposit_for_inbound"
	TEST_REGISTER_OPERATOR_FOR_OUTBOUND         = "pell_register_operator_for_outbound"
	TEST_DELEGATE                               = "pell_delegate"
	TEST_WITHDRAWAL                             = "pell_withdrawal"
	TEST_UNDELEGATE                             = "pell_undelegate"
	TEST_RECHARGE_PELL_TOKEN                    = "pell_pell_recharge"
	TEST_RECHARGE_GAS                           = "pell_recharge_gas"
	TEST_DVS_CREATE_REGISTRY_ROUTER             = "pell_dvs_create_registry_router"
	TEST_DVS_CREATE_GROUP_ON_PELL               = "pell_dvs_create_group_on_pell"
	TEST_DVS_ADD_SUPPORTED_CHAIN                = "pell_dvs_add_supported_chain"
	TEST_DVS_SYNC_GROUP                         = "pell_dvs_sync_group"
	TEST_DVS_CREATE_GROUP                       = "pell_dvs_create_group"
	TEST_DVS_SET_OPERATOR_SET_PARAMS            = "pell_dvs_set_operator_set_params"
	TEST_DVS_SET_GROUP_EJECTION_PARAMS          = "pell_dvs_set_group_ejection_params"
	TEST_DVS_SET_EJECTION_COOLDOWN              = "pell_dvs_set_ejection_cooldown"
	TEST_DVS_REGISTER_OPERATOR_BEFORE_SYNCGROUP = "pell_dvs_register_operator_before_syncgroup"
	TEST_DVS_REGISTER_OPERATOR                  = "pell_dvs_register_operator"
	TEST_DVS_ADD_POOLS                          = "pell_dvs_add_pools"
	TEST_DVS_REMOVE_POOLS                       = "pell_dvs_remove_pools"
	TEST_DVS_MODIFY_POOL_PARAMS                 = "pell_dvs_modify_pool_params"
	TEST_DVS_UPDATE_OPERATORS                   = "pell_dvs_update_operators"
	TEST_DVS_UPDATE_OPERATORS_FOR_GROUP         = "pell_dvs_update_operators_for_group"
	TEST_DVS_DEREGISTER_OPERATOR                = "pell_dvs_deregister_operator"
	TEST_DVS_REGISTER_OPERATOR_WITH_CHURN       = "pell_dvs_register_operator_with_churn"
	TEST_DVS_EJECT_OPERATORS                    = "pell_dvs_eject_operators"
	TEST_BRIDGE_PELL_INBOUND                    = "pell_bridge_pell_inbound"
	TEST_BRIDGE_PELL_OUTBOUND                   = "pell_bridge_pell_outbound"
	TEST_DVS_SYNC_GROUP_FAILED                  = "pell_dvs_sync_group_failed"
	// LST Token dual staking
	TEST_LST_SET_VOTING_POWER_RATIO                  = "lst_set_voting_power_ratio"
	TEST_LST_CREATE_REGISTRY_ROUTER                  = "lst_create_registry_router"
	TEST_LST_CREATE_GROUP                            = "lst_create_group"
	TEST_LST_REGISTER_OPERATOR_TO_DELEGATION_MANAGER = "lst_register_operator_to_delegation_manager"
	TEST_LST_REGISTER_OPERATOR                       = "lst_register_operator"
	TEST_LST_DEPOSIT                                 = "lst_deposit_and_delegate_to_operator"
	TEST_LST_DELEGATE                                = "lst_delegate_to_operator"
	TEST_LST_OPERATE_POOLS                           = "lst_operate_pools"
)

// AllE2ETests is an ordered list of all e2e tests
var AllE2ETests = []runner.E2ETest{
	runner.NewE2ETest(
		TEST_UPGRADE_SYSTEM_CONTRACT,
		"pell contract: TestPellDepositForInBound",
		[]runner.ArgDefinition{
			{Description: "test pell deposit for inbound tx", DefaultValue: "10000000000000000008"},
		},
		TestUpgradeSystemContract,
	),
	runner.NewE2ETest(
		TEST_DEPOSIT_FOR_INBOUND,
		"pell contract: TestPellDepositForInBound",
		[]runner.ArgDefinition{
			{Description: "test pell deposit for inbound tx", DefaultValue: "10000000000000000008"},
		},
		TestDeposit,
	),
	runner.NewE2ETest(
		TEST_REGISTER_OPERATOR_FOR_OUTBOUND,
		"pell contract: TestPellRegisterOperatorForOutbound",
		[]runner.ArgDefinition{
			{Description: "test pell register operator for outbound tx", DefaultValue: "10000000000000000008"},
		},
		TestRegisterOperator,
	),
	runner.NewE2ETest(
		TEST_DELEGATE,
		"pell contract: TestPellDelegate",
		[]runner.ArgDefinition{
			{Description: "test pell delegate for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDelegateTo,
	),
	runner.NewE2ETest(
		TEST_WITHDRAWAL,
		"pell contract: TestPellWithdrawal",
		[]runner.ArgDefinition{
			{Description: "test pell withdrawal for both chain", DefaultValue: "10000000000000000008"},
		},
		TestQueueWithdrawals,
	),
	runner.NewE2ETest(
		TEST_UNDELEGATE,
		"pell contract: TestUndelegate",
		[]runner.ArgDefinition{
			{Description: "test pell undelegate for both chain", DefaultValue: "10000000000000000008"},
		},
		TestUndelegate,
	),
	runner.NewE2ETest(
		TEST_RECHARGE_PELL_TOKEN,
		"pell contract: TEST_RECHARGE_PELL_TOKEN",
		[]runner.ArgDefinition{
			{Description: "test pell recharge pell token for both chain", DefaultValue: "10000000000000000008"},
		},
		TestPellRechargeToken,
	),
	runner.NewE2ETest(
		TEST_RECHARGE_GAS,
		"pell contract: TEST_RECHARGE_GAS",
		[]runner.ArgDefinition{
			{Description: "test pell recharge gas token for both chain", DefaultValue: "10000000000000000008"},
		},
		TestGasRechargeToken,
	),

	// DVS tests
	runner.NewE2ETest(
		TEST_DVS_CREATE_REGISTRY_ROUTER,
		"pell contract: TestDVSCreateRegistryRouter",
		[]runner.ArgDefinition{
			{Description: "test pell create registry router for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSCreateRegistryRouter,
	),
	runner.NewE2ETest(
		TEST_DVS_ADD_SUPPORTED_CHAIN,
		"pell contract: TestDVSAddSupportedChain",
		[]runner.ArgDefinition{
			{Description: "test pell add supported chain for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSAddSupportedChain,
	),
	runner.NewE2ETest(
		TEST_DVS_CREATE_GROUP_ON_PELL,
		"pell contract: TestDVSCreateGroupOnPell",
		[]runner.ArgDefinition{
			{Description: "test pell create group on pell for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSCreateGroupOnPell,
	),
	runner.NewE2ETest(
		TEST_DVS_SYNC_GROUP,
		"pell contract: TestDVSSyncGroup",
		[]runner.ArgDefinition{
			{Description: "test pell sync group for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSSyncGroup,
	),
	runner.NewE2ETest(
		TEST_DVS_CREATE_GROUP,
		"pell contract: TestDVSCreateGroup",
		[]runner.ArgDefinition{
			{Description: "test pell create group for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSCreateGroup,
	),
	runner.NewE2ETest(
		TEST_DVS_SET_OPERATOR_SET_PARAMS,
		"pell contract: TestDVSCreateGroup",
		[]runner.ArgDefinition{
			{Description: "test pell set operator set params for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSSetOperatorSetParams,
	),
	runner.NewE2ETest(
		TEST_DVS_SET_GROUP_EJECTION_PARAMS,
		"pell contract: TestDVSSetGroupEjectionParams",
		[]runner.ArgDefinition{
			{Description: "test pell set group ejection params for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSSetGroupEjectionParams,
	),
	runner.NewE2ETest(
		TEST_DVS_SET_EJECTION_COOLDOWN,
		"pell contract: TestDVSSetEjectionCooldown",
		[]runner.ArgDefinition{
			{Description: "test pell set ejection cooldown for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSSetEjectionCooldown,
	),
	runner.NewE2ETest(
		TEST_DVS_REGISTER_OPERATOR_BEFORE_SYNCGROUP,
		"pell contract: TestRegisterOperatorBeforeSyncGroup",
		[]runner.ArgDefinition{
			{Description: "test pell dvs register operator before sync group for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSRegisterOperatorBeforeSyncGroup,
	),
	runner.NewE2ETest(
		TEST_DVS_REGISTER_OPERATOR,
		"pell contract: TestRegisterOperator",
		[]runner.ArgDefinition{
			{Description: "test pell dvs register operator for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSRegisterOperator,
	),
	runner.NewE2ETest(
		TEST_DVS_ADD_POOLS,
		"pell contract: TestDVSAddPools",
		[]runner.ArgDefinition{
			{Description: "test pell dvs add pools for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSAddPools,
	),
	runner.NewE2ETest(
		TEST_DVS_REMOVE_POOLS,
		"pell contract: TestDVSRemovePools",
		[]runner.ArgDefinition{
			{Description: "test pell dvs remove pools for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSRemovePools,
	),
	runner.NewE2ETest(
		TEST_DVS_MODIFY_POOL_PARAMS,
		"pell contract: TestDVSModifyPoolParams",
		[]runner.ArgDefinition{
			{Description: "test pell dvs modify pool params for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSModifyPoolParams,
	),
	runner.NewE2ETest(
		TEST_DVS_UPDATE_OPERATORS,
		"pell contract: TestDVSUpdateOperators",
		[]runner.ArgDefinition{
			{Description: "test pell dvs update operators for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSUpdateOperators,
	),
	runner.NewE2ETest(
		TEST_DVS_UPDATE_OPERATORS_FOR_GROUP,
		"pell contract: TestDVSUpdateOperatorsForGroup",
		[]runner.ArgDefinition{
			{Description: "test pell dvs update operators for group for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSUpdateOperatorsForGroup,
	),
	runner.NewE2ETest(
		TEST_DVS_DEREGISTER_OPERATOR,
		"pell contract: TestDVSDeregisterOperator",
		[]runner.ArgDefinition{
			{Description: "test pell dvs deregister operator for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSDeregisterOperator,
	),
	runner.NewE2ETest(
		TEST_DVS_REGISTER_OPERATOR_WITH_CHURN,
		"pell contract: TestDVSRegisterOperatorWithChurn",
		[]runner.ArgDefinition{
			{Description: "test pell dvs register operator with churn for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSRegisterOperatorWithChurn,
	),
	runner.NewE2ETest(
		TEST_DVS_EJECT_OPERATORS,
		"pell contract: TestDVSEjectOperators",
		[]runner.ArgDefinition{
			{Description: "test pell dvs eject operators for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSEjectOperators,
	),

	// bridge pell
	runner.NewE2ETest(
		TEST_BRIDGE_PELL_INBOUND,
		"pell contract: TestBridgePellInbound",
		[]runner.ArgDefinition{
			{Description: "test pell bridge inbound", DefaultValue: "10000000000000000008"},
		},
		TestBridgePellInbound,
	),
	runner.NewE2ETest(
		TEST_BRIDGE_PELL_OUTBOUND,
		"pell contract: TestBridgePellOutbound",
		[]runner.ArgDefinition{
			{Description: "test pell bridge outbound", DefaultValue: "10000000000000000008"},
		},
		TestBridgePellOutbound,
	),
	runner.NewE2ETest(
		TEST_DVS_SYNC_GROUP_FAILED,
		"pell contract: TestDVSSyncGroupFailed",
		[]runner.ArgDefinition{
			{Description: "test pell sync group data failed for both chain", DefaultValue: "10000000000000000008"},
		},
		TestDVSSyncGroupFailed,
	),

	// LST Token dual staking
	runner.NewE2ETest(
		TEST_LST_SET_VOTING_POWER_RATIO,
		"pell contract: TestLSTSetVotingPowerRatio",
		[]runner.ArgDefinition{
			{Description: "test lst set voting power ratio", DefaultValue: "10000000000000000008"},
		},
		TestLSTSetVotingPowerRatio,
	),
	runner.NewE2ETest(
		TEST_LST_CREATE_REGISTRY_ROUTER,
		"pell contract: TestLSTCreateRegistryRouter",
		[]runner.ArgDefinition{
			{Description: "test lst create registry router", DefaultValue: "10000000000000000008"},
		},
		TestLSTCreateRegistryRouter,
	),
	runner.NewE2ETest(
		TEST_LST_CREATE_GROUP,
		"pell contract: TestLSTCreateGroup",
		[]runner.ArgDefinition{
			{Description: "test lst create group", DefaultValue: "10000000000000000008"},
		},
		TestLSTCreateGroup,
	),
	runner.NewE2ETest(
		TEST_LST_REGISTER_OPERATOR_TO_DELEGATION_MANAGER,
		"pell contract: TestLSTRegisterOperatorToDelegationManager",
		[]runner.ArgDefinition{
			{Description: "test lst register operator to delegation manager", DefaultValue: "10000000000000000008"},
		},
		TestLSTRegisterOperatorToDelegationManager,
	),
	runner.NewE2ETest(
		TEST_LST_REGISTER_OPERATOR,
		"pell contract: TestLSTRegisterOperator",
		[]runner.ArgDefinition{
			{Description: "test lst register operator", DefaultValue: "10000000000000000008"},
		},
		TestLSTRegisterOperator,
	),
	runner.NewE2ETest(
		TEST_LST_DEPOSIT,
		"pell contract: TestLSTDeposit",
		[]runner.ArgDefinition{
			{Description: "test lst deposit", DefaultValue: "10000000000000000008"},
		},
		TestLSTDeposit,
	),
	runner.NewE2ETest(
		TEST_LST_DELEGATE,
		"pell contract: TestLSTDelegate",
		[]runner.ArgDefinition{
			{Description: "test lst delegate", DefaultValue: "10000000000000000008"},
		},
		TestLSTDelegate,
	),
	runner.NewE2ETest(
		TEST_LST_OPERATE_POOLS,
		"pell contract: TestLSTOperatePools",
		[]runner.ArgDefinition{
			{Description: "test lst operate pools", DefaultValue: "10000000000000000008"},
		},
		TestLSTOperatePools,
	),
}
