package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ZhanibekTau/go-sdk/pkg/config"
	"github.com/ZhanibekTau/go-sdk/pkg/constants"
	"github.com/ZhanibekTau/go-sdk/pkg/exception"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"log"
	"net/http"
	"strconv"
	"time"
)

func FormattedTextErrorResponse(c *gin.Context, statusCode int, message string, context map[string]any) {
	TextErrorResponse(c, statusCode, message, context)
	FormattedResponse(c)
}

func TextErrorResponse(c *gin.Context, statusCode int, message string, context map[string]any) {
	AppExceptionResponse(c, exception.NewAppException(statusCode, errors.New(message), context))
}

func FormattedErrorResponse(c *gin.Context, statusCode int, err error, context map[string]any) {
	ErrorResponse(c, statusCode, err, context)
	FormattedResponse(c)
}

func ErrorResponse(c *gin.Context, statusCode int, err error, context map[string]any) {
	AppExceptionResponse(c, exception.NewAppException(statusCode, err, context))
}

func FormattedAppExceptionResponse(c *gin.Context, exception *exception.AppException) {
	AppExceptionResponse(c, exception)
	FormattedResponse(c)
}

func AppExceptionResponse(c *gin.Context, exception *exception.AppException) {
	c.Set("exception", exception)
	c.Status(exception.Code)
}

func SuccessResponse(c *gin.Context, data any) {
	c.Set("data", data)
}

func SuccessCreatedResponse(c *gin.Context, data any) {
	c.Set("data", data)
	c.Set("status_code", http.StatusCreated)
}

func SuccessDeletedResponse(c *gin.Context, data any) {
	c.Set("data", data)
	c.Set("status_code", http.StatusNoContent)
}

func FormattedSuccessResponse(c *gin.Context, data any) {
	SuccessResponse(c, data)
	FormattedResponse(c)
}

func FormattedResponse(c *gin.Context) {
	start := time.Now()

	appExceptionObject, exists := c.Get("exception")
	fmt.Printf("%+v\n", appExceptionObject)

	if !exists {
		data, _ := c.Get("data")
		response := struct {
			Success bool        `json:"success"`
			Data    interface{} `json:"data"`
		}{
			true,
			data,
		}

		jsonBytes, err := json.Marshal(response)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal JSON"})

			return
		}

		c.Writer.Status()
		statusCode, ex := c.Get("status_code")

		if !ex {
			SetColors(c, http.StatusOK, start)
			c.Data(http.StatusOK, "application/json", jsonBytes)
		} else {
			SetColors(c, statusCode.(int), start)
			c.Data(statusCode.(int), "application/json", jsonBytes)
		}

		return
	}

	appException := exception.AppException{}
	mapstructure.Decode(appExceptionObject, &appException)
	fmt.Printf("%+v\n", appException)
	serviceName := "UNKNOWN (maybe you not used RequestMiddleware)"
	requestId := "UNKNOWN (maybe you not used RequestMiddleware)"
	value, exists := c.Get("app_info")

	if exists {
		appInfo := value.(*config.AppInfo)
		serviceName = appInfo.ServiceName
		requestId = appInfo.RequestId
	}

	responseData := gin.H{
		"status":       appException.Code,
		"error":        appException.GetErrorType(),
		"message":      appException.Error.Error(),
		"request_id":   requestId,
		"hostname":     serviceName,
		"service_code": appException.ServiceCode,
		"details":      appException.Context,
	}

	response := struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
	}{
		false,
		responseData,
	}

	jsonBytes, err := json.Marshal(response)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal JSON"})

		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		// Добавляем заголовки запроса
		mapHeaders := make(map[string]any)
		for key, values := range c.Request.Header {
			for _, value := range values {
				mapHeaders[fmt.Sprintf("header_%s", key)] = value
			}
		}
		scope.SetContext("header", mapHeaders)

		// Добавляем Query параметры
		mapQueries := make(map[string]any)
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				mapQueries[fmt.Sprintf("query_%s", key)] = value
			}
		}
		scope.SetContext("query", mapQueries)

		// Захватываем ошибку
		scope.SetContext("error", responseData)

		sentry.CaptureMessage("Http Status Code  - " + strconv.Itoa(appException.Code) + ". Url - " + c.Request.URL.Path)
	})

	SetColors(c, appException.Code, start)
	c.Data(appException.Code, "application/json", jsonBytes)
}

func SetColors(c *gin.Context, statusCode int, start time.Time) {
	methodColor := constants.Green
	if statusCode >= 400 && statusCode < 500 {
		methodColor = constants.Yellow
	} else if statusCode >= 500 {
		methodColor = constants.Red
	}

	log.Printf("%s[%s%s%s%s] \"%s\" - status - %s%d%s, size %d bytes in %v second%s",
		constants.Cyan,
		methodColor, c.Request.Method, constants.Reset,
		constants.Cyan, c.Request.URL.Path,
		methodColor, c.Status, constants.Reset,
		c.Request.ContentLength, time.Since(start).Seconds(), constants.Reset)
}
