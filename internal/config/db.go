package config

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	DatabaseDriverPostgres = "postgres"
	DatabaseDriverMySQL    = "mysql"
)

func DatabaseURL() string {
	return GetEnv("DATABASE_URL", "")
}

func ConnectDB(dsn string) *gorm.DB {
	return ConnectDBWithDriver(dsn, "")
}

func ConnectDBWithDriver(dsn, driver string) *gorm.DB {
	if dsn == "" {
		log.Fatal("❌ DATABASE_URL is not set")
	}

	driver = NormalizeDatabaseDriver(driver, dsn)

	var db *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = openGormDB(driver, dsn)

		if err == nil {
			log.Printf("✅ %s database connected", driver)
			break
		}

		log.Println("⏳ Waiting for database...", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("❌ DB connection failed after retries: %v", err)
	}

	// ✅ Connection pool config
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(envInt("DB_MAX_OPEN_CONNS", 5))
	sqlDB.SetMaxIdleConns(envInt("DB_MAX_IDLE_CONNS", 2))
	sqlDB.SetConnMaxLifetime(time.Duration(envInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute)

	return db
}

func openGormDB(driver, dsn string) (*gorm.DB, error) {
	switch driver {
	case DatabaseDriverPostgres:
		sqlDB, err := sql.Open("pgx", dsn)
		if err != nil {
			return nil, err
		}
		return gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	case DatabaseDriverMySQL:
		mysqlDSN, err := MySQLGormDSN(dsn)
		if err != nil {
			return nil, err
		}
		return gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database driver %q", driver)
	}
}

func NormalizeDatabaseDriver(driver, dsn string) string {
	driver = strings.ToLower(strings.TrimSpace(driver))
	switch driver {
	case "postgresql":
		return DatabaseDriverPostgres
	case "mariadb":
		return DatabaseDriverMySQL
	case DatabaseDriverPostgres, DatabaseDriverMySQL:
		return driver
	}

	if parsed, err := url.Parse(dsn); err == nil {
		switch strings.ToLower(parsed.Scheme) {
		case "postgres", "postgresql":
			return DatabaseDriverPostgres
		case "mysql", "mariadb":
			return DatabaseDriverMySQL
		}
	}

	return DatabaseDriverPostgres
}

func SupportedDatabaseDriver(driver string) bool {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case DatabaseDriverPostgres, "postgresql", DatabaseDriverMySQL, "mariadb":
		return true
	default:
		return false
	}
}

func MySQLGormDSN(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "mysql" && parsed.Scheme != "mariadb") {
		return raw, nil
	}

	user := parsed.User.Username()
	if password, ok := parsed.User.Password(); ok {
		user += ":" + password
	}

	network := "tcp"
	host := parsed.Host
	if strings.HasPrefix(host, "unix(") {
		network = "unix"
		host = strings.TrimPrefix(strings.TrimSuffix(host, ")"), "unix(")
	} else if strings.HasPrefix(host, "tcp(") {
		host = strings.TrimPrefix(strings.TrimSuffix(host, ")"), "tcp(")
	}

	dbName := strings.TrimPrefix(parsed.Path, "/")
	query := parsed.Query()
	if query.Get("parseTime") == "" {
		query.Set("parseTime", "true")
	}
	if query.Get("charset") == "" {
		query.Set("charset", "utf8mb4")
	}

	return fmt.Sprintf("%s@%s(%s)/%s?%s", user, network, host, dbName, query.Encode()), nil
}

func MigrationDatabaseURL(raw, driver string) string {
	if NormalizeDatabaseDriver(driver, raw) != DatabaseDriverMySQL {
		return raw
	}

	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "mysql" && parsed.Scheme != "mariadb") {
		return raw
	}

	query := parsed.Query()
	if query.Get("multiStatements") == "" {
		query.Set("multiStatements", "true")
	}
	parsed.RawQuery = query.Encode()
	if parsed.Host != "" && !strings.HasPrefix(parsed.Host, "tcp(") && !strings.HasPrefix(parsed.Host, "unix(") {
		parsed.Host = "tcp(" + parsed.Host + ")"
	}
	return parsed.String()
}
