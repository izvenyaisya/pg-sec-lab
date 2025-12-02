package verifier

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"pg-sec-lab/internal/generator"
	"pg-sec-lab/internal/policy"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func Verify(ctx context.Context, p *policy.Policy, conn *pgx.Conn) error {
	rand.Seed(time.Now().UnixNano())
	testSchema := fmt.Sprintf("pgseclabtest%d", rand.Intn(100000))

	log.Printf("Creating test schema: %s\n", testSchema)
	if _, err := conn.Exec(ctx, fmt.Sprintf("CREATE SCHEMA %s", testSchema)); err != nil {
		return fmt.Errorf("failed to create test schema: %w", err)
	}

	defer func() {
		log.Printf("Cleaning up test schema: %s\n", testSchema)
		conn.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", testSchema))
	}()

	if _, err := conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", testSchema)); err != nil {
		return fmt.Errorf("failed to set search_path: %w", err)
	}

	if err := createTestTables(ctx, conn); err != nil {
		return fmt.Errorf("failed to create test tables: %w", err)
	}

	tenantA := uuid.New()
	tenantB := uuid.New()

	if err := insertTestData(ctx, conn, tenantA, tenantB); err != nil {
		return fmt.Errorf("failed to insert test data: %w", err)
	}

	sql, err := generator.GenerateSQL(p)
	if err != nil {
		return fmt.Errorf("failed to generate SQL: %w", err)
	}

	// Replace public schema with test schema
	sql = strings.ReplaceAll(sql, `"public".`, fmt.Sprintf(`"%s".`, testSchema))

	log.Println("Applying generated policies...")
	if _, err := conn.Exec(ctx, sql); err != nil {
		return fmt.Errorf("failed to apply policies: %w", err)
	}

	if err := verifyRLS(ctx, conn, p, testSchema); err != nil {
		return fmt.Errorf("RLS verification failed: %w", err)
	}

	log.Println("✅ All verification checks passed!")

	return nil
}

func createTestTables(ctx context.Context, conn *pgx.Conn) error {
	schema := `
		CREATE TABLE customers (
			id uuid PRIMARY KEY,
			tenant_id uuid NOT NULL,
			email text NOT NULL
		);

		CREATE TABLE orders (
			id uuid PRIMARY KEY,
			tenant_id uuid NOT NULL,
			amount numeric NOT NULL
		);
	`

	if _, err := conn.Exec(ctx, schema); err != nil {
		return err
	}

	return nil
}

func insertTestData(ctx context.Context, conn *pgx.Conn, tenantA, tenantB uuid.UUID) error {
	queries := []string{
		fmt.Sprintf("INSERT INTO customers (id, tenant_id, email) VALUES ('%s', '%s', 'user1@tenantA.com')", uuid.New(), tenantA),
		fmt.Sprintf("INSERT INTO customers (id, tenant_id, email) VALUES ('%s', '%s', 'user2@tenantA.com')", uuid.New(), tenantA),
		fmt.Sprintf("INSERT INTO customers (id, tenant_id, email) VALUES ('%s', '%s', 'user1@tenantB.com')", uuid.New(), tenantB),
		fmt.Sprintf("INSERT INTO orders (id, tenant_id, amount) VALUES ('%s', '%s', 100.00)", uuid.New(), tenantA),
		fmt.Sprintf("INSERT INTO orders (id, tenant_id, amount) VALUES ('%s', '%s', 200.00)", uuid.New(), tenantB),
	}

	for _, q := range queries {
		if _, err := conn.Exec(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

func verifyRLS(ctx context.Context, conn *pgx.Conn, p *policy.Policy, testSchema string) error {
	if !p.Tenants.Enabled {
		log.Println("Tenants not enabled, skipping RLS verification")
		return nil
	}

	log.Println("Verifying RLS is enabled on tables...")

	// Check that RLS is enabled on tables in the test schema
	for tableName := range p.Tables {
		// Replace public schema with test schema
		parts := strings.Split(tableName, ".")
		tblName := parts[len(parts)-1]
		testTableName := fmt.Sprintf("%s.%s", testSchema, tblName)

		var rlsEnabled bool
		query := `SELECT relrowsecurity FROM pg_class WHERE oid = $1::regclass`
		if err := conn.QueryRow(ctx, query, testTableName).Scan(&rlsEnabled); err != nil {
			return fmt.Errorf("failed to check RLS status for %s: %w", testTableName, err)
		}

		if !rlsEnabled {
			return fmt.Errorf("RLS not enabled on table %s", testTableName)
		}
		log.Printf("✅ RLS enabled on %s\n", testTableName)
	}

	log.Println("Verifying RLS policies exist...")

	// Check that policies exist on tables in test schema
	for tableName, tp := range p.Tables {
		if !tp.RLS.Enabled {
			continue
		}

		var policyCount int
		query := `SELECT COUNT(*) FROM pg_policies WHERE schemaname = $1 AND tablename = $2`
		// Extract just table name without schema
		parts := strings.Split(tableName, ".")
		tblName := parts[len(parts)-1]

		if err := conn.QueryRow(ctx, query, testSchema, tblName).Scan(&policyCount); err != nil {
			return fmt.Errorf("failed to check policies for %s: %w", tableName, err)
		}

		if policyCount == 0 {
			return fmt.Errorf("no policies found for table %s in schema %s", tblName, testSchema)
		}
		log.Printf("✅ Found %d policies on %s.%s\n", policyCount, testSchema, tblName)
	}

	log.Println("✅ RLS verification successful - all tables have RLS enabled with policies")

	return nil
}
