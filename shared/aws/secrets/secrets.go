// Package secrets to use secrets functions from aws sdk go in our services
package secrets

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

var (
	secretsClient secretsmanagerClientInterface = &secretsmanager.Client{}

	// ErrEmptySecretValue secret value is empty
	ErrEmptySecretValue = errors.New("secret value is empty")
	// ErrEmptySecretName secret name is empty
	ErrEmptySecretName = errors.New("secret key is empty")
	// ErrForceMockFailError error when force mock to fail
	ErrForceMockFailError = errors.New("force mock to fail")
	// ErrForceMockMarkedDeletion error when force mock to fail marked for deletion
	ErrForceMockMarkedDeletion = errors.New("it was marked for deletion")
	// ErrForceMockAlreadyExists error when force already exist error
	ErrForceMockAlreadyExists = errors.New("already exists")
	// ForceMockFail force the mock to fail
	ForceMockFail = false
	// ForceMockMarkedDeletion force marked for deletion
	ForceMockMarkedDeletion = false
	// ForceMockAlreadyExists force already exist error
	ForceMockAlreadyExists = false
)

func configSecretsClient() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	otelaws.AppendMiddlewares(&cfg.APIOptions)

	secretsClient = secretsmanager.NewFromConfig(cfg)
}

func init() {
	configSecretsClient()
}

// Get fetch the current value of the secretName from aws secrets manager
func Get(ctx context.Context, secretName string) (string, error) {
	str, err := get(ctx, secretName)
	return strings.TrimSpace(str), err
}

func get(ctx context.Context, secretName string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := secretsClient.GetSecretValue(ctx, input)
	if err != nil {
		return "", err
	}

	if result.SecretString != nil {
		return aws.ToString(result.SecretString), nil
	}

	decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))

	len, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
	if err != nil {
		return "", err
	}

	decodedBinarySecret := string(decodedBinarySecretBytes[:len])

	return decodedBinarySecret, nil
}

// Update change the current value of the secretName from aws secrets manager
func Update(ctx context.Context, secretName, secretValue string) error {
	if secretValue == "" {
		return ErrEmptySecretValue
	}

	input := &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(secretName),
		SecretString: aws.String(secretValue),
	}

	_, err := secretsClient.UpdateSecret(ctx, input)

	return err
}

// Set the value of the secretName from aws secrets manager
func Set(ctx context.Context, secretName, secretValue string) error {
	if secretValue == "" {
		return ErrEmptySecretValue
	}

	input := &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(secretValue),
	}

	_, err := secretsClient.CreateSecret(ctx, input)

	return err
}

// Put update the value of the secretName from aws secrets manager
func Put(ctx context.Context, secretName, secretValue string) error {
	if secretValue == "" {
		return ErrEmptySecretValue
	}

	input := &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(secretName),
		SecretString: aws.String(secretValue),
	}

	_, err := secretsClient.PutSecretValue(ctx, input)

	return err
}

// Delete deletes the secret with the secretName from aws secrets manager
func Delete(ctx context.Context, secretName string) error {
	if secretName == "" {
		return ErrEmptySecretName
	}

	input := &secretsmanager.DeleteSecretInput{
		SecretId: aws.String(secretName),
	}

	_, err := secretsClient.DeleteSecret(ctx, input)

	return err
}

// Restore restores the secret with the secretName from aws secrets manager
func Restore(ctx context.Context, secretName string) error {
	if secretName == "" {
		return ErrEmptySecretName
	}

	input := &secretsmanager.RestoreSecretInput{
		SecretId: aws.String(secretName),
	}

	_, err := secretsClient.RestoreSecret(ctx, input)

	return err
}

// List returns a list of the secrets existents
func List(ctx context.Context, input *secretsmanager.ListSecretsInput) (*secretsmanager.ListSecretsOutput, error) {
	return secretsClient.ListSecrets(ctx, input)
}
