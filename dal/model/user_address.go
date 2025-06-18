package model

import (
	"gorm.io/plugin/soft_delete"
	"time"
)

type UserAddress struct {
	ID            int64                 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	UserId        int64                 `gorm:"column:user_id;NOT NULL"`
	UserName      string                `gorm:"column:user_name;NOT NULL"`
	UserPhone     string                `gorm:"column:user_phone;NOT NULL"`
	Default       int                   `gorm:"column:default;default:0;NOT NULL"`
	ProvinceName  string                `gorm:"column:province_name;NOT NULL"`
	CityName      string                `gorm:"column:city_name;NOT NULL"`
	RegionName    string                `gorm:"column:region_name;NOT NULL"`
	DetailAddress string                `gorm:"column:detail_address;NOT NULL"`
	IsDel         soft_delete.DeletedAt `gorm:"softDelete:flag"`
	CreatedAt     time.Time             `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt     time.Time             `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (m *UserAddress) TableName() string {
	return "user_address"
}
