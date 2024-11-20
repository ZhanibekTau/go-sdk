package middleware

import (
	"github.com/ZhanibekTau/go-sdk/pkg/config"
	gin2 "github.com/ZhanibekTau/go-sdk/pkg/gin"
	"github.com/gin-gonic/gin"
)

// RequestMiddleware Middleware заполняющий данные запроса
func RequestMiddleware(baseConfig *config.BaseConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		gin2.SetAppInfo(c, baseConfig)
		c.Next()
	}
}
