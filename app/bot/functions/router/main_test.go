package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	"bitbucket.org/truora/scrap-services/devops/bot/storage"
	"bitbucket.org/truora/scrap-services/devops/bot/storage/conversation"
	"bitbucket.org/truora/scrap-services/devops/models"
	"bitbucket.org/truora/scrap-services/logger"
	"github.com/stretchr/testify/require"
	"shared/aws/apigateway"
	"shared/aws/cache"
	"shared/aws/client"
	"shared/aws/secrets"
	"shared/aws/sns"
)

func BenchmarkHandler(b *testing.B) {
	c := require.New(b)

	setMockedClient()

	defer deactivateMockedClient()

	for i := 0; i < b.N; i++ {
		response, err := apiGatewayHandler(context.Background(), &request{
			APIGatewayProxyRequest: &apigateway.Request{},
		})
		c.NoError(err)
		c.Equal(http.StatusOK, response.StatusCode)
	}
}

func TestGetCommandCallback(t *testing.T) {
	c := require.New(t)

	command, err := getCommand(&models.CallbackMessage{Command: "command"})
	c.NoError(err)
	c.Equal("command", command)
}

func TestGetCommandError(t *testing.T) {
	c := require.New(t)

	_, err := getCommand(&models.CallbackMessage{})
	c.ErrorIs(err, ErrMessageEmpty)
}

func TestGetMessage(t *testing.T) {
	c := require.New(t)

	cache.InitMock()

	err := cache.Add(context.Background(), "BOTdata_id", "data", 0)
	c.NoError(err)

	message, err := getMessage(context.Background(), &event{CallbackMessage: &models.CallbackMessage{Data: "command, BOTdata_id"}})
	c.NoError(err)
	c.Equal("data", message.Data)
}

func TestGetMessageError(t *testing.T) {
	c := require.New(t)

	_, err := getMessage(context.Background(), &event{})
	c.ErrorIs(err, ErrUnknownTelegramEvent)

	_, err = getMessage(context.Background(), &event{CallbackMessage: &models.CallbackMessage{}})
	c.ErrorIs(err, ErrInvalidCallback)
}

func TestCreateTelegramClientFailed(t *testing.T) {
	c := require.New(t)

	secrets.InitSecretsMock()
	client.ActivateMock()

	response, err := apiGatewayHandler(context.Background(), &request{
		APIGatewayProxyRequest: &apigateway.Request{Body: `{"message":{"text":"/hi dummy_email@dummy.com"}}`},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)
}

func TestApiGatewayHandlerUnauthorizedCommand(t *testing.T) {
	c := require.New(t)

	setMockedClient()

	defer deactivateMockedClient()

	sns.InitSNSMock()

	response, err := apiGatewayHandler(context.Background(), &request{
		APIGatewayProxyRequest: &apigateway.Request{Body: `{"message":{"text":"/hi dummy_email@dummy.com"}}`},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)

	assignNewWebhook = true

	response, err = apiGatewayHandler(context.Background(), &request{
		APIGatewayProxyRequest: &apigateway.Request{Body: `{"message":{"text":"/hi dummy_email@dummy.com"}}`},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)
}

func TestApiGatewayHandlerAuthorizedCommand(t *testing.T) {
	c := require.New(t)

	setMockedClient()

	defer deactivateMockedClient()

	storage.InitDynamoMock()
	sns.InitSNSMock()

	err := storage.SaveUser(context.Background(), models.From{Email: "dummy_email@dummy.com", ID: int64(000)})
	c.NoError(err)

	// when user is not verified
	response, err := apiGatewayHandler(context.Background(), &request{
		APIGatewayProxyRequest: &apigateway.Request{Body: `{"message":{"text":"/deployTerraformStaging -b master checks/core","id":"000"}}`},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)

	err = storage.VerifyUserEmail(context.Background(), "dummy_email@dummy.com")
	c.NoError(err)

	response, err = apiGatewayHandler(context.Background(), &request{
		APIGatewayProxyRequest: &apigateway.Request{Body: `{"message":{"text":"/deployTerraformStaging -b master checks/core","id":"000"}}`},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)

	storage.ActiveForceFailure()

	defer storage.DeactiveForceFailure()

	response, err = apiGatewayHandler(context.Background(), &request{
		APIGatewayProxyRequest: &apigateway.Request{Body: `{"message":{"text":"/deployTerraformStaging -b master checks/core","id":"000"}}`},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)
}

func TestApiGatewayHandlerErrorInit(t *testing.T) {
	c := require.New(t)

	setMockedClient()

	defer deactivateMockedClient()

	response, err := apiGatewayHandler(context.Background(), &request{
		APIGatewayProxyRequest: &apigateway.Request{},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)
}

func TestApiGatewayHandlerErrorUnmarshall(t *testing.T) {
	c := require.New(t)

	setMockedClient()

	defer deactivateMockedClient()

	response, err := apiGatewayHandler(context.Background(), &request{
		APIGatewayProxyRequest: &apigateway.Request{Body: "{-}"},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)
}

func TestApiGatewayHandlerErrorGettingCommands(t *testing.T) {
	c := require.New(t)

	cache.InitMock()
	sns.InitSNSMock()
	setMockedClient()

	defer deactivateMockedClient()

	ctx := context.Background()
	log := logger.New("test")
	buf := bytes.NewBufferString("")
	log.Output = buf

	ctx = logger.Set(ctx, log)

	response, err := apiGatewayHandler(ctx, &request{
		APIGatewayProxyRequest: &apigateway.Request{Body: `{"message":null}`},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)

	output := buf.String()
	c.NotContains(output, "process_router_request_failed")

	buf.Reset()

	response, err = apiGatewayHandler(ctx, &request{
		APIGatewayProxyRequest: &apigateway.Request{Body: `{"message":{"text":""}}`},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)

	output = buf.String()
	c.NotContains(output, "process_router_request_failed")

	buf.Reset()

	response, err = apiGatewayHandler(ctx, &request{
		APIGatewayProxyRequest: &apigateway.Request{Body: `{"message":{"text":"  "}}`},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)

	output = buf.String()
	c.Contains(output, "process_router_request_failed")
	c.Contains(output, ErrMissingArgs.Error())
}

func TestApiGatewayHandlerUseCachedConversationData(t *testing.T) {
	c := require.New(t)

	cache.InitMock()
	sns.InitSNSMock()
	setMockedClient()

	defer deactivateMockedClient()

	message := models.Message{
		From: models.From{
			ID: 0,
		},
	}
	conversationState := &models.ConversationState{
		Data: "123",
	}

	err := conversation.StoreConversationState(context.Background(), message, conversationState)
	c.NoError(err)

	ctx := context.Background()
	log := logger.New("test")
	buf := bytes.NewBufferString("")
	log.Output = buf

	ctx = logger.Set(ctx, log)

	response, err := apiGatewayHandler(ctx, &request{
		APIGatewayProxyRequest: &apigateway.Request{Body: `{"message":{"text":""}}`},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)

	output := buf.String()
	c.NotContains(output, "process_router_request_failed")

	expectedError := errors.New("expected err")
	oldGetConversationState := getConversationState
	getConversationState = func(ctx context.Context, message models.Message) (*models.ConversationState, error) {
		return nil, expectedError
	}

	defer func() {
		getConversationState = oldGetConversationState
	}()

	buf.Reset()

	response, err = apiGatewayHandler(ctx, &request{
		APIGatewayProxyRequest: &apigateway.Request{Body: `{"message":{"text":""}}`},
	})
	c.NoError(err)
	c.Equal(http.StatusOK, response.StatusCode)

	output = buf.String()
	c.Contains(output, "process_router_request_failed")
	c.Contains(output, expectedError.Error())
}

func setMockedClient() {
	secrets.InitSecretsMock()
	secrets.SetMockedSecret("betty-bot-token", "token")

	client.ActivateMock()

	client.AddMockedResponse(http.MethodPost, "https://api.telegram.org/bottoken/getMe", http.StatusOK, `{"ok": true}`)
	client.AddMockedResponse(http.MethodPost, "https://api.telegram.org/bottoken/setWebhook", http.StatusOK, `{"ok": true}`)
	client.AddMockedResponse(http.MethodPost, "https://api.telegram.org/bottoken/sendMessage", http.StatusOK, `{"ok": true}`)
}

func deactivateMockedClient() {
	client.DeactivateMock()
	secrets.DeactivateMock()
}
