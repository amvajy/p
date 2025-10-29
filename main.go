package main

import (
	"log"
	"net/http"

	"pxe-manager/api"
	"pxe-manager/config"
	"pxe-manager/database"
)

func main() {
	cfg := config.LoadConfig()

	// 初始化数据库
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()

	// 运行迁移
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 设置路由
	router := api.SetupRouter(db, cfg)

	log.Printf("PXE管理系统启动在 %s", cfg.ServerAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
}