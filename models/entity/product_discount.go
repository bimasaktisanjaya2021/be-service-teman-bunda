package entity

import (
	"time"
)

type ProductDiscount struct {
	Id         string    `gorm:"primaryKey;column:id;"`
	IdProduct  string    `gorm:"column:id_product;"`
	Percentage float64   `gorm:"column:percentage;"`
	Nominal    float64   `gorm:"column:nominal;"`
	FlagPromo  string    `gorm:"column:flag_promo;"`
	StartDate  time.Time `gorm:"column:start_date;"`
	EndDate    time.Time `gorm:"column:end_date;"`
}

func (ProductDiscount) TableName() string {
	return "products_discount"
}
