package redis

import (
	"fmt"
	"github.com/baobao233/gorder/common/handler/factory"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"time"
)

const (
	confName      = "redis"
	localSupplier = "local"
)

var (
	singleton = factory.NewSingleton(supplier)
)

func Init() {
	conf := viper.GetStringMap(confName)
	for supplierName := range conf {
		Client(supplierName)
	}
}

// LocalClient 因为我们只有一个 Client，因此为了避免每次都需要写 string，直接实例化一个 LocalClient
func LocalClient() *redis.Client {
	return Client(localSupplier)
}

func Client(name string) *redis.Client {
	return singleton.Get(name).(*redis.Client)
}

func supplier(key string) any {
	keyName := confName + "." + key // eg. key 是 global.yaml 中的 redis 下的 local
	// 定义一个和 conf 一样的 struct，通过 viper 把配置 marshall 到里面去
	type Section struct {
		IP           string        `mapstructure:"ip"`
		Port         string        `mapstructure:"port"`
		PoolSize     int           `mapstructure:"pool_size"`
		MaxConn      int           `mapstructure:"max_conn"`
		ConnTimeout  time.Duration `mapstructure:"conn_timeout"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
	}

	var c Section
	if err := viper.UnmarshalKey(keyName, &c); err != nil {
		panic(err)
	}

	return redis.NewClient(&redis.Options{
		Network:         "tcp",
		Addr:            fmt.Sprintf("%s:%s", c.IP, c.Port),
		PoolSize:        c.PoolSize,
		MaxActiveConns:  c.MaxConn,
		ConnMaxLifetime: c.ConnTimeout * time.Millisecond,
		ReadTimeout:     c.ReadTimeout * time.Millisecond,
		WriteTimeout:    c.WriteTimeout * time.Millisecond,
	})
}
