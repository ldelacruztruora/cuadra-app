package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func BenchmarkScanSuccess(b *testing.B) {
	c := require.New(b)

	InitMock()

	for i := 0; i < b.N; i++ {
		_, _, err := Scan(context.Background(), 0, "*", 10)
		c.NoError(err)
	}
}

func TestHDel(t *testing.T) {
	c := require.New(t)

	InitMock()

	key := "key"
	field := "field"
	value := "value"
	_, err := redisClient.HSet(context.Background(), key, field, value).Result()
	c.NoError(err)

	err = HDel(context.Background(), key, field)
	c.NoError(err)

	_, err = redisClient.HGet(context.Background(), key, field).Result()
	c.True(errors.Is(err, redis.Nil)) // key should be removed
}

func TestHGet_UnexpectedError(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	_, err := HGet(context.Background(), "key", "field")
	c.Error(err)
}

func TestHGet_KeyNotExists(t *testing.T) {
	c := require.New(t)

	InitMock()

	_, err := HGet(context.Background(), "key", "field")
	c.Equal(ErrKeyNotExists, err)
}

func TestHGet(t *testing.T) {
	c := require.New(t)

	InitMock()

	key := "key"
	value := "value"
	field := "field"
	_, err := redisClient.HSet(context.Background(), key, field, value).Result()
	c.NoError(err)

	val, err := HGet(context.Background(), key, field)
	c.NoError(err)
	c.Equal(val, value)
}

func TestHGetWithMock(t *testing.T) {
	c := require.New(t)

	InitClusterMock()

	key := "key"
	field := "field"

	_, err := HGet(context.Background(), key, field)
	c.NoError(err)

	ForceError = errors.New("some-error")

	defer func() {
		ForceError = nil
	}()

	_, err = HGet(context.Background(), key, field)
	c.NotNil(err)
	c.Equal("some-error", err.Error())
}

func TestHDel_OK(t *testing.T) {
	c := require.New(t)

	InitClusterMock()

	err := HDel(context.Background(), "key", "field")
	c.NoError(err)
}

func TestHDel_Error(t *testing.T) {
	c := require.New(t)

	InitClusterMock()

	mockErr := errors.New("force error")
	ForceError = mockErr

	defer func() {
		ForceError = nil
	}()

	err := HDel(context.Background(), "key", "field")
	c.ErrorIs(err, mockErr)
}

func TestHSet_Error(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	err := HSet(context.Background(), "key", "field", "value")
	c.Error(err)
}

func TestHSet(t *testing.T) {
	c := require.New(t)

	InitMock()

	key := "key"
	value := "value"
	field := "field"
	err := HSet(context.Background(), key, field, value)
	c.NoError(err)

	val, err := redisClient.HGet(context.Background(), key, field).Result()
	c.NoError(err)
	c.Equal(value, val)
}

func TestHSetWithMock(t *testing.T) {
	c := require.New(t)

	InitClusterMock()

	key := "key"
	value := "value"
	field := "field"
	err := HSet(context.Background(), key, field, value)
	c.NoError(err)

	ForceError = errors.New("some")

	defer func() { ForceError = nil }()

	err = HSet(context.Background(), key, field, value)
	c.NotNil(err)
	c.Equal("some", err.Error())
}

func TestHIncrBy(t *testing.T) {
	c := require.New(t)

	InitMock()

	key := "key"
	field := "field"
	afterValue, err := HIncrBy(context.Background(), key, field, 1)
	c.Equal(int64(1), afterValue)
	c.NoError(err)

	afterValue, err = HIncrBy(context.Background(), key, field, 1)
	c.Equal(int64(2), afterValue)
	c.NoError(err)

	val, err := HGet(context.Background(), key, field)
	c.NoError(err)
	c.Equal("2", val)
}

func TestHIncrByError(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	key := "key"
	field := "field"
	_, err := HIncrBy(context.Background(), key, field, 1)
	c.Error(err)
}

func TestScan(t *testing.T) {
	c := require.New(t)

	InitMock()

	c.NoError(Set(context.Background(), "key1", 0))
	c.NoError(Set(context.Background(), "key2", 0))

	keys, _, err := Scan(context.Background(), 0, "*", 10)
	c.Nil(err)
	c.Equal(ScanSpectedKeys, keys)
}

func TestDel_Ok(t *testing.T) {
	c := require.New(t)

	delKeys := []string{"key1", "key2"}

	InitMock()
	c.NoError(Set(context.Background(), "key1", 0))
	c.NoError(Set(context.Background(), "key2", 0))

	err := Del(context.Background(), delKeys...)
	c.Nil(err)

	ErrKeyNotFound := errors.New("ERR no such key")
	_, err = MockServer.Get("key1")
	c.Equal(ErrKeyNotFound, err)
	_, err = MockServer.Get("key2")
	c.Equal(ErrKeyNotFound, err)
}

func TestForEachNode_NonClusterClient(t *testing.T) {
	c := require.New(t)

	InitMock() // init non cluster client

	f := func(cli *redis.Client) error {
		return nil
	}

	err := ForEachNode(context.Background(), f)
	c.Equal(ErrClientIsNoClusterClient, err)
}

func TestForEachNode(t *testing.T) {
	c := require.New(t)

	InitClusterMock()

	f := func(cli *redis.Client) error {
		return nil
	}

	err := ForEachNode(context.Background(), f)
	c.Nil(err)
}

func TestPersists(t *testing.T) {
	c := require.New(t)

	InitMock()

	ctx := context.Background()
	err := Set(ctx, "some", "value") // set a value with 1 second of expiration
	c.Nil(err)

	val, err := Get(ctx, "some")
	c.Nil(err)
	c.Equal("value", val)

	_, err = Persist(ctx, "some")
	c.Nil(err)

	time.AfterFunc(2*time.Second, func() { c.Equal("value", val) }) // after 2 seconds the value should exists
}

func TestIsCluster(t *testing.T) {
	c := require.New(t)

	InitMock() // init non cluster client

	ctx := context.Background()

	isCluster := IsClusterClient(ctx)
	c.False(isCluster)

	InitClusterMock()

	isCluster = IsClusterClient(ctx)
	c.True(isCluster)
}

func TestDel_Err(t *testing.T) {
	c := require.New(t)

	delKeys := []string{"key1", "key2"}

	InitMockWithoutServer()

	err := Del(context.Background(), delKeys...)
	c.NotNil(err)
}

func TestScanError(t *testing.T) {
	c := require.New(t)

	InitClusterMock()

	redisMockCluster.WithForEachNodeError(errors.New("dummy"), func() {
		_, _, err := Scan(context.Background(), 0, "*", 10)
		c.NotNil(err)
	})
}

func TestSet_MiniRedis(t *testing.T) {
	c := require.New(t)

	InitMock()

	err := Set(context.Background(), "some", 0)
	c.Nil(err)
	val, err := Get(context.Background(), "some")
	c.Nil(err)

	c.Equal("0", val)
}

func TestScanCluster_InvalidClient(t *testing.T) {
	c := require.New(t)

	InitMock() // this mock is not a cluster, scan cluster will be fail

	redisMockCluster.WithError(errors.New("dummy"), func() {
		_, err := scanCluster(context.Background(), mockScanClientFn, 0, "*", 10)
		c.Equal(ErrClientIsNoClusterClient, err)
	})
}

func TestScanCluster_HIncrByMock(t *testing.T) {
	c := require.New(t)

	InitClusterMock()

	cmd := redisMockCluster.HIncrBy(context.Background(), "unknown", "dummy", 1)

	_, err := cmd.Result()
	c.NoError(err)

	ForceError = errors.New("some")

	defer func() { ForceError = nil }()

	cmd = redisMockCluster.HIncrBy(context.Background(), "unknown", "dummy", 1)

	_, err = cmd.Result()
	c.Equal(ForceError, err)
}

func TestScanCluster_GetMock(t *testing.T) {
	c := require.New(t)

	InitClusterMock()

	cmd := redisMockCluster.Get(context.Background(), "unknown")

	value, err := cmd.Result()
	c.NoError(err)
	c.Equal(GetValue, value)

	ForceError = errors.New("some")

	defer func() { ForceError = nil }()

	cmd = redisMockCluster.Get(context.Background(), "unknown")

	_, err = cmd.Result()
	c.Equal(ForceError, err)
}

func TestScanCluster_HGetAllMock(t *testing.T) {
	c := require.New(t)

	InitClusterMock()

	cmd := redisMockCluster.HGetAll(context.Background(), "unknown")

	_, err := cmd.Result()
	c.NoError(err)

	ForceError = errors.New("some")

	defer func() { ForceError = nil }()

	cmd = redisMockCluster.HGetAll(context.Background(), "unknown")

	_, err = cmd.Result()
	c.Equal(ForceError, err)
}

func TestScanCluster_OK(t *testing.T) {
	c := require.New(t)

	mock := &mockClientCluster{forceCallForEachFn: true} // force to call mock scan client fn
	redisClient = mock

	keys, err := scanCluster(context.Background(), mockScanClientFn, 0, "*", 0)
	c.Nil(err)
	c.ElementsMatch(ScanSpectedKeys, keys) // returned keys by scan are filtred using a map, can be in distinct order
}

func TestScanCluster_Dedup(t *testing.T) {
	c := require.New(t)

	mock := &mockClientCluster{forceCallForEachFn: true} // force to call mock scan client fn
	redisClient = mock

	prevKeys := ScanSpectedKeys

	defer func() { ScanSpectedKeys = prevKeys }()

	ScanSpectedKeys = []string{"key1", "key1", "key2", "key2", "key2"}

	keys, err := scanCluster(context.Background(), mockScanClientFn, 0, "*", 0)
	c.Nil(err)
	c.Len(keys, 2)
	c.ElementsMatch(keys, []string{"key1", "key2"}) // returned keys should be 2, dedup
}

func TestScanCluster_NodeError(t *testing.T) {
	c := require.New(t)

	clusterMock := &mockClientCluster{forceCallForEachFn: true}
	redisClient = clusterMock

	ForceError = errors.New("dummy")

	keys, err := scanCluster(context.Background(), mockScanClientFn, 0, "*", 0)
	c.Empty(keys)
	c.NotNil(err)
}

func TestScanCluster_PublicFunc(t *testing.T) {
	c := require.New(t)

	InitClusterMock()

	keys, _, err := Scan(context.Background(), 0, "*", 0)
	c.Nil(err)
	c.Empty(keys)
}

func TestScanCluster_Error(t *testing.T) {
	c := require.New(t)

	clusterMock := &mockClientCluster{forceCallForEachFn: true}
	redisClient = clusterMock

	initialErr := errors.New("dummy")
	redisMockCluster.WithError(initialErr, func() {
		keys, err := scanCluster(context.Background(), mockScanClientFn, 0, "*", 0)
		c.Equal(initialErr, err)
		c.Empty(keys)
	})
}

func TestAdd(t *testing.T) {
	c := require.New(t)

	InitMock()
	c.NoError(Add(context.Background(), "TestAddkey", "value", 1*time.Minute))

	value, err := MockServer.Get("TestAddkey")
	c.NoError(err)
	c.Equal("value", value)
}

func TestAddNoExpiration(t *testing.T) {
	c := require.New(t)

	InitMock()
	c.NoError(Add(context.Background(), "TestAddNoExpirationkey", "value", 0))

	value, err := MockServer.Get("TestAddNoExpirationkey")
	c.NoError(err)
	c.Equal("value", value)
}

func TestAddOnce(t *testing.T) {
	c := require.New(t)

	InitMock()

	added, err := AddOnce(context.Background(), "keyTestAddOnce", "keyTestAddOnceValue", 1*time.Minute)
	c.NoError(err)

	c.True(added)

	notAdded, err := AddOnce(context.Background(), "keyTestAddOnce", "keyTestAddOnceValue", 1*time.Minute)
	c.NoError(err)
	c.False(notAdded)

	value, err := MockServer.Get("keyTestAddOnce")
	c.NoError(err)
	c.Equal("keyTestAddOnceValue", value)
}

func TestAddOnceNoExpiration(t *testing.T) {
	c := require.New(t)

	InitMock()

	added, err := AddOnce(context.Background(), "keyTestAddOnceNoExpiration", "value", 0)
	c.True(added)
	c.NoError(err)
	added, err = AddOnce(context.Background(), "keyTestAddOnceNoExpiration", "value", 0)
	c.False(added)
	c.NoError(err)
	value, err := MockServer.Get("keyTestAddOnceNoExpiration")
	c.NoError(err)
	c.Equal("value", value)
}

func TestContains(t *testing.T) {
	c := require.New(t)

	InitMock()
	c.NoError(Set(context.Background(), "keyTestContains", 0))
	c.True(Contains(context.Background(), "keyTestContains"))

	value, err := MockServer.Get("keyTestContains")
	c.NoError(err)
	c.Equal("0", value)
}

func TestContainsWithError(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	result, err := Contains(context.Background(), "key")
	c.False(result)
	c.NotNil(err)
}

func TestGet(t *testing.T) {
	c := require.New(t)

	InitMock()

	c.NoError(Set(context.Background(), "TestGetkey", "value"))
	value, err := Get(context.Background(), "TestGetkey")

	c.Nil(err)
	c.Equal("value", value)
	value, err = MockServer.Get("TestGetkey")
	c.NoError(err)
	c.Equal("value", value)
}

func TestGetFailed(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	val, err := Get(context.Background(), "TestGetFailedkey")
	c.NotNil(err)
	c.Equal(val, "")
}

func TestGetKeyNotExists(t *testing.T) {
	c := require.New(t)

	InitMock()

	val, err := Get(context.Background(), "TestGetKeyNotExists")
	c.NotNil(err)
	c.Equal("", val)
	c.Equal(ErrKeyNotExists, err)
}

func TestSet(t *testing.T) {
	c := require.New(t)

	InitMock()

	err := Set(context.Background(), "TestSetKey", "val")
	c.Nil(err)
	value, err := MockServer.Get("TestSetKey")
	c.NoError(err)
	c.Equal("val", value)
}

func TestSetFailed(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	err := Set(context.Background(), "key", "val")
	c.NotNil(err)
}

func TestIncr(t *testing.T) {
	c := require.New(t)

	InitMock()

	count, err := Incr(context.Background(), "counter1TestIncr")
	c.Equal(nil, err)
	c.Equal(int64(1), count)
	count, err = Incr(context.Background(), "counter1TestIncr")
	c.Equal(nil, err)
	c.Equal(int64(2), count)
}

func TestIncrBy(t *testing.T) {
	c := require.New(t)

	InitMock()

	count, err := IncrBy(context.Background(), "counter1TestIncr", 3)
	c.Equal(nil, err)
	c.Equal(int64(3), count)
	count, err = IncrBy(context.Background(), "counter1TestIncr", 3)
	c.Equal(nil, err)
	c.Equal(int64(6), count)
}

func TestDecr(t *testing.T) {
	c := require.New(t)

	InitMock()

	count, err := Decr(context.Background(), "counter1TestDecr")
	c.Equal(nil, err)
	c.Equal(int64(-1), count)
	count, err = Decr(context.Background(), "counter1TestDecr")
	c.Equal(nil, err)
	c.Equal(int64(-2), count)
}

func TestIncrWithError(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	count, err := Incr(context.Background(), "key")
	c.NotNil(err)
	c.Equal(count, int64(0))
}

func TestIncrByWithError(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	count, err := IncrBy(context.Background(), "key", 3)
	c.NotNil(err)
	c.Equal(int64(0), count)
}

func TestDecrWithError(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	count, err := Decr(context.Background(), "key")
	c.NotNil(err)
	c.Equal(count, int64(0))
}

func TestGetPipeliner(t *testing.T) {
	c := require.New(t)

	InitMockPipeliner()

	pipeliner := GetPipeliner()

	c.NotNil(pipeliner)
	c.NoError(pipeliner.Close())
}

func TestMockPipeliner(t *testing.T) {
	c := require.New(t)

	InitMockPipeliner()

	pipeliner := GetPipeliner()

	defer func() {
		c.NoError(pipeliner.Close())
	}()

	inc := pipeliner.Incr(context.Background(), "test")
	c.NotNil(inc)

	dec := pipeliner.Decr(context.Background(), "test")
	c.NotNil(dec)

	_, err := pipeliner.Exec(context.Background())
	c.Nil(err)

	set := redisMock.Set(context.Background(), "test", nil, time.Microsecond)
	c.NotNil(set)

	setNX := redisMock.SetNX(context.Background(), "test", nil, time.Microsecond)
	c.NotNil(setNX)
}

func TestHGetAll(t *testing.T) {
	c := require.New(t)

	InitMock()

	vals, err := HGetAll(context.Background(), "non_existing_key")
	c.Empty(vals)
	c.NoError(err)

	mapKey := "map"
	field := "field"
	value := "12"

	field2 := "field2"
	value2 := "13"

	err = HSet(context.Background(), mapKey, field, value)
	c.NoError(err)

	err = HSet(context.Background(), mapKey, field2, value2)
	c.NoError(err)

	vals, err = HGetAll(context.Background(), mapKey)
	c.NoError(err)

	c.EqualValues(map[string]string{
		field:  value,
		field2: value2,
	}, vals)
}

func TestHGetAllServerError(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	_, err := HGetAll(context.Background(), "some_key")
	c.Error(err)
}

func TestMockPipeliner_Error(t *testing.T) {
	c := require.New(t)

	InitMockPipeliner()

	testError := errors.New("test error")
	ForceError = testError

	pipeliner := GetPipeliner()

	defer func() {
		c.NoError(pipeliner.Close())
	}()

	_, err := pipeliner.Incr(context.Background(), "test").Result()
	c.Equal(testError, err)

	_, err = pipeliner.Decr(context.Background(), "test").Result()
	c.Equal(testError, err)

	_, err = pipeliner.Exec(context.Background())
	c.Equal(testError, err)
}

func TestExists(t *testing.T) {
	c := require.New(t)

	InitMock()

	key := "test"
	_, err := redisClient.Set(context.Background(), key, "1", 1*time.Second).Result()
	c.NoError(err)

	exists, err := Exists(context.Background(), key)
	c.NoError(err)
	c.True(exists)
}

func TestExistsFalse(t *testing.T) {
	c := require.New(t)

	InitMock()

	key := "test"

	exists, err := Exists(context.Background(), key)
	c.NoError(err)
	c.False(exists)
}

func TestExistsError(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	key := "test"

	exists, err := Exists(context.Background(), key)
	c.Error(err)
	c.False(exists)
}

func TestExpire(t *testing.T) {
	c := require.New(t)

	InitMock()

	key := "key"

	_, err := redisClient.HSet(context.Background(), key, "foo", "var").Result()
	c.NoError(err)

	_, err = redisClient.Expire(context.Background(), key, 1*time.Hour).Result()
	c.NoError(err)
}

func TestExpireWithMock(t *testing.T) {
	c := require.New(t)

	InitClusterMock()

	key := "key"

	err := Expire(context.Background(), key, 1*time.Hour)
	c.NoError(err)

	ForceError = errors.New("some-error")

	defer func() {
		ForceError = nil
	}()

	err = Expire(context.Background(), key, 1*time.Hour)
	c.Equal("some-error", err.Error())
}

func TestEval(t *testing.T) {
	c := require.New(t)

	InitMock()

	err := Eval(context.Background(), `return redis.call("SET", KEYS[1], ARGV[1])`, []string{"foo"}, "var")
	c.NoError(err)

	val, err := Get(context.Background(), "foo")
	c.NoError(err)

	c.Equal("var", val)
}

func TestTTLKeyNotExist(t *testing.T) {
	c := require.New(t)

	InitMock()

	ttl, err := TTL(context.Background(), "non_existing_key")
	c.Equal(time.Duration(0), ttl)
	c.Equal(ErrKeyNotExists, err)
}

func TestTTL(t *testing.T) {
	c := require.New(t)

	InitMock()

	key := "key"
	value := "12"

	err := Add(context.Background(), key, value, 1*time.Minute)
	c.NoError(err)

	ttl, err := TTL(context.Background(), key)
	c.NoError(err)
	c.Equal(1*time.Minute, ttl)
}

func TestTTLServerError(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	_, err := TTL(context.Background(), "some_key")
	c.Error(err)
}

func TestType(t *testing.T) {
	c := require.New(t)

	InitMock()

	key := "key"
	value := "12"

	err := Add(context.Background(), key, value, 1*time.Minute)
	c.NoError(err)

	typ, err := Type(context.Background(), key)
	c.NoError(err)

	c.Equal("string", typ)

	key = "key2"

	err = HSet(context.Background(), key, "field", value)
	c.NoError(err)

	typ, err = Type(context.Background(), key)
	c.NoError(err)

	c.Equal("hash", typ)

	key = "key3"

	err = AddToUnorderedSet(context.Background(), key, value)
	c.NoError(err)

	typ, err = Type(context.Background(), key)
	c.NoError(err)

	c.Equal("set", typ)
}
