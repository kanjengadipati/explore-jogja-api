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

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
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
}
