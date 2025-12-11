package logger

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ZhanibekTau/go-sdk/pkg/notifier"
	"github.com/ZhanibekTau/go-sdk/pkg/notifier/errors"
)

type Logger struct {
	Service  string
	Env      string
	Notifier *notifier.TelegramNotifier
}

func NewLogger(service, env string, notifier *notifier.TelegramNotifier) *Logger {
	return &Logger{
		Service:  service,
		Env:      env,
		Notifier: notifier,
	}
}

func (l *Logger) Error(method string, msg string, err error) {
	appErr := &errors.AppError{
		Service: l.Service,
		Method:  method,
		Message: msg,
		Env:     l.Env,
		Err:     err,
	}

	logData := map[string]any{
		"level":   "error",
		"service": l.Service,
		"env":     l.Env,
		"method":  method,
		"message": msg,
		"error":   errString(err),
		"time":    time.Now().Format(time.RFC3339),
	}

	jsonLog, _ := json.Marshal(logData)
	log.Println(string(jsonLog))

	if l.Notifier != nil {
		_ = l.Notifier.Send(appErr.Error())
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
