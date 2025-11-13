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
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 注册测试
func TestRegister(t *testing.T) {

	os.Setenv("JWT_SECRET", "my_strong_secret_key!")
	t.Run("注册成功", func(t *testing.T) {
		cleanup(testDB) // 每次测试前清理数据库
		// 昵称为空，应默认使用学号
		reqBody := `{"username":"1234567890","password":"testpassword", "nickname": ""}`
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
		assert.Equal(t, "1234567890", user["nickname"], "昵称为空时应默认为学号")
	})

	t.Run("用户名已存在", func(t *testing.T) {
		cleanup(testDB)
		createUser("1234567890", "existingpass") // 先创建一个用户

		reqBody := `{"username":"1234567890","password":"testpassword"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
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

		assert.Equal(t, http.StatusBadRequest, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_PARAM_INVALID), resp["code"])
		data, ok := resp["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "输入错误，请输入十位学号", data["error"])
	})

	t.Run("注册失败 (昵称违规)", func(t *testing.T) {
		if os.Getenv("SILICONFLOW_API_KEY") == "" {
			t.Skip("SILICONFLOW_API_KEY 环境变量未设置, 跳过 AI 违规测试")
		}
		cleanup(testDB)
		// "我恨这个世界" 在 app_test.go 中被视为违规
		reqBody := `{"username":"9876543210","password":"testpassword", "nickname": "我恨这个世界，我要跳楼了"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/register", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_PARAM_INVALID), resp["code"])
		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, "昵称包含不当内容，请修改", data["error"])
	})

	// --- [新] AI 审核测试 ---
	t.Run("注册失败 (AI服务错误 - 昵称太长)", func(t *testing.T) {
		cleanup(testDB)
		longNickname := string(make([]byte, 1001)) // 1001 字节
		reqBody := gin.H{
			"username": "9876543211",
			"password": "testpassword",
			"nickname": longNickname,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_PARAM_INVALID), resp["code"])
		data, _ := resp["data"].(map[string]interface{})
		// 验证错误是否从 ai_service 传递上来
		assert.Equal(t, "内容长度不能超过1000个字符", data["error"])
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

		// [!! 修正 !!] 登录成功应该返回 200 OK
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

		assert.Equal(t, http.StatusUnauthorized, w.Code)
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

		assert.Equal(t, http.StatusUnauthorized, w.Code)
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

		// [!! 修正 !!] 成功获取应返回 200 OK
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

		// 未授权应返回 401
		assert.Equal(t, http.StatusUnauthorized, w.Code)
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
			"nickname":  newNickname,
			"avatar_id": newAvatarID,
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
		assert.Equal(t, float64(newAvatarID), data["avatar_id"])

		// 验证数据库
		var updatedUser model.User
		testDB.First(&updatedUser, user.ID)
		assert.Equal(t, newNickname, updatedUser.Nickname)
		assert.Equal(t, newAvatarID, *updatedUser.AvatarID)
	})

	// AI 审核测试
	t.Run("更新失败 (昵称违规)", func(t *testing.T) {
		if os.Getenv("SILICONFLOW_API_KEY") == "" {
			t.Skip("SILICONFLOW_API_KEY 环境变量未设置, 跳过 AI 违规测试")
		}
		cleanup(testDB)
		user := createUser("3000000002", "password")
		token := createToken(user.ID)

		reqBody, _ := json.Marshal(gin.H{
			"nickname": "我恨这个世界，我要跳楼了", // 违规昵称
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/user", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_PARAM_INVALID), resp["code"])
		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, "昵称包含不当内容，请修改", data["error"])
	})

	t.Run("更新失败 (AI服务错误 - 内容为空)", func(t *testing.T) {
		cleanup(testDB)
		user := createUser("3000000003", "password")
		token := createToken(user.ID)

		reqBody, _ := json.Marshal(gin.H{
			"nickname": " ", // 空昵称
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/user", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_PARAM_INVALID), resp["code"])
		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, "内容不能为空", data["error"])
	})

	t.Run("未授权 (无Token)", func(t *testing.T) {
		cleanup(testDB)
		reqBody := `{"nickname":"test"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/user", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		// 不携带 token
		testRouter.ServeHTTP(w, req)

		// 未授权应返回 401
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_UNAUTHORIZED), resp["code"])
		assert.Equal(t, apperr.GetMsg(apperr.ERROR_UNAUTHORIZED), resp["message"])
	})
}
