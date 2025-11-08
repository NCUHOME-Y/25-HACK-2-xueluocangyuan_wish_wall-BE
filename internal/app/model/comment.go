package model

import "time"

// Comment represents a comment on a wish.
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	WishID    uint      `gorm:"not null;index" json:"wishId"`
	UserID    uint      `gorm:"not null;index" json:"userId"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}
