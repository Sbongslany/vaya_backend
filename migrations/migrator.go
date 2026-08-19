package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var migrationFS embed.FS

func Up(db *sql.DB) error {
	goose.SetBaseFS(migrationFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, "sql")
}

func Down(db *sql.DB) error {
	goose.SetBaseFS(migrationFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Down(db, "sql")
}
