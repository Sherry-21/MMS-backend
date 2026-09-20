package model

import "time"

// Product represents a product catalog entity
type Product struct {
	ID          int64            `gorm:"primaryKey;autoIncrement" json:"id"`
	CategoryID  int64            `gorm:"not null;index" json:"category_id"`
	Name        string           `gorm:"type:varchar(255);not null" json:"name"`
	Description string           `gorm:"type:text" json:"description"`
	CreatedAt   time.Time        `gorm:"autoCreateTime" json:"created_at"`
	Category    *Category        `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Variants    []ProductVariant `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
}

func (Product) TableName() string {
	return "products"
}
