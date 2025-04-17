package main

import (
	tmcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "pellclientd",
	Short: "PellClient CLI",
}

var rootArgs = rootArguments{}

type rootArguments struct {
	pellCoreHome string
}

func setHomeDir() error {
	var err error
	rootArgs.pellCoreHome, err = RootCmd.Flags().GetString(tmcli.HomeFlag)
	return err
}
