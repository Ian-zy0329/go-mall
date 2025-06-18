package model

import (
	"gorm.io/plugin/soft_delete"
	"time"
)

type User struct {
	ID        int64                 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	Nickname  string                `gorm:"column:nickname;NOT NULL"`
	LoginName string                `gorm:"column:login_name;NOT NULL"`
	Password  string                `gorm:"column:password;NOT NULL"`
	Verified  string                `gorm:"column:verified;default:0;NOT NULL"`
	Avatar    string                `gorm:"column:avatar;NOT NULL"`
	Slogan    string                `gorm:"column:slogan;NOT NULL"`
	IsDel     soft_delete.DeletedAt `gorm:"softDelete:flag"`
	CreatedAt time.Time             `gorm:"column:created_at;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt time.Time             `gorm:"column:updated_at;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (u *User) TableName() string {
	return "users"
}
