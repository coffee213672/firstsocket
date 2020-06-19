package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// RedisConfig Normal Config
type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	Index        int
	MaxIdleConns int
	MaxConns     int
	WriteTimeout int
	ReadTimeout  int
	IdleTimeout  int
	Wait         bool
}

// Config 現在只有redis config
type Config struct {
	Redis *RedisConfig
}

// Env Global config model
var Env Config

// InitConfig 載入設定
func InitConfig() {
	// Redis Default set
	{
		viper.SetDefault("redis.host", "localhost")
		viper.SetDefault("redis.port", "6379")
		viper.SetDefault("redis.password", "123456")
		viper.SetDefault("redis.index", "0")
		viper.SetDefault("redis.maxidleconns", "10")
		viper.SetDefault("redis.maxconns", "10")
		viper.SetDefault("redis.wait", "true")
		viper.SetDefault("redis.writetimeout", "100")
		viper.SetDefault("redis.readtimeout", "1000")
		viper.SetDefault("redis.idletimeout", "300000")
	}

	// Load config.yml
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	if err := viper.Unmarshal(&Env); err != nil {
		panic(err)
	}

}
