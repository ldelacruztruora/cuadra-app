package storage

import (
	"context"
	"testing"

	"bitbucket.org/truora/scrap-services/devops/models"
	"github.com/stretchr/testify/require"
	"github.com/truora/minidyn"
)

func TestSaveUser(t *testing.T) {
	c := require.New(t)

	ctx := context.Background()

	InitDynamoMock()

	expectedUser := &models.From{
		ID:            int64(123456),
		PhoneNumber:   "305xyxdsu",
		FirstName:     "Beto",
		LastName:      "Gomez",
		Email:         "bgomez@truora.com",
		BitbucketID:   "bgomez@truora.com",
		EmailVerified: true,
	}

	err := SaveUser(ctx, *expectedUser)
	c.NoError(err)

	err = SaveUser(ctx, *expectedUser)
	c.Error(err)
	c.Contains(err.Error(), "ConditionalCheckFailedException")
}

func TestGetUserByemail(t *testing.T) {
	c := require.New(t)

	ctx := context.Background()

	InitDynamoMock()

	expectedUser := GetDummyUser(int64(123456), "dummy-email", "dummy-name", "dummy-last-name", "dummy-phone-number", "dummy-bitbucket-id")

	err := SaveUser(ctx, *expectedUser)
	c.NoError(err)

	user, err := GetUser(ctx, "dummy-email")
	c.NoError(err)
	c.Equal(user.Email, expectedUser.Email)
}

// func TestAddIndexFails(t *testing.T) {
// 	c := require.New(t)

// 	client := minidyn.NewClient()

// 	minidyn.ActiveForceFailure(client)

// 	defer minidyn.DeactiveForceFailure(client)

// 	c.Panics(func() { addIndex(client, "fakeTable") })
// }

func TestGetVerifiedUserByEmail(t *testing.T) {
	c := require.New(t)
	ctx := context.Background()

	InitDynamoMock()

	expectedUser := &models.From{
		ID:            int64(123456),
		PhoneNumber:   "305xyxdsu",
		FirstName:     "Beto",
		LastName:      "Gomez",
		Email:         "bgomez@truora.com",
		EmailVerified: true,
	}

	err := SaveUser(ctx, *expectedUser)
	c.NoError(err)

	user, err := GetVerifiedUser(ctx, "bgomez@truora.com", int64(123456))
	c.NoError(err)
	c.Equal(user.Email, expectedUser.Email)
}

func TestGetVerifiedUserByEmailError(t *testing.T) {
	c := require.New(t)

	ctx := context.Background()

	InitDynamoMock()

	expectedUser := &models.From{
		ID:            int64(123456),
		PhoneNumber:   "305xyxdsu",
		FirstName:     "Beto",
		LastName:      "Gomez",
		Email:         "bgomez@truora.com",
		EmailVerified: false,
	}

	err := SaveUser(ctx, *expectedUser)
	c.NoError(err)

	_, err = GetVerifiedUser(ctx, "bgomez@truora.com", int64(123456))
	c.Error(err)
	c.EqualError(err, ErrUserNotFound.Error())

	ActiveForceFailure()

	_, err = GetVerifiedUser(ctx, "bgomez@truora.com", int64(123456))
	c.Error(err)
	c.EqualError(err, minidyn.ErrForcedFailure.Error())
}

func TestGetUserError(t *testing.T) {
	c := require.New(t)

	ctx := context.Background()

	InitDynamoMock()

	user, err := GetUser(ctx, "dummy-email")
	c.Empty(user)
	c.EqualError(err, ErrUserNotFound.Error())
}

func TestGetUserErrorPerformingQuery(t *testing.T) {
	c := require.New(t)

	ctx := context.Background()

	InitDynamoMock()
	ActiveForceFailure()

	defer DeactiveForceFailure()

	_, err := GetUser(ctx, "dummy-email")
	c.EqualError(err, minidyn.ErrForcedFailure.Error())
}

func TestVerifyEmail(t *testing.T) {
	c := require.New(t)

	ctx := context.Background()

	InitDynamoMock()

	user := GetDummyUser(int64(123456), "dummy-email@example.com", "dummy-name", "dummy-last-name", "dummy-phone-number", "dummy-bitbucket-id")

	err := SaveUser(ctx, *user)
	c.NoError(err)

	err = VerifyUserEmail(context.Background(), "dummy-email@example.com")
	c.NoError(err)

	userUpdated, err := GetUser(context.Background(), "dummy-email@example.com")
	c.NoError(err)
	c.True(userUpdated.EmailVerified)
	c.Equal(int64(0), userUpdated.ExpirationTime)
}

func TestGetTelegramUser(t *testing.T) {
	c := require.New(t)

	ctx := context.Background()

	InitDynamoMock()

	user := GetDummyUser(int64(123456), "dummy-email@example.com", "dummy-name", "dummy-last-name", "dummy-phone-number", "dummy-bitbucket-id")

	err := SaveUser(ctx, *user)
	c.NoError(err)

	err = VerifyUserEmail(context.Background(), "dummy-email@example.com")
	c.NoError(err)

	user, err = GetTelegramUser(context.Background(), int64(123456))
	c.NoError(err)

	c.Equal(int64(int64(123456)), user.ID)
}

func TestGetTelegramUserByBitbucketID(t *testing.T) {
	c := require.New(t)

	ctx := context.Background()

	InitDynamoMock()

	user := GetDummyUser(int64(123456), "dummy-email@example.com", "dummy-name", "dummy-last-name", "dummy-phone-number", "dummy-bitbucket-account-id")

	err := SaveUser(ctx, *user)
	c.NoError(err)

	err = VerifyUserEmail(context.Background(), "dummy-email@example.com")
	c.NoError(err)

	user, err = GetTelegramUserByBitbucketID(context.Background(), "dummy-bitbucket-account-id")
	c.NoError(err)

	c.Equal(int64(int64(123456)), user.ID)
}

func TestErrorGetTelegramUserByBitbucketID(t *testing.T) {
	c := require.New(t)

	InitDynamoMock()

	_, err := GetTelegramUserByBitbucketID(context.Background(), "dummy-bitbucket-account-id")
	c.Error(err)
	c.EqualError(ErrUserNotFound, err.Error())
}

func TestGetTelegramUserFail(t *testing.T) {
	c := require.New(t)

	ctx := context.Background()

	InitDynamoMock()

	user := GetDummyUser(int64(123456), "dummy-email@example.com", "dummy-name", "dummy-last-name", "dummy-phone-number", "dummy-bitbucket-id")

	err := SaveUser(ctx, *user)
	c.NoError(err)

	err = VerifyUserEmail(context.Background(), "dummy-email@example.com")
	c.NoError(err)

	_, err = GetTelegramUser(context.Background(), int64(1234567))
	c.Error(err)
	c.EqualError(ErrUserNotFound, err.Error())

	ActiveForceFailure()

	defer DeactiveForceFailure()

	_, err = GetTelegramUser(context.Background(), int64(1234567))
	c.Error(err)
	c.EqualError(minidyn.ErrForcedFailure, err.Error())
}

func TestPutUser(t *testing.T) {
	c := require.New(t)

	InitDynamoMock()

	user := GetDummyUser(int64(123456), "dummy-email@example.com", "dummy-name", "dummy-last-name", "dummy-phone-number", "dummy-bitbucket-id")

	err := PutUser(context.Background(), user)
	c.NoError(err)
}

func TestPutUserError(t *testing.T) {
	c := require.New(t)

	ctx := context.Background()

	InitDynamoMock()

	user := GetDummyUser(int64(123456), "", "dummy-name", "dummy-last-name", "dummy-phone-number", "dummy-bitbucket-id")

	err := PutUser(ctx, user)
	c.Equal(ErrMissingEmail, err)

	user.Email = "dummy-email@example.com"

	ActiveForceFailure()

	defer DeactiveForceFailure()

	err = PutUser(ctx, user)
	c.Equal(minidyn.ErrForcedFailure, err)
	c.Panics(func() { checkErr(err) })
}
