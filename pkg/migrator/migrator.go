package migrator

import (
	"database/sql"
	"fmt"

	"github.com/gocasters/rankr/pkg/database"
	migrate "github.com/rubenv/sql-migrate"

	_ "github.com/lib/pq"
)

type Migrator struct {
	dialect    string
	dbConfig   database.Config
	migrations *migrate.FileMigrationSource
}

func New(dbConfig database.Config, path string) Migrator {

	migrations := &migrate.FileMigrationSource{
		Dir: path,
	}
	return Migrator{dbConfig: dbConfig, dialect: "postgres", migrations: migrations}
}

func (m Migrator) Up() error {

	connStr := database.BuildDSN(m.dbConfig)

	db, err := sql.Open(m.dialect, connStr)
	if err != nil {
		return fmt.Errorf("can't open %s db: %w", m.dialect, err)
	}
	defer db.Close()

	n, err := migrate.Exec(db, m.dialect, m.migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("can't apply migrations up: %w", err)
	}

	fmt.Printf("Applied %d migrations!\n", n)
	return nil
}

func (m Migrator) Down() error {

	connStr := database.BuildDSN(m.dbConfig)

	db, err := sql.Open(m.dialect, connStr)
	if err != nil {
		return fmt.Errorf("can't open %s db: %w", m.dialect, err)
	}
	defer db.Close()

	n, err := migrate.Exec(db, m.dialect, m.migrations, migrate.Down)
	if err != nil {
		return fmt.Errorf("can't apply migrations down: %w", err)
	}

	fmt.Printf("Rolled back %d migrations!\n", n)
	return nil
}
