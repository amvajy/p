package database

import (
	"database/sql"
	"fmt"
	"log"

	"pxe-manager/config"

	_ "modernc.org/sqlite"
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
	// modernc.org/sqlite 使用驱动名 "sqlite"，无需 CGO
	// 显式设置关键 PRAGMA，确保 WAL、外键与忙等待
	dsn := dbPath
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// 设置 PRAGMA
	_, _ = db.Exec("PRAGMA foreign_keys = ON")
	_, _ = db.Exec("PRAGMA busy_timeout = 5000")
	_, _ = db.Exec("PRAGMA journal_mode = WAL")
	// 验证 WAL 模式
	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		return nil, err
	}
	log.Printf("SQLite journal mode: %s", journalMode)
	return db, nil
}