package cache

import (
	"context"
	"errors"
)

var (
	// ErrNilPair means than the nil pair is invalid
	ErrNilPair = errors.New("invalid nil pair")
)

// MSetPair represents struct pair for be used in MSet method
type MSetPair struct {
	Key   string
	Value interface{}
}

// NewMSetPair takes a key and a value, validate them and if are valid return a new mSet pair
func NewMSetPair(key string, value interface{}) (*MSetPair, error) {
	return &MSetPair{
		Key:   key,
		Value: value,
	}, nil
}

// generateRawMSetPairs generate unidimencional array based in mset pairs, in the form [key1, value1, key2, value2...]
func generateRawMSetPairs(pairs []*MSetPair) ([]interface{}, error) {
	var mSetPairs []interface{}

	for _, pair := range pairs {
		if pair == nil {
			return nil, ErrNilPair
		}

		mSetPairs = append(mSetPairs, pair.Key)
		mSetPairs = append(mSetPairs, pair.Value)
	}

	return mSetPairs, nil
}

// MSet makes a multiple set to storage, example: MSet("key1",  "Hello", "key2",  "World")
// Sets the given keys to their respective values. MSET replaces existing values with new values,
// just as regular SET.
// MSET is atomic, so all given keys are set at once. It is not possible for clients to see that
// some of the keys were updated while others are unchanged.
func MSet(ctx context.Context, pairs []*MSetPair) error {
	rawMSetPairs, err := generateRawMSetPairs(pairs)
	if err != nil {
		return err
	}

	cmd := redisClient.MSet(ctx, rawMSetPairs...)
	_, err = cmd.Result()

	return err
}
