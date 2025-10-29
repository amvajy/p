package database

import (
	"database/sql"
	"fmt"
	"log"

	"pxe-manager/config"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(cfg *config.Config) (*sql.DB, error) {
	switch cfg.Database.Driver {
	case "sqlite":
		return initSQLite(cfg.Database.SQLitePath)
	case "mysql":
		return sql.Open("mysql", cfg.Database.MySQLDSN)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", cfg.Database.Driver)
	}
}

func initSQLite(dbPath string) (*sql.DB, error) {
	dsn := dbPath + "?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	// 验证 WAL 模式
	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		return nil, err
	}
	log.Printf("SQLite journal mode: %s", journalMode)
	return db, nil
}
