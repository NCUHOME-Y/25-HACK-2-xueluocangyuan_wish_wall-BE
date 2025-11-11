// 等后续logger.go封装好zap日志库后，再把本文件中的log替换为zap
package database

import (
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"

	//  导入 Gorm 和 Gorm 的 MySQL 驱动
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	//"log"  暂时用标准库 log 替代 zap，防止 panic
	"go.uber.org/zap"
	"os" // 读取环境变量
	"time"
)

var DB *gorm.DB

// InitDB 负责初始化数据库连接
func InitDB() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:your_password@tcp(127.0.0.1:3306)/wish_wall?charset=utf8mb4&parseTime=True&loc=Local"
		zap.S().Warn("环境变量MYSQL_DSN未设置,使用默认值连接数据库")
	}
	var err error
	const maxRetries = 10
	const retryDelay = 3 * time.Second

	// 2. 循环重试连接
	for i := 0; i < maxRetries; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			//添加gorm日志配置等
		})

		if err == nil {
			// 3. 连接成功
			zap.S().Info("数据库连接成功！")

			// 运行 Gorm 自动迁移
			err = DB.AutoMigrate(
				&model.User{},
				&model.Wish{},
				&model.Like{},
				&model.Comment{},
				&model.WishTag{},
			)
			if err != nil {
				// 迁移失败，这是个严重问题，直接 panic
				zap.S().Fatalf("错误：数据库迁移失败: %v", err)
			}
			zap.S().Info("数据库迁移成功！")

			// 成功，退出函数
			return
		}

		// 4. 连接失败，记录日志并等待
		zap.S().Warnw("数据库连接失败，正在重试...",
			"attempt", i+1,
			"maxAttempts", maxRetries,
			"error", err,
		)
		time.Sleep(retryDelay)
	}

	// 5. 达到最大重试次数后仍然失败
	zap.S().Fatalf("错误：数据库连接失败（已达最大重试次数）: %v", err)
}
