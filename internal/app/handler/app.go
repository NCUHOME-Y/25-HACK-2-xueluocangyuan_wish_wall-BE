package handler

import (
	"net/http"
	"os"

	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err" //诶诶还有命名冲突
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/service"

	"github.com/gin-gonic/gin"
)

func GetAppState(c *gin.Context) {
	//从环境变量读取激活的活动配置
	activeActivity := os.Getenv("ACTIVE_ACTIVITY")

	// 检查是否配置
	if activeActivity == "" {
		logger.Log.Error("获取应用状态失败：环境变量 ACTIVE_ACTIVITY 未设置")
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR, // 标准错误码是10004
			"message": "服务器配置错误，请联系管理员",
			"data":    gin.H{},
		})
		return
	}

	// 返回当前应用状态
	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data": gin.H{
			"activeActivity": activeActivity, //v2
		},
	})

}

func TestAI(c *gin.Context) {
	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"error": "gimme json: {\"content\": \"...\"}"})
		return
	}

	// 直接调用你的 AI Service
	isViolating, aiErr := service.CheckContent(req.Content)
	if aiErr != nil {
		c.JSON(200, gin.H{"error": aiErr.Error()})
		return
	}

	// 把 AI 的原始结果直接返回给你
	c.JSON(200, gin.H{
		"input_content": req.Content,
		"is_violating":  isViolating, // (true=违规, false=安全)
	})
}
