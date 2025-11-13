package handler

import (
	"net/http"
	"strconv"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 请求参数说明（与前端接口定义一致）：page、pageSize
// 返回当前登录用户自己发布的愿望列表（包含基础信息及该用户是否对每条愿望已点赞）
func GetMyWishes(c *gin.Context, db *gorm.DB) {
	// 从上下文获取用户ID（由认证中间件设置）
	userIDInterface, exists := c.Get("userID")
	if !exists {
		logger.Log.Error("获取我的愿望失败:未找到用户ID")
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    apperr.ERROR_UNAUTHORIZED,
			"message": apperr.GetMsg(apperr.ERROR_UNAUTHORIZED),
			"data":    gin.H{},
		})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		logger.Log.Error("获取我的愿望失败:用户ID类型转换错误")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	//  解析分页参数（使用前端约定的 page 和 pageSize）
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		logger.Log.Warnw("获取我的愿望失败：页码无效", "page", pageStr, "error", err)
	    c.JSON(http.StatusBadRequest, gin.H{
		    "code":    apperr.ERROR_PARAM_INVALID,
		    "message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
		    "data":    gin.H{"error": "页码无效"},
	    })
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		logger.Log.Warnw("获取我的愿望失败：分页大小无效", "pageSize", pageSizeStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "分页大小无效"},
		})
		return
	}
	offset := (page - 1) * pageSize

	//  统计总数
	var total int64
	if err := db.Model(&model.Wish{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		logger.Log.Errorw("获取我的愿望失败：统计总数出错", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	//  查询愿望列表
	var wishes []model.Wish
	if err := db.Where("user_id = ?", userID).
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&wishes).Error; err != nil {
		logger.Log.Errorw("获取我的愿望失败：查询愿望出错", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	//  查询该用户针对这些愿望的点赞状态（如果有愿望）
	likedMap := make(map[uint]bool)
	if len(wishes) > 0 {
		wishIDs := make([]uint, 0, len(wishes))
		for _, w := range wishes {
			wishIDs = append(wishIDs, w.ID)
		}
		var likes []model.Like
		if err := db.Where("user_id = ? AND wish_id IN ?", userID, wishIDs).Find(&likes).Error; err != nil {
			// 点赞查询失败不会影响主流程，只记录日志
			logger.Log.Errorw("获取我的愿望：查询点赞状态出错", "userID", userID, "error", err)
		} else {
			for _, l := range likes {
				likedMap[l.WishID] = true
			}
		}
	}

	// 构造响应数据
	items := make([]gin.H, 0, len(wishes))
	for _, w := range wishes {
		item := gin.H{
			"id":           w.ID,
			"content":      w.Content,
			"background":   w.Background,
			"isPublic":     w.IsPublic,
			"likeCount":    w.LikeCount,
			"commentCount": w.CommentCount,
			"createdAt":    w.CreatedAt,
			"updatedAt":    w.UpdatedAt,
		}
		if likedMap[w.ID] {
			item["liked"] = true
		} else {
			item["liked"] = false
		}
		items = append(items, item)
	}

	//  返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data": gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
			"wishes":   items,
		},
	})
}
