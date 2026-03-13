package db

import "testing"

func TestMySQLDSN(t *testing.T) {
	t.Parallel()

	cfg := MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "1234",
		Database: "social_backend",
		Params: map[string]string{
			"parseTime": "true",
			"loc":       "UTC",
		},
	}

	dsn := cfg.DSN()
	want := "root:1234@tcp(localhost:3306)/social_backend?loc=UTC&parseTime=true"
	if dsn != want {
		t.Fatalf("unexpected dsn: got %q want %q", dsn, want)
	}
}
