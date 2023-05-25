package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	c := require.New(t)

	InitMock()
	c.NoError(Init([]string{}, true))

	InitMock()
	c.NoError(Init([]string{}, false))
}

func TestInitPipeliner(t *testing.T) {
	c := require.New(t)

	InitMockPipeliner()

	c.NoError(Init([]string{}, true))
	c.NoError(Init([]string{}, false))
}

func TestInitFromEnv(t *testing.T) {
	c := require.New(t)

	InitMock()
	c.NoError(InitFromEnv())
}

func Test_newRedisClient(t *testing.T) {
	c := require.New(t)

	InitMock()

	c.NotNil(newRedisClient([]string{}, false))
	c.Equal(0, len(redisMock.shards))

	InitMock()

	c.NotNil(newRedisClient([]string{}, true))
	c.Equal(0, len(redisMock.shards))
}

func TestGetClient(t *testing.T) {
	c := require.New(t)

	InitMock()

	client := GetClient()
	c.Equal(redisClient, client)
}

func BenchmarkInit(b *testing.B) {
	InitMock()

	for n := 0; n < b.N; n++ {
		err := Init([]string{}, true)
		if err != nil {
			b.Fatal(err)
		}
	}
}
