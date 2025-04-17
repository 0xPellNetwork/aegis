package main

import (
	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/cmd/pelle2e/multi"
)

var asciiArt = `
             _ _      ___      
            | | |    |__ \     
  _ __   ___| | | ___   ) |___ 
 | '_ \ / _ \ | |/ _ \ / // _ \
 | |_) |  __/ | |  __// /|  __/
 | .__/ \___|_|_|\___|____\___|
 | |                           
 |_|                           
`

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pelle2e",
		Short: asciiArt,
	}
	cmd.AddCommand(
		multi.NewMultiCmd(),
	)

	return cmd
}
