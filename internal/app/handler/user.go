package handler

import (
	"net/http"
	"regexp"
	"time"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/service" 
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
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
	Nickname string `json:"nickname"` // 允许为空，在后续默认赋值
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

var isStudentId = regexp.MustCompile(`^[0-9]{10}$`)

// Register 是 /api/register 接口的 Gin handler
func Register(c *gin.Context, db *gorm.DB) {
	var req RegisterRequest

	//  绑定 JSON 请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Warnw("注册请求参数绑定失败", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "用户名和密码均不能为空"},
		})
		return
	}

	//  验证业务逻辑（学号格式）
	if !isStudentId.MatchString(req.Username) {
		logger.Log.Warnw("注册失败：学号格式不正确", "username", req.Username)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "输入错误，请输入十位学号"},
		})
		return
	}

	// 如果昵称为空，默认使用用户名
	if req.Nickname == "" {
		req.Nickname = req.Username
	}

	
	// AI审核昵称
	isViolating, aiErr := service.CheckContent(req.Nickname)
	if aiErr != nil {
		// AI 服务本身出错 (例如内容为空/过长，或 API key 问题)
		// ai_service 内部会处理空字符串等，所以这里返回 400
		logger.Log.Warnw("注册时昵称审核服务失败或输入无效", "nickname", req.Nickname, "error", aiErr)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": aiErr.Error()}, // 将 AI service 的错误返回
		})
		return
	}
	if isViolating {
		// AI 判定昵称违规
		logger.Log.Warnw("注册失败：昵称违规", "nickname", req.Nickname)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "昵称包含不当内容，请修改"},
		})
		return
	}
	

	// 检查用户是否已存在
	var existingUser model.User
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		logger.Log.Warnw("注册失败：用户名已存在", "username", req.Username)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "该学号已被注册"},
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		logger.Log.Errorw("注册时查询用户失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	//  加密密码
	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if hashErr != nil {
		logger.Log.Errorw("注册时密码加密失败", "error", hashErr)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	//  创建新用户
	newUser := model.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Role:     "user", //  GORM 模型里 Role 默认是 "user"
	}

	if createErr := db.Create(&newUser).Error; createErr != nil {
		logger.Log.Errorw("创建用户到数据库失败", "error", createErr)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	logger.Log.Infow("新用户注册成功", "username", newUser.Username, "userID", newUser.ID)

	//  生成 Token
	token, tokenErr := util.GenerateToken(newUser.ID)
	if tokenErr != nil {
		logger.Log.Errorw("注册成功但生成 Token 失败", "error", tokenErr)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// 返回成功响应
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
		c.JSON(http.StatusBadRequest, gin.H{
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
			logger.Log.Infow("登陆失败。用户不存在", "username", req.Username)
		} else {
			logger.Log.Errorw("登录时查询用户失败", "error", err)
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    apperr.ERROR_LOGIN_FAILED,
			"message": apperr.GetMsg(apperr.ERROR_LOGIN_FAILED),
			"data":    gin.H{"error": "用户名或密码错误"},
		})
		return

	}
	//验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Log.Infow("登录失败。密码错误", "username", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{
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
		c.JSON(http.StatusInternalServerError, gin.H{
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
		c.JSON(http.StatusUnauthorized, gin.H{
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
		c.JSON(http.StatusBadRequest, gin.H{
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
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}

	//更新用户信息
	if req.Nickname != nil {
		// --- 3. [新] AI 审核昵称 ---
		isViolating, aiErr := service.CheckContent(*req.Nickname)
		if aiErr != nil {
			// AI 服务出错或输入无效
			logger.Log.Warnw("UpdateUser 昵称审核服务失败或输入无效", "nickname", *req.Nickname, "error", aiErr)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    apperr.ERROR_PARAM_INVALID,
				"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
				"data":    gin.H{"error": aiErr.Error()},
			})
			return
		}
		if isViolating {
			// AI 判定昵称违规
			logger.Log.Warnw("UpdateUser 失败：昵称违规", "nickname", *req.Nickname)
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    apperr.ERROR_PARAM_INVALID,
				"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
				"data":    gin.H{"error": "昵称包含不当内容，请修改"},
			})
			return
		}
		// 审核通过
		user.Nickname = *req.Nickname
		// --- AI 审核结束 ---
	}

	if req.AvatarID != nil {
		user.AvatarID = req.AvatarID
	}
	if err := db.Save(&user).Error; err != nil {
		logger.Log.Errorw("UpdateUser: 更新用户信息失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
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