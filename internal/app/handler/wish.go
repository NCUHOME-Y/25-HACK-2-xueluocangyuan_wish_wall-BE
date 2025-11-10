package handler

import (
	"net/http"
	"strconv"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LikeWish handles POST /api/wishes/:id/like - toggle like on a wish
func LikeWish(c *gin.Context, db *gorm.DB) {
	// 1. Get wish ID from URL parameter
	wishIDStr := c.Param("id")
	wishID, err := strconv.ParseUint(wishIDStr, 10, 32)
	if err != nil {
		logger.Log.Warnw("点赞失败：愿望ID无效", "wishID", wishIDStr, "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "愿望ID无效"},
		})
		return
	}

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

	// 3. Check if wish exists
	var wish model.Wish
	if err := db.First(&wish, wishID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Log.Warnw("点赞失败：愿望不存在", "wishID", wishID)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_WISH_NOT_FOUND,
				"message": apperr.GetMsg(apperr.ERROR_WISH_NOT_FOUND),
				"data":    gin.H{},
			})
		} else {
			logger.Log.Errorw("点赞失败：查询愿望出错", "wishID", wishID, "error", err)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_SERVER_ERROR,
				"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
				"data":    gin.H{},
			})
		}
		return
	}

	// 4. Check if like already exists
	var existingLike model.Like
	likeExists := db.Where("wish_id = ? AND user_id = ?", wishID, userID).First(&existingLike).Error == nil

	// 5. Toggle like
	if likeExists {
		// Unlike: delete the like record and decrement like_count
		if err := db.Delete(&existingLike).Error; err != nil {
			logger.Log.Errorw("取消点赞失败", "wishID", wishID, "userID", userID, "error", err)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_LIKE_FAILED,
				"message": apperr.GetMsg(apperr.ERROR_LIKE_FAILED),
				"data":    gin.H{},
			})
			return
		}

		// Decrement like_count
		if err := db.Model(&wish).Update("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
			logger.Log.Errorw("取消点赞失败：更新点赞数出错", "wishID", wishID, "error", err)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_SERVER_ERROR,
				"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
				"data":    gin.H{},
			})
			return
		}

		// Reload wish to get updated like_count
		if err := db.First(&wish, wishID).Error; err != nil {
			logger.Log.Errorw("取消点赞失败：重新加载愿望出错", "wishID", wishID, "error", err)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_SERVER_ERROR,
				"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
				"data":    gin.H{},
			})
			return
		}

		// Return success response using RespondLike
		RespondLike(c, wish.LikeCount, false, wish.ID)

	} else {
		// Like: create a new like record and increment like_count
		newLike := model.Like{
			WishID: uint(wishID),
			UserID: userID,
		}

		if err := db.Create(&newLike).Error; err != nil {
			logger.Log.Errorw("点赞失败：创建点赞记录出错", "wishID", wishID, "userID", userID, "error", err)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_LIKE_FAILED,
				"message": apperr.GetMsg(apperr.ERROR_LIKE_FAILED),
				"data":    gin.H{},
			})
			return
		}

		// Increment like_count
		if err := db.Model(&wish).Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
			logger.Log.Errorw("点赞失败：更新点赞数出错", "wishID", wishID, "error", err)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_SERVER_ERROR,
				"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
				"data":    gin.H{},
			})
			return
		}

		// Reload wish to get updated like_count
		if err := db.First(&wish, wishID).Error; err != nil {
			logger.Log.Errorw("点赞失败：重新加载愿望出错", "wishID", wishID, "error", err)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_SERVER_ERROR,
				"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
				"data":    gin.H{},
			})
			return
		}

		// Return success response using RespondLike
		RespondLike(c, wish.LikeCount, true, wish.ID)
	}
}
