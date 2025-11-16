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

// GetPublicWishes handles GET /api/wishes/public
// 支持分页：?page=1&pageSize=10
// 支持按标签过滤：?tag=xxxxx
// 请求参数说明（与前端接口定义一致）：page（可选）、pageSize（可选）、tag（可选）
func GetPublicWishes(c *gin.Context, db *gorm.DB) {
	// parse pagination params (page, pageSize) with defaults
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	tag := c.Query("tag")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		logger.Log.Warnw("获取公共愿望墙失败：页码无效", "page", pageStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "页码无效"},
		})
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		logger.Log.Warnw("获取公共愿望墙失败：分页大小无效", "pageSize", pageSizeStr, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperr.ERROR_PARAM_INVALID,
			"message": apperr.GetMsg(apperr.ERROR_PARAM_INVALID),
			"data":    gin.H{"error": "分页大小无效"},
		})
		return
	}
	offset := (page - 1) * pageSize

	// build base query for public wishes
	baseQuery := db.Model(&model.Wish{}).Where("is_public = ?", true)

	// if tag filter provided, join wish_tags and filter by tag_name
	if tag != "" {
		// join with wish_tags table; count and find should both use the join
		baseQuery = baseQuery.Joins("JOIN wish_tags wt ON wt.wish_id = wishes.id AND wt.tag_name = ?", tag)
	}

	// count total distinct wishes matching query
	var total int64
	// Use a subquery to count distinct wish IDs for better performance
	if tag != "" {
		// Subquery: select distinct wish IDs matching the tag
		var count int64
		subQuery := db.Model(&model.Wish{}).
			Select("wishes.id").
			Joins("JOIN wish_tags wt ON wt.wish_id = wishes.id AND wt.tag_name = ?", tag).
			Where("is_public = ?", true).
			Group("wishes.id")
		if err := db.Table("(?) as sub", subQuery).Count(&count).Error; err != nil {
			logger.Log.Errorw("获取公共愿望墙失败：统计总数出错", "tag", tag, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    apperr.ERROR_SERVER_ERROR,
				"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
				"data":    gin.H{},
			})
			return
		}
		total = count
	} else {
		// No tag filter, just count public wishes
		if err := db.Model(&model.Wish{}).Where("is_public = ?", true).Count(&total).Error; err != nil {
			logger.Log.Errorw("获取公共愿望墙失败：统计总数出错", "tag", tag, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    apperr.ERROR_SERVER_ERROR,
				"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
				"data":    gin.H{},
			})
			return
		}
	}

	// query wishes with ordering and pagination
	var wishes []model.Wish
	if err := baseQuery.
		Select("wishes.*").
		Order("wishes.created_at desc").
		Offset(offset).
		Limit(pageSize).
		Preload("User").
		Find(&wishes).Error; err != nil {
		logger.Log.Errorw("获取公共愿望墙失败：查询愿望出错", "tag", tag, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperr.ERROR_SERVER_ERROR,
			"message": apperr.GetMsg(apperr.ERROR_SERVER_ERROR),
			"data":    gin.H{},
		})
		return
	}

	// if logged in, determine which of these wishes the current user liked
	userIDInterface, loggedIn := c.Get("userID")
	var likedMap map[uint]bool
	if loggedIn {
		if uid, ok := userIDInterface.(uint); ok && len(wishes) > 0 {
			wishIDs := make([]uint, 0, len(wishes))
			for _, w := range wishes {
				wishIDs = append(wishIDs, w.ID)
			}
			var likes []model.Like
			if err := db.Where("user_id = ? AND wish_id IN ?", uid, wishIDs).Find(&likes).Error; err != nil {
				logger.Log.Errorw("获取公共愿望墙失败：查询点赞状态出错", "userID", uid, "error", err)
			} else {
				likedMap = make(map[uint]bool, len(likes))
				for _, l := range likes {
					likedMap[l.WishID] = true
				}
			}
		}
	}

	// construct response items
	items := make([]gin.H, 0, len(wishes))
	for _, w := range wishes {
		item := gin.H{
			"id":           w.ID,
			"content":      w.Content,
			"background":   w.Background,
			"isPublic":     w.IsPublic,
			"likeCount":    w.LikeCount,
			"commentCount": w.CommentCount,
			"userId":       w.UserID,
			"userNickname": w.User.Nickname,
			"userAvatar":   w.User.AvatarID,
			"createdAt":    w.CreatedAt,
			"updatedAt":    w.UpdatedAt,
		}
		if loggedIn && likedMap != nil && likedMap[w.ID] {
			item["liked"] = true
		} else {
			item["liked"] = false
		}
		items = append(items, item)
	}

	// return response using page and pageSize to match frontend API
	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data": gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
			"tag":      tag,
			"wishes":   items,
		},
	})
}
