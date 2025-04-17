package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/pkg/constant"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "version description from git describe --tags",
	RunE:  Version,
}

func Version(_ *cobra.Command, _ []string) error {
	fmt.Printf(constant.Version)
	return nil
}
