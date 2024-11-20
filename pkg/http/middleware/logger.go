package middleware

import (
	"github.com/ZhanibekTau/go-sdk/pkg/exception"
	gin2 "github.com/ZhanibekTau/go-sdk/pkg/gin"
	"github.com/ZhanibekTau/go-sdk/pkg/logger"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

// LoggerMiddleware Middleware для логирования ответа и отправки ошибок в сентри
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		appInfo := gin2.GetAppInfo(c)

		for _, err := range c.Errors {
			sentry.CaptureException(err)
			logger.FormattedErrorWithAppInfo(appInfo, err.Error())
		}

		appExceptionObject, exists := c.Get("exception")

		if exists {
			appException := exception.AppException{}
			mapstructure.Decode(appExceptionObject, &appException)
			sentry.CaptureException(appException.Error)
			logger.FormattedErrorWithAppInfo(appInfo, appException.Error.Error())
		}
	}
}
