// Package migrations provides database migration support for OpenTest
package migrations

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	_ "github.com/lib/pq"
)

// Migration represents a single database migration
type Migration struct {
	Version     int
	Name        string
	Description string
	Up          func(db *sql.DB) error
	Down        func(db *sql.DB) error
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version     int       `json:"version"`
	Name        string    `json:"name"`
	AppliedAt   time.Time `json:"applied_at,omitempty"`
	Applied     bool      `json:"applied"`
	Description string    `json:"description,omitempty"`
}

// Migrator handles database migrations
type Migrator struct {
	db         *sql.DB
	migrations []Migration
	tableName  string
}

// MigratorConfig configures the migrator
type MigratorConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

// NewMigrator creates a new Migrator with the given database connection
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: GetAllMigrations(),
		tableName:  "schema_migrations",
	}
}

// NewMigratorWithConfig creates a new Migrator with the given configuration
func NewMigratorWithConfig(cfg MigratorConfig) (*Migrator, error) {
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.Password, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return NewMigrator(db), nil
}

// Close closes the database connection
func (m *Migrator) Close() error {
	return m.db.Close()
}

// ensureMigrationTable creates the schema_migrations table if it doesn't exist
func (m *Migrator) ensureMigrationTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, m.tableName)

	_, err := m.db.Exec(query)
	return err
}

// getAppliedVersions returns a map of applied migration versions
func (m *Migrator) getAppliedVersions() (map[int]time.Time, error) {
	query := fmt.Sprintf("SELECT version, applied_at FROM %s", m.tableName)
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]time.Time)
	for rows.Next() {
		var version int
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, err
		}
		applied[version] = appliedAt
	}

	return applied, rows.Err()
}

// recordMigration records a migration as applied
func (m *Migrator) recordMigration(version int, name string) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (version, name, applied_at) VALUES ($1, $2, $3)",
		m.tableName,
	)
	_, err := m.db.Exec(query, version, name, time.Now())
	return err
}

// removeMigration removes a migration record
func (m *Migrator) removeMigration(version int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE version = $1", m.tableName)
	_, err := m.db.Exec(query, version)
	return err
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	if err := m.ensureMigrationTable(); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	applied, err := m.getAppliedVersions()
	if err != nil {
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	// Sort migrations by version
	sorted := make([]Migration, len(m.migrations))
	copy(sorted, m.migrations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version < sorted[j].Version
	})

	for _, migration := range sorted {
		if _, ok := applied[migration.Version]; ok {
			continue // Already applied
		}

		fmt.Printf("Applying migration %d: %s...\n", migration.Version, migration.Name)
		if err := migration.Up(m.db); err != nil {
			return fmt.Errorf("failed to apply migration %d (%s): %w",
				migration.Version, migration.Name, err)
		}

		if err := m.recordMigration(migration.Version, migration.Name); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}
		fmt.Printf("Applied migration %d: %s\n", migration.Version, migration.Name)
	}

	return nil
}

// UpTo runs migrations up to and including the specified version
func (m *Migrator) UpTo(targetVersion int) error {
	if err := m.ensureMigrationTable(); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	applied, err := m.getAppliedVersions()
	if err != nil {
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	sorted := make([]Migration, len(m.migrations))
	copy(sorted, m.migrations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version < sorted[j].Version
	})

	for _, migration := range sorted {
		if migration.Version > targetVersion {
			break
		}
		if _, ok := applied[migration.Version]; ok {
			continue
		}

		fmt.Printf("Applying migration %d: %s...\n", migration.Version, migration.Name)
		if err := migration.Up(m.db); err != nil {
			return fmt.Errorf("failed to apply migration %d (%s): %w",
				migration.Version, migration.Name, err)
		}

		if err := m.recordMigration(migration.Version, migration.Name); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}
		fmt.Printf("Applied migration %d: %s\n", migration.Version, migration.Name)
	}

	return nil
}

// Down rolls back the specified number of migrations
func (m *Migrator) Down(steps int) error {
	if err := m.ensureMigrationTable(); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	applied, err := m.getAppliedVersions()
	if err != nil {
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	// Get applied versions in descending order
	var appliedVersions []int
	for v := range applied {
		appliedVersions = append(appliedVersions, v)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(appliedVersions)))

	// Build migration lookup
	migrationMap := make(map[int]Migration)
	for _, mig := range m.migrations {
		migrationMap[mig.Version] = mig
	}

	rolledBack := 0
	for _, version := range appliedVersions {
		if rolledBack >= steps {
			break
		}

		migration, ok := migrationMap[version]
		if !ok {
			return fmt.Errorf("migration %d not found in registered migrations", version)
		}

		fmt.Printf("Rolling back migration %d: %s...\n", migration.Version, migration.Name)
		if err := migration.Down(m.db); err != nil {
			return fmt.Errorf("failed to rollback migration %d (%s): %w",
				migration.Version, migration.Name, err)
		}

		if err := m.removeMigration(version); err != nil {
			return fmt.Errorf("failed to remove migration record %d: %w", version, err)
		}
		fmt.Printf("Rolled back migration %d: %s\n", migration.Version, migration.Name)
		rolledBack++
	}

	return nil
}

// Status returns the status of all migrations
func (m *Migrator) Status() ([]MigrationStatus, error) {
	if err := m.ensureMigrationTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure migration table: %w", err)
	}

	applied, err := m.getAppliedVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied versions: %w", err)
	}

	var statuses []MigrationStatus
	for _, migration := range m.migrations {
		status := MigrationStatus{
			Version:     migration.Version,
			Name:        migration.Name,
			Description: migration.Description,
			Applied:     false,
		}

		if appliedAt, ok := applied[migration.Version]; ok {
			status.Applied = true
			status.AppliedAt = appliedAt
		}

		statuses = append(statuses, status)
	}

	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Version < statuses[j].Version
	})

	return statuses, nil
}

// Reset runs all down migrations and then all up migrations
func (m *Migrator) Reset() error {
	statuses, err := m.Status()
	if err != nil {
		return err
	}

	// Count applied migrations
	appliedCount := 0
	for _, s := range statuses {
		if s.Applied {
			appliedCount++
		}
	}

	// Roll back all
	if appliedCount > 0 {
		if err := m.Down(appliedCount); err != nil {
			return fmt.Errorf("failed to rollback all migrations: %w", err)
		}
	}

	// Apply all
	return m.Up()
}

// CurrentVersion returns the current schema version
func (m *Migrator) CurrentVersion() (int, error) {
	if err := m.ensureMigrationTable(); err != nil {
		return 0, err
	}

	query := fmt.Sprintf("SELECT COALESCE(MAX(version), 0) FROM %s", m.tableName)
	var version int
	err := m.db.QueryRow(query).Scan(&version)
	return version, err
}
