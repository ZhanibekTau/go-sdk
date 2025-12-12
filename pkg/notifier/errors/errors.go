package errors

import "fmt"

type AppError struct {
	Service string
	Method  string
	Message string
	Env     string
	Err     error
}

func (a *AppError) Error() string {
	if a.Err != nil {
		return fmt.Sprintf("[%s] %s.%s: %s | %v", a.Env, a.Service, a.Method, a.Message, a.Err)
	}

	return fmt.Sprintf("[%s] %s.%s: %s", a.Env, a.Service, a.Method, a.Message)
}
