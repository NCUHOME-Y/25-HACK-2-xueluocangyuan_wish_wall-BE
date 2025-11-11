package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// parseResponse 在 user_test.go 中定义，避免重复声明

// createUser 在 user_test.go 中定义，避免重复声明

// createUserWithRole 在测试数据库中创建一个指定角色的用户
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

// createWish 在测试数据库中创建一个心愿
func createWish(userID uint, content string) *model.Wish {
	wish := model.Wish{
		UserID:  userID,
		Content: content,
	}
	if err := testDB.Create(&wish).Error; err != nil {
		logger.Log.Fatalf("创建测试心愿失败: %v", err)
	}
	return &wish
}

// createComment 在测试数据库中创建一个评论（并手动更新 wish 计数）
func createComment(userID, wishID uint, content string) *model.Comment {
	comment := model.Comment{
		UserID:  userID,
		WishID:  wishID,
		Content: content,
	}
	// 手动模拟事务（
	testDB.Create(&comment)
	testDB.Model(&model.Wish{}).Where("id = ?", wishID).UpdateColumn("comment_count", gorm.Expr("comment_count + 1"))
	return &comment
}

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

		// 检查 "幽灵用户" Bug 是否修复

		respUser, _ := data["user"].(map[string]interface{})
		assert.Equal(t, float64(user.ID), respUser["id"])
		assert.Equal(t, user.Nickname, respUser["nickname"])

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
		assert.Equal(t, float64(apperr.ERROR_PARAM_INVALID), resp["code"])
		assert.Equal(t, apperr.GetMsg(apperr.ERROR_PARAM_INVALID), resp["message"])
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

// TestUpdateComment
func TestUpdateComment(t *testing.T) {
	cleanup(testDB)
	author := createUserWithRole("author", "pass", "user")
	authorToken := createToken(author.ID)
	wish := createWish(author.ID, "test wish")
	comment := createComment(author.ID, wish.ID, "original content")

	t.Run("作者更新成功", func(t *testing.T) {
		reqBody := gin.H{"content": "Updated by author"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		// (警告：假设路由是 PUT /api/comments/:id。你的 router.go 里没有注册这个!)
		req, _ := http.NewRequest("PUT", "/api/comments/"+strconv.Itoa(int(comment.ID)), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authorToken)
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		resp := parseResponse(t, w)
		assert.Equal(t, float64(apperr.SUCCESS), resp["code"])
		// 使用统一的消息源
		assert.Equal(t, apperr.GetMsg(apperr.SUCCESS), resp["message"]) //

		data, _ := resp["data"].(map[string]interface{})
		assert.Equal(t, "Updated by author", data["content"])

		// 检查 "幽灵用户" Bug 是否修复
		respUser, _ := data["user"].(map[string]interface{})
		assert.Equal(t, float64(author.ID), respUser["id"])

		// 检查数据库
		var updatedComment model.Comment
		testDB.First(&updatedComment, comment.ID)
		assert.Equal(t, "Updated by author", updatedComment.Content)
	})
}
