package handler

import (
	"net/http"
	"strconv"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	service "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/service"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateWishRequest 用于发布新愿望的请求体
type CreateWishRequest struct {
	Content  string `json:"content" binding:"required"`
	IsPublic *bool  `json:"isPublic"`
}

// WishSummary 用于返回给前端的简短愿望信息
type WishSummary struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"userId"`
	Content      string `json:"content"`
	IsPublic     bool   `json:"isPublic"`
	LikeCount    int    `json:"likeCount"`
	CommentCount int    `json:"commentCount"`
	Background   string `json:"background"`
	Nickname     string `json:"nickname"`
}

// CreateWish 发布新愿望（V1 模式）
func CreateWish(c *gin.Context, db *gorm.DB) {
	var req CreateWishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warnw("CreateWish: 参数绑定失败", "error", err)
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_PARAM_INVALID, "message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID), "data": gin.H{"error": err.Error()}})
		return
	}

	userIDI, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_UNAUTHORIZED, "message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED), "data": gin.H{}})
		return
	}
	userID := userIDI.(uint)

	// 内容审核
	isViolating, aiErr := service.CheckContent(req.Content)
	if aiErr != nil {
		logger.Log.Warnw("CreateWish: 内容审核失败", "userID", userID, "error", aiErr)
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_PARAM_INVALID, "message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID), "data": gin.H{"error": aiErr.Error()}})
		return
	}
	if isViolating {
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_PARAM_INVALID, "message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID), "data": gin.H{"error": "内容未通过审核"}})
		return
	}

	// 默认公开
	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	wish := model.Wish{
		UserID:   userID,
		Content:  req.Content,
		IsPublic: isPublic,
	}

	if err := db.Create(&wish).Error; err != nil {
		logger.Log.Errorw("CreateWish: 创建 wish 失败", "error", err)
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_SERVER_ERROR, "message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR), "data": gin.H{}})
		return
	}

	resp := WishSummary{
		ID:           wish.ID,
		UserID:       wish.UserID,
		Content:      wish.Content,
		IsPublic:     wish.IsPublic,
		LikeCount:    wish.LikeCount,
		CommentCount: wish.CommentCount,
		Background:   wish.Background,
	}

	c.JSON(http.StatusOK, gin.H{"code": apperr.SUCCESS, "message": apperr.GetMsg(apperr.SUCCESS), "data": resp})
}

// GetPublicWishes 列出公开愿望，支持分页
func GetPublicWishes(c *gin.Context, db *gorm.DB) {
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := db.Model(&model.Wish{}).Where("is_public = ?", true).Count(&total).Error; err != nil {
		logger.Log.Errorw("GetPublicWishes: 计数失败", "error", err)
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_SERVER_ERROR, "message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR), "data": gin.H{}})
		return
	}

	var wishes []model.Wish
	if err := db.Where("is_public = ?", true).Preload("User").Order("created_at desc").Offset(offset).Limit(pageSize).Find(&wishes).Error; err != nil {
		logger.Log.Errorw("GetPublicWishes: 查询失败", "error", err)
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_SERVER_ERROR, "message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR), "data": gin.H{}})
		return
	}

	items := make([]WishSummary, 0, len(wishes))
	for _, w := range wishes {
		items = append(items, WishSummary{
			ID:           w.ID,
			UserID:       w.UserID,
			Content:      w.Content,
			IsPublic:     w.IsPublic,
			LikeCount:    w.LikeCount,
			CommentCount: w.CommentCount,
			Background:   w.Background,
			Nickname:     w.User.Nickname,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": apperr.SUCCESS, "message": apperr.GetMsg(apperr.SUCCESS), "data": gin.H{"total": total, "page": page, "pageSize": pageSize, "items": items}})
}

// GetMyWishes 列出当前用户的愿望（已登录）
func GetMyWishes(c *gin.Context, db *gorm.DB) {
	userIDI, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_UNAUTHORIZED, "message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED), "data": gin.H{}})
		return
	}
	userID := userIDI.(uint)

	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := db.Model(&model.Wish{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		logger.Log.Errorw("GetMyWishes: 计数失败", "error", err)
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_SERVER_ERROR, "message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR), "data": gin.H{}})
		return
	}

	var wishes []model.Wish
	if err := db.Where("user_id = ?", userID).Preload("User").Order("created_at desc").Offset(offset).Limit(pageSize).Find(&wishes).Error; err != nil {
		logger.Log.Errorw("GetMyWishes: 查询失败", "error", err)
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_SERVER_ERROR, "message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR), "data": gin.H{}})
		return
	}

	items := make([]WishSummary, 0, len(wishes))
	for _, w := range wishes {
		items = append(items, WishSummary{
			ID:           w.ID,
			UserID:       w.UserID,
			Content:      w.Content,
			IsPublic:     w.IsPublic,
			LikeCount:    w.LikeCount,
			CommentCount: w.CommentCount,
			Background:   w.Background,
			Nickname:     w.User.Nickname,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": apperr.SUCCESS, "message": apperr.GetMsg(apperr.SUCCESS), "data": gin.H{"total": total, "page": page, "pageSize": pageSize, "items": items}})
}

// GetInteractions 返回某个愿望的互动详情：likeCount, commentCount, 当前用户是否已点赞
func GetInteractions(c *gin.Context, db *gorm.DB) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_PARAM_INVALID, "message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID), "data": gin.H{}})
		return
	}
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Log.Warnw("GetInteractions: id 解析失败", "id", idStr, "error", err)
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_PARAM_INVALID, "message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID), "data": gin.H{}})
		return
	}
	wishID := uint(id64)

	var wish model.Wish
	if err := db.First(&wish, wishID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_PARAM_INVALID, "message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID), "data": gin.H{}})
			return
		}
		logger.Log.Errorw("GetInteractions: 查询 wish 失败", "error", err)
		c.JSON(http.StatusOK, gin.H{"code": apperr.ERROR_SERVER_ERROR, "message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR), "data": gin.H{}})
		return
	}

	// 检查当前用户是否已点赞（如果登录）
	liked := false
	if userIDI, ok := c.Get("userID"); ok {
		userID := userIDI.(uint)
		var like model.Like
		if err := db.Where("wish_id = ? AND user_id = ?", wishID, userID).First(&like).Error; err == nil {
			liked = true
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": apperr.SUCCESS, "message": apperr.GetMsg(apperr.SUCCESS), "data": gin.H{"wishId": wishID, "likeCount": wish.LikeCount, "commentCount": wish.CommentCount, "liked": liked}})
}
