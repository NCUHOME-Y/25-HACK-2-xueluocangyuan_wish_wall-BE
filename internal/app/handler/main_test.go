package handler_test

import (
	"os"
	"testing"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/router"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	testDB     *gorm.DB    // 全局测试数据库连接
	testRouter *gin.Engine // 全局测试路由
)

// TestMain 设置测试环境
func TestMain(m *testing.M) {
	//设置Gin为测试模式（减少不必要的日志）
	gin.SetMode(gin.TestMode)
	// 初始化日志系统
	logger.InitLogger()

	// internal/app/handler 目录，.env 在三层目录之上
	if err := godotenv.Load("../../../.env"); err != nil {
		logger.Log.Fatalf("加载 .env 文件失败 : %v", err)
	}

	//连接测试数据库
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		logger.Log.Fatalf("MYSQL_TEST_DSN 环境变量未设置，请检查 .env 文件")
	}
	var err error
	testDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Fatalf("测试数据库连接失败: %v", err)
	}

	//自动迁移
	err = testDB.AutoMigrate(
		&model.User{},
		&model.Wish{},
		&model.Like{},
		&model.Comment{},
		&model.WishTag{},
	)
	if err != nil {
		logger.Log.Fatalf("测试数据库迁移失败: %v", err)
	}

	//设置测试路由
	// 因为 .env 已加载, os.Getenv("ACTIVE_ACTIVITY") 现在可以读到 "v1" 了
	testRouter = router.SetupRouter(testDB)

	//运行测试
	exitCode := m.Run()

	//退出
	cleanup(testDB)

	os.Exit(exitCode)
}
func cleanup(db *gorm.DB) {
	//删除所有表数据,从外键开始删
	db.Exec("DELETE FROM wish_tags")
	db.Exec("DELETE FROM comments")
	db.Exec("DELETE FROM likes")
	db.Exec("DELETE FROM wishes")
	db.Exec("DELETE FROM users")
}
