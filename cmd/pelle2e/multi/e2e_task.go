package multi

import (
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pell-chain/pellcore/e2e/e2etests"
	"github.com/pell-chain/pellcore/e2e/runner"
)

// define all pell e2e task
var regularPellTest = []string{
	e2etests.TEST_UPGRADE_SYSTEM_CONTRACT,
	e2etests.TEST_DEPOSIT_FOR_INBOUND,
	e2etests.TEST_REGISTER_OPERATOR_FOR_OUTBOUND,
	e2etests.TEST_DELEGATE,
	e2etests.TEST_WITHDRAWAL,
	// TODO: wait for the contract update
	//e2etests.TEST_RECHARGE_PELL_TOKEN,
	//e2etests.TEST_RECHARGE_GAS,
	e2etests.TEST_DVS_CREATE_REGISTRY_ROUTER,
	e2etests.TEST_DVS_CREATE_GROUP_ON_PELL,
	// register operator before sync group
	e2etests.TEST_DVS_REGISTER_OPERATOR_BEFORE_SYNCGROUP,
	e2etests.TEST_DVS_ADD_SUPPORTED_CHAIN,
	e2etests.TEST_DVS_SYNC_GROUP,
	// deregister operator after sync group
	e2etests.TEST_DVS_DEREGISTER_OPERATOR,
	e2etests.TEST_DVS_CREATE_GROUP,
	e2etests.TEST_DVS_SET_OPERATOR_SET_PARAMS,
	e2etests.TEST_DVS_SET_GROUP_EJECTION_PARAMS,
	e2etests.TEST_DVS_SET_EJECTION_COOLDOWN,
	e2etests.TEST_DVS_REGISTER_OPERATOR,
	e2etests.TEST_DVS_ADD_POOLS,
	e2etests.TEST_DVS_REMOVE_POOLS,
	e2etests.TEST_DVS_MODIFY_POOL_PARAMS,
	e2etests.TEST_DVS_UPDATE_OPERATORS,
	e2etests.TEST_DVS_UPDATE_OPERATORS_FOR_GROUP,
	e2etests.TEST_UNDELEGATE,
	e2etests.TEST_DVS_DEREGISTER_OPERATOR,
	e2etests.TEST_DVS_REGISTER_OPERATOR_WITH_CHURN,
	e2etests.TEST_DVS_EJECT_OPERATORS,
	e2etests.TEST_BRIDGE_PELL_INBOUND,
	e2etests.TEST_BRIDGE_PELL_OUTBOUND,
	e2etests.TEST_DVS_SYNC_GROUP_FAILED,
	// LST Token dual staking
	e2etests.TEST_LST_SET_VOTING_POWER_RATIO,
	e2etests.TEST_LST_CREATE_REGISTRY_ROUTER,
	e2etests.TEST_LST_CREATE_GROUP,
	e2etests.TEST_LST_REGISTER_OPERATOR_TO_DELEGATION_MANAGER,
	e2etests.TEST_LST_REGISTER_OPERATOR,
	e2etests.TEST_LST_DEPOSIT,
	e2etests.TEST_LST_DELEGATE,
	e2etests.TEST_LST_OPERATE_POOLS,
}

// run e2e all task
func e2eRun(runner *runner.Runner) {
	runner.Logger.Print(runningLogo())
	testStartTime := time.Now()
	// run tests
	var eg errgroup.Group

	eg.Go(pellTestRoutine(runner, regularPellTest...))

	if err := eg.Wait(); err != nil {
		runner.Logger.Print("❌ %v", err)
		runner.Logger.Print("❌ e2e tests failed after %s", time.Since(testStartTime).String())
		os.Exit(1)
	}

	runner.Logger.Print("✅ e2e tests completed in %s", time.Since(testStartTime).String())
}

func runningLogo() string {
	return `      ____            _             _                            _                   
  ___|___ \ ___   ___| |_ __ _ _ __| |_   _ __ _   _ _ __  _ __ (_)_ __   __ _       
 / _ \ __) / _ \ / __| __/ _` + "`" + ` | '__| __| | '__| | | | '_ \| '_ \| | '_ \ / _` + "`" + ` |      
|  __// __|  __/ \__ | || (_| | |  | |_  | |  | |_| | | | | | | | | | | | (_| |_ _ _ 
 \___|_____\___| |___/\__\__,_|_|   \__| |_|   \__,_|_| |_|_| |_|_|_| |_|\__, (_(_(_)
                                                                         |___/       `
}
