// Package cache contains functions to manage the cache library
package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	// ErrKeyNotExists the key does not exist in the cache
	ErrKeyNotExists = redis.Nil
	// ErrClientIsNoClusterClient means than the client is no a cluster client
	ErrClientIsNoClusterClient = errors.New("redis client is no a cluster client")
)

// scanClientFn is the type for scan functions than receive a redis client as argument. Used for mock tests
type scanClientFn func(ctx context.Context, c RedisClientInterface, cursor uint64, match string, count int64) ([]string, uint64, error)

// hasForEachNodeFunc defines all implementations with for each node function
type hasForEachNodeFunc interface {
	ForEachNode(fn func(client *redis.Client) error) error
}

// Add adds an object to the cache and sets an expiration, if expiration <= 0 is given it's automatically set to 1 hour
func Add(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if expiration == 0 {
		expiration = 1 * time.Hour
	}

	return redisClient.Set(ctx, key, value, expiration).Err()
}

// AddOnce adds an object only once to the cache and sets an expiration, if expiration <= 0 is given it's automatically set to 1 hour
// If the object already exists it returns false, meaning the object was not set
func AddOnce(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if expiration == 0 {
		expiration = 1 * time.Hour
	}

	return redisClient.SetNX(ctx, key, value, expiration).Result()
}

// IsClusterClient returns true if the current client is cluster or not
func IsClusterClient(ctx context.Context) bool {
	_, isCluster := redisClient.(hasForEachNodeFunc) // only redis cluster clients has the method ForEachNode

	return isCluster
}

// Contains returns true if the cache contains the given key
func Contains(ctx context.Context, key string) (bool, error) {
	ret, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return ret == 1, nil
}

// HDel delete fields in cache map
func HDel(ctx context.Context, key string, fields ...string) error {
	_, err := redisClient.HDel(ctx, key, fields...).Result()
	return err
}

// Exists test if a given key exists in cache
func Exists(ctx context.Context, key string) (bool, error) {
	res, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return res == 1, nil
}

// HGet gets the value of a key in a map
func HGet(ctx context.Context, key, field string) (string, error) {
	keyValue, err := redisClient.HGet(ctx, key, field).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrKeyNotExists
	}

	if err != nil {
		return "", err
	}

	return keyValue, nil
}

// HGetAll gets the value of a all keys in a map
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	values, err := redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	return values, nil
}

// HSet sets the value of a key in a map
func HSet(ctx context.Context, key, field, value string) error {
	_, err := redisClient.HSet(ctx, key, field, value).Result()
	return err
}

// Get the value of key. If the key does not exist the cache.ErrKeyNotExists
// is returned. An error is returned if the value stored at key is not a
// string, because GET only handles string values.
func Get(ctx context.Context, key string) (string, error) {
	keyValue, err := redisClient.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrKeyNotExists
	}

	if err != nil {
		return "", err
	}

	return keyValue, nil
}

// Incr returns the value of the counter in cache increased once
func Incr(ctx context.Context, key string) (int64, error) {
	counter, err := redisClient.Incr(ctx, key).Result()
	if err != nil {
		return int64(0), err
	}

	return counter, nil
}

// IncrBy returns the value of the counter in cache increased by the input value
func IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	counter, err := redisClient.IncrBy(ctx, key, value).Result()
	if err != nil {
		return int64(0), err
	}

	return counter, nil
}

// Set set the value and return the status of the execution
func Set(ctx context.Context, key string, value interface{}) error {
	cmd := redisClient.Set(ctx, key, value, time.Second)
	_, err := cmd.Result()

	return err
}

// Del drops multiple keys from redis
func Del(ctx context.Context, keys ...string) error {
	cmd := redisClient.Del(ctx, keys...)

	return cmd.Err()
}

func scan(ctx context.Context, c RedisClientInterface, cursor uint64, match string, count int64) ([]string, uint64, error) {
	cmd := c.Scan(ctx, cursor, match, count)
	vals, cursor, err := cmd.Result()

	return vals, cursor, err
}

func scanAll(ctx context.Context, scan scanClientFn, c RedisClientInterface, match string) ([]string, error) {
	var cursor uint64
	var allKeys []string

	for {
		var keys []string
		var err error

		keys, cursor, err = scan(ctx, c, cursor, match, 10)
		if err != nil {
			return nil, err
		}

		allKeys = append(allKeys, keys...)

		if cursor == 0 {
			break
		}
	}

	return allKeys, nil
}

// GetPipeliner handler redis pipeline
func GetPipeliner() redis.Pipeliner {
	return redisClient.Pipeline()
}

// HIncrBy increase a field with a given value
func HIncrBy(ctx context.Context, key, field string, value int64) (int64, error) {
	return redisClient.HIncrBy(ctx, key, field, value).Result()
}

// Scan functions find the keys by match. If the client is a cluster, all keys of all cluster are returned,
// and the cursor will be set to 0, this means than all has been scaned
func Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	_, isCluster := redisClient.(hasForEachNodeFunc) // only redis cluster clients has the metod ForEachNode
	if isCluster {
		keys, err := scanCluster(ctx, scan, cursor, match, count)
		return keys, 0, err
	}

	return scan(ctx, redisClient, cursor, match, count)
}

// ForEachNode execute a given fuction for each cluster client node, if client is not a cluster, will be fail
func ForEachNode(ctx context.Context, f func(c *redis.Client) error) error {
	clusterClient, ok := redisClient.(hasForEachNodeFunc) // only redis cluster clients has the method ForEachNode
	if !ok {
		return ErrClientIsNoClusterClient
	}

	return clusterClient.ForEachNode(f)
}

// Persist removes the timeout from a key
func Persist(ctx context.Context, key string) (bool, error) {
	return redisClient.Persist(ctx, key).Result()
}

func scanCluster(ctx context.Context, scan scanClientFn, cursor uint64, match string, count int64) ([]string, error) {
	var uniqueKeys sync.Map // used for remove duplicated keys
	var keys []string

	clusterClient, ok := redisClient.(hasForEachNodeFunc) // only redis cluster clients has the method ForEachNode
	if !ok {
		return []string{}, ErrClientIsNoClusterClient
	}

	err := clusterClient.ForEachNode(func(c *redis.Client) error {
		clusterKeys, err := scanAll(ctx, scan, c, match)
		if err != nil {
			return err
		}

		for _, key := range clusterKeys {
			uniqueKeys.Store(key, true) // remove duplicated info
		}

		return nil
	})

	uniqueKeys.Range(func(key interface{}, value interface{}) bool {
		keys = append(keys, fmt.Sprintf("%v", key))
		return true
	})

	return keys, err
}

// Decr returns the value of the counter in cache decreated once
func Decr(ctx context.Context, key string) (int64, error) {
	counter, err := redisClient.Decr(ctx, key).Result()
	if err != nil {
		return int64(0), err
	}

	return counter, nil
}

// Expire is for explicitly expiring keys
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	_, err := redisClient.Expire(ctx, key, expiration).Result()
	return err
}

// Eval runs the specified Lua script in Redis
func Eval(ctx context.Context, script string, keys []string, args ...interface{}) error {
	_, err := redisClient.Eval(ctx, script, keys, args).Result()

	return err
}

// TTL returns the remaining time to live of a key that has a timeout
func TTL(ctx context.Context, key string) (time.Duration, error) {
	remainingTTL, err := redisClient.TTL(ctx, key).Result()
	if err != nil {
		return time.Duration(0), err
	}

	if remainingTTL < 0 {
		return time.Duration(0), ErrKeyNotExists
	}

	return remainingTTL, nil
}

// Type returns the type of the value stored at key in form of a string
func Type(ctx context.Context, key string) (string, error) {
	return redisClient.Type(ctx, key).Result()
}
