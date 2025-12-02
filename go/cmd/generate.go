package cmd

import (
	"fmt"
	"os"

	"pg-sec-lab/internal/generator"
	"pg-sec-lab/internal/policy"

	"github.com/spf13/cobra"
)

var (
	policyFile string
	outFile    string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate SQL from policy file",
	Long:  `Generate SQL scripts for roles, RLS policies, and data masking based on policy.yaml`,
	RunE:  runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVar(&policyFile, "policy", "policy.yaml", "path to policy file")
	generateCmd.Flags().StringVar(&outFile, "out", "", "output SQL file (default: stdout)")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	p, err := policy.Load(policyFile)
	if err != nil {
		return fmt.Errorf("failed to load policy: %w", err)
	}

	sql, err := generator.GenerateSQL(p)
	if err != nil {
		return fmt.Errorf("failed to generate SQL: %w", err)
	}

	if outFile == "" {
		fmt.Println(sql)
	} else {
		if err := os.WriteFile(outFile, []byte(sql), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("SQL generated successfully: %s\n", outFile)
	}

	return nil
}
