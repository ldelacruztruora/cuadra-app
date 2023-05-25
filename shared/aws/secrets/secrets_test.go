package secrets

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/require"
)

func TestDeactivateMock(t *testing.T) {
	c := require.New(t)

	InitSecretsMock()

	defer DeactivateMock()

	c.NotNil(mockSecrets)
	c.NotNil(mockBinarySecrets)

	DeactivateMock()
	c.Nil(mockSecrets)
	c.Nil(mockBinarySecrets)
}

func TestDeleteMocketSecret(t *testing.T) {
	c := require.New(t)

	InitSecretsMock()

	defer DeactivateMock()

	SetMockedSecret("mock", "mock")

	DeleteMocketSecret("mock")

	_, err := Get(context.Background(), "mock")
	c.Contains(err.Error(), aws.ToString(ErrNotFound.Message))
}

func TestGetStringSecret(t *testing.T) {
	c := require.New(t)

	InitSecretsMock()

	defer DeactivateMock()

	SetMockedSecret("mock", "test:test")

	data, err := Get(context.Background(), "mock")
	c.Nil(err)
	c.Equal("test:test", data)

	_, err = Get(context.Background(), "404")
	c.Contains(err.Error(), aws.ToString(ErrNotFound.Message))
}

func BenchmarkGetStringSecret(b *testing.B) {
	InitSecretsMock()

	defer DeactivateMock()

	SetMockedSecret("mock", "test:test")

	for n := 0; n < b.N; n++ {
		_, err := Get(context.Background(), "mock")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestGetBinarySecret(t *testing.T) {
	c := require.New(t)

	InitSecretsMock()

	defer DeactivateMock()

	SetMockedBinarySecret("binaryMock", "test:test")

	data, err := Get(context.Background(), "binaryMock")
	c.Nil(err)
	c.Equal("test:test", data)

	_, err = Get(context.Background(), "404")
	c.Contains(err.Error(), aws.ToString(ErrNotFound.Message))
}

func TestUpdateSecret(t *testing.T) {
	c := require.New(t)

	InitSecretsMock()

	defer DeactivateMock()

	SetMockedSecret("mock", "test")

	data, err := Get(context.Background(), "mock")
	c.NoError(err)
	c.Equal("test", data)

	err = Update(context.Background(), "mock", "dummy")
	c.NoError(err)

	data, err = Get(context.Background(), "mock")
	c.NoError(err)
	c.Equal("dummy", data)

	err = Update(context.Background(), "token", "dummy")
	c.Contains(err.Error(), aws.ToString(ErrNotFound.Message))

	err = Update(context.Background(), "token", "")
	c.Equal(ErrEmptySecretValue, err)
}

func TestSetSecret(t *testing.T) {
	c := require.New(t)

	InitSecretsMock()

	defer DeactivateMock()

	err := Set(context.Background(), "mock", "")
	c.Equal(ErrEmptySecretValue, err)

	ForceMockFail = true

	err = Set(context.Background(), "mock", "dummy")
	c.Equal(ErrForceMockFailError, err)

	ForceMockFail = false
	ForceMockAlreadyExists = true

	err = Set(context.Background(), "mock", "dummy")
	c.Equal(ErrForceMockAlreadyExists, err)

	ForceMockAlreadyExists = false

	err = Set(context.Background(), "mock", "dummy")
	c.NoError(err)

	data, err := Get(context.Background(), "mock")
	c.NoError(err)
	c.Equal("dummy", data)

	err = Set(context.Background(), "mock", "dummy")
	c.Contains(err.Error(), ErrResourceExistsException.Error())
}

func TestPutSecret(t *testing.T) {
	c := require.New(t)

	InitSecretsMock()

	defer DeactivateMock()

	err := Put(context.Background(), "mock", "")
	c.Equal(ErrEmptySecretValue, err)

	ForceMockFail = true

	err = Put(context.Background(), "mock", "dummy")
	c.Equal(ErrForceMockFailError, err)

	ForceMockFail = false

	err = Put(context.Background(), "mock", "dummy")
	c.NoError(err)

	data, err := Get(context.Background(), "mock")
	c.NoError(err)
	c.Equal("dummy", data)

	err = Put(context.Background(), "mock", "dummy")
	c.Contains(err.Error(), ErrResourceExistsException.Error())
}

func TestDeleteSecret(t *testing.T) {
	c := require.New(t)

	InitSecretsMock()

	defer DeactivateMock()

	err := Delete(context.Background(), "")
	c.ErrorIs(ErrEmptySecretName, err)

	ForceMockFail = true

	err = Delete(context.Background(), "mock")
	c.ErrorIs(ErrForceMockFailError, err)

	ForceMockFail = false

	err = Put(context.Background(), "mock", "dummy")
	c.NoError(err)

	err = Delete(context.Background(), "mock")
	c.NoError(err)

	_, err = Get(context.Background(), "mock")
	c.Contains(err.Error(), aws.ToString(ErrNotFound.Message))
}

func TestRestoreSecret(t *testing.T) {
	c := require.New(t)

	InitSecretsMock()

	defer DeactivateMock()

	err := Restore(context.Background(), "")
	c.Equal(ErrEmptySecretName, err)

	ForceMockFail = true

	err = Restore(context.Background(), "mock")
	c.Equal(ErrForceMockFailError, err)

	ForceMockFail = false

	err = Restore(context.Background(), "mock")
	c.NoError(err)
}

func TestList(t *testing.T) {
	c := require.New(t)

	InitSecretsMock()

	defer DeactivateMock()

	_, err := List(context.Background(), nil)
	c.Nil(err)

	input := secretsmanager.ListSecretsInput{NextToken: aws.String("tokeeen")}

	secrets, err := List(context.Background(), &input)
	c.Nil(err)
	c.Len(secrets.SecretList, 5)
}

func TestListFailed(t *testing.T) {
	c := require.New(t)

	InitSecretsMock()

	ForceMockFail = true

	defer func() {
		DeactivateMock()

		ForceMockFail = false
	}()

	secrets, err := List(context.Background(), nil)
	c.Equal(ErrForceMockFailError, err)
	c.Nil(secrets)
}
