package models

import (
	"time"

)

// CallbackMessage received from Telegram when a callback is triggered
type CallbackMessage struct {
	ID             string
	From           From
	Message        Message
	Data           string
	Command        string
	AdditionalData map[string]string
}

// Message is the entire information about a message
type Message struct {
	ID     int            `json:"message_id"`
	Text   string         `json:"text"`
	Images []*PhotoUpload `json:"photo"`
	From   From
	Chat   Chat
}

// PhotoUpload is the different sizes for an uploaded image
type PhotoUpload struct {
	FileID string `json:"file_id"`
}

// From is where the message is coming
type From struct {
	ID             int64             `json:"id"`
	IsBot          bool              `json:"is_bot"`
	Email          string            `json:"email"`
	FirstName      string            `json:"first_name"`
	LastName       string            `json:"last_name"`
	Username       string            `json:"username"`
	EmailVerified  bool              `json:"email_verified"`
	ExpirationTime int64             `json:"expiration_time"`
	CreationDate   time.Time         `json:"creation_date"`
	PhoneNumber    string            `json:"phone_number"`
}

// Chat contains information about chat
type Chat struct {
	ID          int64  `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	Type        string `json:"type"`
}
