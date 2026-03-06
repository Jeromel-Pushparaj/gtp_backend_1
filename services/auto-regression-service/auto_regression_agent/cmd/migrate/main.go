package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"gitlab.com/tekion/development/toc/poc/opentest/internal/config"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/migrations"
)

var (
	configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create migrator
	migrator, err := migrations.NewMigratorWithConfig(migrations.MigratorConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		Database: cfg.Database.Name,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create migrator: %v\n", err)
		os.Exit(1)
	}
	defer migrator.Close()

	switch command {
	case "up":
		if err := runUp(migrator, args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
			os.Exit(1)
		}

	case "down":
		if err := runDown(migrator, args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Rollback failed: %v\n", err)
			os.Exit(1)
		}

	case "status":
		if err := runStatus(migrator); err != nil {
			fmt.Fprintf(os.Stderr, "Status check failed: %v\n", err)
			os.Exit(1)
		}

	case "reset":
		if err := runReset(migrator); err != nil {
			fmt.Fprintf(os.Stderr, "Reset failed: %v\n", err)
			os.Exit(1)
		}

	case "version":
		if err := runVersion(migrator); err != nil {
			fmt.Fprintf(os.Stderr, "Version check failed: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("OpenTest Database Migration Tool")
	fmt.Println()
	fmt.Println("Usage: opentest-migrate [options] <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up              Run all pending migrations")
	fmt.Println("  up <version>    Run migrations up to specified version")
	fmt.Println("  down <steps>    Rollback the specified number of migrations")
	fmt.Println("  status          Show migration status")
	fmt.Println("  reset           Rollback all and re-run migrations")
	fmt.Println("  version         Show current schema version")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -config string  Path to configuration file (default: configs/config.yaml)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  opentest-migrate up                 # Run all pending migrations")
	fmt.Println("  opentest-migrate up 3               # Run migrations up to version 3")
	fmt.Println("  opentest-migrate down 1             # Rollback last migration")
	fmt.Println("  opentest-migrate status             # Show migration status")
}

func runUp(m *migrations.Migrator, args []string) error {
	if len(args) > 0 {
		version, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid version number: %s", args[0])
		}
		fmt.Printf("Running migrations up to version %d...\n", version)
		return m.UpTo(version)
	}

	fmt.Println("Running all pending migrations...")
	return m.Up()
}

func runDown(m *migrations.Migrator, args []string) error {
	steps := 1
	if len(args) > 0 {
		var err error
		steps, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid steps number: %s", args[0])
		}
	}

	fmt.Printf("Rolling back %d migration(s)...\n", steps)
	return m.Down(steps)
}

func runStatus(m *migrations.Migrator) error {
	statuses, err := m.Status()
	if err != nil {
		return err
	}

	fmt.Println("\nMigration Status")
	fmt.Println("================")
	fmt.Printf("%-10s %-35s %-10s %s\n", "Version", "Name", "Status", "Applied At")
	fmt.Println("------------------------------------------------------------------------")

	for _, s := range statuses {
		status := "Pending"
		appliedAt := "-"
		if s.Applied {
			status = "Applied"
			appliedAt = s.AppliedAt.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("%-10d %-35s %-10s %s\n", s.Version, s.Name, status, appliedAt)
	}

	return nil
}

func runReset(m *migrations.Migrator) error {
	fmt.Println("Resetting database (rollback all and re-run)...")
	return m.Reset()
}

func runVersion(m *migrations.Migrator) error {
	version, err := m.CurrentVersion()
	if err != nil {
		return err
	}

	if version == 0 {
		fmt.Println("No migrations have been applied")
	} else {
		fmt.Printf("Current schema version: %d\n", version)
	}

	return nil
}
