package handler

import (
	"net/http"

	apperr "github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/err"
	"github.com/gin-gonic/gin"
)

// RespondLike sends a standardized JSON response for like operations.
// It returns the response in the format expected by the frontend:
// { code: number, message: string, data: { likeCount: number, liked: boolean, wishId: number } }
func RespondLike(c *gin.Context, likeCount int, liked bool, wishID uint) {
	c.JSON(http.StatusOK, gin.H{
		"code":    apperr.SUCCESS,
		"message": apperr.GetMsg(apperr.SUCCESS),
		"data": gin.H{
			"likeCount": likeCount,
			"liked":     liked,
			"wishId":    wishID,
		},
	})
}
