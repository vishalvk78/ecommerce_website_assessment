package model

import "time"

type Order struct {
	ID         int        `gorm:"primaryKey" json:"id"`
	ProductID  int        `gorm:"not null" json:"product_id"`
	Quantity   int        `gorm:"not null" json:"quantity"`
	TotalPrice float64    `gorm:"not null" json:"total_price"`
	UserID     int        `gorm:"not null" json:"user_id"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  *time.Time `gorm:"index" json:"-"`
}
