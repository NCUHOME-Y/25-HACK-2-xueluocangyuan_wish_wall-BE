package model

import "time"

// Like represents a like on a wish.
type Like struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	WishID    uint      `gorm:"not null;uniqueIndex:idx_user_wish" json:"wishId"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_wish" json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
}

// LikeData 对应前端 Data 接口：当前点赞相关的具体数据
type LikeData struct {
	// 当前总点赞数
	LikeCount int  `json:"likeCount"`
	// 当前点赞状态（true=已点赞，false=未点赞）
	Liked bool `json:"liked"`
	// 愿望ID
	WishID uint `json:"wishId"`
}

// LikeResponse 对应前端 Response 接口的基础实现
type LikeResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    LikeData `json:"data"`
	// 注意：前端接口允许任意额外字段（index signature），
	// 若需要在服务端返回额外字段，可以扩展此结构或直接在 handler 中返回 gin.H。
}

// NewLikeResponse 创建一个标准的点赞响应（默认 code=200）
func NewLikeResponse(likeCount int, liked bool, wishID uint) LikeResponse {
	return LikeResponse{
		Code:    200,
		Message: "success",
		Data: LikeData{
			LikeCount: likeCount,
			Liked:     liked,
			WishID:    wishID,
		},
	}
}
