package models

import (
	"time"
)

type Item struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"` // "details"
	Price        float64   `json:"price"`
	DeliveryTime string    `json:"delivery_time"` // "time to build"
	ImageURL     string    `json:"image_url"`
	Status       string    `json:"status"` // "available", "out_of_stock", "archived"
	CreatedAt    time.Time `json:"created_at"`
}

type Order struct {
	ID              int       `json:"id"`
	OrderRef        string    `json:"order_ref"` // Public "A7X9..." ID
	ItemID          int       `json:"item_id"`
	ItemTitle       string    `json:"item_title"` // For display convenience
	ItemImageURL    string    `json:"item_image_url"` // For display convenience
	Quantity        int       `json:"quantity"`
	CustomerName    string    `json:"customer_name"`
	CustomerEmail   string    `json:"customer_email"`
	CustomerAddress string    `json:"customer_address"`
	Status          string    `json:"status"`
	Notes           string    `json:"notes"`
	MagicToken      string    `json:"magic_token"`
	MagicTokenExpiry time.Time `json:"magic_token_expiry"`
	CreatedAt       time.Time `json:"created_at"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"` // Store hashed password
}
