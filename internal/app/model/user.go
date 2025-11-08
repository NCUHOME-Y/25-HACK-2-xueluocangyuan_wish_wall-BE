// 25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model/user.go
package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;size:64;not null" json:"username"` // 用户名，唯一且非空
	Password string `gorm:"not null;size:255" json:"-"`                   // 密码，非空
	Nickname string `gorm:"size:64" json:"nickname"`                      // 昵称
	Avatar   string `gorm:"size:255" json:"avatar"`                       // 头像URL

	Role string `gorm:"size:16;default:'user'" json:"role"` // 角色，默认为"user"
	//v2修改了，不再需要HasCompletedV2Review字段
}
