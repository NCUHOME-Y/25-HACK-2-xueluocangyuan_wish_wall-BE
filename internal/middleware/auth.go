package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/util"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token, x-token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// 2. LoggerMiddleware
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// (使用 'apperr' 作为别名)
		defer func() {// 捕获 Panic
			if r := recover(); r != nil {
				logger.Log.Errorw("服务器崩溃 (Panic)", "error", r)

				c.JSON(http.StatusOK, gin.H{
					"code":    apperr.ERROR_SERVER_ERROR,                // code: 10
					"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR), // "服务器内部错误"
					"data":    gin.H{},                                  // data 是空对象
				})
			}
		}()
		c.Next()
	}
}

func JWTAuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_UNAUTHORIZED,                // code: 3
				"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED), // "登录已过期..."
				"data":    gin.H{"error": "未提供 Authorization Header"},
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_UNAUTHORIZED,
				"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
				"data":    gin.H{"error": "缺少 'Bearer ' 前缀"},
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, parseErr := util.ParseToken(tokenString)

		if parseErr != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    apperr.ERROR_UNAUTHORIZED,
				"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
				"data":    gin.H{"error": parseErr.Error()}, // 包含具体错误
			})
			c.Abort()
			return
		}

		// 验证成功
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
