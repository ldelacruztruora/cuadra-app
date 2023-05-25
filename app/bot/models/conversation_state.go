package models

// ConversationState data stored in Betty cache to follow up conversations
type ConversationState struct {
	Command        string            `json:"command"`
	Data           string            `json:"data"`
	AdditionalData map[string]string `json:"additional_data"`
}
