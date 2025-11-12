package handler

import (
	"net/http"
	"time"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/service"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


// 在保存到数据库前，会调用 service.CheckContent 进行内容 AI 审核
func CreateWish(c *gin.Context, db *gorm.DB) {
	// 1. 检查登录用户
	userIDInterface, exists := c.Get("userID")
	if !exists {
		logger.Log.Error("创建愿望失败:未找到用户ID")
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		logger.Log.Error("创建愿望失败:用户ID类型转换错误")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	//  绑定输入
	var req struct {
		Content    string `json:"content" binding:"required"`
		Background string `json:"background"`
		IsPublic   *bool  `json:"isPublic"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warnw("创建愿望失败：参数绑定错误", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "请求参数不合法"},
		})
		return
	}

	//  AI 内容审核（在保存前调用）
	isViolating, aiErr := service.CheckContent(req.Content)
	if aiErr != nil {
		// 审核过程中出现明确错误（如内容为空/过长，或 AI 无法判断等），把错误信息返回给客户端
		logger.Log.Warnw("创建愿望被拒绝：内容审核出错或无法判断", "userID", userID, "error", aiErr)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": aiErr.Error()},
		})
		return
	}
	if isViolating {
		// AI 明确判定为不安全内容
		logger.Log.Infow("创建愿望被拒绝：AI 判定不安全", "userID", userID)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "内容未通过审核"},
		})
		return
	}

	//  构造 Wish 并保存
	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}
	wish := model.Wish{
		UserID:     userID,
		Content:    req.Content,
		Background: req.Background,
		IsPublic:   isPublic,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := db.Create(&wish).Error; err != nil {
		logger.Log.Errorw("创建愿望失败：保存到数据库出错", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	//  返回成功
	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data": gin.H{
			"wishID":    wish.ID,
			"createdAt": wish.CreatedAt,
		},
	})
}
