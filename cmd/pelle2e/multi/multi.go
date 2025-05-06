package multi

import (
	"context"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/e2e/config"
	"github.com/0xPellNetwork/aegis/e2e/runner"
	"github.com/0xPellNetwork/aegis/e2e/utils"
)

const (
	flagWaitForHeight    = "wait-for"
	flagConfigOut        = "config-out"
	flagVerbose          = "verbose"
	flagTestAdmin        = "test-admin"
	flagSetupOnly        = "setup-only"
	flagSkipSetup        = "skip-setup"
	flagSkipBitcoinSetup = "skip-bitcoin-setup"
	flagSkipHeaderProof  = "skip-header-proof"
	flagInit             = "init"
	flagConfigFile       = "config"
)

var (
	TestTimeout = 15 * time.Minute
)

// NewMultiCmd returns the local command
// which runs the E2E tests locally on the machine with localnet for each blockchain
func NewMultiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multi",
		Short: "Run multi E2E tests",
		Run:   multiE2ETest,
	}
	cmd.Flags().Int64(flagWaitForHeight, 0, "block height for tests to begin, ex. --wait-for 100")
	cmd.Flags().Bool(flagVerbose, false, "set to true to enable verbose logging")
	cmd.Flags().Bool(flagSetupOnly, false, "set to true to only setup the networks")
	cmd.Flags().String(flagConfigOut, "", "config file to write the deployed contracts from the setup")
	cmd.Flags().Bool(flagSkipSetup, false, "set to true to skip setup")
	cmd.Flags().Bool(flagSkipHeaderProof, false, "set to true to skip header proof tests")
	cmd.Flags().String(flagInit, "", "export config file path")
	cmd.Flags().String(flagConfigFile, "", "config file path")

	return cmd
}

func multiE2ETest(cmd *cobra.Command, _ []string) {
	// fetch flags
	waitForHeight, err := cmd.Flags().GetInt64(flagWaitForHeight)
	if err != nil {
		panic(err)
	}
	verbose, err := cmd.Flags().GetBool(flagVerbose)
	if err != nil {
		panic(err)
	}
	setupOnly, err := cmd.Flags().GetBool(flagSetupOnly)
	if err != nil {
		panic(err)
	}
	skipSetup, err := cmd.Flags().GetBool(flagSkipSetup)
	if err != nil {
		panic(err)
	}

	skipHeaderProof, err := cmd.Flags().GetBool(flagSkipHeaderProof)
	if err != nil {
		panic(err)
	}

	initPath, err := cmd.Flags().GetString(flagInit)
	if err != nil {
		panic(err)
	}

	configPath, err := cmd.Flags().GetString(flagConfigFile)
	if err != nil {
		panic(err)
	}

	logger := runner.NewLogger(verbose, color.FgWhite, "setup")

	// initialize tests config
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}

	if initPath != "" {
		if err := conf.Export(initPath); err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	// initialize context
	ctx, cancel := context.WithCancel(context.Background())

	// wait for a specific height on PellChain
	if waitForHeight != 0 {
		utils.WaitForBlockHeight(ctx, waitForHeight, conf.Rpcs.PellCoreRpc, logger)
	}

	// initialize runner with config
	e2eRunner := runner.NewFromConfig(conf)

	e2eRunner.CtxCancel = cancel

	if !skipHeaderProof {
		if err := e2eRunner.TxServer.EnableVerificationFlags(utils.FungibleAdminName); err != nil {
			panic(err)
		}
	}

	// wait for keygen to be completed
	// if setup is skipped, we assume that the keygen is already completed
	if !skipSetup {
		WaitKeygenHeight(ctx, e2eRunner.PellClients.XmsgClient, logger)
	}

	// if setup only, quit
	if setupOnly {
		logger.Print("✅ the localnet has been setup")
		os.Exit(0)
	}

	// run tests
	e2eRun(e2eRunner)

	os.Exit(0)
}
