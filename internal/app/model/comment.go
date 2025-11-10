package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment represents a comment on a wish.
type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	WishID    uint           `gorm:"not null;index" json:"wishId"`
	UserID    uint           `gorm:"not null;index" json:"userId"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Wish Wish `gorm:"foreignKey:WishID" json:"wish,omitempty"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "comments"
}
