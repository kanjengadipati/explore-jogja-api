package config

import "testing"

func TestNormalizeDatabaseDriver(t *testing.T) {
	tests := []struct {
		name   string
		driver string
		dsn    string
		want   string
	}{
		{
			name:   "explicit postgres alias",
			driver: "postgresql",
			want:   DatabaseDriverPostgres,
		},
		{
			name:   "explicit mysql",
			driver: DatabaseDriverMySQL,
			want:   DatabaseDriverMySQL,
		},
		{
			name: "infer postgres URL",
			dsn:  "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
			want: DatabaseDriverPostgres,
		},
		{
			name: "infer mysql URL",
			dsn:  "mysql://root:secret@localhost:3306/auth_db?parseTime=true",
			want: DatabaseDriverMySQL,
		},
		{
			name: "default postgres",
			want: DatabaseDriverPostgres,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeDatabaseDriver(tt.driver, tt.dsn); got != tt.want {
				t.Fatalf("NormalizeDatabaseDriver() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMySQLGormDSNConvertsURL(t *testing.T) {
	got, err := MySQLGormDSN("mysql://root:secret@localhost:3306/auth_db?multiStatements=true")
	if err != nil {
		t.Fatalf("MySQLGormDSN() unexpected error: %v", err)
	}

	want := "root:secret@tcp(localhost:3306)/auth_db?charset=utf8mb4&multiStatements=true&parseTime=true"
	if got != want {
		t.Fatalf("MySQLGormDSN() = %q, want %q", got, want)
	}
}

func TestMigrationDatabaseURLAddsMySQLMultiStatements(t *testing.T) {
	got := MigrationDatabaseURL("mysql://root:secret@localhost:3306/auth_db?parseTime=true", DatabaseDriverMySQL)
	want := "mysql://root:secret@tcp(localhost:3306)/auth_db?multiStatements=true&parseTime=true"
	if got != want {
		t.Fatalf("MigrationDatabaseURL() = %q, want %q", got, want)
	}
}
