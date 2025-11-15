// internal/router/router.go
package router

import (
	"os"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/handler"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter 负责配置所有 API 路由
func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()

	//  注册全局中间件
	r.Use(middleware.CORSMiddleware())//跨域资源共享（CORS）中间件。允许或拒绝来自不同域名的前端页面访问你的 API。
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())//这个中间件会“接住”这个崩溃，防止整个服务器停止服务，并通常会返回一个 500 错误给客户端。

	//  创建 /api 根路由组
	api := r.Group("/api")
	{	// 匿名函数可以使用它被定义时所在作用域的变量（这里就是 db）。
		// 注册 (提升到公共区域，防止 ACTIVE_ACTIVITY 未设置时 404)
		api.POST("/register", func(c *gin.Context) { handler.Register(c, db) })//调用 handler.Register 函数，并把 gin.Context 和数据库连接 db 传递给它。

		// 登录 (V1 和 V2 都需要)
		api.POST("/login", func(c *gin.Context) { handler.Login(c, db) })
		// 获取应用状态 (V1 和 V2 都需要)
		api.GET("/app-state", handler.GetAppState)
		// 内部 AI 测试 (V1 和 V2 都保留)
		api.POST("/test-ai", func(c *gin.Context) { handler.TestAI(c) })

		// 公共：获取某个愿望的评论列表
		api.GET("/wishes/:id/comments", func(c *gin.Context) { handler.ListCommentsByWish(c, db) })

		// 公共：获取公共愿望列表（可带 Token，用于 liked 状态；不强制，可选鉴权）
		public := api.Group("/")
		public.Use(middleware.JWTOptionalAuthMiddleware())
		{
			public.GET("/wishes/public", func(c *gin.Context) { handler.GetPublicWishes(c, db) })
		}

		//受保护的基础路由 (V1 和 V2 都需要)
		auth := api.Group("/")
		auth.Use(middleware.JWTAuthMiddleware())
		{
			// 获取用户信息 (V1 和 V2 都需要)
			auth.GET("/user/me", func(c *gin.Context) { handler.GetUserMe(c, db) })
			// 查看个人星河 (V2 "只读" 的核心功能)
			auth.GET("/wishes/me", func(c *gin.Context) { handler.GetMyWishes(c, db) })
			// 兼容测试用评论创建路由 (无论活动状态都提供)
			auth.POST("/comments", func(c *gin.Context) { handler.CreateComment(c, db) })
		}

		// V1 / V2 动态功能路由

		activity := os.Getenv("ACTIVE_ACTIVITY")

		if activity == "v1" {

			// V1 受保护路由
			{
				// 更新用户信息 (V1 允许)
				auth.PUT("/user", func(c *gin.Context) { handler.UpdateUser(c, db) })

				// 发布新愿望
				auth.POST("/wishes", func(c *gin.Context) {
					handler.CreateWish(c, db)
				})

				// 删除愿望
				auth.DELETE("/wishes/:id", func(c *gin.Context) {
					 handler.DeleteWish(c, db) // (确保 handler.DeleteWish 存在)
				})

				// 点赞/取消点赞
				auth.POST("/wishes/:id/like", func(c *gin.Context) {
					handler.LikeWish(c, db)
				})

				// 获取愿望互动详情
				auth.GET("/wishes/:id/interactions", func(c *gin.Context) {
					handler.GetInteractions(c, db)
				})

				// 创建评论或回复 
				auth.POST("/wishes/:id/comment", func(c *gin.Context) { handler.CreateComment(c, db) })
				//auth.PUT("/comments/:id", func(c *gin.Context) { handler.UpdateComment(c, db) })
				auth.DELETE("/comments/:id", func(c *gin.Context) { handler.DeleteComment(c, db) })

				auth.POST("/comments/reply", func(c *gin.Context) { handler.CreateReplyAI(c, db) })
			}

		} else {

			// V2 只读模式

			{
				// V2 模式下, PUT /user (更新用户信息) 被禁用
				// V2 模式下, POST /wishes (发布愿望) 被禁用
				// V2 模式下, DELETE /wishes/:id (删除愿望) 被禁用
				// V2 模式下, POST /wishes/:id/like (点赞) 被禁用
				// V2 模式下, GET /wishes/:id/interactions (看详情) 被禁用
				// V2 模式下, POST /wishes/:id/comment (评论) 被禁用
			}
			// V2 模式下，只有上面注册的“基础路由” - 登录 和 查看个人心愿
		}
	}

	return r
}
