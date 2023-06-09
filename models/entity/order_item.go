package entity

import "time"

type OrderItem struct {
	Id                  string    `gorm:"primaryKey;column:id;"`
	IdOrder             string    `gorm:"column:id_order;"`
	IdProduct           string    `gorm:"column:id_product;"`
	NoSku               string    `gorm:"column:no_sku;"`
	ProductName         string    `gorm:"column:product_name;"`
	PictureUrl          string    `gorm:"column:picture_url;"`
	Thumbnail           string    `gorm:"column:thumbnail;"`
	FlagPromo           string    `gorm:"column:flag_promo;"`
	Description         string    `gorm:"column:description;"`
	Weight              float64   `gorm:"column:weight;"`
	Volume              float64   `gorm:"column:volume;"`
	Qty                 int       `gorm:"column:qty;"`
	Price               float64   `gorm:"column:price;"`
	PriceBeforeDiscount float64   `gorm:"column:price_before_discount;"`
	PriceAfterDiscount  float64   `gorm:"column:price_after_discount;"`
	TotalPrice          float64   `gorm:"column:total_price;"`
	CreatedAt           time.Time `gorm:"column:created_at;"`
}

func (OrderItem) TableName() string {
	return "orders_items"
}
