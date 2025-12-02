package cmd

import (
	"context"
	"fmt"
	"log"

	"pg-sec-lab/internal/policy"
	"pg-sec-lab/internal/verifier"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
)

var (
	verifyPolicyFile string
	dsn              string
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify policies on test database",
	Long:  `Apply generated policies to a test database and verify RLS and permissions work correctly`,
	RunE:  runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().StringVar(&verifyPolicyFile, "policy", "policy.yaml", "path to policy file")
	verifyCmd.Flags().StringVar(&dsn, "dsn", "", "database connection string (required)")
	verifyCmd.MarkFlagRequired("dsn")
}

func runVerify(cmd *cobra.Command, args []string) error {
	p, err := policy.Load(verifyPolicyFile)
	if err != nil {
		return fmt.Errorf("failed to load policy: %w", err)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close(ctx)

	log.Println("Starting policy verification...")

	if err := verifier.Verify(ctx, p, conn); err != nil {
		log.Printf("❌ Verification failed: %v\n", err)
		return err
	}

	log.Println("✅ All verification checks passed!")
	return nil
}
