package storage

import (
	"bitbucket.org/truora/scrap-services/devops/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/truora/minidyn"
)

// InitDynamoMock initializes dynamodb client with a mock
func InitDynamoMock() {
	client := minidyn.NewClient()

	_, err := client.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.String(bettyTableUsers),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("email"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("bitbucket_account_id"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("N"),
			},
		},
		BillingMode: aws.String("PAY_PER_REQUEST"),
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("email"),
				KeyType:       aws.String("HASH"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String(bitbucketAccountIDIndex),
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String(dynamodb.ProjectionTypeAll),
				},
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("bitbucket_account_id"),
						KeyType:       aws.String("HASH"),
					},
				},
			},
			{
				IndexName: aws.String(idIndex),
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String(dynamodb.ProjectionTypeAll),
				},
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("id"),
						KeyType:       aws.String("HASH"),
					},
				},
			},
		},
	})
	checkErr(err)

	dynamoClient = client
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// ActiveForceFailure force fake dynamodb to fail
func ActiveForceFailure() {
	minidyn.ActiveForceFailure(dynamoClient)
}

// DeactiveForceFailure remove forcing fake dynamodb to fail
func DeactiveForceFailure() {
	minidyn.DeactiveForceFailure(dynamoClient)
}

// GetDummyUser returns a dummy user for the unit tests
func GetDummyUser(userID int64, email, firstName, lastName, phoneNumber string, bitbucketID string) *models.From {
	return &models.From{
		ID:          userID,
		PhoneNumber: phoneNumber,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
		BitbucketID: bitbucketID,
	}
}
