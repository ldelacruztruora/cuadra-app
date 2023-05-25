package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// CreateTopic creates a sns topic with the given name
func CreateTopic(ctx context.Context, name string) (string, error) {
	result, err := SNSClient.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(name),
	})
	if err != nil {
		return "", err
	}

	return *result.TopicArn, nil
}

// SetTopicAttribute assign the topic attribute
func SetTopicAttribute(ctx context.Context, arn, key, val string) error {
	_, err := SNSClient.SetTopicAttributes(ctx, &sns.SetTopicAttributesInput{
		TopicArn:       aws.String(arn),
		AttributeName:  aws.String(key),
		AttributeValue: aws.String(val),
	})

	return err
}

// SubscribeLambda subscribes a lambda to a sns topic
func SubscribeLambda(ctx context.Context, arn string, lambdaARN string) (string, error) {
	input := sns.SubscribeInput{
		Endpoint:              aws.String(lambdaARN),
		Protocol:              aws.String("lambda"),
		ReturnSubscriptionArn: true,
		TopicArn:              aws.String(arn),
	}

	result, err := SNSClient.Subscribe(ctx, &input)
	if err != nil {
		return "", errorManager(err)
	}

	return *result.SubscriptionArn, nil
}

// DeleteTopic delete the specified topic
func DeleteTopic(ctx context.Context, arn string) error {
	input := &sns.DeleteTopicInput{
		TopicArn: aws.String(arn),
	}

	_, err := SNSClient.DeleteTopic(ctx, input)
	if err != nil {
		return errorManager(err)
	}

	return nil
}
