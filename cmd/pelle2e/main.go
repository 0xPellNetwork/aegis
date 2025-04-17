package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"

	"github.com/0xPellNetwork/aegis/cmd/pelle2e/common"
)

func main() {
	// enable color output
	color.NoColor = false

	// initialize root command
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	common.SetCosmosConfig()
}
