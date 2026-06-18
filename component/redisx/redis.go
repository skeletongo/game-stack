package redisx

import (
	"sync"

	"github.com/dobyte/due/v2/etc"
	"github.com/dobyte/due/v2/log"
	"github.com/redis/go-redis/v9"
)

var (
	once     sync.Once
	instance Redis
	_config  = new(Config)
)

type (
	Redis  = redis.UniversalClient
	Script = redis.Script
)

type Config struct {
	Addrs      []string `json:"addrs"`
	DB         int      `json:"db"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	MaxRetries int      `json:"maxRetries"`
	Prefix     string   `json:"prefix"`
}

// Instance 获取单例
func Instance() Redis {
	once.Do(func() {
		instance = NewInstance("etc.redis.default")
		if err := etc.Get("etc.redis.default").Scan(_config); err != nil {
			log.Fatalf("load redis config failed: %v", err)
		}
	})

	return instance
}

// GetConfig 获取配置
func GetConfig() *Config {
	if instance == nil {
		Instance()
	}
	return _config
}

// GetPrefix 获取默认前缀
func GetPrefix() string {
	return GetConfig().Prefix
}

// NewInstance 新建实例
func NewInstance[T string | Config | *Config](config T) Redis {
	var (
		v    any = config
		conf     = new(Config)
	)

	switch c := v.(type) {
	case string:
		if err := etc.Get(c).Scan(conf); err != nil {
			log.Fatalf("load redis config failed: %v", err)
		}
	case Config:
		conf = &c
	case *Config:
		conf = c
	}

	cli := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:      conf.Addrs,
		DB:         conf.DB,
		Username:   conf.Username,
		Password:   conf.Password,
		MaxRetries: conf.MaxRetries,
	})

	return cli
}
