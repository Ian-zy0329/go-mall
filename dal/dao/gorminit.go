package dao

import (
	"github.com/Ian-zy0329/go-mall/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var _DbMaster *gorm.DB
var _DbSlave *gorm.DB

func DB() *gorm.DB {
	return _DbSlave
}

func DBMaster() *gorm.DB {
	return _DbMaster
}

func initDB(option config.DbConnectOption) *gorm.DB {
	db, err := gorm.Open(mysql.Open(option.DSN), &gorm.Config{Logger: NewGormLogger()})
	if err != nil {
		panic(err)
	}
	sqlDb, _ := db.DB()
	sqlDb.SetMaxOpenConns(option.MaxOpenConn)
	sqlDb.SetMaxIdleConns(option.MaxIdleConn)
	sqlDb.SetConnMaxLifetime(option.MaxLifeTime)
	if err = sqlDb.Ping(); err != nil {
		panic(err)
	}
	return db
}

func init() {
	_DbMaster = initDB(config.Database.Master)
	_DbSlave = initDB(config.Database.Slave)
}

func SetDBMasterConn(conn *gorm.DB) {
	_DbMaster = conn
}

func SetDBSlaveConn(conn *gorm.DB) {
	_DbSlave = conn
}
