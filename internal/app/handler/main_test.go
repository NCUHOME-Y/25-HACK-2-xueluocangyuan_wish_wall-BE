package handler_test

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/util"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/router"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
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

	// internal/app/handler 目录，
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

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		logger.Log.Fatalf("无法解析 JSON 响应: %v", w.Body.String())
	}
	return resp
}

func createUser(username, password string) *model.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := model.User{
		Username: username,
		Password: string(hashedPassword),
		Nickname: username,
		Role:     "user",
	}
	if err := testDB.Create(&user).Error; err != nil {
		logger.Log.Fatalf("创建测试用户失败: %v", err)
	}
	return &user
}

func createToken(userID uint) string {
	os.Setenv("JWT_SECRET", "my_strong_secret_key!")
	token, err := util.GenerateToken(userID)
	if err != nil {
		logger.Log.Fatalf("生成测试Token失败: %v", err)
	}
	return token
}

func createUserWithRole(username, password, role string) *model.User {
	u := createUser(username, password)
	if role != "user" {
		if err := testDB.Model(&u).Update("role", role).Error; err != nil {
			logger.Log.Fatalf("设置用户角色失败: %v", err)
		}
		u.Role = role
	}
	return u
}

func createWish(userID uint, content string) *model.Wish {
	wish := model.Wish{
		UserID:   userID,
		Content:  content,
		IsPublic: true, // 默认为 public
	}
	if err := testDB.Create(&wish).Error; err != nil {
		logger.Log.Fatalf("创建测试心愿失败: %v", err)
	}
	return &wish
}

func createComment(userID, wishID uint, content string) *model.Comment {
	comment := model.Comment{
		UserID:  userID,
		WishID:  wishID,
		Content: content,
	}
	testDB.Create(&comment)
	testDB.Model(&model.Wish{}).Where("id = ?", wishID).UpdateColumn("comment_count", gorm.Expr("comment_count + 1"))
	return &comment
}
