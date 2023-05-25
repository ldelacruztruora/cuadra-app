package sns

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateTopic(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	arn, err := CreateTopic(context.Background(), "my-topic")
	c.Equal("arn:aws:sns:us-east-1:230572311368:my-topic", arn)
	c.Nil(err)
}
