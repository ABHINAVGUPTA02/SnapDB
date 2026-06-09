package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "snapdb",
	Short: "SnapDB - A professional-grade database backup & restore tool",
	Long:  `SnapDB is a high-performance CLI tool for database backup, restore, compression, scheduling, and cloud integration.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to SnapDB! Run `snapdb --help` to explore commands.")
	},
}

func SetArguments(cmdArgs []string) {
	rootCmd.SetArgs(cmdArgs)
}

// Execute runs the root command
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
