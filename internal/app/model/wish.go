package model

import (
	"time"

	"gorm.io/gorm"
)

type Wish struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"not null;index" json:"userId"`
	// 冗余字段：便于列表直接展示作者昵称与头像，无需额外联表
	UserNickname string         `gorm:"size:50;not null;default:''" json:"userNickname"`
	UserAvatarID *uint          `json:"userAvatar"`
	Content      string         `gorm:"type:text;not null" json:"content"`
	IsPublic     bool           `gorm:"not null;default:true" json:"isPublic"`
	Background   string         `gorm:"size:50;default:'default'" json:"background"`
	LikeCount    int            `gorm:"not null;default:0" json:"likeCount"`
	CommentCount int            `gorm:"not null;default:0" json:"commentCount"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Tags     []WishTag `gorm:"foreignKey:WishID" json:"tags,omitempty"`
	Likes    []Like    `gorm:"foreignKey:WishID" json:"likes,omitempty"`
	Comments []Comment `gorm:"foreignKey:WishID" json:"comments,omitempty"`
}

type WishTag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	WishID    uint      `gorm:"not null;index" json:"wishId"`
	TagName   string    `gorm:"size:20;not null;index" json:"tagName"`
	CreatedAt time.Time `json:"createdAt"`

	// 关联关系
	Wish Wish `gorm:"foreignKey:WishID" json:"wish,omitempty"`
}

// TableName 指定表名
func (Wish) TableName() string {
	return "wishes"
}

func (WishTag) TableName() string {
	return "wish_tags"
}
