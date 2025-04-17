package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/cmd/pelltool/config"
	"github.com/0xPellNetwork/aegis/cmd/pelltool/filterdeposit"
)

var rootCmd = &cobra.Command{
	Use:   "pelltool",
	Short: "utility tool for pell-chain",
}

func init() {
	rootCmd.AddCommand(filterdeposit.NewFilterDepositCmd())
	rootCmd.PersistentFlags().String(config.FlagConfig, "", "custom config file: --config filename.json")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
