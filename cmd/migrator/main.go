package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var dsn, migrationsPath, migrationsTable string

	flag.StringVar(&dsn, "dsn", "", "PostgreSQL DSN (e.g., postgres://user:password@localhost:5432/dbname?sslmode=disable)")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "schema_migrations", "name of migrations table")
	flag.Parse()

	if dsn == "" {
		panic("dsn is required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	dsnWithTable := fmt.Sprintf("%s&x-migrations-table=%s", dsn, migrationsTable)

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsnWithTable,
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}
}
