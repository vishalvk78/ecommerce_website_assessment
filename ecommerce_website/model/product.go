package model

import (
	"time"

	"gorm.io/gorm"
)

type Products struct {
	gorm.Model
	ID          int     `json:"id" gorm:"primary_key"`
	Category    string  `json:"category"`
	ProductName string  `json:"productname"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Attributes  string  `json:"attributes"`
	Inventory   int     `json:"inventory"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type CreateProduct struct {
	Category    string  `json:"category"`
	ProductName string  `json:"productname"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Attributes  string  `json:"attributes"`
	Inventory   int     `json:"inventory"`
}

type UpdateProducts struct {
	Category    string  `json:"category"`
	ProductName string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Attributes  string  `json:"attributes"`
	Inventory   int     `json:"inventory"`
}

type PID struct {
	ProductID string `json:"productid"`
}
