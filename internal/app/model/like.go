package model

import (
	"time"

	"gorm.io/gorm"
)

// Like represents a like on a wish.
type Like struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	WishID    uint           `gorm:"not null;uniqueIndex:idx_user_wish" json:"wishId"`
	UserID    uint           `gorm:"not null;uniqueIndex:idx_user_wish" json:"userId"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Wish Wish `gorm:"foreignKey:WishID" json:"wish,omitempty"`
}

// TableName 指定表名
func (Like) TableName() string {
	return "likes"
}
