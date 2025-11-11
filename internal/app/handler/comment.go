package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	service "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/service"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateCommentRequest 前端发起新评论的请求结构
type CreateCommentRequest struct {
	WishID  uint   `json:"wishId" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// (UpdateCommentRequest 结构体已被删除)

// UserShort 用于在评论返回中携带简单用户信息
type UserShort struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	AvatarID *uint  `json:"avatarId"`
}

// CommentResponse 注释返回结构
type CommentResponse struct {
	ID        uint      `json:"id"`
	WishID    uint      `json:"wishId"`
	UserID    uint      `json:"userId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	User      UserShort `json:"user"`
}

// CreateComment 创建评论：用户需登录（中间件将 userID 写入上下文）
// - 校验请求体
// - 校验愿望是否存在
// - 创建评论并将 wish.comment_count +1
func CreateComment(c *gin.Context, db *gorm.DB) {
	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warnw("CreateComment: 参数绑定失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": err.Error()},
		})
		return
	}

	userIDi, ok := c.Get("userID")
	if !ok {
		logger.Log.Warn("CreateComment: 未找到 userID 上下文")
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}
	userID := userIDi.(uint)

	// 校验愿望是否存在
	var wish model.Wish
	if err := db.First(&wish, req.WishID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Log.Infow("CreateComment: wish 未找到", "wishId", req.WishID)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_PARAM_INVALID,
				"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
				"data":    gin.H{},
			})
			return
		}
		logger.Log.Errorw("CreateComment: 查询 wish 失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_WISH_NOT_FOUND,               
			"message": apperr.GetMsg(apperr.ERROR_WISH_NOT_FOUND),
			"data":    gin.H{},
		})
		return
	}

	// 检查是否允许评论
if !wish.IsPublic && wish.UserID != userID {
    logger.Log.Infow("CreateComment: 评论被拒绝，尝试评论私有愿望", "wishId", req.WishID, "userID", userID)
    c.JSON(http.StatusOK, gin.H{
        "code":    apperr.ERROR_FORBIDDEN_COMMENT, // <-- 对应 code 13
        "message": apperr.GetMsg(apperr.ERROR_FORBIDDEN_COMMENT),
        "data":    gin.H{},
    })
    return
}

	isViolating, aiErr := service.CheckContent(req.Content)
	if aiErr != nil {
		// AI 服务本身出错（如内容为空/过长 或无法判断）
		logger.Log.Warnw("创建评论被拒绝：内容审核出错", "userID", userID, "error", aiErr)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": aiErr.Error()},
		})
		return
	}
	if isViolating {
		// AI 明确判定为不安全内容
		logger.Log.Infow("创建评论被拒绝:AI 判定不安全", "userID", userID)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "内容未通过审核"},
		})
		return
	}

	// 创建评论（使用事务，确保 comment_count 与 comment 保持一致）
	var comment model.Comment
	if err := db.Transaction(func(tx *gorm.DB) error {
		comment = model.Comment{
			WishID:  req.WishID,
			UserID:  userID,
			Content: req.Content,
		}
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Wish{}).Where("id = ?", req.WishID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Log.Errorw("CreateComment: 创建评论失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_COMMENT_FAILED,
			"message": apperr.GetMsg(apperr.ERROR_COMMENT_FAILED),
			"data":    gin.H{},
		})
		return
	}

	// 为避免“幽灵用户”，在事务成功后重新查询并预加载 User
	if err := db.Preload("User").First(&comment, comment.ID).Error; err != nil {
		// 预加载失败不是致命错误，但要记录日志
		logger.Log.Warnw("CreateComment: 重新查询评论并预加载用户失败，可能返回无用户信息", "error", err, "commentID", comment.ID)
	}

	resp := CommentResponse{
		ID:        comment.ID,
		WishID:    comment.WishID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		User: UserShort{
			ID:       comment.User.ID,
			Nickname: comment.User.Nickname,
			AvatarID: comment.User.AvatarID,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data":    resp,
	})
}

// DeleteComment 删除评论：仅作者或管理员或评论人可删除
func DeleteComment(c *gin.Context, db *gorm.DB) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{},
		})
		return
	}
	idUint64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Log.Warnw("DeleteComment: id 解析失败", "id", idStr, "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_COMMENT_NOT_FOUND, 
			"message": apperr.GetMsg(apperr.ERROR_COMMENT_NOT_FOUND), 
			"data":    gin.H{},
		})
		return
	}
	commentID := uint(idUint64)

	userIDi, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}
	userID := userIDi.(uint)

	// 查询评论
	var comment model.Comment
	if err := db.First(&comment, commentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_PARAM_INVALID,
				"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
				"data":    gin.H{},
			})
			return
		}
		logger.Log.Errorw("DeleteComment: 查询评论失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// 如果不是作者，需要判断是否为管理员
	if comment.UserID != userID {
		var user model.User
		if err := db.First(&user, userID).Error; err != nil {
			logger.Log.Errorw("DeleteComment: 查询用户失败", "error", err, "userID", userID)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_UNAUTHORIZED,
				"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
				"data":    gin.H{},
			})
			return
		}
		if user.Role != "admin" {
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_PARAM_INVALID,
				"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
				"data":    gin.H{},
			})
			return
		}
	}

	// 删除并减少 wish.comment_count（事务）
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&comment).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Wish{}).Where("id = ?", comment.WishID).
			UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - ?, 0)", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Log.Errorw("DeleteComment: 删除评论事务失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data":    gin.H{},
	})
}

// ListCommentsByWish 列出某个愿望的评论，支持分页
// 路径示例：GET /wishes/:wishId/comments?page=1&pageSize=20
func ListCommentsByWish(c *gin.Context, db *gorm.DB) {
	wishIDStr := c.Param("wishId")
	if wishIDStr == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{},
		})
		return
	}
	wishIDUint64, err := strconv.ParseUint(wishIDStr, 10, 64)
	if err != nil {
		logger.Log.Warnw("ListCommentsByWish: wishId 解析失败", "wishId", wishIDStr, "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
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
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}
	offset := (page - 1) * pageSize

	// 检查 wish 是否存在（可选）
	var wish model.Wish
	if err := db.First(&wish, wishID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_PARAM_INVALID,
				"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
				"data":    gin.H{},
			})
			return
		}
		logger.Log.Errorw("ListCommentsByWish: 查询 wish 失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	var total int64
	var comments []model.Comment

	if err := db.Model(&model.Comment{}).Where("wish_id = ?", wishID).Count(&total).Error; err != nil {
		logger.Log.Errorw("ListCommentsByWish: 计数失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// 查询并预加载用户信息
	if err := db.Where("wish_id = ?", wishID).
		Preload("User").
		Order("created_at asc").
		Offset(offset).
		Limit(pageSize).
		Find(&comments).Error; err != nil {
		logger.Log.Errorw("ListCommentsByWish: 查询评论失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	respComments := make([]CommentResponse, 0, len(comments))
	for _, cm := range comments {
		respComments = append(respComments, CommentResponse{
			ID:        cm.ID,
			WishID:    cm.WishID,
			UserID:    cm.UserID,
			Content:   cm.Content,
			CreatedAt: cm.CreatedAt,
			User: UserShort{
				ID:       cm.User.ID,
				Nickname: cm.User.Nickname,
				AvatarID: cm.User.AvatarID,
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data": gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
			"items":    respComments,
		},
	})
}

// (UpdateComment 函数已被删除)
