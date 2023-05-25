package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"shared/app/bot/models"

	"bitbucket.org/truora/scrap-services/deployments/approval"
	"bitbucket.org/truora/scrap-services/shared/awscore"
	"bitbucket.org/truora/scrap-services/shared/env"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

const (
	bettyTableUsers = "betty-users"

	idIndex                 = "id_index"
	bitbucketAccountIDIndex = "bitbucket_account_id_index"
)

var (
	useAssumeRole = env.GetBool("USE_ASSUME_ROLE", false)
	roleToAssume  = env.GetString("ROLE_TO_ASSUME", "arn:aws:iam::031975712270:role/access_betty_users")
	dynamoClient  dynamodbiface.DynamoDBAPI

	// ErrUserNotFound when user is not in the table
	ErrUserNotFound = errors.New("user not found")
	// ErrMissingEmail when user email is missing
	ErrMissingEmail = errors.New("missing email")
)

func init() {
	if useAssumeRole {
		sess := awscore.NewSession()
		creds := stscreds.NewCredentials(sess, roleToAssume)
		config := &aws.Config{Credentials: creds}

		client := dynamodb.New(sess, config)
		dynamoClient = client

		return
	}

	client := dynamodb.New(session.Must(session.NewSession()))
	dynamoClient = client
}

// GetUser find user by email
func GetUser(ctx context.Context, email string) (*models.From, error) {
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":email": {
				S: aws.String(email),
			},
		},
		KeyConditionExpression: aws.String("email = :email"),
		TableName:              aws.String(bettyTableUsers),
	}

	result, err := dynamoClient.QueryWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if *result.Count == 0 {
		return nil, ErrUserNotFound
	}

	var user *models.From

	err = dynamodbattribute.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetTelegramUserByBitbucketID find user by bitbucket ID
func GetTelegramUserByBitbucketID(ctx context.Context, bitbucketID string) (*models.From, error) {
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":bitbucket_account_id": {
				S: aws.String(bitbucketID),
			},
			":email_verified": {
				BOOL: aws.Bool(true),
			},
		},
		FilterExpression:       aws.String("email_verified = :email_verified"),
		KeyConditionExpression: aws.String("bitbucket_account_id = :bitbucket_account_id"),
		IndexName:              aws.String(bitbucketAccountIDIndex),
		TableName:              aws.String(bettyTableUsers),
	}

	result, err := dynamoClient.QueryWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if *result.Count == 0 {
		return nil, ErrUserNotFound
	}

	var user *models.From

	err = dynamodbattribute.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetTelegramUser find user by Telegram ID
func GetTelegramUser(ctx context.Context, id int64) (*models.From, error) {
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id": {
				N: aws.String(fmt.Sprint(id)),
			},
			":email_verified": {
				BOOL: aws.Bool(true),
			},
		},
		FilterExpression:       aws.String("email_verified = :email_verified"),
		KeyConditionExpression: aws.String("id = :id"),
		IndexName:              aws.String(idIndex),
		TableName:              aws.String(bettyTableUsers),
	}

	result, err := dynamoClient.QueryWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if *result.Count == 0 {
		return nil, ErrUserNotFound
	}

	var user *models.From

	err = dynamodbattribute.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetVerifiedUser find verified user by email
func GetVerifiedUser(ctx context.Context, email string, id int64) (*models.From, error) {
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":email":          {S: aws.String(email)},
			":email_verified": {BOOL: aws.Bool(true)},
			":telegram_id":    {N: aws.String(fmt.Sprint(id))},
		},
		FilterExpression:       aws.String("email_verified = :email_verified AND id = :telegram_id"),
		KeyConditionExpression: aws.String("email = :email"),
		TableName:              aws.String(bettyTableUsers),
	}

	result, err := dynamoClient.QueryWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if *result.Count == 0 {
		return nil, ErrUserNotFound
	}

	var user *models.From

	err = dynamodbattribute.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// SaveUser saves user to dynamodb
func SaveUser(ctx context.Context, user models.From) error {
	user.CreationDate = time.Now()
	user.ExpirationTime = time.Now().Add(5 * time.Minute).Unix()
	user.UserRole = approval.RoleDevelopers

	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return err
	}

	params := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(bettyTableUsers),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":current_time":   {N: aws.String(fmt.Sprintf("%d", time.Now().Unix()))},
			":email_verified": {BOOL: aws.Bool(false)},
		},
		ConditionExpression: aws.String("attribute_not_exists(expiration_time) OR expiration_time < :current_time AND email_verified = :email_verified"),
	}

	_, err = dynamoClient.PutItemWithContext(ctx, params)

	return err
}

// VerifyUserEmail saves user to dynamodb
func VerifyUserEmail(ctx context.Context, email string) error {
	params := &dynamodb.UpdateItemInput{
		TableName: aws.String(bettyTableUsers),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {S: aws.String(email)},
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":email_verified": {BOOL: aws.Bool(true)},
			":ttl":            {N: aws.String("0")},
		},
		UpdateExpression: aws.String("SET email_verified = :email_verified, expiration_time = :ttl"),
	}
	_, err := dynamoClient.UpdateItemWithContext(ctx, params)

	return err
}

// PutUser saves user to dynamodb
func PutUser(ctx context.Context, user *models.From) error {
	if user.Email == "" {
		return ErrMissingEmail
	}

	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return err
	}

	params := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(bettyTableUsers),
	}

	_, err = dynamoClient.PutItemWithContext(ctx, params)

	return err
}
