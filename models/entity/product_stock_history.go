package entity

import "time"

type ProductStockHistory struct {
	IdProduct   string    `gorm:"column:id_product;"`
	TxDate      time.Time `gorm:"column:tx_date;"`
	StockInQty  int       `gorm:"column:stock_in_qty;"`
	StockOutQty int       `gorm:"column:stock_out_qty;"`
	StockOpname int       `gorm:"column:stock_opname;"`
	StockFinal  int       `gorm:"column:stock_final;"`
	Description string    `gorm:"column:description;"`
	CreatedAt   time.Time `gorm:"column:created_at;"`
}

func (ProductStockHistory) TableName() string {
	return "products_stock_history"
}
