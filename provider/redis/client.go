package redis

import (
	"github.com/protomesh/go-app"
	"github.com/redis/go-redis/v9"
)

type RedisDependency interface {
}

type RedisClient[D RedisDependency] struct {
	Host   app.Config `config:"host,str" default:"localhost:6379" usage:"Redis host"`
	Client redis.UniversalClient
}

func (rc *RedisClient[D]) Initialize() {

	redisHost := rc.Host.StringVal()

	if len(redisHost) > 0 {
		rc.Client = redis.NewClient(&redis.Options{Addr: redisHost})
	}
}
