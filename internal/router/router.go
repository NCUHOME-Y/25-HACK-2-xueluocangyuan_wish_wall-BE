// internal/router/router.go
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
	//  手动控制中间件
	r := gin.New()

	// 注册“全局”中间件 (CORS, Logger, Recovery)
	//
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	// 3. 创建 /api 根路由组
	api := r.Group("/api")
	{

		// (来自 handler/user.go 和 handler/app.go)
		api.POST("/login", func(c *gin.Context) { handler.Login(c, db) })
		api.GET("/app-state", handler.GetAppState)
		api.POST("/test-ai", func(c *gin.Context) { handler.TestAI(c) })

		// 受保护路由
		auth := api.Group("/")
		auth.Use(middleware.JWTAuthMiddleware()) //
		{

			auth.GET("/user/me", func(c *gin.Context) { handler.GetUserMe(c, db) })
			auth.PUT("/user", func(c *gin.Context) { handler.UpdateUser(c, db) })
		}

		// V1 / V2 动态路由
		//
		activity := os.Getenv("ACTIVE_ACTIVITY")

		if activity == "v1" {

			// V1 公开路由 (挂载到 "api" 组)
			api.POST("/register", func(c *gin.Context) { handler.Register(c, db) }) //

			// --- Comment 路由 (来自 handler/comment.go) ---
			//
			api.GET("/wishes/:wishId/comments", func(c *gin.Context) { handler.ListCommentsByWish(c, db) })

			// --- Like 路由 (来自 handler/like.go) ---
			//
			api.GET("/wishes/:wishId/likes", func(c *gin.Context) { handler.ListLikesByWish(c, db) })

			// V1 受保护路由 (挂载到 "auth" 组)

			// --- Wish 路由 (来自 handler/wish.go) ---
			//
			auth.POST("/wishes/:id/like", func(c *gin.Context) { handler.LikeWish(c, db) })

			// --- Comment 路由 (来自 handler/comment.go) ---
			//
			auth.POST("/comments", func(c *gin.Context) { handler.CreateComment(c, db) })
			//
			auth.PUT("/comments/:id", func(c *gin.Context) { handler.UpdateComment(c, db) })
			//
			auth.DELETE("/comments/:id", func(c *gin.Context) { handler.DeleteComment(c, db) })

			// --- Like 路由 (来自 handler/like.go) ---
			//
			auth.POST("/likes", func(c *gin.Context) { handler.CreateLike(c, db) })
			//
			auth.DELETE("/likes/:id", func(c *gin.Context) { handler.DeleteLike(c, db) })

			/*
			 auth.POST("/wishes", func(c *gin.Context) {
			 	// handler.CreateWish(c, db)
			 })
			 auth.GET("/wishes/me", func(c *gin.Context) {
			 	// handler.GetMyWishes(c, db)
			 })
			 auth.DELETE("/wishes/:id", func(c *gin.Context) {
			 	// handler.DeleteWish(c, db)
			 })
			 auth.GET("/wishes/:id/interactions", func(c *gin.Context) {
			 	// handler.GetInteractions(c, db)
			 })
			*/

		} else {
			//V2
		}
	} // "api" 组的括号在这里结束

	return r
}
