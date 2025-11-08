package main

import (
	"log"
	"os"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/database"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/seeder"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)


func main() {
	if err := godotenv.Load(); err != nil {
		// 在 logger 初始化之前用标准库打印警告
		log.Println("警告：未找到 .env 文件，将使用系统环境变量")
	}

	// 初始化日志
	logger.InitLogger()
	zap.S().Info("日志系统初始化成功")

	database.InitDB()
	zap.S().Info("数据库初始化成功,并且执行AutoMigrate成功")

	// 填充初始数据
	if os.Getenv("GIN_MODE") != "release" {
		zap.S().Info("Main: GIN_MODE 非 'release'，开始执行数据填充...")
		seeder.SeedData(database.DB)
	} else {
		zap.S().Info("Main: GIN_MODE 为 'release'，跳过数据填充。")
	}
	zap.S().Info("Main: 开始依赖注入...")

	//后续在这里添加路由和启动服务器的代码

}
