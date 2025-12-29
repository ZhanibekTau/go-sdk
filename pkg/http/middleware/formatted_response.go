package middleware

import (
	"github.com/ZhanibekTau/go-sdk/pkg/http/helpers"
	"github.com/gin-gonic/gin"
)

// FormattedResponseMiddleware Middleware для обработки ответа
func FormattedResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.IsAborted() {
			return
		}
		
		helpers.FormattedResponse(c)
	}
}
