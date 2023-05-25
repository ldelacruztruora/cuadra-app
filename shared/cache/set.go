package cache

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
)

// OrderedSetOption options to use ordered set operations
type OrderedSetOption string

const (
	// OnlyUpdate Only update elements that already exist. Never add elements.
	OnlyUpdate OrderedSetOption = "XX"

	// OnlyAdd Don't update already existing elements. Always add new elements.
	OnlyAdd OrderedSetOption = "NX"
)

var (
	// ErrUnsupportedOrderedSetOption ordered set is not soported
	ErrUnsupportedOrderedSetOption = errors.New("unsupported ordered set option")
	// ErrNonStringMember means than the set element member is non string
	ErrNonStringMember = errors.New("each member of a set should be a string")
)

// AddToOrderedSetWithOption add the value to the zrange
func AddToOrderedSetWithOption(ctx context.Context, key, value string, score float64, opt OrderedSetOption) error {
	z := redis.Z{
		Score:  score,
		Member: value,
	}

	switch opt {
	case OnlyUpdate:
		_, err := redisClient.ZAddXX(ctx, key, &z).Result()
		return err
	case OnlyAdd:
		_, err := redisClient.ZAddNX(ctx, key, &z).Result()
		return err
	}

	return ErrUnsupportedOrderedSetOption
}

// RemoveFromToOrderedSet remove the value from the zrange
func RemoveFromToOrderedSet(ctx context.Context, key, value string) error {
	_, err := redisClient.ZRem(ctx, key, value).Result()

	return err
}

// GetOrderedSetMin get the first value of the zrange
// always return -1 in score if some goes wrong
func GetOrderedSetMin(ctx context.Context, key string) (element string, score float64, err error) {
	elements, err := redisClient.ZRangeWithScores(ctx, key, 0, 0).Result()
	if err != nil {
		return "", -1, err
	}

	if len(elements) == 0 {
		return "", -1, nil
	}

	strMember, ok := elements[0].Member.(string)
	if !ok {
		return "", -1, ErrNonStringMember
	}

	return strMember, elements[0].Score, nil
}

// GetAllOrderedSetMembers returns all members that belongs to a set
func GetAllOrderedSetMembers(ctx context.Context, key string) ([]string, error) {
	return redisClient.ZRange(ctx, key, 0, -1).Result()
}

// AddToUnorderedSet add the value to an unordered set
func AddToUnorderedSet(ctx context.Context, key string, values ...interface{}) error {
	return redisClient.SAdd(ctx, key, values).Err()
}

// RemoveFromUnorderedSet remove the value from an unordered set
func RemoveFromUnorderedSet(ctx context.Context, key string, value ...interface{}) error {
	return redisClient.SRem(ctx, key, value).Err()
}

// GetAllUnorderedSetMembers returns all members that belongs to an unordered set
func GetAllUnorderedSetMembers(ctx context.Context, key string) ([]string, error) {
	return redisClient.SMembers(ctx, key).Result()
}
