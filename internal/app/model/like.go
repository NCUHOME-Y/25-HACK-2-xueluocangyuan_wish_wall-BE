package model

import "time"

// Like represents a like on a wish.
type Like struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	WishID    uint      `gorm:"not null;uniqueIndex:idx_user_wish" json:"wishId"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_wish" json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
}
