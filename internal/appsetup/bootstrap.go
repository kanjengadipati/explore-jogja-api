package appsetup

import (
	"log"

	"pleco-api/internal/config"
	"pleco-api/internal/seeds"
	migrationFiles "pleco-api/migrations"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"
)

func RunStartupTasks(cfg config.AppConfig, db *gorm.DB) {
	if cfg.AutoRunMigrations {
		if err := RunMigrationsForDriver(cfg.DatabaseURL, cfg.DatabaseDriver); err != nil {
			log.Fatalf("❌ startup migrations failed: %v", err)
		}
		log.Println("✅ startup migrations completed")
	}

	if cfg.AutoRunSeeds {
		RunSeeds(db, cfg)
		log.Println("✅ startup seeds completed")
	}
}

func RunMigrations(dbURL string) error {
	return RunMigrationsForDriver(dbURL, "")
}

func RunMigrationsForDriver(dbURL, driver string) error {
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL is not set")
	}

	driver = config.NormalizeDatabaseDriver(driver, dbURL)
	migrationsPath := "."
	if driver == config.DatabaseDriverMySQL {
		migrationsPath = "mysql"
	}

	sourceDriver, err := iofs.New(migrationFiles.Files, migrationsPath)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, config.MigrationDatabaseURL(dbURL, driver))
	if err != nil {
		return err
	}

	// If the database is in a dirty state, force it back to the previous
	// clean version so the failed migration can be re-applied.
	if dbErr := m.Up(); dbErr != nil {
		if dbErr.Error() != "no change" {
			// Check for dirty state error
			if isDirtyError(dbErr) {
				version, _, verErr := m.Version()
				if verErr == nil && version > 0 {
					log.Printf("⚠️  dirty migration version %d detected — forcing back to %d and retrying", version, version-1)
					if forceErr := m.Force(int(version) - 1); forceErr != nil {
						return forceErr
					}
					// Retry the migration
					if retryErr := m.Up(); retryErr != nil && retryErr != migrate.ErrNoChange {
						return retryErr
					}
					return nil
				}
			}
			return dbErr
		}
	}

	return nil
}

// isDirtyError checks whether the error is a migrate dirty-state error.
func isDirtyError(err error) bool {
	if err == nil {
		return false
	}
	return len(err.Error()) > 5 && err.Error()[:5] == "Dirty"
}

func RunSeeds(db *gorm.DB, cfg config.AppConfig) {
	if db == nil {
		log.Fatal("❌ DB is not initialized before seeding")
	}

	seeds.SeedRoles(db)
	seeds.SeedPermissions(db)
	seeds.SeedRolePermissions(db)
	seeds.SeedAdmin(db, cfg)
	seeds.SeedDestinations(db)
	seeds.SeedPartners(db)
	seeds.SeedModules(db)
	seeds.SeedEvents(db)
}
