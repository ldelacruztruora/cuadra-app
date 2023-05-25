package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/truora/scrap-services/shared/env"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"
)

var (
	newClientFunc = newRedisClient
	redisClient   redis.Cmdable
	rateLimiter   *redis_rate.Limiter
)

// RedisClientInterface defines the redis client interface
type RedisClientInterface interface {
	redis.Cmdable
}

func newRedisClient(shardAddrs []string, cluster bool) RedisClientInterface {
	if cluster {
		return redis.NewClusterClient(&redis.ClusterOptions{Addrs: shardAddrs})
	}

	addr := ""
	if len(shardAddrs) > 0 {
		addr = shardAddrs[0]
	}

	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

// GetClient returns the redis client
func GetClient() RedisClientInterface {
	return redisClient
}

// Init initializes the cache module
func Init(shardAddrs []string, cluster bool) error {
	redisClient = newClientFunc(shardAddrs, cluster)

	go func(client RedisClientInterface) {
		_, err := client.Ping(context.Background()).Result()
		if err != nil {
			fmt.Printf("Connecting to redis failed: %s\n", err.Error())
		}
	}(redisClient)

	rateLimiter = redis_rate.NewLimiter(redisClient)

	DefaultRate(context.Background(), time.Second, 10)

	return nil
}

// InitFromEnv takes the environment variable TRUORA_CACHE_SHARDS and initializes the clients
// The format of the variable is: "shard1;shard2;shard3"
func InitFromEnv() error {
	shards := env.GetString("TRUORA_CACHE_SHARDS", "truora-cache-0001-001.uubxoz.0001.use1.cache.amazonaws.com:6379;truora-cache-0001-002.uubxoz.0001.use1.cache.amazonaws.com:6379")
	isClusterEnv := env.GetString("TRUORA_CACHE_USE_CLUSTER", "true")
	isCluster := isClusterEnv == "true"

	return Init(strings.Split(shards, ";"), isCluster)
}
