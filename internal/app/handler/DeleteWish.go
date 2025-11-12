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


func DeleteWish(c *gin.Context, db *gorm.DB) {
	//  解析愿望ID
	wishIDStr := c.Param("id")
	wishID64, err := strconv.ParseUint(wishIDStr, 10, 32)
	if err != nil {
		logger.Log.Warnw("删除愿望失败:愿望ID无效", "wishID", wishIDStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "愿望ID无效"},
		})
		return
	}
	wishID := uint(wishID64)

	// 获取用户ID (从中间件)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		logger.Log.Error("删除愿望失败：未找到用户ID")
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		logger.Log.Error("删除愿望失败：用户ID类型转换错误")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// 获取当前用户信息 (用于权限校验)
	var currentUser model.User
	if err := db.First(&currentUser, userID).Error; err != nil {
		logger.Log.Errorw("删除愿望失败：无法获取当前用户信息", "userID", userID, "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}

	//  查找愿望并校验权限，然后删除（事务）
	var deletedAt time.Time
	if err := db.Transaction(func(tx *gorm.DB) error {
		var wish model.Wish
		if err := tx.First(&wish, wishID).Error; err != nil {
			return err
		}

		// 权限校验：必须是作者 (isOwner) 或 管理员 (isAdmin)
		isOwner := (wish.UserID == userID)
		isAdmin := (currentUser.Role == "admin")

		if !isOwner && !isAdmin {
			return errors.New("not_authorized") // 修改错误标识
		}

		
		if err := tx.Delete(&wish).Error; err != nil {
			return err
		}
		
		deletedAt = time.Now()
	
		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Log.Warnw("删除愿望失败：愿望不存在", "wishID", wishID)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    apperr.ERROR_WISH_NOT_FOUND,
				"message": apperr.GetMsg(apperr.ERROR_WISH_NOT_FOUND),
				"data":    gin.H{},
			})
			return
		}
		// 捕获新的错误标识
		if err.Error() == "not_authorized" {
			logger.Log.Warnw("删除愿望失败：非愿望所有者或管理员", "wishID", wishID, "userID", userID, "role", currentUser.Role)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    apperr.ERROR_UNAUTHORIZED,
				"message": "没有权限删除该愿望",
				"data":    gin.H{},
			})
			return
		}
		logger.Log.Errorw("删除愿望事务失败", "wishID", wishID, "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// 成功返回
	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data": gin.H{
			"deletedAt":     deletedAt,
			"deletedWishId": wishID,
		},
	})
}