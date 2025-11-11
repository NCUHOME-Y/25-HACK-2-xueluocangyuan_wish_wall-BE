package handler_test

import (
	"bytes"
	
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	
	"github.com/stretchr/testify/assert"
)


// TestGetAppState 测试获取应用状态
func TestGetAppState(t *testing.T) {
	// 清空数据库 (虽然这个 handler 不用 DB)
	cleanup(testDB)

	t.Run("当 ACTIVE_ACTIVITY 被设置为 v1", func(t *testing.T) {

		t.Setenv("ACTIVE_ACTIVITY", "v1")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/app-state", nil)
		testRouter.ServeHTTP(w, req) // testRouter 来自 main_test.go

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(err.SUCCESS), resp["code"])
		assert.Equal(t, err.GetMsg(err.SUCCESS), resp["message"])
		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, "v1", data["activeActivity"])
	})

	t.Run("当 ACTIVE_ACTIVITY 未设置", func(t *testing.T) {
		// 确保环境变量为空
		t.Setenv("ACTIVE_ACTIVITY", "")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/app-state", nil)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(err.ERROR_SERVER_ERROR), resp["code"])
		assert.Equal(t, "服务器配置错误，请联系管理员", resp["message"])
	})
}

// TestTestAI 测试 AI 审查接口
func TestTestAI(t *testing.T) {
	// 这个测试会真的请求 Silicon Flow API
	// 必须在运行测试前设置 `SILICONFLOW_API_KEY`
	if os.Getenv("SILICONFLOW_API_KEY") == "" {
		t.Skip("SILICONFLOW_API_KEY 环境变量未设置, 跳过 TestTestAI")
	}

	cleanup(testDB)

	t.Run("测试安全内容", func(t *testing.T) {

		reqBody := `{"content": "我爱这个世界"}`

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/test-ai", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, false, resp["is_violating"])
		assert.Equal(t, "我爱这个世界", resp["input_content"])
	})

	t.Run("测试违规内容", func(t *testing.T) {

		reqBody := `{"content": "我恨这个世界，我要跳楼了"}`

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/test-ai", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, true, resp["is_violating"])
	})
}
