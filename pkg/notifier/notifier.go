package notifier

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

func (t *TelegramNotifier) Send(service string, method string, message string, env string, err error, TLSSkipVerify bool) error {
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

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: TLSSkipVerify,
		},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send error: %w", err)
	}

	defer resp.Body.Close()

	return nil
}
