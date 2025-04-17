package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pell-chain/pellcore/pkg/crypto/bls"
)

func NewBLSCmd() *cobra.Command {
	bls := &cobra.Command{
		Use:   "bls",
		Short: "BLS commands",
	}

	bls.AddCommand(GenerateBLSKey())
	bls.AddCommand(ShowBLSKey())

	return bls
}

func GenerateBLSKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-bls-key [path]",
		Short: "Generate a BLS key",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			path := args[0]

			keys, err := bls.GenRandomBlsKeys()
			if err != nil {
				return err
			}

			if checkIfKeyExists(path) {
				return fmt.Errorf("key file already exists at %s", path)
			}

			if err := keys.SaveToFile(path, ""); err != nil {
				return fmt.Errorf("failed to save key to file: %w", err)
			}

			fmt.Println(keys)
			return nil
		},
	}
	return cmd
}

func ShowBLSKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-bls-key [path]",
		Short: "Show the BLS key",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			path := args[0]

			keyContent, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read key file: %w", err)
			}

			fmt.Println("Key content: " + string(keyContent))
			return nil
		},
	}
	return cmd
}

func checkIfKeyExists(fileLoc string) bool {
	_, err := os.Stat(fileLoc)
	return !os.IsNotExist(err)
}
