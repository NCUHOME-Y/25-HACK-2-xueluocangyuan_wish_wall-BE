package handler

import (
	"net/http"
	"strconv"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/service"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateCommentAI 是带 AI 内容审核的创建评论处理器
// 路由示例：POST /api/comments (需鉴权)
// 请求体：{ "wishId": 1, "content": "..." }
// 返回遵循现有 CommentResponse 格式
func CreateCommentAI(c *gin.Context, db *gorm.DB) {
	var req struct {
		WishID  uint   `json:"wishId" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warnw("CreateCommentAI: 参数绑定失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": err.Error()},
		})
		return
	}

	userIDI, ok := c.Get("userID")
	if !ok {
		logger.Log.Warn("CreateCommentAI: 未找到 userID 上下文")
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}
	userID, ok := userIDI.(uint)
	if !ok {
		logger.Log.Error("CreateCommentAI: userID 类型断言失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// AI 内容审核： CheckContent 返回两个值 (isViolating, err)
	isViolating, aiErr := service.CheckContent(req.Content)
	if aiErr != nil {
		// 审核过程中出现明确错误（如内容为空/过长，或 AI 无法判断等），把错误信息返回给客户端
		logger.Log.Warnw("CreateCommentAI: 内容审核出错或无法判断", "userID", userID, "error", aiErr)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": aiErr.Error()},
		})
		return
	}
	if isViolating {
		// AI 明确判定为不安全内容，拒绝创建
		logger.Log.Infow("CreateCommentAI: AI 判定不安全，拒绝创建评论", "userID", userID)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "内容未通过审核"},
		})
		return
	}

	//  校验 wish 是否存在，并在事务中创建评论与更新计数
	var comment model.Comment
	if err := db.Transaction(func(tx *gorm.DB) error {
		var wish model.Wish
		if err := tx.First(&wish, req.WishID).Error; err != nil {
			return err
		}

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
		if err == gorm.ErrRecordNotFound {
			logger.Log.Infow("CreateCommentAI: wish 未找到", "wishId", req.WishID)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    apperr.ERROR_PARAM_INVALID,
				"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
				"data":    gin.H{},
			})
			return
		}
		logger.Log.Errorw("CreateCommentAI: 事务失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// 为避免幽灵用户，重新查询并预加载 User（非致命）
	if err := db.Preload("User").First(&comment, comment.ID).Error; err != nil {
		logger.Log.Warnw("CreateCommentAI: 重新查询并预加载用户失败", "commentID", comment.ID, "error", err)
	}

	//  构造返回体（与项目中 CommentResponse 保持一致）
	resp := gin.H{
		"id":        comment.ID,
		"wishId":    comment.WishID,
		"userId":    comment.UserID,
		"content":   comment.Content,
		"createdAt": comment.CreatedAt,
		"user": gin.H{
			"id":       comment.User.ID,
			"nickname": comment.User.Nickname,
			"avatarId": comment.User.AvatarID,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data":    resp,
	})
}

// CreateReplyAI 是带 AI 审核的回复（子评论）创建器
// 请求体示例：{ "wishId": 1, "parentId": 10, "content": "回复内容" }
func CreateReplyAI(c *gin.Context, db *gorm.DB) {
	var req struct {
		WishID   uint   `json:"wishId" binding:"required"`
		ParentID uint   `json:"parentId" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warnw("CreateReplyAI: 参数绑定失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": err.Error()},
		})
		return
	}

	userIDI, ok := c.Get("userID")
	if !ok {
		logger.Log.Warn("CreateReplyAI: 未找到 userID 上下文")
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}
	userID, ok := userIDI.(uint)
	if !ok {
		logger.Log.Error("CreateReplyAI: userID 类型断言失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// AI 审核
	isViolating, aiErr := service.CheckContent(req.Content)
	if aiErr != nil {
		logger.Log.Warnw("CreateReplyAI: 内容审核出错或无法判断", "userID", userID, "error", aiErr)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": aiErr.Error()},
		})
		return
	}
	if isViolating {
		logger.Log.Infow("CreateReplyAI: AI 判定不安全，拒绝创建回复", "userID", userID)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "内容未通过审核"},
		})
		return
	}

	// 创建回复并更新 wish.comment_count（事务）
	var reply model.Comment
	if err := db.Transaction(func(tx *gorm.DB) error {
		// 校验父评论与愿望存在性
		var parent model.Comment
		if err := tx.First(&parent, req.ParentID).Error; err != nil {
			return err
		}
		// 创建回复
		reply = model.Comment{
			WishID:   req.WishID,
			ParentID: &req.ParentID,
			UserID:   userID,
			Content:  req.Content,
		}
		if err := tx.Create(&reply).Error; err != nil {
			return err
		}
		// 更新愿望评论计数
		if err := tx.Model(&model.Wish{}).Where("id = ?", req.WishID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Log.Infow("CreateReplyAI: 父评论或愿望未找到", "parentId", req.ParentID, "wishId", req.WishID)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    apperr.ERROR_PARAM_INVALID,
				"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
				"data":    gin.H{},
			})
			return
		}
		logger.Log.Errorw("CreateReplyAI: 事务失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// 重新查询以预加载用户信息（非致命）
	if err := db.Preload("User").First(&reply, reply.ID).Error; err != nil {
		logger.Log.Warnw("CreateReplyAI: 重新查询并预加载用户失败", "commentID", reply.ID, "error", err)
	}

	resp := gin.H{
		"id":        reply.ID,
		"wishId":    reply.WishID,
		"userId":    reply.UserID,
		"parentId":  req.ParentID,
		"content":   reply.Content,
		"createdAt": reply.CreatedAt,
		"user": gin.H{
			"id":       reply.User.ID,
			"nickname": reply.User.Nickname,
			"avatarId": reply.User.AvatarID,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data":    resp,
	})
}

// GetInteractions 返回某个愿望的互动详情：likeCount, commentCount, 当前用户是否已点赞
// 兼容旧路由：GET /wishes/:id/interactions
func GetInteractions(c *gin.Context, db *gorm.DB) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": apperr.ERROR_PARAM_INVALID, "message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID), "data": gin.H{}})
		return
	}
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Log.Warnw("GetInteractions: id 解析失败", "id", idStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": apperr.ERROR_PARAM_INVALID, "message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID), "data": gin.H{}})
		return
	}
	wishID := uint(id64)

	var wish model.Wish
	if err := db.First(&wish, wishID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"code": apperr.ERROR_PARAM_INVALID, "message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID), "data": gin.H{}})
			return
		}
		logger.Log.Errorw("GetInteractions: 查询 wish 失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": apperr.ERROR_SERVER_ERROR, "message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR), "data": gin.H{}})
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
