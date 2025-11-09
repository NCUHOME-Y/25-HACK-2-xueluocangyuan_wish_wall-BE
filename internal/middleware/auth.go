package middleware

import(
	"net/http"
	"strings"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/util"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c*gin.Context){
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{
				"code":err.ERROR_UNAUTHORIZED,
				"message":"登陆已过期，请重新登录",
				"data":gin.H{},
			})
			c.Abort()//终止请求
			return
		}

		//检查Header格式是否正确
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusOK, gin.H{
				"code":err.ERROR_UNAUTHORIZED,
				"message":"请求头中auth格式有误",
				"data":gin.H{},
			})
			c.Abort()
			return
		}

		//解析token
		tokenString := parts[1]
		claims,paeseErr :=util.ParseToken(tokenString)
		if paeseErr != nil {
			//token过期，签名无效等等
			c.JSON(http.StatusOK, gin.H{
				"code":err.ERROR_UNAUTHORIZED,
				"message":"登录已过期，请重新登录",
				"data":gin.H{"error":paeseErr.Error()},
			})
			c.Abort()
			return
		}
		//Token验证成功，将当前请求的用户信息保存到gin的上下文中
		c.Set("userID", claims.UserID)
		c.Next()
	}
}