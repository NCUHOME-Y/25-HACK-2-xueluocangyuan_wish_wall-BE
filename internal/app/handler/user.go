package handler

import (
	"net/http"
	"time"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err" //诶诶还有命名冲突
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterRequest 定义了注册时前端传来的 JSON 结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateUserRequest struct {
	Nickname *string `json:"nickname"`
	AvatarID *uint   `json:"avatarID"` //指针可以用来区分未传和传了0
}

// UserResponse 定义了注册/登录成功时返回的用户信息
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	AvatarID  *uint     `json:"avatarID"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

// Register 是 /api/register 接口的 Gin handler
func Register(c *gin.Context, db *gorm.DB) {
	var req RegisterRequest

	// 1. 绑定 JSON 请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果前端没传 username 或 password
		logger.Log.Warnw("注册请求参数绑定失败", "error", err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID, // 我们的标准 code: 4
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "用户名和密码均不能为空"},
		})
		return
	}

	// 2. 验证业务逻辑
	if len(req.Username) != 10 {
		logger.Log.Warnw("注册失败：学号不为10位", "username", req.Username)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID, // 我们的标准 code: 4
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "输入错误，请输入十位学号"}, // API 文档里的标准错误
		})
		return
	}

	// 3. 检查用户是否已存在
	var existingUser model.User
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		// 找到了用户，说明已被注册
		logger.Log.Warnw("注册失败：用户名已存在", "username", req.Username)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID, // 同样是参数错误
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "该学号已被注册"}, // 更具体的提示
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		// 如果是数据库查询出错，这是服务器内部错误
		logger.Log.Errorw("注册时查询用户失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR, // 我们的标准 code: 10
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// 4. 加密密码
	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if hashErr != nil {
		logger.Log.Errorw("注册时密码加密失败", "error", hashErr)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR, // 我们的标准 code: 10
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
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
		logger.Log.Errorw("创建用户到数据库失败", "error", createErr)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR, // 我们的标准 code: 10
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	logger.Log.Infow("新用户注册成功", "username", newUser.Username, "userID", newUser.ID)

	// 6. 生成 Token
	token, tokenErr := util.GenerateToken(newUser.ID)
	if tokenErr != nil {
		logger.Log.Errorw("注册成功但生成 Token 失败", "error", tokenErr)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR, // 我们的标准 code: 10
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
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
		"code":    apperr.SUCCESS, // 200
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data": gin.H{
			"token": token,
			"user":  responseUser,
		},
	})
}

func Login(c *gin.Context, db *gorm.DB) {
	var req LoginRequest

	//  绑定 JSON 请求体

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warnw("登录请求参数绑定失败", "error", err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "用户名和密码均不能为空"},
		})
		return
	}
	//查找用户
	var user model.User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			//未找到用户
			logger.Log.Infow("登陆失败。用户不存在", "username", req.Username)
		} else {
			//其他数据库错误
			logger.Log.Errorw("登录时查询用户失败", "error", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_LOGIN_FAILED,
			"message": apperr.GetMsg(apperr.ERROR_LOGIN_FAILED),
			"data":    gin.H{"error": "用户名或密码错误"},
		})
		return

	}
	//验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		//密码错误
		logger.Log.Infow("登录失败。密码错误", "username", req.Username)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_LOGIN_FAILED,
			"message": apperr.GetMsg(apperr.ERROR_LOGIN_FAILED),
			"data":    gin.H{"error": "用户名或密码错误"},
		})
		return
	}
	//生成token
	token, tokenErr := util.GenerateToken(user.ID)
	if tokenErr != nil {
		logger.Log.Errorw("登陆成功但生成Token失败", "error", tokenErr)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}
	//登陆成功，返回token和用户信息
	logger.Log.Infow("用户登录成功", "username", req.Username)

	respondUser := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		AvatarID:  user.AvatarID,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data":    gin.H{"token": token, "user": respondUser},
	})
}

func GetUserMe(c *gin.Context, db *gorm.DB) {
	//从中间件注入的上下文直接获取userID
	userID, _ := c.Get("userID")

	//查找用户
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		logger.Log.Errorw("GetUserMe: 查询用户失败", "userID", userID)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}

	// 返回用户信息
	responseUser := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		AvatarID:  user.AvatarID,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data":    responseUser,
	})
}

func UpdateUser(c *gin.Context, db *gorm.DB) {
	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warnw("更新用户请求参数绑定失败", "error", err.Error())
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": err.Error()},
		})
		return
	}
	//从中间件注入的上下文直接获取userID
	userID, _ := c.Get("userID")
	//查找用户
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		logger.Log.Errorw("UpdateUser: 查询用户失败", "userID", userID)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}

	//更新用户信息
	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.AvatarID != nil {
		user.AvatarID = req.AvatarID
	}
	if err := db.Save(&user).Error; err != nil {
		logger.Log.Errorw("UpdateUser: 更新用户信息失败", "error", err)
		c.JSON(http.StatusOK, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}
	//返回更新后的用户信息
	responseUser := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		AvatarID:  user.AvatarID,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data":    responseUser,
	})
}
