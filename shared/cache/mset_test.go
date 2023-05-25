package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMSet_OK(t *testing.T) {
	c := require.New(t)
	key := "key"
	val := "val"
	pair, err := NewMSetPair(key, val)
	c.Nil(err)
	c.Equal(pair.Key, key)
	c.Equal(pair.Value, val)
}

func TestGenerateRawMSetPairs_OK(t *testing.T) {
	c := require.New(t)

	pair1, err := NewMSetPair("some", 1)
	c.Nil(err)
	pair2, err := NewMSetPair("other", 2)
	c.Nil(err)

	pairs := []*MSetPair{pair1, pair2}

	rawMSetPairs, err := generateRawMSetPairs(pairs)
	c.Nil(err)
	c.Len(rawMSetPairs, len(pairs)*2)
	c.Equal(rawMSetPairs[0], "some")
	c.Equal(rawMSetPairs[1], 1)
	c.Equal(rawMSetPairs[2], "other")
	c.Equal(rawMSetPairs[3], 2)
}

func TestGenerateRawMSetPairs_InvalidNilPair(t *testing.T) {
	c := require.New(t)

	pair1, err := NewMSetPair("some", 1)
	c.Nil(err)
	var pair2 *MSetPair // invalid nil pair

	pairs := []*MSetPair{pair1, pair2}

	rawMSetPairs, err := generateRawMSetPairs(pairs)
	c.Nil(rawMSetPairs)
	c.Equal(ErrNilPair, err)
}

func TestMSet_InvalidNilPair(t *testing.T) {
	c := require.New(t)

	pair1, err := NewMSetPair("some", 1)
	c.Nil(err)
	var pair2 *MSetPair // invalid nil pair

	pairs := []*MSetPair{pair1, pair2}

	err = MSet(context.Background(), pairs)
	c.Equal(ErrNilPair, err)
}

func TestMSet_OK(t *testing.T) {
	c := require.New(t)

	InitMock()
	pair1, err := NewMSetPair("some", 1)
	c.Nil(err)
	pair2, err := NewMSetPair("lastKey", "lastValue") // last key and last value in MSet
	c.Nil(err)

	pairs := []*MSetPair{pair1, pair2}

	err = MSet(context.Background(), pairs)
	c.Nil(err)

	value, err := MockServer.Get("some")
	c.NoError(err)
	c.Equal("1", value)
	value, err = MockServer.Get("lastKey")
	c.NoError(err)
	c.Equal("lastValue", value)
}
