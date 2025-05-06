package multi

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/0xPellNetwork/aegis/e2e/e2etests"
	"github.com/0xPellNetwork/aegis/e2e/runner"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// ethereumTestRoutine runs Ethereum related e2e tests
func pellTestRoutine(
	runner *runner.Runner,
	testNames ...string,
) func() error {
	return func() (err error) {
		// return an error on panic
		// TODO: remove and instead return errors in the tests
		defer func() {
			if r := recover(); r != nil {
				// print stack trace
				stack := make([]byte, 4096)
				n := runtime.Stack(stack, false)
				err = fmt.Errorf("ethereum panic: %v, stack trace %s", r, stack[:n])
			}
		}()
		if len(testNames) == 0 {
			return nil
		}

		runner.Logger.Print("🏃 starting Pell tests")
		startTime := time.Now()

		// run pell test
		// Note: due to the extensive block generation in Ethereum localnet, block header test is run first
		// to make it faster to catch up with the latest block header
		testsToRun, err := runner.GetE2ETestsToRunByName(
			e2etests.AllE2ETests,
			testNames...,
		)
		if err != nil {
			return fmt.Errorf("ethereum tests failed: %v", err)
		}

		if err := runner.RunPellTests(testsToRun); err != nil {
			panic(err)
		}

		runner.Logger.Print("🍾 Ethereum tests completed in %s", time.Since(startTime).String())

		return err
	}
}

// waitKeygenHeight waits for keygen height
func WaitKeygenHeight(
	ctx context.Context,
	xmsgClient xmsgtypes.QueryClient,
	logger *runner.Logger,
) {
	logger.Print("⏳ wait height %v for keygen to be completed", 15)

	for {
		time.Sleep(2 * time.Second)
		response, err := xmsgClient.LastPellHeight(ctx, &xmsgtypes.QueryLastPellHeightRequest{})
		if err != nil {
			logger.Error("xmsgClient.LastPellHeight error: %s", err)
			continue
		}
		if response.Height >= 15 {
			break
		}
		logger.Info("Last PellHeight: %d", response.Height)
	}
}
