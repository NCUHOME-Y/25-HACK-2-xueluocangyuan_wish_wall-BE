package database/*
import(
	"github.com/NCUHOME-YC/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"//导入model包
	"gorm.io/driver/mysql"//导入gorm和gorm mysql驱动
	"gorm.io/gorm"
	
	"zap"
	"os"//读取环境变量
)
var DB *gorm.DB

func InitDB(){
	dsn:=os.Getenv("MYSQL_DSN")//从环境变量中获取数据库连接字符串
	if dsn ==""{//如果环境变量未设置，使用默认值
		dsn = "root:your_password@tcp(127.0.0.1:3306)/wish_wall?charset=utf8mb4&parseTime=True&loc=Local"
		zap.S().Warn("环境变量MYSQL_DSN未设置，使用默认值连接数据库")
	}
	//连接数据库
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		//添加gorm日志配置等
	})
}
	
	*/