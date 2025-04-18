package cache

import (
	"encoding/gob"

	"github.com/gin-gonic/gin"
	"github.com/wangyi1310/mycloud-disk/conf"
	"github.com/wangyi1310/mycloud-disk/pkg/log"
)

func init() {
	gob.Register(map[string]itemWithTTL{})
}

// Store 缓存存储器
var Store Driver = NewMemoStore()

// Init 初始化缓存
func Init() {
	if conf.RedisConfig.Server != "" && gin.Mode() != gin.TestMode {
		Store = NewRedisStore(
			10,
			conf.RedisConfig.Network,
			conf.RedisConfig.Server,
			conf.RedisConfig.User,
			conf.RedisConfig.Password,
			conf.RedisConfig.DB,
		)
	}
}

// Restore restores cache from given disk file
func Restore(persistFile string) {
	if err := Store.Restore(persistFile); err != nil {
		log.Log().Warning("Failed to restore cache from disk: %s", err)
	}
}

func InitSlaveOverwrites() {
	err := Store.Sets(conf.OptionOverwrite, "setting_")
	if err != nil {
		log.Log().Warning("Failed to overwrite database setting: %s", err)
	}
}

// Driver 键值缓存存储容器
type Driver interface {
	// 设置值，ttl为过期时间，单位为秒
	Set(key string, value interface{}, ttl int) error

	// 取值，并返回是否成功
	Get(key string) (interface{}, bool)

	// 批量取值，返回成功取值的map即不存在的值
	Gets(keys []string, prefix string) (map[string]interface{}, []string)

	// 批量设置值，所有的key都会加上prefix前缀
	Sets(values map[string]interface{}, prefix string) error

	// 删除值
	Delete(keys []string, prefix string) error

	// Save in-memory cache to disk
	Persist(path string) error

	// Restore cache from disk
	Restore(path string) error
}

// Set 设置缓存值
func Set(key string, value interface{}, ttl int) error {
	return Store.Set(key, value, ttl)
}

// Get 获取缓存值
func Get(key string) (interface{}, bool) {
	return Store.Get(key)
}

// Deletes 删除值
func Deletes(keys []string, prefix string) error {
	return Store.Delete(keys, prefix)
}

// GetSettings 根据名称批量获取设置项缓存
func GetSettings(keys []string, prefix string) (map[string]string, []string) {
	raw, miss := Store.Gets(keys, prefix)

	res := make(map[string]string, len(raw))
	for k, v := range raw {
		res[k] = v.(string)
	}

	return res, miss
}

// SetSettings 批量设置站点设置缓存
func SetSettings(values map[string]string, prefix string) error {
	var toBeSet = make(map[string]interface{}, len(values))
	for key, value := range values {
		toBeSet[key] = interface{}(value)
	}
	return Store.Sets(toBeSet, prefix)
}
