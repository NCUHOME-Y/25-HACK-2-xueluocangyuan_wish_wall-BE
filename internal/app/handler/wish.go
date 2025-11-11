package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LikeWish handles POST /api/wishes/:id/like - toggle like on a wish
func LikeWish(c *gin.Context, db *gorm.DB) {
	// 1. Get wish ID from URL parameter
	wishIDStr := c.Param("id")
	wishID64, err := strconv.ParseUint(wishIDStr, 10, 32)
	if err != nil {
		logger.Log.Warnw("点赞失败：愿望ID无效", "wishID", wishIDStr, "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "愿望ID无效"},
		})
		return
	}
	wishID := uint(wishID64)

	// 2. Get user ID from JWT context (set by middleware)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		logger.Log.Error("点赞失败：未找到用户ID")
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		logger.Log.Error("点赞失败：用户ID类型转换错误")
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// Variables to return after transaction
	var (
		finalLikeCount int
		finalLiked     bool
	)

	// 3. Run transaction to toggle like atomically
	txErr := db.Transaction(func(tx *gorm.DB) error {
		// Load wish with FOR UPDATE lock to prevent race conditions
		var wish model.Wish
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&wish, wishID).Error; err != nil {
			// propagate ErrRecordNotFound so outer code can map to WISH_NOT_FOUND
			return err
		}

		// Check if like already exists within the transaction
		var existingLike model.Like
		if err := tx.Where("wish_id = ? AND user_id = ?", wishID, userID).First(&existingLike).Error; err == nil {
			// Unlike: delete the like record and decrement like_count
			if err := tx.Delete(&existingLike).Error; err != nil {
				logger.Log.Errorw("取消点赞失败", "wishID", wishID, "userID", userID, "error", err)
				return err
			}
			if err := tx.Model(&wish).Update("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
				logger.Log.Errorw("取消点赞失败：更新点赞数出错", "wishID", wishID, "error", err)
				return err
			}
			// Reload wish to get updated like_count
			if err := tx.First(&wish, wishID).Error; err != nil {
				logger.Log.Errorw("取消点赞失败：重新加载愿望出错", "wishID", wishID, "error", err)
				return err
			}
			finalLikeCount = wish.LikeCount
			finalLiked = false
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			// Query error
			logger.Log.Errorw("查询点赞记录出错", "wishID", wishID, "userID", userID, "error", err)
			return err
		}

		// Like: create a new like record and increment like_count
		newLike := model.Like{
			WishID: wishID,
			UserID: userID,
		}
		if err := tx.Create(&newLike).Error; err != nil {
			logger.Log.Errorw("点赞失败：创建点赞记录出错", "wishID", wishID, "userID", userID, "error", err)
			return err
		}
		if err := tx.Model(&wish).Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
			logger.Log.Errorw("点赞失败：更新点赞数出错", "wishID", wishID, "error", err)
			return err
		}
		// Reload wish to get updated like_count
		if err := tx.First(&wish, wishID).Error; err != nil {
			logger.Log.Errorw("点赞失败：重新加载愿望出错", "wishID", wishID, "error", err)
			return err
		}
		finalLikeCount = wish.LikeCount
		finalLiked = true
		return nil
	})

	// 4. Handle transaction result
	if txErr != nil {
		if errors.Is(txErr, gorm.ErrRecordNotFound) {
			logger.Log.Warnw("点赞失败：愿望不存在", "wishID", wishID)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_WISH_NOT_FOUND,
				"message": apperr.GetMsg(apperr.ERROR_WISH_NOT_FOUND),
				"data":    gin.H{},
			})
			return
		}
		logger.Log.Errorw("点赞事务失败", "wishID", wishID, "userID", userID, "error", txErr)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_LIKE_FAILED,
			"message": apperr.GetMsg(apperr.ERROR_LIKE_FAILED),
			"data":    gin.H{},
		})
		return
	}

	// 5. 成功返回
	RespondLike(c, finalLikeCount, finalLiked, wishID)
}

// RespondLike 返回点赞/取消点赞的标准响应
func RespondLike(c *gin.Context, likeCount int, liked bool, wishID uint) {
	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data": gin.H{
			"wishID":    wishID,
			"likeCount": likeCount,
			"liked":     liked,
		},
	})
}
