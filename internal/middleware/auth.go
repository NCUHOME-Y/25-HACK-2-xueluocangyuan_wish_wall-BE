package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/util"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	// 支持多域名白名单，环境变量 CORS_ALLOWED_ORIGINS 以逗号分隔
	// 例如："https://snowkeptwishes.ncuhos.com,http://localhost:5173"
	// 若为空，使用安全默认：允许上述两个常用来源
	allowedEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowList []string
	if allowedEnv == "" {
		allowList = []string{
			"https://snowkeptwishes.ncuhos.com",
			"http://localhost:5173",
		}
	} else {
		parts := strings.Split(allowedEnv, ",")
		for _, p := range parts {
			o := strings.TrimSpace(p)
			if o != "" {
				allowList = append(allowList, o)
			}
		}
		if len(allowList) == 0 {
			allowList = []string{"https://snowkeptwishes.ncuhos.com", "http://localhost:5173"}
		}
	}

	// 小工具：判断请求 Origin 是否在白名单
	isAllowed := func(origin string) bool {
		if origin == "" {
			return false
		}
		for _, a := range allowList {
			if strings.EqualFold(a, origin) {
				return true
			}
		}
		return false
	}

	return func(c *gin.Context) {
		reqOrigin := c.Request.Header.Get("Origin")
		if isAllowed(reqOrigin) {
			// 只能回显具体 Origin（启用凭证时不能用 *）
			c.Writer.Header().Set("Access-Control-Allow-Origin", reqOrigin)
			c.Writer.Header().Set("Vary", "Origin")
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token, x-token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")

		if c.Request.Method == http.MethodOptions {
			// 预检请求：若来源不在白名单，拒绝；否则放行 204
			if !isAllowed(reqOrigin) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

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
		defer func() { // 捕获 Panic
			if r := recover(); r != nil {
				logger.Log.Errorw("服务器崩溃 (Panic)", "error", r)

				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    apperr.ERROR_SERVER_ERROR, // code: 10
					"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
					"data":    gin.H{},
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
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    apperr.ERROR_UNAUTHORIZED, // code: 3
				"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
				"data":    gin.H{"error": "未提供 Authorization Header"},
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
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
			c.JSON(http.StatusUnauthorized, gin.H{
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

// 检查 Token，如果有效，则注入 "userID"
// 如果无效或不存在，它*不会*报错，而是直接放行 (c.Next())
func JWTOptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			// 未提供 Token，直接放行
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			// Token 格式错误，放
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, parseErr := util.ParseToken(tokenString)

		if parseErr != nil {
			//  Token 过期或无效
			c.Next()
			return
		}

		// 验证成功
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
