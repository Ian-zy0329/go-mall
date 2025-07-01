package config

import "time"

var (
	App      *appConfig
	Database *databaseConfig
	Redis    *redisConfig
)

type redisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type appConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Log  struct {
		FilePath         string `mapstructure:"path"`
		FileMaxSize      int    `mapstructure:"max_size"`
		BackUpFileMaxAge int    `mapstructure:"max_age"`
	}
	Pagination struct {
		DefaultSize int `mapstructure:"default_size"`
		MaxSize     int `mapstructure:"max_size"`
	}
	WechatPay struct {
		AppId           string `mapstructure:"appid"`
		MchId           string `mapstructure:"mchid"`
		PrivateSerialNo string `mapstructure:"private_serial_no"`
		AesKey          string `mapstructur:"aes_key""`
		NotifyUrl       string `mapstructur:"notify_url"`
	}
}

type databaseConfig struct {
	Type   string          `mapstructure:"type"`
	Master DbConnectOption `mapstructure:"master"`
	Slave  DbConnectOption `mapstructure:"slave"`
}

type DbConnectOption struct {
	DSN         string        `mapstructure:"dsn"`
	MaxOpenConn int           `mapstructure:"maxopen""`
	MaxIdleConn int           `mapstructure:"maxidle"`
	MaxLifeTime time.Duration `mapstructure:"maxlifetime"`
}
