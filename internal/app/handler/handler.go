package handler

import "github.com/gin-gonic/gin"

// 定义所有处理器的接口方法

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	RefreshToken(c *gin.Context)
}

type UserHandler interface {
	GetCurrentUser(c *gin.Context)
	UpdateUser(c *gin.Context)
	GetUserProfile(c *gin.Context)
}

type WishHandler interface {
	CreateWish(c *gin.Context)
	GetPublicWishes(c *gin.Context)
	GetMyWishes(c *gin.Context)
	UpdateWish(c *gin.Context)
	DeleteWish(c *gin.Context)
}

type InteractionHandler interface {
	ToggleLike(c *gin.Context)
	CreateComment(c *gin.Context)
	GetWishInteractions(c *gin.Context)
}

