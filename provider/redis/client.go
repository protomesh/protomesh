package redis

import (
	"fmt"
	"log"
	"net"

	"github.com/protomesh/go-app"
	"github.com/redis/go-redis/v9"
)

type RedisDependency interface {
}

type RedisClient[D RedisDependency] struct {
	*app.Injector[D]

	Host app.Config `config:"host,str" default:"localhost:6379" usage:"Redis host"`

	UseSrvRecords app.Config `config:"use.srv.records,bool" default:"false" usage:"Use SRV records to lookup Redis hosts"`

	Client redis.UniversalClient
}

func (rc *RedisClient[D]) Initialize() {

	redisHost := rc.Host.StringVal()

	if len(redisHost) == 0 {
		return
	}

	addrs := []string{}

	if rc.UseSrvRecords.BoolVal() {

		records, err := rc.lookupSrvRecords(redisHost)
		if err != nil {
			log.Panic("Failed to lookup SRV records", "error", err)
		}

		addrs = append(addrs, records...)

	} else {

		addrs = append(addrs, redisHost)

	}

	rc.Client = redis.NewUniversalClient(&redis.UniversalOptions{Addrs: addrs})
}

func (rc *RedisClient[D]) lookupSrvRecords(record string) ([]string, error) {

	log := rc.Log().With("driver", "redis")

	cname, srvs, err := net.LookupSRV("", "", record)
	if err != nil {
		log.Error("Failed to lookup SRV records", "error", err)
		return nil, err
	}

	log.Debug("SRV records", "cname", cname, "srvs", srvs)

	hosts := []string{}

	for _, srv := range srvs {
		log.Debug("SRV record", "srv", srv)
		hosts = append(hosts, fmt.Sprintf("%s:%d", srv.Target, srv.Port))
	}

	return hosts, nil

}
