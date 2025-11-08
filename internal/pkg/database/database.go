
//等后续logger.go封装好zap日志库后，再把本文件中的log替换为zap
package database

import (
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"

	//  导入 Gorm 和 Gorm 的 MySQL 驱动
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	
	//"log"  暂时用标准库 log 替代 zap，防止 panic
	"os"  // 读取环境变量
	"go.uber.org/zap"
)

var DB *gorm.DB

// InitDB 负责初始化数据库连接
func InitDB() {
	dsn := os.Getenv("MYSQL_DSN") 
	if dsn == "" {                
		dsn = "root:your_password@tcp(127.0.0.1:3306)/wish_wall?charset=utf8mb4&parseTime=True&loc=Local"
		
		// 在 logger.go 封装好之前，直接调用 zap.S() 会导致 panic。
		// 先使用标准库 log 来打印警告。
		//log.Println("警告：环境变量 MYSQL_DSN 未设置，使用默认值连接数据库")
		 zap.S().Warn("环境变量MYSQL_DSN未设置,使用默认值连接数据库")
	}
	
	//连接数据库
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		//添加gorm日志配置等
	})
	//检查连接错误
	if err != nil {
		zap.S().Fatalf("错误：数据库连接失败: %v", err)
	}

	zap.S().Info("数据库连接成功！")

	// Gorm 会自动检查 `User` 和 `Wish` 表是否存在
	// 如果不存在，Gorm 会根据你 model/ 目录下的结构体自动创建它们
	err = DB.AutoMigrate(
		&model.User{},
		&model.Wish{},
		&model.Like{},    
		&model.Comment{}, 
		&model.WishTag{}, 
	)

	// 检查迁移错误
	if err != nil {
		zap.S().Fatalf("错误：数据库迁移失败: %v", err)
	}

	zap.S().Info("数据库迁移成功！")
}