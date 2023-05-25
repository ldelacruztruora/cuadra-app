package secrets

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	scTypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
)

var (
	mockSecrets       map[string]string
	mockBinarySecrets map[string]string

	// ErrNotFound when secret is not found
	ErrNotFound = types.NotFoundException{
		Message: aws.String("not found mocked"),
	}
	// ErrResourceExistsException when secret is not found
	ErrResourceExistsException = errors.New("resource exists exception,mocked")

	// DefaultSecrets data dummy of list secrets output
	DefaultSecrets secretsmanager.ListSecretsOutput
)

type secretsmanagerClientInterface interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
	UpdateSecret(ctx context.Context, params *secretsmanager.UpdateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.UpdateSecretOutput, error)
	CreateSecret(ctx context.Context, params *secretsmanager.CreateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error)
	PutSecretValue(ctx context.Context, params *secretsmanager.PutSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error)
	DeleteSecret(ctx context.Context, params *secretsmanager.DeleteSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error)
	RestoreSecret(ctx context.Context, params *secretsmanager.RestoreSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.RestoreSecretOutput, error)
	ListSecrets(ctx context.Context, params *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error)
}

// Define a mock struct to be used
type mockSecretClient struct{}

// InitSecretsMock inits the secretmanager mock for the files library
func InitSecretsMock() {
	if mockSecrets == nil && mockBinarySecrets == nil {
		secretsClient = &mockSecretClient{}

		err := os.Setenv("AWS_REGION", "us-east-1")
		if err != nil {
			panic(err)
		}

		mockSecrets = map[string]string{}
		mockBinarySecrets = map[string]string{}
	}

	DefaultSecrets = secretsmanager.ListSecretsOutput{
		SecretList: []scTypes.SecretListEntry{
			{
				CreatedDate: aws.Time(time.Now().AddDate(-2, 1, 1)),
			},
			{
				CreatedDate: aws.Time(time.Now().AddDate(-1, 1, 1)),
			},
			{
				CreatedDate: aws.Time(time.Now()),
			},
		},
		NextToken: nil,
	}
}

// DeactivateMock Deactivate mock
func DeactivateMock() {
	configSecretsClient()

	mockSecrets = nil
	mockBinarySecrets = nil
	ForceMockFail = false
	ForceMockMarkedDeletion = false
	ForceMockAlreadyExists = false
}

// DeleteMocketSecret Remove keyfor a secret key only testing purposes
func DeleteMocketSecret(key string) {
	delete(mockSecrets, key)
}

// SetMockedSecret assigns a value for a secret key only testing purposes
func SetMockedSecret(key, value string) {
	mockSecrets[key] = value
}

// SetMockedBinarySecret assigns a value as binary for secret key only testing purposes
func SetMockedBinarySecret(key, value string) {
	mockBinarySecrets[key] = base64.StdEncoding.EncodeToString([]byte(value))
}

func (m *mockSecretClient) GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	str, strOk := mockSecrets[aws.ToString(input.SecretId)]
	bin, binOk := mockBinarySecrets[aws.ToString(input.SecretId)]

	if ForceMockFail {
		return nil, ErrForceMockFailError
	}

	if !(strOk || binOk) {
		return &secretsmanager.GetSecretValueOutput{}, &ErrNotFound
	}

	var strSecret *string
	if strOk {
		strSecret = aws.String(str)
	}

	return &secretsmanager.GetSecretValueOutput{
		SecretString: strSecret,
		SecretBinary: []byte(bin),
	}, nil
}

func (m *mockSecretClient) UpdateSecret(ctx context.Context, input *secretsmanager.UpdateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.UpdateSecretOutput, error) {
	secretName := aws.ToString(input.SecretId)
	_, strOk := mockSecrets[secretName]
	_, binOk := mockBinarySecrets[secretName]

	if !(strOk || binOk) {
		return &secretsmanager.UpdateSecretOutput{}, &ErrNotFound
	}

	mockSecrets[secretName] = aws.ToString(input.SecretString)
	mockBinarySecrets[secretName] = aws.ToString(input.SecretString)

	return &secretsmanager.UpdateSecretOutput{
		Name: input.SecretId,
	}, nil
}

func (m *mockSecretClient) PutSecretValue(ctx context.Context, input *secretsmanager.PutSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error) {
	secretName := aws.ToString(input.SecretId)
	_, strOk := mockSecrets[secretName]
	_, binOk := mockBinarySecrets[secretName]

	if ForceMockFail {
		return nil, ErrForceMockFailError
	}

	if strOk || binOk {
		return &secretsmanager.PutSecretValueOutput{}, ErrResourceExistsException
	}

	mockSecrets[secretName] = aws.ToString(input.SecretString)
	mockBinarySecrets[secretName] = aws.ToString(input.SecretString)

	return &secretsmanager.PutSecretValueOutput{
		Name: input.SecretId,
	}, nil
}

func (m *mockSecretClient) CreateSecret(ctx context.Context, input *secretsmanager.CreateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error) {
	secretName := aws.ToString(input.Name)
	_, strOk := mockSecrets[secretName]
	_, binOk := mockBinarySecrets[secretName]

	if ForceMockFail {
		return nil, ErrForceMockFailError
	}

	if ForceMockAlreadyExists {
		return nil, ErrForceMockAlreadyExists
	}

	if strOk || binOk {
		return &secretsmanager.CreateSecretOutput{}, ErrResourceExistsException
	}

	mockSecrets[secretName] = aws.ToString(input.SecretString)
	mockBinarySecrets[secretName] = aws.ToString(input.SecretString)

	return &secretsmanager.CreateSecretOutput{
		Name: input.Name,
	}, nil
}

func (m *mockSecretClient) DeleteSecret(ctx context.Context, input *secretsmanager.DeleteSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error) {
	secretName := aws.ToString(input.SecretId)
	_, strOk := mockSecrets[secretName]
	_, binOk := mockBinarySecrets[secretName]

	if ForceMockFail {
		return nil, ErrForceMockFailError
	}

	if ForceMockMarkedDeletion {
		return nil, ErrForceMockMarkedDeletion
	}

	if !strOk || !binOk {
		return &secretsmanager.DeleteSecretOutput{}, &ErrNotFound
	}

	delete(mockSecrets, secretName)
	delete(mockBinarySecrets, secretName)

	return &secretsmanager.DeleteSecretOutput{
		Name: input.SecretId,
	}, nil
}

func (m *mockSecretClient) RestoreSecret(ctx context.Context, input *secretsmanager.RestoreSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.RestoreSecretOutput, error) {
	secretName := aws.ToString(input.SecretId)
	_, strOk := mockSecrets[secretName]
	_, binOk := mockBinarySecrets[secretName]

	if ForceMockFail {
		return nil, ErrForceMockFailError
	}

	if strOk || binOk {
		return &secretsmanager.RestoreSecretOutput{}, ErrResourceExistsException
	}

	mockSecrets[secretName] = aws.ToString(input.SecretId)
	mockBinarySecrets[secretName] = aws.ToString(input.SecretId)

	return &secretsmanager.RestoreSecretOutput{
		Name: input.SecretId,
	}, nil
}

func (m *mockSecretClient) ListSecrets(ctx context.Context, input *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
	if ForceMockFail {
		return nil, ErrForceMockFailError
	}

	if input == nil || input.NextToken == nil {
		return &DefaultSecrets, nil
	}

	var newSecrets []scTypes.SecretListEntry

	for i := 0; i < 5; i++ {
		newSecrets = append(newSecrets, scTypes.SecretListEntry{CreatedDate: aws.Time(time.Now())})
	}

	return &secretsmanager.ListSecretsOutput{SecretList: newSecrets, NextToken: nil}, nil
}
