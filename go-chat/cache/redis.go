package cache

import (
	"go-chat/config"

	"github.com/go-redis/redis"
)

// Redis Global
var Redis *redis.Client

// InitCache init ...
func InitCache() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     config.Env.Redis.Host + ":" + config.Env.Redis.Port,
		Password: config.Env.Redis.Password,
		DB:       0, // use default DB
	})
}

// GetOnlineUser 取得user name
func GetOnlineUser(key, field string) string {
	strcmd := Redis.HGet(key, field)
	return strcmd.Val()
}

// SetOnlineUser 存user name
func SetOnlineUser(key, field, value string) bool {
	boolcmd := Redis.HSet(key, field, value)
	return boolcmd.Val()
}

// GetAllOnlineUser all
func GetAllOnlineUser(key string) map[string]string {
	rAll := Redis.HGetAll(key)
	return rAll.Val()
}

// DelOnlineUser 下線刪除
func DelOnlineUser(key, field string) {
	Redis.HDel(key, field)
	return
}
