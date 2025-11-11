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

// DeleteWish handles DELETE /api/wishes/:id
// 只有愿望所有者可以删除
// 返回格式遵循 API 文档：data 包含 deletedAt 和 deletedWishId
func DeleteWish(c *gin.Context, db *gorm.DB) {
	// 1. 解析愿望ID
	wishIDStr := c.Param("id")
	wishID64, err := strconv.ParseUint(wishIDStr, 10, 32)
	if err != nil {
		logger.Log.Warnw("删除愿望失败：愿望ID无效", "wishID", wishIDStr, "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "愿望ID无效"},
		})
		return
	}
	wishID := uint(wishID64)

	// 2. 获取用户ID
	userIDInterface, exists := c.Get("userID")
	if !exists {
		logger.Log.Error("删除愿望失败：未找到用户ID")
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		logger.Log.Error("删除愿望失败：用户ID类型转换错误")
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// 3. 查找愿望并校验所有者，然后删除（事务）
	var deletedAt time.Time
	if err := db.Transaction(func(tx *gorm.DB) error {
		var wish model.Wish
		if err := tx.First(&wish, wishID).Error; err != nil {
			return err
		}
		if wish.UserID != userID {
			return errors.New("not_owner")
		}
		// Perform delete (soft delete if model has DeletedAt)
		if err := tx.Delete(&wish).Error; err != nil {
			return err
		}
		// set deletedAt to now for response (GORM may not update the struct's DeletedAt)
		deletedAt = time.Now()
		// 可选：同时删除关联的 likes/comments/tags（视业务而定）
		// tx.Where("wish_id = ?", wishID).Delete(&model.Like{})
		// tx.Where("wish_id = ?", wishID).Delete(&model.Comment{})
		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Log.Warnw("删除愿望失败：愿望不存在", "wishID", wishID)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_WISH_NOT_FOUND,
				"message": apperr.GetMsg(apperr.ERROR_WISH_NOT_FOUND),
				"data":    gin.H{},
			})
			return
		}
		if err.Error() == "not_owner" {
			logger.Log.Warnw("删除愿望失败：非愿望所有者", "wishID", wishID, "userID", userID)
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_UNAUTHORIZED,
				"message": "没有权限删除该愿望",
				"data":    gin.H{},
			})
			return
		}
		logger.Log.Errorw("删除愿望事务失败", "wishID", wishID, "userID", userID, "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// 4. 成功返回（包含 deletedAt 和 deletedWishId）
	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data": gin.H{
			"deletedAt":     deletedAt,
			"deletedWishId": wishID,
		},
	})
}
