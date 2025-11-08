package model

import "time"

// Like represents a like on a wish.
type Like struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	WishID    uint      `gorm:"not null;index" json:"wishId"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
}
