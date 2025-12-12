package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ZhanibekTau/go-sdk/pkg/notifier/errors"
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

func (t *TelegramNotifier) Send(service string, method string, message string, env string, err error) error {
	appErr := &errors.AppError{
		Service: service,
		Method:  method,
		Message: message,
		Env:     env,
		Err:     err,
	}

	body := map[string]string{
		"chat_id": t.ChatID,
		"text":    appErr.Error(),
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body %v", err)
	}
	url := "https://api.telegram.org/bot" + t.Token + "/sendMessage"

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to send error %v", err)
	}

	defer resp.Body.Close()

	return nil
}
