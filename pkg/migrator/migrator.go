package migrator

import (
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func RunMigrations(fs embed.FS, databaseURL string) error {
	d, err := iofs.New(fs, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %w", err)
	}

	migratorURL := databaseURL
	if len(databaseURL) >= 11 && databaseURL[:11] == "postgres://" {
		migratorURL = "pgx5://" + databaseURL[11:]
	} else if len(databaseURL) >= 13 && databaseURL[:13] == "postgresql://" {
		migratorURL = "pgx5://" + databaseURL[13:]
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, migratorURL)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Database schema is up to date (no migrations to apply)")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("Database migrations applied successfully")
	return nil
}
