package sns

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/stretchr/testify/require"
)

func TestListSubscriptionsError(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	ForceMockFailListSubcriptions = true

	var err error
	_, err = SNSClient.ListSubscriptions(context.Background(), &sns.ListSubscriptionsInput{})
	c.Equal(ErrForcedFailure, err)
}

func TestListSubscriptionsByTopicError(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	var err error
	_, err = SNSClient.ListSubscriptionsByTopic(context.Background(), &sns.ListSubscriptionsByTopicInput{TopicArn: aws.String("not-found")})

	var aerr *types.NotFoundException

	c.True(errors.As(err, &aerr))
	c.EqualError(ErrNotFoundException, err.Error())

	ForceMockFail = true

	_, err = SNSClient.ListSubscriptionsByTopic(context.Background(), &sns.ListSubscriptionsByTopicInput{TopicArn: aws.String("not-found")})
	c.Equal(ErrForcedFailure, err)

	defer func() { ForceMockFail = false }()
}

func TestExtraFunctions(t *testing.T) {
	c := require.New(t)

	InitSNSMock()

	topicOutput, err := SNSClient.SetTopicAttributes(context.Background(), nil)
	c.NoError(err)
	c.Nil(topicOutput)

	tagOutput, err := SNSClient.TagResource(context.Background(), nil)
	c.NoError(err)
	c.Nil(tagOutput)
}

func BenchmarkListSubscriptionsError(b *testing.B) {
	c := require.New(b)

	InitSNSMock()

	ForceMockFailListSubcriptions = true
	ctx := context.Background()
	var err error

	for n := 0; n < b.N; n++ {
		_, err = SNSClient.ListSubscriptions(ctx, &sns.ListSubscriptionsInput{})
		c.Equal(ErrForcedFailure, err)
	}
}
