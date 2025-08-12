package migrator

import (
	"database/sql"
	"fmt"
	"github.com/gocasters/rankr/pkg/database"
	migrate "github.com/rubenv/sql-migrate"
	"log"

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

func (m Migrator) Up() {

	connStr := database.BuildDSN(m.dbConfig)

	db, err := sql.Open(m.dialect, connStr)
	if err != nil {
		log.Fatalf("can't open postgres db: %v", err)
	}

	n, err := migrate.Exec(db, m.dialect, m.migrations, migrate.Up)
	if err != nil {
		log.Fatalf("can't apply migrations: %v", err)
	}

	fmt.Printf("Applied %d migrations!\n", n)
}

func (m Migrator) Down() {

	connStr := database.BuildDSN(m.dbConfig)

	db, err := sql.Open(m.dialect, connStr)
	if err != nil {
		log.Fatalf("can't open postgres db: %v", err)
	}

	n, err := migrate.Exec(db, m.dialect, m.migrations, migrate.Down)
	if err != nil {
		log.Fatalf("can't apply migrations: %v", err)
	}

	fmt.Printf("Rolled back %d migrations!\n", n)
}
