package checker

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type InstanceInfo struct {
	Version  string            `json:"version"`
	Settings map[string]string `json:"settings"`
}

type RoleInfo struct {
	Name      string   `json:"name"`
	Login     bool     `json:"login"`
	Superuser bool     `json:"superuser"`
	BypassRLS bool     `json:"bypassrls"`
	Grants    []string `json:"grants"`
}

type TableInfo struct {
	Schema     string `json:"schema"`
	Name       string `json:"name"`
	RLSEnabled bool   `json:"rls_enabled"`
}

type Finding struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type Report struct {
	Instance InstanceInfo `json:"instance"`
	Roles    []RoleInfo   `json:"roles"`
	Tables   []TableInfo  `json:"tables"`
	Findings []Finding    `json:"findings"`
}

func Analyze(ctx context.Context, conn *pgx.Conn) (*Report, error) {
	report := &Report{
		Roles:    []RoleInfo{},
		Tables:   []TableInfo{},
		Findings: []Finding{},
	}

	var err error

	report.Instance, err = getInstanceInfo(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance info: %w", err)
	}

	report.Roles, err = getRoles(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	report.Tables, err = getTables(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	report.Findings = generateFindings(report)

	return report, nil
}

func getInstanceInfo(ctx context.Context, conn *pgx.Conn) (InstanceInfo, error) {
	var version string
	err := conn.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		return InstanceInfo{}, err
	}

	settings := make(map[string]string)
	configParams := []string{
		"ssl",
		"password_encryption",
		"log_connections",
		"log_disconnections",
		"log_statement",
	}

	for _, param := range configParams {
		var value string
		err := conn.QueryRow(ctx,
			"SELECT setting FROM pg_settings WHERE name = $1", param).Scan(&value)
		if err == nil {
			settings[param] = value
		}
	}

	return InstanceInfo{
		Version:  version,
		Settings: settings,
	}, nil
}

func getRoles(ctx context.Context, conn *pgx.Conn) ([]RoleInfo, error) {
	query := `
		SELECT 
			rolname,
			rolcanlogin,
			rolsuper,
			rolbypassrls
		FROM pg_roles
		WHERE rolname NOT LIKE 'pg_%'
		ORDER BY rolname
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []RoleInfo
	for rows.Next() {
		var role RoleInfo
		if err := rows.Scan(&role.Name, &role.Login, &role.Superuser, &role.BypassRLS); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Close rows before making new queries
	rows.Close()

	// Now fetch grants for each role
	for i := range roles {
		grants, err := getRoleGrants(ctx, conn, roles[i].Name)
		if err != nil {
			return nil, err
		}
		roles[i].Grants = grants
	}

	return roles, nil
}

func getRoleGrants(ctx context.Context, conn *pgx.Conn, roleName string) ([]string, error) {
	query := `
		SELECT 
			table_schema || '.' || table_name AS object,
			privilege_type
		FROM information_schema.role_table_grants
		WHERE grantee = $1
		ORDER BY table_schema, table_name, privilege_type
	`

	rows, err := conn.Query(ctx, query, roleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []string
	for rows.Next() {
		var object, privilege string
		if err := rows.Scan(&object, &privilege); err != nil {
			return nil, err
		}
		grants = append(grants, fmt.Sprintf("%s ON %s", privilege, object))
	}

	return grants, rows.Err()
}

func getTables(ctx context.Context, conn *pgx.Conn) ([]TableInfo, error) {
	query := `
		SELECT 
			n.nspname AS schema,
			c.relname AS name,
			c.relrowsecurity AS rls_enabled
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relkind = 'r'
		  AND n.nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		ORDER BY n.nspname, c.relname
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.Schema, &table.Name, &table.RLSEnabled); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

func generateFindings(report *Report) []Finding {
	var findings []Finding

	for _, table := range report.Tables {
		if !table.RLSEnabled {
			findings = append(findings, Finding{
				Severity: "warning",
				Code:     "NO_RLS",
				Message:  fmt.Sprintf("Table %s.%s has no RLS enabled", table.Schema, table.Name),
			})
		}
	}

	if ssl, ok := report.Instance.Settings["ssl"]; ok && strings.ToLower(ssl) == "off" {
		findings = append(findings, Finding{
			Severity: "high",
			Code:     "SSL_DISABLED",
			Message:  "SSL is disabled on this PostgreSQL instance",
		})
	}

	for _, role := range report.Roles {
		if role.Superuser && role.Login {
			findings = append(findings, Finding{
				Severity: "critical",
				Code:     "SUPERUSER_LOGIN",
				Message:  fmt.Sprintf("Role %s is a superuser with login capability", role.Name),
			})
		}

		if role.BypassRLS {
			findings = append(findings, Finding{
				Severity: "warning",
				Code:     "BYPASS_RLS",
				Message:  fmt.Sprintf("Role %s can bypass RLS policies", role.Name),
			})
		}
	}

	return findings
}
