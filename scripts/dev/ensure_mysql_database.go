package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xyun1996/social_backend/pkg/db"
)

func main() {
	mysqlConfig := db.LoadMySQLConfig()
	adminConfig := mysqlConfig
	adminConfig.Database = "mysql"

	sqlDB, err := sql.Open("mysql", adminConfig.DSN())
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		panic(err)
	}

	databaseName := strings.TrimSpace(mysqlConfig.Database)
	if databaseName == "" {
		panic("MYSQL_DATABASE is required")
	}

	if _, err := sqlDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", databaseName)); err != nil {
		panic(err)
	}
}
