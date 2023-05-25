package conversation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"shared/app/bot/models"
	"time"

	"bitbucket.org/truora/scrap-services/shared/cache"
	"bitbucket.org/truora/scrap-services/shared/env"
)

const botConversationDataKey = "BOT-CONVERSATION-DATA:%d"

var (
	conversationMinutes        = env.GetInt64("CONVERSATION_MINUTES", 15)
	conversationExpirationTime = time.Duration(conversationMinutes) * time.Minute

	// ErrConversationNotFound when a conversation is not found in the cache
	ErrConversationNotFound = errors.New("conversation not found")

	marshal = json.Marshal
)

// GetConversationState returns the conversation data stored in cache
func GetConversationState(ctx context.Context, message models.Message) (*models.ConversationState, error) {
	rawData, err := cache.Get(ctx, fmt.Sprintf(botConversationDataKey, message.From.ID))
	if errors.Is(err, cache.ErrKeyNotExists) {
		return nil, ErrConversationNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("read cache data failed: %w", err)
	}

	conversationState := &models.ConversationState{}

	err = json.Unmarshal([]byte(rawData), conversationState)
	if err != nil {
		return nil, fmt.Errorf("cache data json unmarshal error: %w", err)
	}

	return conversationState, nil
}

// StoreConversationState stores the conversation data in cache
func StoreConversationState(ctx context.Context, message models.Message, conversationState *models.ConversationState) error {
	rawData, err := marshal(conversationState)
	if err != nil {
		return err
	}

	return cache.Add(ctx, fmt.Sprintf(botConversationDataKey, message.From.ID), string(rawData), conversationExpirationTime)
}

// DeleteConversationState deletes the conversation data in cache
func DeleteConversationState(ctx context.Context, message models.Message) error {
	return cache.Del(ctx, fmt.Sprintf(botConversationDataKey, message.From.ID))
}
