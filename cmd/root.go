package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "s3magic",
	Short: "S3Magic is a CLI tool for managing AWS S3 buckets and objects.",
	Long: `S3Magic is a powerful and easy-to-use command-line interface (CLI)
tool designed to simplify your interactions with AWS S3.
Manage your buckets and objects efficiently directly from your terminal.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no command is provided
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
