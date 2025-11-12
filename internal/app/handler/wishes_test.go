package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCreateWish 测试创建愿望 (CreatWish.go)
func TestCreateWish(t *testing.T) {
	// 初始化在各子用例中分别进行

	t.Run("创建愿望成功", func(t *testing.T) {
		cleanup(testDB)
		user := createUser("wish_creator_success", "pass")
		token := createToken(user.ID)

		reqBody := gin.H{"content": "我的新愿望", "isPublic": true}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/wishes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])
	})

	t.Run("创建愿望失败 (未授权)", func(t *testing.T) {
		cleanup(testDB)
		reqBody := gin.H{"content": "未授权的愿望"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/wishes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		// 不带 Token
		testRouter.ServeHTTP(w, req)

		// 未授权应返回 401
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("创建愿望失败 (AI 违规)", func(t *testing.T) {
		if os.Getenv("SILICONFLOW_API_KEY") == "" {
			t.Skip("SILICONFLOW_API_KEY 环境变量未设置, 跳过 AI 违规测试")
		}
		cleanup(testDB)
		user := createUser("wish_creator_ai", "pass")
		token := createToken(user.ID)

		reqBody := gin.H{"content": "我恨这个世界，我要跳楼了"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/wishes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_PARAM_INVALID), resp["code"])
		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, "内容未通过审核", data["error"]) //
	})
}

// TestDeleteWish 测试删除愿望 (DeleteWish.go)
func TestDeleteWish(t *testing.T) {
	t.Skip("DELETE /api/wishes/:id 路由当前未启用，跳过此用例")
	cleanup(testDB)
	owner := createUser("wish_owner", "pass")
	ownerToken := createToken(owner.ID)
	otherUser := createUser("other_user", "pass")
	otherToken := createToken(otherUser.ID)

	wish := createWish(owner.ID, "A wish to be deleted")

	t.Run("删除愿望失败 (非作者)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/wishes/"+strconv.Itoa(int(wish.ID)), nil)
		req.Header.Set("Authorization", "Bearer "+otherToken) // 使用路人token
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, "没有权限删除该愿望", resp["message"]) //
	})

	t.Run("删除愿望失败 (愿望不存在)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/wishes/9999", nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_WISH_NOT_FOUND), resp["code"]) //
	})

	t.Run("删除愿望成功 (作者)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/wishes/"+strconv.Itoa(int(wish.ID)), nil)
		req.Header.Set("Authorization", "Bearer "+ownerToken) // 使用 "作者" 的 token
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.SUCCESS), resp["code"]) //
	})
}

// TestGetMyWishes 测试获取个人愿望 (GetMyWish.go)
func TestGetMyWishes(t *testing.T) {
	cleanup(testDB)
	userA := createUser("userA", "pass")
	tokenA := createToken(userA.ID)
	userB := createUser("userB", "pass")

	// 准备数据
	createWish(userA.ID, "Wish A1")
	createWish(userA.ID, "Wish A2")
	createWish(userB.ID, "Wish B1")

	t.Run("获取个人愿望 (隔离性测试)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/wishes/me", nil)
		req.Header.Set("Authorization", "Bearer "+tokenA) // 使用 UserA 的 token
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		data, _ := resp["data"].(map[string]interface{})

		assert.Equal(t, float64(2), data["total"], "User A 应该有 2 条愿望")
		items, _ := data["items"].([]interface{})
		assert.Equal(t, 2, len(items), "返回的 items 数量应为 2")
	})

	t.Run("获取个人愿望 (分页)", func(t *testing.T) {
		cleanup(testDB)
		userC := createUser("userC", "pass")
		tokenC := createToken(userC.ID)
		for i := 0; i < 15; i++ { // 创建15条
			createWish(userC.ID, fmt.Sprintf("Wish %d", i))
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/wishes/me?page=2&pageSize=10", nil) // 请求第2页
		req.Header.Set("Authorization", "Bearer "+tokenC)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		data, _ := resp["data"].(map[string]interface{})

		assert.Equal(t, float64(15), data["total"])
		assert.Equal(t, float64(2), data["page"])
		items, _ := data["items"].([]interface{})
		assert.Equal(t, 5, len(items), "第2页应该只剩5条数据") // 15 = 10 + 5
	})
}

// TestGetPublicWishes 测试获取公共愿望 (GetPublicWish.go)
func TestGetPublicWishes(t *testing.T) {
	cleanup(testDB)
	user := createUser("public_user", "pass")
	token := createToken(user.ID)

	wish1 := createWish(user.ID, "Public Wish 1")
	// (需要手动模拟点赞)
	testDB.Exec("INSERT INTO likes (user_id, wish_id) VALUES (?, ?)", user.ID, wish1.ID)

	t.Run("获取公共愿望 (未登录)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/wishes/public", nil)
		// 不带 Token
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		data, _ := resp["data"].(map[string]interface{})
		items, _ := data["items"].([]interface{})
		firstWish, _ := items[0].(map[string]interface{})

		// 未登录时, liked 字段必须为 false
		assert.Equal(t, false, firstWish["liked"]) //
	})

	t.Run("获取公共愿望 (已登录, 检查 liked 状态)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/wishes/public", nil)
		req.Header.Set("Authorization", "Bearer "+token) // 带 Token
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		data, _ := resp["data"].(map[string]interface{})
		items, _ := data["items"].([]interface{})
		firstWish, _ := items[0].(map[string]interface{})

		// 已登录且点赞过, liked 字段必须为 true
		assert.Equal(t, true, firstWish["liked"])
	})
}
