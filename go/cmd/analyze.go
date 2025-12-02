package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"pg-sec-lab/internal/configcheck"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
)

var (
	analyzeDsn     string
	analyzeOutFile string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze PostgreSQL configuration",
	Long:  `Analyze PostgreSQL instance for security configuration and generate JSON report`,
	RunE:  runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().StringVar(&analyzeDsn, "dsn", "", "database connection string (required)")
	analyzeCmd.Flags().StringVar(&analyzeOutFile, "out", "report.json", "output JSON file")
	analyzeCmd.MarkFlagRequired("dsn")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, analyzeDsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	log.Println("Analyzing PostgreSQL configuration...")

	report, err := configcheck.Analyze(ctx, conn)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if analyzeOutFile == "" {
		fmt.Println(string(jsonData))
	} else {
		if err := os.WriteFile(analyzeOutFile, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		log.Printf("Analysis report saved to: %s\n", analyzeOutFile)
	}

	return nil
}
