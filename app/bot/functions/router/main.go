package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"app/bot/models"

	"bitbucket.org/truora/scrap-services/devops/bot/shared/handler"
	"bitbucket.org/truora/scrap-services/devops/bot/storage/conversation"
	"bitbucket.org/truora/scrap-services/logger"
	"bitbucket.org/truora/scrap-services/shared/apigateway"
	"bitbucket.org/truora/scrap-services/shared/cache"
	"bitbucket.org/truora/scrap-services/shared/env"
	"bitbucket.org/truora/scrap-services/shared/sns"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	awssns "github.com/aws/aws-sdk-go/service/sns"
)

const (
	bettyBotUserName      = "@bettyabot"
	commandPrefix         = "/"
	defaultConversationID = "conversation"
)

var (
	// ErrMessageEmpty when message and command is empty
	ErrMessageEmpty = errors.New("message and command is empty")
	// ErrMissingArgs when arguments are requiered but they are missing
	ErrMissingArgs = errors.New("missing arguments to execute command")
	// ErrInvalidCommand invalid command received
	ErrInvalidCommand = errors.New("invalid command received")
	// ErrInvalidCallback when the callback message data is invalid
	ErrInvalidCallback = errors.New("invalid callback message")
	// ErrUnknownTelegramEvent when the event received from Telegram is not supported
	ErrUnknownTelegramEvent = errors.New("unknown telegram event")

	defaultLogger = logger.New("bot")

	assignNewWebhook = env.GetBool("ASSIGN_WEBHOOK", false)
	botCommandsTopic = env.GetString("BOT_COMMANDS_TOPIC", "arn:aws:sns:us-east-1:031975712270:bot-commands-topic")

	getConversationState = conversation.GetConversationState
)

type event struct {
	Message         *models.Message         `json:"message"`
	CallbackMessage *models.CallbackMessage `json:"callback_query"`
}

type request struct {
	*events.APIGatewayProxyRequest
	logger       *logger.Logger
	startingTime time.Time
	err          error
}

func (req *request) init(ctx context.Context) {
	req.startingTime = time.Now()
	req.logger = logger.Get(ctx)
}

func (req *request) finish(ctx context.Context) {
	req.logger.LogLambdaTime(ctx, logger.APIGatewayObject(req.APIGatewayProxyRequest), req.startingTime, req.err, recover())
}

func (req *request) process(ctx context.Context) error {
	event, telegramClient, err := req.createEventAndClient(ctx)
	if err != nil {
		return err
	}

	message, err := getMessage(ctx, event)
	if err != nil {
		return err
	}

	command, err := getCommand(message)
	if errors.Is(err, ErrMessageEmpty) || errors.Is(err, ErrInvalidCommand) {
		if message.Data != "" {
			return nil
		}

		command, err = setCacheCallbackData(ctx, message)
		if err != nil {
			telegramClient.Logger.Error(ctx, "set_conversation_data_failed", logger.OneMonth, []logger.Object{logger.ErrObject(err)})
			telegramClient.SendText(ctx, message.From.ID, "Cannot continue conversation, please try sending a command")

			return err
		}
	}

	if err != nil {
		telegramClient.Logger.Error(ctx, "getting_command_failed", logger.OneMonth, []logger.Object{logger.ErrObject(err)})
		telegramClient.SendText(ctx, message.From.ID, "Please add arguments to the command  ")

		return err
	}

	err = sendSNS(ctx, command, message)
	if err != nil {
		return fmt.Errorf("error sending SNS message %w", err)
	}

	return nil
}

func (req *request) createEventAndClient(ctx context.Context) (*event, *handler.TelegramClient, error) {
	event := &event{}

	err := json.Unmarshal([]byte(req.APIGatewayProxyRequest.Body), event)
	if err != nil {
		defaultLogger.Error(ctx, "unmarshal_failed", logger.OneMonth, []logger.Object{logger.ErrObject(err)})

		return nil, nil, err
	}

	telegramClient, err := createTelegramClient(ctx)
	if err != nil {
		defaultLogger.Error(ctx, "create_telegram_client_failed", logger.OneMonth, []logger.Object{logger.ErrObject(err)})

		return nil, nil, err
	}

	return event, telegramClient, nil
}

func createTelegramClient(ctx context.Context) (*handler.TelegramClient, error) {
	if assignNewWebhook {
		return handler.NewTelegramClientWithWebhook(ctx)
	}

	return handler.NewTelegramClient(ctx)
}

func getMessage(ctx context.Context, event *event) (*models.CallbackMessage, error) {
	if event.Message != nil {
		return &models.CallbackMessage{Message: *event.Message, From: event.Message.From}, nil
	}

	if event.CallbackMessage == nil {
		return &models.CallbackMessage{}, ErrUnknownTelegramEvent
	}

	callbackDataID := strings.Split(event.CallbackMessage.Data, " ")
	if len(callbackDataID) != 2 {
		return &models.CallbackMessage{}, ErrInvalidCallback
	}

	var err error

	event.CallbackMessage.Command = callbackDataID[0]
	event.CallbackMessage.Data = callbackDataID[1]

	if strings.HasPrefix(event.CallbackMessage.Data, "BOT") {
		event.CallbackMessage.Data, err = cache.Get(ctx, callbackDataID[1])
	}

	return event.CallbackMessage, err
}

func getCommand(message *models.CallbackMessage) (string, error) {
	if message.Command != "" {
		return message.Command, nil
	}

	text := message.Message.Text
	if text == "" {
		return "", ErrMessageEmpty
	}

	commandAndArgs := strings.Split(text, " ")
	if len(commandAndArgs) == 0 || commandAndArgs[0] == "" {
		return "", ErrMissingArgs
	}

	command := commandAndArgs[0]
	if !strings.HasPrefix(text, command) {
		return "", ErrInvalidCommand
	}

	cmdTrimmed := strings.TrimSuffix(strings.ToLower(command), bettyBotUserName)
	cmd := strings.TrimPrefix(strings.ToLower(cmdTrimmed), "/")

	return cmd, nil
}

func setCacheCallbackData(ctx context.Context, message *models.CallbackMessage) (string, error) {
	conversationState, err := getConversationState(ctx, message.Message)
	if errors.Is(err, conversation.ErrConversationNotFound) {
		return "", nil
	}

	if err != nil {
		defaultLogger.Error(ctx, "get_conversation_state_failed", logger.OneMonth, []logger.Object{logger.ErrObject(err)})

		return "", err
	}

	message.ID = defaultConversationID
	message.Command = conversationState.Command
	message.Data = conversationState.Data
	message.AdditionalData = conversationState.AdditionalData

	return conversationState.Command, nil
}

func sendSNS(ctx context.Context, command string, message *models.CallbackMessage) error {
	messageAttributes := make(map[string]*awssns.MessageAttributeValue)

	messageAttributes[command] = &awssns.MessageAttributeValue{ // cmd
		DataType:    aws.String("String"),
		StringValue: aws.String(command),
	}

	input := awssns.PublishInput{
		TopicArn:          aws.String(botCommandsTopic),
		MessageAttributes: messageAttributes,
	}

	return sns.PublishJSONWithInput(ctx, input, message, false)
}

func apiGatewayHandler(ctx context.Context, req *request) (*apigateway.Response, error) {
	req.init(ctx)

	defer req.finish(ctx)

	err := req.process(ctx)
	if err != nil && !errors.Is(err, ErrUnknownTelegramEvent) {
		req.err = err

		req.logger.Error(ctx, "process_router_request_failed", logger.OneMonth, []logger.Object{logger.ErrObject(req.err)})
	}

	return &apigateway.Response{StatusCode: http.StatusOK}, nil
}

func main() {
	defaultLogger.Must(context.Background(), cache.InitFromEnv(), logger.OneDay)
	lambda.Start(apiGatewayHandler)
}
