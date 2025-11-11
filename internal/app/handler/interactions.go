package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetInteractions handles GET /api/wishes/:id/interactions
// 返回符合 API 文档的互动结构：wishInfo、likes（含 userList、totalCount、currentUserLiked）、comments（支持分页与回复）
func GetInteractions(c *gin.Context, db *gorm.DB) {
	// 1. 解析愿望ID
	wishIDStr := c.Param("id")
	wishID64, err := strconv.ParseUint(wishIDStr, 10, 32)
	if err != nil {
		logger.Log.Warnw("获取互动信息失败：愿望ID无效", "wishID", wishIDStr, "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "愿望ID无效"},
		})
		return
	}
	wishID := uint(wishID64)

	// 2. 解析评论分页参数（用于 comments.pagination）
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 3. 加载愿望基本信息
	var wish model.Wish
	if err := db.Select("id, content, user_id, is_public, like_count, comment_count").First(&wish, wishID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Log.Warnw("获取互动信息失败：愿望不存在", "wishID", wishID)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_WISH_NOT_FOUND,
				"message": apperr.GetMsg(apperr.ERROR_WISH_NOT_FOUND),
				"data":    gin.H{},
			})
			return
		}
		logger.Log.Errorw("获取互动信息失败：查询愿望出错", "wishID", wishID, "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// 4. 准备 wishInfo 返回字段
	wishInfo := gin.H{
		"id":       wish.ID,
		"content":  wish.Content,
		"userId":   wish.UserID,
		"isPublic": wish.IsPublic,
	}

	// 5. 查询 likes：totalCount，userList（按时间降序，限制数量），currentUserLiked（如果用户已登录）
	totalLikes := wish.LikeCount
	likeListLimit := 20

	var likesResult []struct {
		UserID   uint   `json:"userId"`
		Nickname string `json:"nickname"`
		AvatarID *uint  `json:"avatarId"`
		LikedAt  time.Time
	}
	if err := db.Table("likes").
		Select("likes.user_id as user_id, users.nickname as nickname, users.avatar_id as avatar_id, likes.created_at as liked_at").
		Joins("left join users on users.id = likes.user_id").
		Where("likes.wish_id = ?", wishID).
		Order("likes.created_at desc").
		Limit(likeListLimit).
		Scan(&likesResult).Error; err != nil {
		logger.Log.Errorw("获取互动信息：查询点赞用户列表出错", "wishID", wishID, "error", err)
		likesResult = []struct {
			UserID   uint   `json:"userId"`
			Nickname string `json:"nickname"`
			AvatarID *uint  `json:"avatarId"`
			LikedAt  time.Time
		}{}
	}

	// currentUserLiked
	currentUserLiked := false
	userIDInterface, loggedIn := c.Get("userID")
	if loggedIn {
		if uid, ok := userIDInterface.(uint); ok {
			var cnt int64
			if err := db.Model(&model.Like{}).Where("wish_id = ? AND user_id = ?", wishID, uid).Count(&cnt).Error; err == nil {
				currentUserLiked = cnt > 0
			} else {
				logger.Log.Errorw("获取互动信息：查询当前用户点赞状态出错", "wishID", wishID, "userID", uid, "error", err)
			}
		}
	}

	// 构造 likes.userList 响应
	likesUserList := make([]gin.H, 0, len(likesResult))
	for _, lr := range likesResult {
		likedAtStr := ""
		if !lr.LikedAt.IsZero() {
			likedAtStr = lr.LikedAt.Format(time.RFC3339)
		}
		likesUserList = append(likesUserList, gin.H{
			"userId":   lr.UserID,
			"nickname": lr.Nickname,
			"avatarId": lr.AvatarID,
			"likedAt":  likedAtStr,
		})
	}

	likesResp := gin.H{
		"totalCount":       totalLikes,
		"userList":         likesUserList,
		"currentUserLiked": currentUserLiked,
	}

	// 6. 查询 comments：只查询顶级评论（parent_id IS NULL），并分页；预加载用户和回复（回复按创建时间升序）
	var totalTopComments int64
	if err := db.Model(&model.Comment{}).Where("wish_id = ? AND parent_id IS NULL", wishID).Count(&totalTopComments).Error; err != nil {
		logger.Log.Errorw("获取互动信息：统计顶级评论总数失败", "wishID", wishID, "error", err)
		totalTopComments = 0
	}

	var topComments []*model.Comment
	if err := db.Where("wish_id = ? AND parent_id IS NULL", wishID).
		Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Order("created_at asc")
		}).
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&topComments).Error; err != nil {
		logger.Log.Errorw("获取互动信息：查询顶级评论失败", "wishID", wishID, "error", err)
		// 返回空 comments 部分，但仍返回 wishInfo & likes
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.SUCCESS,
			"message": apperr.GetMsg(apperr.SUCCESS),
			"data": gin.H{
				"wishInfo": wishInfo,
				"likes":    likesResp,
				"comments": gin.H{
					"list":       []gin.H{},
					"pagination": gin.H{"page": page, "pageSize": pageSize, "total": totalTopComments},
				},
			},
		})
		return
	}

	// 构造 comments.list 响应（包括 replies）
	commentItems := make([]gin.H, 0, len(topComments))
	var uid uint
	if loggedIn {
		if t, ok := userIDInterface.(uint); ok {
			uid = t
		}
	}
	for _, cm := range topComments {
		// build replies
		repliesItems := make([]gin.H, 0, len(cm.Replies))
		for _, rp := range cm.Replies {
			userNickname := ""
			var userAvatarID *uint
			if rp.User != nil {
				userNickname = rp.User.Nickname
				userAvatarID = rp.User.AvatarID
			}
			isOwnReply := false
			if loggedIn && uid != 0 && uid == rp.UserID {
				isOwnReply = true
			}
			repliesItems = append(repliesItems, gin.H{
				"id":           rp.ID,
				"content":      rp.Content,
				"userId":       rp.UserID,
				"userNickname": userNickname,
				"userAvatar":   userAvatarID,
				"likeCount":    rp.LikeCount,
				"createdAt":    rp.CreatedAt,
				"isOwn":        isOwnReply,
			})
		}

		userNickname := ""
		var userAvatarID *uint
		if cm.User != nil {
			userNickname = cm.User.Nickname
			userAvatarID = cm.User.AvatarID
		}
		isOwn := false
		if loggedIn && uid != 0 && uid == cm.UserID {
			isOwn = true
		}

		item := gin.H{
			"id":           cm.ID,
			"content":      cm.Content,
			"userId":       cm.UserID,
			"userNickname": userNickname,
			"userAvatar":   userAvatarID,
			"likeCount":    cm.LikeCount,
			"createdAt":    cm.CreatedAt,
			"isOwn":        isOwn,
			"replies":      repliesItems,
		}
		commentItems = append(commentItems, item)
	}

	commentsResp := gin.H{
		"list": commentItems,
		"pagination": gin.H{
			"page":     page,
			"pageSize": pageSize,
			"total":    totalTopComments,
		},
	}

	// 7. 返回最终响应
	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data": gin.H{
			"wishInfo": wishInfo,
			"likes":    likesResp,
			"comments": commentsResp,
		},
	})
}
