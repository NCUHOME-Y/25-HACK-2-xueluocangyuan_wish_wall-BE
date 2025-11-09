package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token, x-token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

		c.Next()
	}
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
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

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.JSON(401, gin.H{
				"code":      401,
				"message":   "未提供认证令牌",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// 移除 "Bearer " 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// 解析 token
		claims, err := parseToken(token)
		if err != nil {
			c.JSON(401, gin.H{
				"code":      401,
				"message":   "认证令牌无效",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", claims.UserID)
		c.Set("studentID", claims.StudentID)
		c.Next()
	}
}

// JWT 声明结构
type Claims struct {
	UserID    uint   `json:"userID"`
	StudentID string `json:"studentID"`
	jwt.RegisteredClaims
}

// parseToken 解析 JWT token
func parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil // 从配置中读取
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RateLimit 频率限制中间件
func RateLimit(key string, limit int, window time.Duration) gin.HandlerFunc {
	// 简单的内存限流器（生产环境建议使用 Redis）
	visits := make(map[string][]time.Time)

	return func(c *gin.Context) {
		identifier := key
		if key == "ip" {
			identifier = c.ClientIP()
		} else if key == "user" {
			if userID, exists := c.Get("userID"); exists {
				identifier = fmt.Sprintf("user:%v", userID)
			} else {
				c.JSON(401, gin.H{
					"code":      401,
					"message":   "需要认证",
					"timestamp": time.Now().Format(time.RFC3339),
				})
				c.Abort()
				return
			}
		}

		now := time.Now()
		if visits[identifier] == nil {
			visits[identifier] = make([]time.Time, 0)
		}

		// 清理过期的访问记录
		validVisits := make([]time.Time, 0)
		for _, visit := range visits[identifier] {
			if now.Sub(visit) < window {
				validVisits = append(validVisits, visit)
			}
		}
		visits[identifier] = validVisits

		// 检查是否超过限制
		if len(visits[identifier]) >= limit {
			c.JSON(429, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
				"details": gin.H{
					"retry_after": window.Seconds(),
				},
				"timestamp": now.Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// 记录本次访问
		visits[identifier] = append(visits[identifier], now)
		c.Next()
	}
}
