
package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Nickname  string         `gorm:"size:50;not null;default:''" json:"nickname"`
	AvatarID  *uint          `gorm:"default:null" json:"avatarId"`
	WishValue int            `gorm:"not null;default:0" json:"wishValue"`
	Bio       *string        `gorm:"type:text" json:"bio,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username  string         `gorm:"size:50;not null;uniqueIndex" json:"username"`

	// 关联关系
	Wishes    []Wish     `gorm:"foreignKey:UserID" json:"wishes,omitempty"`
	Likes     []Like     `gorm:"foreignKey:UserID" json:"likes,omitempty"`
	Comments  []Comment  `gorm:"foreignKey:UserID" json:"comments,omitempty"`
	
  
  Role string `gorm:"size:16;default:'user'" json:"role"` // 角色，默认为"user"

}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
