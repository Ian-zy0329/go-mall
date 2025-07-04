package config

import (
	"bytes"
	"embed"
	"github.com/spf13/viper"
	"os"
)

//go:embed *.yaml
var configs embed.FS

func init() {
	env := os.Getenv("env")
	redisStockAddr := os.Getenv("REDIS_STOCK_ADDR")
	vp := viper.New()
	configFileStream, err := configs.ReadFile("application." + env + ".yaml")
	if err != nil {
		panic(err)
	}
	vp.SetConfigType("yaml")
	err = vp.ReadConfig(bytes.NewReader(configFileStream))
	if err != nil {
		panic(err)
	}
	vp.UnmarshalKey("app", &App)
	vp.UnmarshalKey("database", &Database)
	vp.UnmarshalKey("redis", &Redis)
	vp.UnmarshalKey("redis_stock_service", &RedisStockServiceConfig)
	RedisStockServiceConfig.Addr = redisStockAddr
}
