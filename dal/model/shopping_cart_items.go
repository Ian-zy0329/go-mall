package model

import (
	"gorm.io/plugin/soft_delete"
	"time"
)

type ShoppingCartItem struct {
	CartItemId   int64                 `gorm:"column:cart_item_id;primary_key;AUTO_INCREMENT"`
	UserId       int64                 `gorm:"column:user_id;NOT NULL"`
	CommodityId  int64                 `gorm:"column:commodity_id;NOT NULL"`
	CommodityNum int                   `gorm:"column:commodity_num;default:1;NOT NULL"`
	IsDel        soft_delete.DeletedAt `gorm:"softDelete:flag"`
	CreatedAt    time.Time             `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt    time.Time             `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (ShoppingCartItem) TableName() string {
	return "shopping_cart_items"
}
