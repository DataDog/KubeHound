package storedb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	_ "modernc.org/sqlite"
)

var _ Provider = (*SQLiteProvider)(nil)

// SQLiteProvider implements the storedb.Provider interface using SQLite.
type SQLiteProvider struct {
	db *sql.DB
}

// NewSQLiteProvider creates a new SQLite provider instance.
func NewSQLiteProvider(ctx context.Context, cfg *config.KubehoundConfig) (*SQLiteProvider, error) {
	path := cfg.SQLite.Path
	if path == "" {
		path = config.DefaultSQLitePath
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database %s: %w", path, err)
	}

	// Set connection pool to 1 for SQLite (serialized writes)
	db.SetMaxOpenConns(1)

	// Run PRAGMAs
	for _, line := range strings.Split(schemaPragmas, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, line); err != nil {
			db.Close()
			return nil, fmt.Errorf("sqlite pragma %q: %w", line, err)
		}
	}

	return &SQLiteProvider{db: db}, nil
}

func (p *SQLiteProvider) Name() string {
	return "sqlite"
}

func (p *SQLiteProvider) HealthCheck(ctx context.Context) (bool, error) {
	if err := p.db.PingContext(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// Prepare creates all tables and indices.
func (p *SQLiteProvider) Prepare(ctx context.Context) error {
	l := log.Logger(ctx)

	l.Info("Creating SQLite schema")
	if _, err := p.db.ExecContext(ctx, schemaDDL); err != nil {
		return fmt.Errorf("sqlite schema creation: %w", err)
	}

	l.Info("Creating SQLite indices")
	if _, err := p.db.ExecContext(ctx, schemaIndices); err != nil {
		return fmt.Errorf("sqlite index creation: %w", err)
	}

	return nil
}

// Clean deletes data for a specific run/cluster combination.
func (p *SQLiteProvider) Clean(ctx context.Context, runID string, clusterName string) error {
	tables := collections.GetCollections()
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE run_id = ? AND cluster_name = ?", table)
		if _, err := p.db.ExecContext(ctx, query, runID, clusterName); err != nil {
			return fmt.Errorf("sqlite clean table %s: %w", table, err)
		}
	}

	return nil
}

// Reader returns the underlying *sql.DB handle for queries.
func (p *SQLiteProvider) Reader() any {
	return p.db
}

// BulkWriter creates a synchronous writer that implements AsyncWriter.
func (p *SQLiteProvider) BulkWriter(ctx context.Context, collection collections.Collection, opts ...WriterOption) (AsyncWriter, error) {
	return NewSQLiteWriter(p.db, collection), nil
}

// Close closes the database connection.
func (p *SQLiteProvider) Close(ctx context.Context) error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
