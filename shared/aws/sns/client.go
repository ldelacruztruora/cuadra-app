// Package sns library to use aws-sdk-go-v2/service/sns
package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

var (
	// SNSClient is the client for the SNS service
	SNSClient snsClientInterface = &sns.Client{}
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sns.NewFromConfig(cfg)
	SNSClient = client
}
