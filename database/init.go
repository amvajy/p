package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// RunMigrations loads and executes SQL migrations in order.
func RunMigrations(db *sql.DB) error {
	files := []string{
		"migrations/001_initial.sql",
	}
	for _, f := range files {
		if err := execSQLFile(db, f); err != nil {
			return fmt.Errorf("执行迁移 %s 失败: %w", f, err)
		}
	}
	return nil
}

func execSQLFile(db *sql.DB, path string) error {
	abs, _ := filepath.Abs(path)
	b, err := ioutil.ReadFile(abs)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(b))
	return err
}
