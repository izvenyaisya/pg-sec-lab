package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pg-sec-lab",
	Short: "PostgreSQL Security Lab - Policy-as-Code for PostgreSQL",
	Long: `pg-sec-lab is a CLI tool for managing PostgreSQL security through
declarative policies. It supports role management, RLS policies,
data masking, and security configuration analysis.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}
