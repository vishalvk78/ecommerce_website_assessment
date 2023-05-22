package model

import (
	"time"

	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	ID        uint   `json:"id" gorm:"primary_key"`
	FullName  string `json:"fullname"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Role      string `json:"role"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UpdateUser struct {
	FullName string `json:"fullname"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	FullName string `json:"fullname"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

type RegisterResponse struct {
	Token string `json:"token"`
}

type WebTracking struct {
	gorm.Model
	ProductID   string `json:"productid"`
	Event       string `json:"event"`
	ProductName string `json:"product_name"`
	Category    string `json:"category"`
	Attributes  string `json:"attributes"`
}

type Web struct {
	ProductID string `json:"productid"`
}
