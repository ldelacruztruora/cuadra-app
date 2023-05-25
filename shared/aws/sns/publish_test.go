package sns

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublishJSON(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	err := PublishJSON(context.Background(), "my-topic", map[string]interface{}{})
	c.Nil(err)

	err = PublishJSON(context.Background(), "my-topic", map[string]interface{}{})
	c.Nil(err)
}

func TestPublishCompressedJSON(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	err := PublishCompressedJSON(context.Background(), "my-topic", map[string]interface{}{})
	c.Nil(err)
}

func TestPublishBatchCompressedJSON(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	err := PublishBatchCompressedJSON(context.Background(), "my-topic", []map[string]interface{}{
		{"": ""},
		{"": ""},
		{"": ""},
	})
	c.Nil(err)
}

func TestPublishBatchCompressedJSONFail(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	ForceMockFail = true

	defer func() {
		ForceMockFail = false
	}()

	err := PublishBatchCompressedJSON(context.Background(), "my-topic", []map[string]interface{}{
		{"": ""},
		{"": ""},
		{"": ""},
	})
	c.Equal(ErrForcedFailure, err)
}

func TestPublishBatchCompressedJSONNotFound(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	ForceTopicNotFound = true

	defer func() {
		ForceTopicNotFound = false
	}()

	err := PublishBatchCompressedJSON(context.Background(), "my-topic", []map[string]interface{}{
		{"": ""},
		{"": ""},
		{"": ""},
	})
	c.Equal(ErrTopicNotFound, err)
}

func TestPublishBatchCompressedJSONTooLong(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	ForceMessageTooLong = true

	defer func() {
		ForceMessageTooLong = false
	}()

	err := PublishBatchCompressedJSON(context.Background(), "my-topic", []map[string]interface{}{
		{"": ""},
		{"": ""},
		{"": ""},
	})
	c.Contains(err.Error(), "Message too long")
}

func TestPublishBatchCompressedJSONForcePanic(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	ForcePanic = true

	defer func() {
		ForcePanic = false
	}()

	c.PanicsWithError(
		"forced panic",
		func() {
			_ = PublishBatchCompressedJSON(context.Background(), "my-topic", []map[string]interface{}{
				{"": ""},
				{"": ""},
				{"": ""},
			})
		},
	)
}

func TestPublishBatchCompressedJSONEnableStatefulMocks(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	EnableStatefulMocks = true

	defer func() {
		EnableStatefulMocks = false
	}()

	err := PublishBatchCompressedJSON(context.Background(), "my-topic", []map[string]interface{}{
		{"": ""},
		{"": ""},
		{"": ""},
	})
	c.Equal(ErrTopicNotFound, err)
}

func TestPublishJSONFail(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	ForceMockFail = true

	defer func() {
		ForceMockFail = false
	}()

	err := PublishJSON(context.Background(), "my-topic", map[string]interface{}{})
	c.Equal(ErrForcedFailure, err)
}

func TestPublishAutoCompressJSONFail(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	ForceMessageTooLong = true

	defer func() { ForceMessageTooLong = false }()

	err := PublishJSON(context.Background(), "my-topic", map[string]interface{}{})
	c.True(IsMessageTooLong(err))
}

func TestPublishCompressedJSONFail(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	ForceMockFail = true

	defer func() {
		ForceMockFail = false
	}()

	err := PublishCompressedJSON(context.Background(), "my-topic", map[string]interface{}{})
	c.Equal(ErrForcedFailure, err)
}

func TestPublish_TopicNotFound(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	ForceTopicNotFound = true
	err := PublishJSON(context.Background(), "my-topic", map[string]interface{}{})
	c.Equal(ErrTopicNotFound, err)

	ForceTopicNotFound = false
}

func TestPublish_WithStatefulMocks(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	err := os.Setenv("AWS_REGION", "us-east-1")
	c.NoError(err)

	EnableStatefulMocks = true
	topic := "arn:aws:sns:us-east-1:230572311368:my-topic"

	err = PublishJSON(context.Background(), topic, map[string]interface{}{})
	c.Equal(ErrTopicNotFound, err)

	_, err = CreateTopic(context.Background(), "my-topic")
	c.NoError(err)

	EnableStatefulMocks = false

	err = PublishJSON(context.Background(), topic, map[string]interface{}{})
	c.NoError(err)
}

func TestPublishCompress_TopicNotFound(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	ForceTopicNotFound = true
	err := PublishCompressedJSON(context.Background(), "my-topic", map[string]interface{}{})
	c.Equal(ErrTopicNotFound, err)
}
