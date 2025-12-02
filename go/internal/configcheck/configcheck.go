package configcheck

import (
	"context"

	"pg-sec-lab/pkg/checker"

	"github.com/jackc/pgx/v5"
)

// Re-export types from pkg/checker for backward compatibility
type InstanceInfo = checker.InstanceInfo
type RoleInfo = checker.RoleInfo
type TableInfo = checker.TableInfo
type Finding = checker.Finding
type Report = checker.Report

// Analyze delegates to the public checker package
func Analyze(ctx context.Context, conn *pgx.Conn) (*Report, error) {
	return checker.Analyze(ctx, conn)
}
