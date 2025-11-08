package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
)

// InitLogger 初始化全局 zap 日志记录器
func InitLogger() {
	var (
		logger *zap.Logger
		err    error
	)

	// 根据 GIN_MODE 环境变量判断运行模式：release -> 生产；否则视为开发模式
	if os.Getenv("GIN_MODE") == "release" {
		// 生产环境，使用更严格的生产配置（JSON 格式）
		logger, err = zap.NewProduction()
	} else {
		// 开发环境，使用便于阅读的开发配置
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatalf("无法初始化日志记录器: %v", err)
	}

	zap.ReplaceGlobals(logger)
	zap.S().Info("zap logger 初始化成功")
}
