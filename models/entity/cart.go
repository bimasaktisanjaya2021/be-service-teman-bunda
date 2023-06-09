package entity

import "time"

type Cart struct {
	Id        string    `gorm:"primaryKey;column:id;"`
	IdUser    string    `gorm:"column:id_user;"`
	IdProduct string    `gorm:"column:id_product;"`
	Product   Product   `gorm:"foreignKey:IdProduct"`
	Qty       int       `gorm:"column:qty;"`
	CreatedAt time.Time `gorm:"column:created_at;"`
}

func (Cart) TableName() string {
	return "cart"
}
