package sns

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"shared/shared/aws/sns/gzip"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/gofrs/uuid"
)

// Compression compression type for sns
type Compression string

const (
	// ContentEncodingSNSAttributeKey is the key for the content encoding attribute
	ContentEncodingSNSAttributeKey = "CONTENT-ENCODING"

	// CompressionGzip compresses the message using gzip
	CompressionGzip Compression = "gzip"
)

var (
	// ErrTopicNotFound error when the specified sns topic was not found
	ErrTopicNotFound = errors.New("sns topic not found")
	// ErrEmptyMessage error when no message is sent
	ErrEmptyMessage = errors.New("sns message is empty")
	// ErrInvalidInput the entrie is not a slice
	ErrInvalidInput = errors.New("unexpected input. Input is not an slice")
)

// PublishJSON publishes a message in JSON format to the given topic
func PublishJSON(ctx context.Context, topicARN string, v interface{}) error {
	return publishSNSJSON(ctx, topicARN, v, false)
}

// PublishCompressedJSON publishes a message compress to the given topic
func PublishCompressedJSON(ctx context.Context, topicARN string, v interface{}) error {
	return publishSNSJSON(ctx, topicARN, v, true)
}

func publishSNSJSON(ctx context.Context, topicARN string, v interface{}, compress bool) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return Publish(ctx, topicARN, data, compress)
}

// Publish function to publish binary data via SNS
func Publish(ctx context.Context, topicARN string, data []byte, compress bool) error {
	var err error
	if compress {
		data, err = compressData(data)
		if err != nil {
			return err
		}
	}

	_, err = SNSClient.Publish(ctx, &sns.PublishInput{
		Message:  aws.String(string(data)),
		TopicArn: aws.String(topicARN),
	})
	if err != nil {
		return errorManager(err)
	}

	return err
}

// PublishBatchCompressedJSON publishes a message compress to the given topic
func PublishBatchCompressedJSON(ctx context.Context, topicARN string, v interface{}) error {
	return publishBatchSNSJSON(ctx, topicARN, v, true)
}

func publishBatchSNSJSON(ctx context.Context, topicARN string, v interface{}, compress bool) error {
	entries := reflect.ValueOf(v)
	if entries.Kind() != reflect.Slice {
		return ErrInvalidInput
	}

	batch := make([][]byte, entries.Len())

	for i := 0; i < entries.Len(); i++ {
		message, err := json.Marshal(entries.Index(i).Interface())
		if err != nil {
			return err
		}

		batch[i] = message
	}

	return PublishBatch(ctx, topicARN, batch, compress)
}

func getPublishBatchEntries(batch [][]byte, compress bool) ([]types.PublishBatchRequestEntry, error) {
	var err error

	entries := make([]types.PublishBatchRequestEntry, len(batch))

	for i, message := range batch {
		if compress {
			message, err = compressData(message)
			if err != nil {
				return nil, err
			}
		}

		id := uuid.Must(uuid.NewV4())

		entry := types.PublishBatchRequestEntry{
			Id:      aws.String(id.String()),
			Message: aws.String(string(message)),
		}

		entries[i] = entry
	}

	return entries, nil
}

// PublishBatch function to binary batch via SNS
func PublishBatch(ctx context.Context, topicARN string, batch [][]byte, compress bool) error {
	entries, err := getPublishBatchEntries(batch, compress)
	if err != nil {
		return err
	}

	_, err = SNSClient.PublishBatch(ctx, &sns.PublishBatchInput{
		PublishBatchRequestEntries: entries,
		TopicArn:                   aws.String(topicARN),
	})
	if err != nil {
		return errorManager(err)
	}

	return err
}

func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	gz, err := gzip.Compressor.Compress(&buf)
	if err != nil {
		return nil, err
	}

	_, err = gz.Write(data)
	if err != nil {
		return nil, err
	}

	err = gz.Close()
	if err != nil {
		return nil, err
	}

	return []byte(base64.RawStdEncoding.EncodeToString(buf.Bytes())), nil
}

// IsMessageTooLong check if aws error is equal to "Message too long"
func IsMessageTooLong(err error) bool {
	var invalidParameterException *types.InvalidParameterException

	return errors.As(err, &invalidParameterException) && strings.Contains(*invalidParameterException.Message, "Invalid parameter: Message too long")
}

func errorManager(err error) error {
	var notFoundException *types.NotFoundException
	if errors.As(err, &notFoundException) {
		return ErrTopicNotFound
	}

	return err
}
