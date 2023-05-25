package cache

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"

	"github.com/stretchr/testify/require"
)

func TestAddToOrderedSetWithOption(t *testing.T) {
	c := require.New(t)
	ctx := context.Background()

	InitMock()

	err := AddToOrderedSetWithOption(ctx, "round", "a", 1, OnlyAdd)
	c.NoError(err)

	err = AddToOrderedSetWithOption(ctx, "round", "b", 2, OnlyAdd)
	c.NoError(err)

	err = AddToOrderedSetWithOption(ctx, "round", "c", 3, OnlyAdd)
	c.NoError(err)

	results, err := redisClient.ZRange(context.Background(), "round", 0, 3).Result()
	c.NoError(err)
	c.Equal("a", results[0])

	err = AddToOrderedSetWithOption(ctx, "round", "c", 1, OnlyAdd)
	c.NoError(err)

	results, err = redisClient.ZRange(context.Background(), "round", 0, 3).Result()
	c.NoError(err)
	c.Equal("a", results[0])

	err = AddToOrderedSetWithOption(ctx, "round", "c", 1, OnlyUpdate)
	c.NoError(err)

	err = AddToOrderedSetWithOption(ctx, "round", "a", 3, OnlyUpdate)
	c.NoError(err)

	results, err = redisClient.ZRange(context.Background(), "round", 0, 3).Result()
	c.NoError(err)
	c.Equal("c", results[0])

	err = AddToOrderedSetWithOption(ctx, "round", "a", 3, OrderedSetOption(""))
	c.Equal(ErrUnsupportedOrderedSetOption, err)
}

func TestRemoveFromToOrderedSet(t *testing.T) {
	c := require.New(t)
	ctx := context.Background()

	InitMock()

	err := RemoveFromToOrderedSet(ctx, "round", "a")
	c.NoError(err)

	err = AddToOrderedSetWithOption(ctx, "round", "a", 1, OnlyAdd)
	c.NoError(err)

	err = RemoveFromToOrderedSet(ctx, "round", "a")
	c.NoError(err)

	results, err := redisClient.ZRange(context.Background(), "round", 0, 1).Result()
	c.NoError(err)
	c.Empty(results)
}

func TestGetOrderedSetMin(t *testing.T) {
	c := require.New(t)
	ctx := context.Background()

	InitMock()

	member, score, err := GetOrderedSetMin(ctx, "round")
	c.NoError(err)
	c.Empty(member)
	c.Equal(float64(-1), score)

	err = AddToOrderedSetWithOption(ctx, "round", "a", 1, OnlyAdd)
	c.NoError(err)

	member, score, err = GetOrderedSetMin(ctx, "round")
	c.NoError(err)
	c.Equal("a", member)
	c.Equal(float64(1), score)

	err = AddToOrderedSetWithOption(ctx, "round", "b", 0, OnlyAdd)
	c.NoError(err)

	member, score, err = GetOrderedSetMin(ctx, "round")
	c.NoError(err)
	c.Equal("b", member)
	c.Equal(float64(0), score)
}

func TestGetOrderedSetMinError(t *testing.T) {
	c := require.New(t)

	InitMockWithoutServer()

	member, score, err := GetOrderedSetMin(context.Background(), "key")
	c.Empty(member)
	c.Equal(float64(-1), score)
	c.Error(err)
}

func TestGetAllOrderedSetMembers(t *testing.T) {
	c := require.New(t)

	InitMock()

	key := "key"
	memberA := "a"
	memberB := "b"

	quantity, err := redisClient.ZAdd(context.Background(), key,
		&redis.Z{
			Member: memberA,
			Score:  1,
		},
		&redis.Z{
			Member: memberB,
			Score:  2,
		}).Result()
	c.NoError(err)
	c.Equal(int64(2), quantity)

	members, err := GetAllOrderedSetMembers(context.Background(), key)
	c.NoError(err)
	c.Len(members, 2)
	c.EqualValues([]string{memberA, memberB}, members)
}

func TestAddToUnorderedSet(t *testing.T) {
	c := require.New(t)
	ctx := context.Background()

	InitMock()

	err := AddToUnorderedSet(ctx, "round", "a")
	c.NoError(err)

	err = AddToUnorderedSet(ctx, "round", "a")
	c.NoError(err)

	results, err := redisClient.SMembers(context.Background(), "round").Result()
	c.NoError(err)
	c.Len(results, 1)
}

func TestRemoveFromUnorderedSet(t *testing.T) {
	c := require.New(t)
	ctx := context.Background()

	InitMock()

	err := RemoveFromUnorderedSet(ctx, "round", "a")
	c.NoError(err)

	err = AddToUnorderedSet(ctx, "round", "a")
	c.NoError(err)

	err = RemoveFromUnorderedSet(ctx, "round", "a")
	c.NoError(err)

	results, err := redisClient.SMembers(context.Background(), "round").Result()
	c.NoError(err)

	c.Empty(results)
}

func TestGetAllUnorderedSetMembers(t *testing.T) {
	c := require.New(t)
	ctx := context.Background()

	InitMock()

	_, err := GetAllUnorderedSetMembers(ctx, "round")
	c.NoError(err)

	err = AddToUnorderedSet(ctx, "round", "a")
	c.NoError(err)

	err = AddToUnorderedSet(ctx, "round", "b")
	c.NoError(err)

	results, err := GetAllUnorderedSetMembers(ctx, "round")
	c.NoError(err)

	c.Len(results, 2)

	c.Contains(results, "a")
	c.Contains(results, "b")
}

func BenchmarkAddToOrderedSetWithOptionSuccess(b *testing.B) {
	c := require.New(b)
	ctx := context.Background()

	InitMock()

	for n := 0; n < b.N; n++ {
		err := AddToOrderedSetWithOption(ctx, "round", "a", 1, OnlyAdd)
		c.NoError(err)
	}
}
