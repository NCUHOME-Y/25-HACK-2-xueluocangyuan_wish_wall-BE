package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment 表示对愿望的评论
type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	WishID    uint           `gorm:"not null;index" json:"wishId"`
	ParentID  *uint          `gorm:"index;default:null" json:"parentId,omitempty"`
	UserID    uint           `gorm:"not null;index" json:"userId"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	LikeCount int            `gorm:"not null;default:0" json:"likeCount"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	Wish    *Wish      `gorm:"foreignKey:WishID" json:"wish,omitempty"`
	User    *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Replies []*Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comments"
}
