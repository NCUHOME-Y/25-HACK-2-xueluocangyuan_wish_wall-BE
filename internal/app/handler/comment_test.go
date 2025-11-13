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

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCreateComment
func TestCreateComment(t *testing.T) {
	cleanup(testDB) //
	user := createUserWithRole("commenter", "pass", "user")
	token := createToken(user.ID)
	wish := createWish(user.ID, "test wish")

	t.Run("创建评论成功", func(t *testing.T) {
		reqBody := gin.H{"wishId": wish.ID, "content": "My new comment"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()

		req, _ := http.NewRequest("POST", "/api/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req) // testRouter 来自 main_test.go

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)

		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])
		// 使用统一的消息源
		assert.Equal(t, apperr.GetMsg(apperr.SUCCESS), resp["message"]) //

		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, "My new comment", data["content"])

		// 响应已扁平化，不再返回嵌套 user 对象
		assert.Equal(t, float64(user.ID), data["userId"])
		assert.Equal(t, user.Nickname, data["userNickname"])
		// 可选：校验其他新字段
		if v, ok := data["isOwn"].(bool); ok {
			assert.True(t, v)
		}
		if lc, ok := data["likeCount"].(float64); ok {
			assert.Equal(t, float64(0), lc)
		}

		// 检查数据库
		var updatedWish model.Wish
		testDB.First(&updatedWish, wish.ID)
		assert.Equal(t, 1, updatedWish.CommentCount, "心愿的评论数应该增加到 1")
	})

	t.Run("愿望不存在", func(t *testing.T) {
		reqBody := gin.H{"wishId": 9999, "content": "..."} //
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_PARAM_INVALID), resp["code"])
		// 测试硬编码的错误消息
		assert.Equal(t, apperr.GetMsg(apperr.ERROR_PARAM_INVALID), resp["message"]) //
	})

	t.Run("评论失败 (评论他人的私有愿望)", func(t *testing.T) {
		// 1. 创建一个 "otherUser" 和他的 "privateWish"
		otherUser := createUser("privateWishOwner", "pass")
		privateWish := model.Wish{UserID: otherUser.ID, Content: "private", IsPublic: false}
		testDB.Create(&privateWish)
		// 由于模型设置了 default:true，显式将 is_public 更新为 false 以避免默认值覆盖
		testDB.Model(&privateWish).Update("is_public", false)

		// 2. 尝试用 "user" (token) 去评论 "privateWish"
		reqBody := gin.H{"wishId": privateWish.ID, "content": "I see you"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token) // 使用 'commenter' (user) 的 token
		testRouter.ServeHTTP(w, req)

		// 3. 断言 401 Unauthorized 和特定错误码 13
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_FORBIDDEN_COMMENT), resp["code"])
	})

	t.Run("创建评论失败 (AI 判定违规)", func(t *testing.T) {
		// 确保 AI Key 存在
		if os.Getenv("SILICONFLOW_API_KEY") == "" {
			t.Skip("SILICONFLOW_API_KEY 环境变量未设置, 跳过 AI 违规测试")
		}

		// "我恨这个世界" 在 app_test.go 中被视为违规
		reqBody := gin.H{"wishId": wish.ID, "content": "我恨这个世界，我要跳楼了"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code) // 应该是 400
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_PARAM_INVALID), resp["code"])
		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, "内容未通过审核", data["error"]) // 验证 handler 返回的特定错误
	})

	t.Run("创建评论失败 (AI 服务错误 - 内容为空)", func(t *testing.T) {
		reqBody := gin.H{"wishId": wish.ID, "content": " "} // 空内容
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/comments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code) // 应该是 400
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.ERROR_PARAM_INVALID), resp["code"])
		data, _ := resp["data"].(map[string]interface{})
		// 验证错误是否从 ai_service 传递上来
		assert.Equal(t, "内容不能为空", data["error"])
	})
}

// TestDeleteComment
func TestDeleteComment(t *testing.T) {
	cleanup(testDB)
	author := createUserWithRole("author", "pass", "user")
	authorToken := createToken(author.ID)
	otherUser := createUserWithRole("otherUser", "pass", "user")
	otherToken := createToken(otherUser.ID)
	admin := createUserWithRole("adminUser", "pass", "admin")
	adminToken := createToken(admin.ID)

	wish := createWish(author.ID, "test wish")
	comment := createComment(author.ID, wish.ID, "comment by author")

	t.Run("非作者删除失败", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, _ := http.NewRequest("DELETE", "/api/comments/"+strconv.Itoa(int(comment.ID)), nil)
		req.Header.Set("Authorization", "Bearer "+otherToken)
		testRouter.ServeHTTP(w, req)

		resp := parseResponse(t, w)
		// 现在未授权删除返回 ERROR_FORBIDDEN_DELETE
		assert.Equal(t, float64(apperr.ERROR_FORBIDDEN_DELETE), resp["code"])
		assert.Equal(t, apperr.GetMsg(apperr.ERROR_FORBIDDEN_DELETE), resp["message"])
	})

	t.Run("管理员删除成功", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/comments/"+strconv.Itoa(int(comment.ID)), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		testRouter.ServeHTTP(w, req)

		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])
		// 使用统一的消息源
		assert.Equal(t, apperr.GetMsg(apperr.SUCCESS), resp["message"])

		// 检查数据库
		var updatedWish model.Wish
		testDB.First(&updatedWish, wish.ID)
		assert.Equal(t, 0, updatedWish.CommentCount, "评论数应减为 0")
	})

	t.Run("作者删除成功", func(t *testing.T) {
		// 重新创建评论
		comment2 := createComment(author.ID, wish.ID, "comment 2")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/api/comments/"+strconv.Itoa(int(comment2.ID)), nil)
		req.Header.Set("Authorization", "Bearer "+authorToken)
		testRouter.ServeHTTP(w, req)

		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])
		assert.Equal(t, apperr.GetMsg(apperr.SUCCESS), resp["message"])

		// 检查数据库
		var updatedWish model.Wish
		testDB.First(&updatedWish, wish.ID)
		assert.Equal(t, 0, updatedWish.CommentCount, "评论数应再次减为 0")
	})

	t.Run("心愿主人删除评论成功 (非评论作者)", func(t *testing.T) {
		// author 是心愿主人 (使用 authorToken)
		// otherUser 是评论作者
		wishByAuthor := createWish(author.ID, "wish by author")
		commentByOther := createComment(otherUser.ID, wishByAuthor.ID, "comment by other")

		// 确保评论计数为 1
		var wishCheck model.Wish
		testDB.First(&wishCheck, wishByAuthor.ID)
		assert.Equal(t, 1, wishCheck.CommentCount)

		w := httptest.NewRecorder()
		// 使用 心愿主人(author) 的 token 去删除 otherUser 的评论
		req, _ := http.NewRequest("DELETE", "/api/comments/"+strconv.Itoa(int(commentByOther.ID)), nil)
		req.Header.Set("Authorization", "Bearer "+authorToken)
		testRouter.ServeHTTP(w, req)

		// 断言成功
		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])

		// 检查数据库
		testDB.First(&wishCheck, wishByAuthor.ID)
		assert.Equal(t, 0, wishCheck.CommentCount, "评论数应减为 0")
	})
}

// TestListCommentsByWish
func TestListCommentsByWish(t *testing.T) {
	cleanup(testDB)
	user := createUserWithRole("lister", "pass", "user")
	wish := createWish(user.ID, "wish for list")

	// 创建 25 条评论
	for i := 0; i < 25; i++ {
		createComment(user.ID, wish.ID, fmt.Sprintf("Comment %d", i+1))
	}

	t.Run("获取第一页", func(t *testing.T) {
		w := httptest.NewRecorder()

		req, _ := http.NewRequest("GET", "/api/wishes/"+strconv.Itoa(int(wish.ID))+"/comments?page=1&pageSize=20", nil)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])

		assert.Equal(t, apperr.GetMsg(apperr.SUCCESS), resp["message"]) //

		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(25), data["total"])
		assert.Equal(t, float64(1), data["page"])

		items, _ := data["items"].([]interface{})
		assert.Equal(t, 20, len(items), "第一页应返回 20 条")

		// 检查 Preload("User") 是否生效
		firstComment, _ := items[0].(map[string]interface{})
		commentUser, _ := firstComment["user"].(map[string]interface{})
		assert.Equal(t, user.Nickname, commentUser["nickname"])
	})
}
