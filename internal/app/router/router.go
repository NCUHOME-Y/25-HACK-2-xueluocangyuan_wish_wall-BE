package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/handler"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/middleware"
)

// SetupRouter 设置路由
func SetupRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	wishHandler *handler.WishHandler,
	interactionHandler *handler.InteractionHandler,
) *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// API 路由组
	api := r.Group("/api")
	{
		// 认证路由
		setupAuthRoutes(api, authHandler)

		// 用户路由
		setupUserRoutes(api, userHandler)

		// 愿望路由
		setupWishRoutes(api, wishHandler)

		// 互动路由
		setupInteractionRoutes(api, interactionHandler)

	}

	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	return r
}

// setupAuthRoutes 认证相关路由
func setupAuthRoutes(api *gin.RouterGroup, authHandler *handler.AuthHandler) {
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)

		// 需要认证的路由
		auth.Use(middleware.JWTAuth())
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/refresh", authHandler.RefreshToken)
	}
}

// setupUserRoutes 用户相关路由
func setupUserRoutes(api *gin.RouterGroup, userHandler *handler.UserHandler) {
	user := api.Group("/user")
	{
		// 需要认证
		user.Use(middleware.JWTAuth())
		user.GET("/me", userHandler.GetCurrentUser)
		user.PUT("", userHandler.UpdateUser)
		user.GET("/profile", userHandler.GetUserProfile)
	}
}

// setupWishRoutes 愿望相关路由
func setupWishRoutes(api *gin.RouterGroup, wishHandler *handler.WishHandler) {
	wishes := api.Group("/wishes")
	{
		// 公开路由 - 不需要认证
		wishes.GET("/public", wishHandler.GetPublicWishes)

		// 需要认证的路由
		authWishes := wishes.Group("")
		authWishes.Use(middleware.JWTAuth())
		{
			authWishes.POST("", wishHandler.CreateWish)
			authWishes.GET("/me", wishHandler.GetMyWishes)
			authWishes.PUT("/:id", wishHandler.UpdateWish)
			authWishes.DELETE("/:id", wishHandler.DeleteWish)
		}
	}
}

// setupInteractionRoutes 互动相关路由
func setupInteractionRoutes(api *gin.RouterGroup, interactionHandler *handler.InteractionHandler) {
	interactions := api.Group("")
	{
		// 点赞路由
		interactions.POST("/wishes/:id/like", middleware.JWTAuth(), interactionHandler.ToggleLike)

		// 评论路由
		interactions.POST("/wishes/:id/comment", middleware.JWTAuth(), interactionHandler.CreateComment)

		// 获取互动列表路由 - 部分公开，部分需要认证
		interactions.GET("/wishes/:id/interactions", interactionHandler.GetWishInteractions)
	}
}
