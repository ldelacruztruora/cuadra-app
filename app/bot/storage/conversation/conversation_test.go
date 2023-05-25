package conversation

import (
	"context"
	"errors"
	"shared/app/bot/models"
	"testing"

	"bitbucket.org/truora/scrap-services/shared/cache"
	"github.com/stretchr/testify/require"
)

func TestConversationState(t *testing.T) {
	c := require.New(t)

	cache.InitMock()

	message := models.Message{
		From: models.From{
			ID: 0,
		},
	}
	expectedConversationState := &models.ConversationState{
		Data:    "123",
		Command: "command",
	}

	conversationState, err := GetConversationState(context.Background(), message)
	c.Nil(conversationState)
	c.Equal(ErrConversationNotFound, err)

	err = StoreConversationState(context.Background(), message, expectedConversationState)
	c.NoError(err)

	conversationState, err = GetConversationState(context.Background(), message)
	c.Equal(expectedConversationState, conversationState)
	c.NoError(err)

	err = DeleteConversationState(context.Background(), message)
	c.NoError(err)

	conversationState, err = GetConversationState(context.Background(), message)
	c.Nil(conversationState)
	c.Equal(ErrConversationNotFound, err)
}

func TestStoreConversationStateError(t *testing.T) {
	c := require.New(t)

	cache.InitMock()

	expectedErr := errors.New("expected error")
	message := models.Message{
		From: models.From{
			ID: 0,
		},
	}

	oldMarshal := marshal
	marshal = func(v any) ([]byte, error) {
		return nil, expectedErr
	}

	t.Cleanup(func() {
		marshal = oldMarshal
	})

	err := StoreConversationState(context.Background(), message, nil)
	c.Equal(expectedErr, err)
}
