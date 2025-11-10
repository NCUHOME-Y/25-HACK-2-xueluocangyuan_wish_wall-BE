package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateLikeRequest 请求体：为某个愿望点赞
type CreateLikeRequest struct {
	WishID uint `json:"wishId" binding:"required"`
}

// LikeResponse 返回给前端的点赞信息（包含用户简要信息）
type LikeResponse struct {
	ID        uint      `json:"id"`
	WishID    uint      `json:"wishId"`
	UserID    uint      `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	User      UserShort `json:"user"`
}

// LikeWish 为愿望添加点赞（当前用户）
// - 要求登录（中间件在上下文中注入 userID）
// - 幂等：如果已点赞，返回已有点赞的信息或友好提示
// - 使用事务增加 Wish.LikeCount 并创建 Like 记录
func CreateLike(c *gin.Context, db *gorm.DB) {
	var req CreateLikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warnw("CreateLike: 参数绑定失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": "请求参数错误",
			"data":    gin.H{"error": err.Error()},
		})
		return
	}

	userIDI, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": "未登录或 token 无效",
			"data":    gin.H{},
		})
		return
	}
	userID := userIDI.(uint)

	// 检查 wish 是否存在
	var wish model.Wish
	if err := db.First(&wish, req.WishID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_PARAM_INVALID,
				"message": "愿望不存在",
				"data":    gin.H{},
			})
			return
		}
		logger.Log.Errorw("CreateLike: 查询 wish 失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": "服务器内部错误",
			"data":    gin.H{},
		})
		return
	}

	// 检查是否已点赞
	var existing model.Like
	if err := db.Where("wish_id = ? AND user_id = ?", req.WishID, userID).First(&existing).Error; err == nil {
		// 已存在
		var user model.User
		_ = db.First(&user, existing.UserID).Error
		resp := LikeResponse{
			ID:        existing.ID,
			WishID:    existing.WishID,
			UserID:    existing.UserID,
			CreatedAt: existing.CreatedAt,
			User: UserShort{
				ID:       user.ID,
				Nickname: user.Nickname,
				AvatarID: user.AvatarID,
			},
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.SUCCESS,
			"message": "已点赞",
			"data":    resp,
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		logger.Log.Errorw("CreateLike: 查询现有点赞失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": "服务器内部错误",
			"data":    gin.H{},
		})
		return
	}

	// 事务：创建 like 并增加 wish.like_count
	var like model.Like
	if err := db.Transaction(func(tx *gorm.DB) error {
		like = model.Like{
			WishID: req.WishID,
			UserID: userID,
		}
		if err := tx.Create(&like).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Wish{}).Where("id = ?", req.WishID).
			UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Log.Errorw("CreateLike: 事务失败，创建点赞失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": "点赞失败",
			"data":    gin.H{},
		})
		return
	}

	// 读取用户信息以便返回
	var user model.User
	_ = db.First(&user, like.UserID).Error

	resp := LikeResponse{
		ID:        like.ID,
		WishID:    like.WishID,
		UserID:    like.UserID,
		CreatedAt: like.CreatedAt,
		User: UserShort{
			ID:       user.ID,
			Nickname: user.Nickname,
			AvatarID: user.AvatarID,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": "点赞成功",
		"data":    resp,
	})
}

// Unlike (DeleteLike) 删除某条 like：仅作者或管理员可删除
// 支持通过 like id 删除：DELETE /likes/:id
func DeleteLike(c *gin.Context, db *gorm.DB) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": "缺少 like id",
			"data":    gin.H{},
		})
		return
	}
	idUint64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Log.Warnw("DeleteLike: id 解析失败", "id", idStr, "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": "like id 无效",
			"data":    gin.H{},
		})
		return
	}
	likeID := uint(idUint64)

	userIDI, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": "未登录或 token 无效",
			"data":    gin.H{},
		})
		return
	}
	userID := userIDI.(uint)

	var like model.Like
	if err := db.First(&like, likeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_PARAM_INVALID,
				"message": "点赞记录不存在",
				"data":    gin.H{},
			})
			return
		}
		logger.Log.Errorw("DeleteLike: 查询 like 失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": "服务器内部错误",
			"data":    gin.H{},
		})
		return
	}

	// 权限：作者或管理员可以删除
	if like.UserID != userID {
		var user model.User
		if err := db.First(&user, userID).Error; err != nil {
			logger.Log.Errorw("DeleteLike: 查询用户失败", "error", err, "userID", userID)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_UNAUTHORIZED,
				"message": "权限校验失败",
				"data":    gin.H{},
			})
			return
		}
		if user.Role != "admin" {
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_PARAM_INVALID,
				"message": "没有权限删除该点赞",
				"data":    gin.H{},
			})
			return
		}
	}

	// 事务删除并减少 wish.like_count
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&like).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Wish{}).Where("id = ?", like.WishID).
			UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - ?, 0)", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Log.Errorw("DeleteLike: 事务删除失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": "取消点赞失败",
			"data":    gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": "取消点赞成功",
		"data":    gin.H{},
	})
}

// ListLikesByWish 列出某个愿望的点赞用户（分页）
// GET /wishes/:wishId/likes?page=1&pageSize=20
func ListLikesByWish(c *gin.Context, db *gorm.DB) {
	wishIDStr := c.Param("wishId")
	if wishIDStr == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": "缺少 wishId",
			"data":    gin.H{},
		})
		return
	}
	wishIDUint64, err := strconv.ParseUint(wishIDStr, 10, 64)
	if err != nil {
		logger.Log.Warnw("ListLikesByWish: wishId 解析失败", "wishId", wishIDStr, "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": "wishId 无效",
			"data":    gin.H{},
		})
		return
	}
	wishID := uint(wishIDUint64)

	// 分页参数
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 200 {
			pageSize = v
		}
	}
	offset := (page - 1) * pageSize

	// 可选：确认 wish 存在
	var wish model.Wish
	if err := db.First(&wish, wishID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_PARAM_INVALID,
				"message": "愿望不存在",
				"data":    gin.H{},
			})
			return
		}
		logger.Log.Errorw("ListLikesByWish: 查询 wish 失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": "服务器内部错误",
			"data":    gin.H{},
		})
		return
	}

	var total int64
	if err := db.Model(&model.Like{}).Where("wish_id = ?", wishID).Count(&total).Error; err != nil {
		logger.Log.Errorw("ListLikesByWish: 计数失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": "服务器内部错误",
			"data":    gin.H{},
		})
		return
	}

	var likes []model.Like
	if err := db.Where("wish_id = ?", wishID).
		Preload("User").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&likes).Error; err != nil {
		logger.Log.Errorw("ListLikesByWish: 查询 likes 失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": "查询失败",
			"data":    gin.H{},
		})
		return
	}

	respItems := make([]LikeResponse, 0, len(likes))
	for _, l := range likes {
		respItems = append(respItems, LikeResponse{
			ID:        l.ID,
			WishID:    l.WishID,
			UserID:    l.UserID,
			CreatedAt: l.CreatedAt,
			User: UserShort{
				ID:       l.User.ID,
				Nickname: l.User.Nickname,
				AvatarID: l.User.AvatarID,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": "获取点赞列表成功",
		"data": gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
			"items":    respItems,
		},
	})
}
