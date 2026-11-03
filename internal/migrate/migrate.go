// package migrate applies the embedded goose migrations.
package migrate

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	// goose needs a database/sql driver; the app itself uses pgxpool directly.
	_ "github.com/jackc/pgx/v5/stdlib"

	migrations "github.com/nielwyn/murmur/sql"
)

func Up(dbURL string) error {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("opening migration connection: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("setting goose dialect: %w", err)
	}

	return goose.Up(db, "schema")
}
