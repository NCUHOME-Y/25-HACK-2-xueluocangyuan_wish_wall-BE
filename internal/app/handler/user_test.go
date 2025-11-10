package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// 清理测试数据库中的用户数据
func parseResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		// 如果 JSON 解析失败，测试应该立即失败
		logger.Log.Fatalf("无法解析 JSON 响应: %v", w.Body.String())
	}
	return resp
}

// 在测试数据库中创建一个用户
func createUser(username, password string) *model.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := model.User{
		Username: username,
		Password: string(hashedPassword),
		Nickname: username, // 默认昵称
		Role:     "user",
	}
	// testDB 是在 main_test.go 中定义的全局变量
	if err := testDB.Create(&user).Error; err != nil {
		logger.Log.Fatalf("创建测试用户失败: %v", err)
	}
	return &user
}

//为指定ID生成一个Token

func createToken(userID uint) string {
	// 确保设置了JWT_SECRET，以便token生成器能工作
	os.Setenv("JWT_SECRET", "my_strong_secret_key!")
	token, err := util.GenerateToken(userID)
	if err != nil {
		logger.Log.Fatalf("生成测试Token失败: %v", err)
	}
	return token
}

// 注册测试
func TestRegister(t *testing.T) {

	os.Setenv("JWT_SECRET", "my_strong_secret_key!")
	t.Run("注册成功", func(t *testing.T) {
		cleanup(testDB) // 每次测试前清理数据库
		reqBody := `{"username":"1234567890","password":"testpassword"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req) // testRouter 是在 main_test.go 中定义的

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)

		// 验证你修改后的标准响应
		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])
		assert.Equal(t, apperr.GetMsg(apperr.SUCCESS), resp["message"])

		// 验证 data 结构
		data, ok := resp["data"].(map[string]interface{})
		assert.True(t, ok, "data 字段应该是 map[string]interface{}")
		assert.NotEmpty(t, data["token"], "token 不能为空")
		user, ok := data["user"].(map[string]interface{})
		assert.True(t, ok, "user 字段应该是 map[string]interface{}")
		assert.Equal(t, "1234567890", user["username"])
	})

	t.Run("用户名已存在", func(t *testing.T) {
		cleanup(testDB)
		createUser("1234567890", "existingpass") // 先创建一个用户

		reqBody := `{"username":"1234567890","password":"testpassword"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_PARAM_INVALID), resp["code"])
		// 检查 data.error
		data, ok := resp["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "该学号已被注册", data["error"])
	})

	t.Run("学号不为10位", func(t *testing.T) {
		cleanup(testDB)
		reqBody := `{"username":"123","password":"testpassword"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_PARAM_INVALID), resp["code"])
		data, ok := resp["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "输入错误，请输入十位学号", data["error"])
	})
}

// TestLogin 测试登录功能
func TestLogin(t *testing.T) {
	os.Setenv("JWT_SECRET", "my_strong_secret_key!")

	t.Run("登录成功", func(t *testing.T) {
		cleanup(testDB)
		user := createUser("1000000001", "doublegood_password") // 创建用户

		reqBody := `{"username":"1000000001","password":"doublegood_password"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)

		// 验证标准响应
		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])
		assert.Equal(t, apperr.GetMsg(apperr.SUCCESS), resp["message"])

		data, _ := resp["data"].(map[string]interface{})
		assert.NotEmpty(t, data["token"])
		respUser, _ := data["user"].(map[string]interface{})
		assert.Equal(t, user.Username, respUser["username"])
	})

	t.Run("用户不存在", func(t *testing.T) {
		cleanup(testDB)
		reqBody := `{"username":"1000000002","password":"doublebad_password"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_LOGIN_FAILED), resp["code"])
		assert.Equal(t, apperr.GetMsg(apperr.ERROR_LOGIN_FAILED), resp["message"]) // 验证标准错误消息
	})

	t.Run("密码错误", func(t *testing.T) {
		cleanup(testDB)
		createUser("1000000001", "doublegood_password") // 创建用户

		reqBody := `{"username":"1000000001","password":"wrong_password"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_LOGIN_FAILED), resp["code"])
		assert.Equal(t, apperr.GetMsg(apperr.ERROR_LOGIN_FAILED), resp["message"])
	})
}

// TestGetUserMe 测试获取用户信息 (受保护路由)
func TestGetUserMe(t *testing.T) {
	t.Run("获取成功", func(t *testing.T) {
		cleanup(testDB)
		user := createUser("2000000001", "password")
		token := createToken(user.ID) // 生成 token

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/user/me", nil)
		req.Header.Set("Authorization", "Bearer "+token) // 携带 token
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)

		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])
		assert.Equal(t, apperr.GetMsg(apperr.SUCCESS), resp["message"])

		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, user.Username, data["username"])
		assert.Equal(t, user.Nickname, data["nickname"])
	})

	t.Run("未授权 (无Token)", func(t *testing.T) {
		cleanup(testDB)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/user/me", nil)
		// 不携带 token
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_UNAUTHORIZED), resp["code"])
		assert.Equal(t, apperr.GetMsg(apperr.ERROR_UNAUTHORIZED), resp["message"])
	})
}

// TestUpdateUser 测试更新用户信息 (受保护路由)
func TestUpdateUser(t *testing.T) {
	t.Run("更新成功", func(t *testing.T) {
		cleanup(testDB)
		user := createUser("3000000001", "password")
		token := createToken(user.ID)

		newNickname := "我的新昵称"
		var newAvatarID uint = 10
		// gin.H 在测试中构建 json 更方便
		reqBody, _ := json.Marshal(gin.H{
			"nickname": newNickname,
			"avatarID": newAvatarID,
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/user", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)

		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])
		assert.Equal(t, apperr.GetMsg(apperr.SUCCESS), resp["message"])

		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, newNickname, data["nickname"])
		// JSON 解析数字默认为 float64，需要转换
		assert.Equal(t, float64(newAvatarID), data["avatarID"])

		// 验证数据库
		var updatedUser model.User
		testDB.First(&updatedUser, user.ID)
		assert.Equal(t, newNickname, updatedUser.Nickname)
		assert.Equal(t, newAvatarID, *updatedUser.AvatarID)
	})

	t.Run("未授权 (无Token)", func(t *testing.T) {
		cleanup(testDB)
		reqBody := `{"nickname":"test"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/user", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		// 不携带 token
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_UNAUTHORIZED), resp["code"])
		assert.Equal(t, apperr.GetMsg(apperr.ERROR_UNAUTHORIZED), resp["message"])
	})
}

//但是在测试中从resp取回data或user时，需要使用实际类型
