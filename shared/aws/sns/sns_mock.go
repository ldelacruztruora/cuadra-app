package sns

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
)

var (
	// EnableStatefulMocks it enable a stateful behavior with the mocks
	EnableStatefulMocks = false

	// ForceMockFail is variable to force mock response failure
	ForceMockFail = false
	// ErrForcedFailure is error when ForceMockFail is set to true
	ErrForcedFailure = errors.New("forced failure response")
	// ForceTopicNotFound is variable to force mock to respond with resource not found error
	ForceTopicNotFound = false
	// ForceMockFailListSubcriptions forces mock failure for ListSubscriptions function
	ForceMockFailListSubcriptions = false
	// ForceFailCreation forces mock failure for CreateTopic function
	ForceFailCreation = false
	// ErrCreationForcedFailure is error when ForceCreationMockFail is set to true
	ErrCreationForcedFailure = errors.New("forced failure response")
	// ForceThrottledException forces mock failure for Subscribe function
	ForceThrottledException = false
	// ForceMessageTooLong is variable to force mock response message too long
	ForceMessageTooLong = false
	// ErrMessageTooLongForcedFailure is error when ForceMessageTooLong is set to true
	ErrMessageTooLongForcedFailure = errors.New("forced failure message too long")
	// ForcePanic forces sns to panic
	ForcePanic = false
	// ErrForcedPanic is an error when sns is forced to panic
	ErrForcedPanic = errors.New("forced panic")

	// ErrNotFoundException not found exception error
	ErrNotFoundException = &types.NotFoundException{
		Message: aws.String("forced topic not found"),
	}
	// ErrInvalidParameterException invalid parameter exception
	ErrInvalidParameterException = &types.InvalidParameterException{
		Message: aws.String("forced Invalid parameter: Message too long"),
	}
	// ErrThrottledException throttled exception
	ErrThrottledException = &types.ThrottledException{
		Message: aws.String("forced throttled exception"),
	}

	mockedSubscriptions map[string][]types.Subscription
	mutex               = sync.Mutex{}
)

type snsClientInterface interface {
	ListTopics(ctx context.Context, params *sns.ListTopicsInput, optFns ...func(*sns.Options)) (*sns.ListTopicsOutput, error)
	GetTopicAttributes(ctx context.Context, params *sns.GetTopicAttributesInput, optFns ...func(*sns.Options)) (*sns.GetTopicAttributesOutput, error)
	ListSubscriptions(ctx context.Context, params *sns.ListSubscriptionsInput, optFns ...func(*sns.Options)) (*sns.ListSubscriptionsOutput, error)
	CreateTopic(ctx context.Context, params *sns.CreateTopicInput, optFns ...func(*sns.Options)) (*sns.CreateTopicOutput, error)
	TagResource(ctx context.Context, params *sns.TagResourceInput, optFns ...func(*sns.Options)) (*sns.TagResourceOutput, error)
	SetTopicAttributes(ctx context.Context, params *sns.SetTopicAttributesInput, optFns ...func(*sns.Options)) (*sns.SetTopicAttributesOutput, error)
	Subscribe(ctx context.Context, params *sns.SubscribeInput, optFns ...func(*sns.Options)) (*sns.SubscribeOutput, error)
	DeleteTopic(ctx context.Context, params *sns.DeleteTopicInput, optFns ...func(*sns.Options)) (*sns.DeleteTopicOutput, error)
	Unsubscribe(ctx context.Context, params *sns.UnsubscribeInput, optFns ...func(*sns.Options)) (*sns.UnsubscribeOutput, error)
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
	PublishBatch(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error)
	ListSubscriptionsByTopic(ctx context.Context, params *sns.ListSubscriptionsByTopicInput, optFns ...func(*sns.Options)) (*sns.ListSubscriptionsByTopicOutput, error)
}

// InitSNSMock initializes mock client for sns
func InitSNSMock() {
	mutex.Lock()
	defer mutex.Unlock()

	SNSClient = &mockSNSClient{}

	mockedSubscriptions = map[string][]types.Subscription{}

	ForceMockFail = false
	ForceTopicNotFound = false
	ForceMockFailListSubcriptions = false
	ForceFailCreation = false
	ForceThrottledException = false
	ForcePanic = false
	EnableStatefulMocks = false
}

type mockSNSClient struct{}

// Publish mock response for sns
func (m *mockSNSClient) Publish(ctx context.Context, input *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
	if EnableStatefulMocks {
		if _, ok := mockedSubscriptions[*input.TopicArn]; ok {
			return &sns.PublishOutput{}, nil
		}

		return nil, ErrNotFoundException
	}

	if ForceMockFail {
		return nil, ErrForcedFailure
	}

	if ForceTopicNotFound {
		return nil, ErrNotFoundException
	}

	if ForceMessageTooLong {
		return nil, ErrInvalidParameterException
	}

	if ForcePanic {
		panic(ErrForcedPanic)
	}

	return &sns.PublishOutput{}, nil
}

// PublishBatch mock response for sns
func (m *mockSNSClient) PublishBatch(ctx context.Context, input *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error) {
	if EnableStatefulMocks {
		if _, ok := mockedSubscriptions[*input.TopicArn]; ok {
			return &sns.PublishBatchOutput{}, nil
		}

		return nil, ErrNotFoundException
	}

	if ForceMockFail {
		return nil, ErrForcedFailure
	}

	if ForceTopicNotFound {
		return nil, ErrNotFoundException
	}

	if ForceMessageTooLong {
		return nil, ErrInvalidParameterException
	}

	if ForcePanic {
		panic(ErrForcedPanic)
	}

	return &sns.PublishBatchOutput{}, nil
}

// CreateTopic mock create topic for sns
func (m *mockSNSClient) CreateTopic(ctx context.Context, input *sns.CreateTopicInput, opts ...func(*sns.Options)) (*sns.CreateTopicOutput, error) {
	if ForceFailCreation {
		return nil, ErrCreationForcedFailure
	}

	return &sns.CreateTopicOutput{
		TopicArn: aws.String("arn:aws:sns:us-east-1:230572311368:my-topic"),
	}, nil
}

// ListSubscriptions mock response for sns
func (m *mockSNSClient) ListSubscriptions(ctx context.Context, input *sns.ListSubscriptionsInput, opts ...func(*sns.Options)) (*sns.ListSubscriptionsOutput, error) {
	if ForceMockFailListSubcriptions {
		return nil, ErrForcedFailure
	}

	return &sns.ListSubscriptionsOutput{
		Subscriptions: []types.Subscription{
			{
				TopicArn: aws.String("arn:aws:sns:us-east-1:230572311368:co-check-personal_identity"),
				Endpoint: aws.String("arn:aws:lambda:us-east-1:230572311368:function:sisben-service_collect"),
			},
			{
				TopicArn: aws.String("arn:aws:sns:us-east-1:230572311368:co-check-criminal_record"),
				Endpoint: aws.String("arn:aws:lambda:us-east-1:230572311368:function:ponal-service_collect"),
			},
			{
				TopicArn: aws.String("arn:aws:sns:us-east-1:230572311368:co-check-personal_identity-name"),
				Endpoint: aws.String("arn:aws:lambda:us-east-1:230572311368:function:procuraduria-service_collect"),
			},
			{
				TopicArn: aws.String("arn:aws:sns:us-east-1:230572311368:co-check-criminal_record"),
				Endpoint: aws.String("arn:aws:lambda:us-east-1:230572311368:function:jepms-service_collect"),
			},
			{
				TopicArn: aws.String("arn:aws:sns:us-east-1:230572311368:co-check-criminal_record-name"),
				Endpoint: aws.String("arn:aws:lambda:us-east-1:230572311368:function:unsc-service_collect:current"),
			},
			{
				TopicArn: aws.String("arn:aws:sns:us-east-1:230572311368:co-check-international_background-company_name"),
				Endpoint: aws.String("arn:aws:lambda:us-east-1:230572311368:function:csl-service_collect:current"),
			},
			{
				TopicArn: aws.String("arn:aws:sns:us-east-1:230572311368:co-check-international_background-name"),
				Endpoint: aws.String("arn:aws:lambda:us-east-1:230572311368:function:fbi-service_collect:current"),
			},
			{
				TopicArn: aws.String("arn:aws:sns:us-east-1:230572311368:co-check-legal_background"),
				Endpoint: aws.String("arn:aws:lambda:us-east-1:230572311368:function:expedientes-service_collect:current"),
			},
			{
				TopicArn: aws.String("arn:aws:sns:us-east-1:230572311368:co-check-by-national-id"),
				Endpoint: aws.String("dummyarn"),
			},
			{
				TopicArn: aws.String("arn:aws:sns:us-east-1:230572311368:co-check-legal_background"),
				Endpoint: aws.String("arn:aws:lambda:us-east-1:230572311368:func__:expedientes-service_collect:current"),
			},
			{
				TopicArn: aws.String("arn:aws:sns:us-east-1:230572311368:co-check-by-national-id"),
			},
		},
	}, nil
}

func (m *mockSNSClient) Subscribe(ctx context.Context, input *sns.SubscribeInput, opts ...func(*sns.Options)) (*sns.SubscribeOutput, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if ForceTopicNotFound {
		return nil, ErrNotFoundException
	}

	if ForceThrottledException {
		return nil, ErrThrottledException
	}

	topic := aws.ToString(input.TopicArn)

	_, ok := mockedSubscriptions[topic]
	if !ok {
		mockedSubscriptions[topic] = []types.Subscription{}
	}

	uuid := fmt.Sprintf("%06d", rand.Intn(1000000)) // #nosec G404

	mockedSubscriptions[topic] = append(mockedSubscriptions[topic], types.Subscription{
		Endpoint:        input.Endpoint,
		Protocol:        input.Protocol,
		SubscriptionArn: aws.String(aws.ToString(input.TopicArn) + ":" + uuid),
	})

	return &sns.SubscribeOutput{
		SubscriptionArn: aws.String(aws.ToString(input.TopicArn) + ":" + uuid),
	}, nil
}

func (m *mockSNSClient) GetTopicAttributes(ctx context.Context, input *sns.GetTopicAttributesInput, opts ...func(*sns.Options)) (*sns.GetTopicAttributesOutput, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if ForceTopicNotFound {
		return nil, ErrNotFoundException
	}

	subscriptions, ok := mockedSubscriptions[aws.ToString(input.TopicArn)]
	if !ok {
		return nil, ErrNotFoundException
	}

	return &sns.GetTopicAttributesOutput{
		Attributes: map[string]string{
			"SubscriptionsConfirmed": strconv.Itoa(len(subscriptions)),
		},
	}, nil
}

func (m *mockSNSClient) ListSubscriptionsByTopic(ctx context.Context, input *sns.ListSubscriptionsByTopicInput, opt ...func(*sns.Options)) (*sns.ListSubscriptionsByTopicOutput, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if ForceMockFail {
		return nil, ErrForcedFailure
	}

	subscriptions, ok := mockedSubscriptions[aws.ToString(input.TopicArn)]
	if !ok {
		return nil, ErrNotFoundException
	}

	return &sns.ListSubscriptionsByTopicOutput{
		Subscriptions: subscriptions,
	}, nil
}

func (m *mockSNSClient) Unsubscribe(ctx context.Context, input *sns.UnsubscribeInput, opt ...func(*sns.Options)) (*sns.UnsubscribeOutput, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if ForceTopicNotFound {
		return nil, ErrNotFoundException
	}

	if ForceThrottledException {
		return nil, ErrThrottledException
	}

	arn := input.SubscriptionArn
	splitted := strings.Split(aws.ToString(arn), ":")
	topic := strings.Join(splitted[:len(splitted)-1], ":")

	subscriptions, ok := mockedSubscriptions[topic]
	if !ok {
		return nil, ErrNotFoundException
	}

	nsubs := make([]types.Subscription, 0, len(subscriptions)-1)

	for _, s := range subscriptions {
		if aws.ToString(s.SubscriptionArn) == aws.ToString(input.SubscriptionArn) {
			continue
		}

		nsubs = append(nsubs, s)
	}

	mockedSubscriptions[topic] = nsubs

	return &sns.UnsubscribeOutput{}, nil
}

func (m *mockSNSClient) ListTopics(ctx context.Context, input *sns.ListTopicsInput, opt ...func(*sns.Options)) (*sns.ListTopicsOutput, error) {
	if ForceThrottledException {
		return nil, ErrThrottledException
	}

	topics := []types.Topic{}

	for arn := range mockedSubscriptions {
		topics = append(topics, types.Topic{
			TopicArn: aws.String(arn),
		})
	}

	return &sns.ListTopicsOutput{NextToken: new(string), Topics: topics}, nil
}

func (m *mockSNSClient) DeleteTopic(ctx context.Context, input *sns.DeleteTopicInput, opt ...func(*sns.Options)) (*sns.DeleteTopicOutput, error) {
	mutex.Lock()
	defer mutex.Unlock()

	_, ok := mockedSubscriptions[*input.TopicArn]
	if !ok {
		return nil, ErrNotFoundException
	}

	delete(mockedSubscriptions, *input.TopicArn)

	return &sns.DeleteTopicOutput{}, nil
}

func (m *mockSNSClient) SetTopicAttributes(ctx context.Context, params *sns.SetTopicAttributesInput, optFns ...func(*sns.Options)) (*sns.SetTopicAttributesOutput, error) {
	return nil, nil
}

func (m *mockSNSClient) TagResource(ctx context.Context, params *sns.TagResourceInput, optFns ...func(*sns.Options)) (*sns.TagResourceOutput, error) {
	return nil, nil
}
