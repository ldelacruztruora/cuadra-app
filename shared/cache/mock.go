package cache

import (
	"context"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
)

var (
	redisMock = &mockClient{}

	redisMockCluster = &mockClientCluster{}

	// MockServer Redis mock server
	MockServer *miniredis.Miniredis

	// ForceError is forced error response
	ForceError error

	// ForEachNodeError forces cluster client foreach function to return an error
	ForEachNodeError error

	// ScanSpectedKeys expected keys for the function Scan
	ScanSpectedKeys = []string{
		"key1",
		"key2",
	}

	// GetAllMap expected map for the function HGetAll
	GetAllMap = map[string]string{}

	// GetValue expected string for the function Get
	GetValue = "dummy"
)

func mockScanClientFn(ctx context.Context, c RedisClientInterface, cursor uint64, match string, count int64) ([]string, uint64, error) {
	if ForceError != nil {
		return nil, 0, ForceError
	}

	return ScanSpectedKeys, 0, nil
}

type mockClient struct {
	RedisClientInterface
	shards []string
}

// Ping mock client response
func (c *mockClient) Ping(ctx context.Context) *redis.StatusCmd {
	return &redis.StatusCmd{}
}

// Set mock client response
func (c *mockClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return &redis.StatusCmd{}
}

// SetNX mock client response
func (c *mockClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	_ = c.Set(ctx, key, value, expiration)
	return redis.NewBoolResult(true, ForceError)
}

type mockClientCluster struct {
	RedisClientInterface
	// force to ForEachNode to call the given function, if this is true, take
	// in account than the client sended to ForEachNode function will be nil
	// @see ForEachNode implementation
	forceCallForEachFn bool
}

type mockPipeliner struct {
	redis.Pipeliner
}

func (c *mockPipeliner) Incr(ctx context.Context, key string) *redis.IntCmd {
	return redis.NewIntResult(1, ForceError)
}

// Close mock client close
func (c *mockPipeliner) Close() error {
	return nil
}

func (c *mockPipeliner) Decr(ctx context.Context, key string) *redis.IntCmd {
	return redis.NewIntResult(1, ForceError)
}

func (c *mockPipeliner) Exec(ctx context.Context) ([]redis.Cmder, error) {
	return nil, ForceError
}

// WithError sets an error to be returned during the mock operations
func (c *mockClientCluster) WithError(err error, cb func()) {
	ForceError = err

	defer func() {
		ForceError = nil
	}()

	cb()
}

// HIncrBy mock client response
func (c *mockClientCluster) HIncrBy(ctx context.Context, key string, field string, value int64) *redis.IntCmd {
	if ForceError != nil {
		return redis.NewIntResult(1, ForceError)
	}

	return redis.NewIntResult(0, nil)
}

// HSet mock client response
func (c *mockClientCluster) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	if ForceError != nil {
		return redis.NewIntResult(1, ForceError)
	}

	return redis.NewIntResult(0, nil)
}

// HDel mock client response
func (c *mockClientCluster) HDel(ctx context.Context, key string, values ...string) *redis.IntCmd {
	if ForceError != nil {
		return redis.NewIntResult(1, ForceError)
	}

	return redis.NewIntResult(0, nil)
}

// Get mock client response
func (c *mockClientCluster) Get(ctx context.Context, key string) *redis.StringCmd {
	if ForceError != nil {
		return redis.NewStringResult("", ForceError)
	}

	return redis.NewStringResult(GetValue, nil)
}

// HGet mock client response
func (c *mockClientCluster) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	if ForceError != nil {
		return redis.NewStringStringMapResult(nil, ForceError)
	}

	return redis.NewStringStringMapResult(GetAllMap, nil)
}

// HGet mock client response
func (c *mockClientCluster) HGet(ctx context.Context, key string, field string) *redis.StringCmd {
	if ForceError != nil {
		stringCmd := redis.NewStringCmd(ctx, ForceError)

		stringCmd.SetErr(ForceError)

		return stringCmd
	}

	return redis.NewStringCmd(ctx, "dummy")
}

// Expire mock client response
func (c *mockClientCluster) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	if ForceError != nil {
		return redis.NewBoolResult(false, ForceError)
	}

	return redis.NewBoolResult(true, nil)
}

func (c *mockClient) Pipeline() redis.Pipeliner {
	return &mockPipeliner{}
}

// WithError sets an error to be returned during the mock operations
func (c *mockClientCluster) WithForEachNodeError(err error, cb func()) {
	ForEachNodeError = err

	defer func() {
		ForEachNodeError = nil
	}()

	cb()
}

func (c *mockClientCluster) ForEachNode(fn func(client *redis.Client) error) error {
	if c.forceCallForEachFn {
		// nil client is sended, take in account in your tests if enable force fn call
		if err := fn(&redis.Client{}); err != nil {
			return err
		}
	}

	return ForEachNodeError
}

func newMockClientCluster(shardAddrs []string, cluster bool) RedisClientInterface {
	return &mockClientCluster{}
}

func newMockPipeliner(shardAddrs []string, cluster bool) RedisClientInterface {
	return &mockClient{}
}

// Ping mock client cluster response
func (c *mockClientCluster) Ping(ctx context.Context) *redis.StatusCmd {
	return &redis.StatusCmd{}
}

// InitMock initializes mock client for cache
func InitMock() {
	newClientFunc = newRedisClient

	var err error

	MockServer, err = miniredis.Run()
	if err != nil {
		panic(err)
	}

	_ = Init([]string{MockServer.Addr()}, false)
}

// InitMockPipeliner initializes pipeliner mock client for cache
func InitMockPipeliner() {
	newClientFunc = newMockPipeliner

	_ = Init([]string{""}, false)
}

// InitMockWithoutServer initializes mock client for cache without server
func InitMockWithoutServer() {
	newClientFunc = newRedisClient

	_ = Init([]string{"notserver:6379"}, false)
}

// InitClusterMock initializes cluster mock client for cache
func InitClusterMock() {
	ForceError = nil
	newClientFunc = newMockClientCluster
	_ = Init([]string{}, true)
}
