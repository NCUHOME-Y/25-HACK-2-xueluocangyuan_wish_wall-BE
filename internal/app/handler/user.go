package handler

import (
	"net/http"
	"time"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/util"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterRequest 定义了注册时前端传来的 JSON 结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse 定义了注册/登录成功时返回的用户信息
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	AvatarID  *uint     `json:"avatarId"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

// Register 是 /api/register 接口的 Gin handler
func Register(c *gin.Context, db *gorm.DB) {
	var req RegisterRequest

	// 1. 绑定 JSON 请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果前端没传 username 或 password
		zap.S().Warnw("注册请求参数绑定失败", "error", err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code": apperr.ERROR_PARAM_INVALID, // 我们的标准 code: 4
			"msg":  apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data": gin.H{"error": "用户名和密码均不能为空"},
		})
		return
	}

	// 2. 验证业务逻辑 
	if len(req.Username) != 10 {
		zap.S().Warnw("注册失败：学号不为10位", "username", req.Username)
		c.JSON(http.StatusOK, gin.H{
			"code": apperr.ERROR_PARAM_INVALID, // 我们的标准 code: 4
			"msg":  apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data": gin.H{"error": "输入错误，请输入十位学号"}, // API 文档里的标准错误
		})
		return
	}

	// 3. 检查用户是否已存在
	var existingUser model.User
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		// 找到了用户，说明已被注册
		zap.S().Warnw("注册失败：用户名已存在", "username", req.Username)
		c.JSON(http.StatusOK, gin.H{
			"code": apperr.ERROR_PARAM_INVALID, // 同样是参数错误
			"msg":  apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data": gin.H{"error": "该学号已被注册"}, // 更具体的提示
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		// 如果是数据库查询出错，这是服务器内部错误
		zap.S().Errorw("注册时查询用户失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code": apperr.ERROR_SERVER_ERROR, // 我们的标准 code: 10
			"msg":  apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data": gin.H{},
		})
		return
	}

	// 4. 加密密码
	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if hashErr != nil {
		zap.S().Errorw("注册时密码加密失败", "error", hashErr)
		c.JSON(http.StatusOK, gin.H{
			"code": apperr.ERROR_SERVER_ERROR, // 我们的标准 code: 10
			"msg":  "服务器内部错误",
			"data": gin.H{},
		})
		return
	}

	// 5. 创建新用户
	newUser := model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Username, // 你的 GORM 模型里 Nickname 是 not null，我们先用学号作为默认昵称
		Role:     "user",       // 你的 GORM 模型里 Role 默认是 "user"
		// AvatarID, WishValue 等字段用数据库默认值 (default)
	}

	if createErr := db.Create(&newUser).Error; createErr != nil {
		zap.S().Errorw("创建用户到数据库失败", "error", createErr)
		c.JSON(http.StatusOK, gin.H{
			"code": apperr.ERROR_SERVER_ERROR, // 我们的标准 code: 10
			"msg":  "注册失败，请稍后重试",
			"data": gin.H{},
		})
		return
	}

	zap.S().Infow("新用户注册成功", "username", newUser.Username, "userID", newUser.ID)

	// 6. 生成 Token
	token, tokenErr := util.GenerateToken(newUser.ID)
	if tokenErr != nil {
		zap.S().Errorw("注册成功但生成 Token 失败", "error", tokenErr)
		c.JSON(http.StatusOK, gin.H{
			"code": apperr.ERROR_SERVER_ERROR, // 我们的标准 code: 10
			"msg":  "注册成功，但登录失败",
			"data": gin.H{},
		})
		return
	}

	// 7. 返回成功响应
	// 准备返回给前端的用户信息 
	responseUser := UserResponse{
		ID:        newUser.ID,
		Username:  newUser.Username,
		Nickname:  newUser.Nickname,
		AvatarID:  newUser.AvatarID,
		Role:      newUser.Role,
		CreatedAt: newUser.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": apperr.SUCCESS, // 200
		"msg":  "注册成功",
		"data": gin.H{
			"token": token,
			"user":  responseUser,
		},
	})
}
