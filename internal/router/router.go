// internal/router/router.go
package router

import (
	"os" 

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/handler"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter 是你唯一的路由配置
func SetupRouter(db *gorm.DB) *gin.Engine {
	//  使用 gin.New()，手动控制中间件
	r := gin.New()

	// 注册“全局”中间件 (CORS, Logger, Recovery)
	
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	// 创建 /api 根路由组
	api := r.Group("/api")
	{
		// 永远在线的“公共路由” (V1 和 V2 都需要) 
		api.POST("/login", func(c *gin.Context) { handler.Login(c, db) })
		api.GET("/app-state", handler.GetAppState)
		api.POST("/test-ai", func(c *gin.Context) {
			handler.TestAI(c)
		})

		// 永远在线的“受保护路由” (V1 和 V2 都需要) 
		auth := api.Group("/")
		auth.Use(middleware.JWTAuthMiddleware())
		{
			// (V1 和 V2 都需要这些)
			auth.GET("/user/me", func(c *gin.Context) { handler.GetUserMe(c, db) })
			auth.PUT("/user", func(c *gin.Context) { handler.UpdateUser(c, db) })
		}

	
		activity := os.Getenv("ACTIVE_ACTIVITY")

		if activity == "v1" {
			// V1 开启时：挂载 V1 的所有接口 

			// V1 公开路由 (挂载到 "api" 组)
			api.POST("/register", func(c *gin.Context) { handler.Register(c, db) })
			api.GET("/wishes/public", func(c *gin.Context) {
				// handler.GetPublicWishes(c, db) 
			})

			// V1 受保护路由 (挂载到 "auth" 组)
			auth.POST("/wishes", func(c *gin.Context) {
				// handler.CreateWish(c, db) 
			})
			auth.GET("/wishes/me", func(c *gin.Context) {
				// handler.GetMyWishes(c, db) 
			})
			auth.DELETE("/wishes/:id", func(c *gin.Context) {
				// handler.DeleteWish(c, db) 
			})
			auth.POST("/wishes/:id/like", func(c *gin.Context) {
				// handler.LikeWish(c, db) 
			})
			auth.POST("/wishes/:id/comment", func(c *gin.Context) {
				// handler.CreateComment(c, db) 
			})
			auth.GET("/wishes/:id/interactions", func(c *gin.Context) {
				// handler.GetInteractions(c, db) 
			})

		} else {
			// --- V2 开启时 (或默认)：只挂载 V2 的接口 ---

			// V2 受保护路由 (挂载到 "auth" 组)
			// (handler.CompleteV2Review 路由已根据请求移除)
		}
	} // "api" 组的括号在这里结束

	return r
}