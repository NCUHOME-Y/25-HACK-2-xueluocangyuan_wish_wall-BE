package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/stretchr/testify/assert"
)

func TestLikeWish(t *testing.T) {
	cleanup(testDB)
	user := createUser("like_user", "pass")
	token := createToken(user.ID)
	wish := createWish(user.ID, "A wish to be liked")

	t.Run("点赞成功 (第一次)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/wishes/"+strconv.Itoa(int(wish.ID))+"/like", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])

		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, true, data["liked"], "Liked 状态应为 true")
		assert.Equal(t, float64(1), data["likeCount"], "LikeCount 应为 1")
	})

	t.Run("取消点赞成功 (第二次)", func(t *testing.T) {
		//  (基于上一个测试) 已经是点赞状态
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/wishes/"+strconv.Itoa(int(wish.ID))+"/like", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])

		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, false, data["liked"], "Liked 状态应为 false")
		assert.Equal(t, float64(0), data["likeCount"], "LikeCount 应为 0")
	})

	t.Run("点赞失败 (愿望不存在)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/wishes/9999/like", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_WISH_NOT_FOUND), resp["code"])
	})

	t.Run("点赞失败 (未授权)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/wishes/"+strconv.Itoa(int(wish.ID))+"/like", nil)
		// 不带 Token
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_UNAUTHORIZED), resp["code"])
	})
}