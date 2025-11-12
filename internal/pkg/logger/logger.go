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

	// 根据 GIN_MODE 环境变量判断运行模式：release 
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
	// 初始化包级 SugaredLogger 变量，方便其他包统一调用
	Log = logger.Sugar()
	Log.Info("zap logger 初始化成功")
}

// Log 是包级导出的 SugaredLogger，供项目其他包直接调用
var Log *zap.SugaredLogger

// GetLogger 返回底层 *zap.Logger
func GetLogger() *zap.Logger {
	return zap.L()
}
