// 25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model/wish.go
package model

import "gorm.io/gorm"

type Wish struct {
	gorm.Model
	
	UserID   uint   `gorm:"not null;index" json:"user_id"`                    // 用户ID，外键关联User表
	Content  string `gorm:"not null;size:512" json:"content"`                 // 愿望内容
	IsPublic bool   `gorm:"not null;default:true" json:"is_public"`           // 是否公开
	Status   string `gorm:"size:32;not null;default:'pending'" json:"status"` // 愿望状态，默认为 "pending"
}
