package notifier

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type TelegramNotifier struct {
	Token  string
	ChatID string
}

func NewTelegramNotifier(token, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		Token:  token,
		ChatID: chatID,
	}
}

func (t *TelegramNotifier) Send(message string) error {
	body := map[string]string{
		"chat_id": t.ChatID,
		"text":    message,
	}

	jsonBody, _ := json.Marshal(body)

	url := "https://api.telegram.org/bot" + t.Token + "/sendMessage"

	_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	return err
}
