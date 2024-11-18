package gin

import (
	"encoding/json"
	"fmt"
	"github.com/ZhanibekTau/go-sdk/pkg/config"
	"github.com/ZhanibekTau/go-sdk/pkg/constants"
	"github.com/ZhanibekTau/go-sdk/pkg/exception"
	"github.com/ZhanibekTau/go-sdk/pkg/gin/validation"
	"github.com/ZhanibekTau/go-sdk/pkg/http/helpers"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-errors/errors"
	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	timeout "github.com/vearne/gin-timeout"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"net/http"
	"strconv"
	"time"
)

// InitRouter Базовая инициализация gin
func InitRouter(baseConfig *config.BaseConfig) *gin.Engine {
	if baseConfig.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Options
	router := gin.New()
	prefix := baseConfig.SwaggerPrefix

	if prefix == "" {
		prefix = baseConfig.Name
	}

	if prefix == "" {
		prefix = "swagger"
	}

	router.GET("/"+prefix+"/api-docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "404 page not found"})
	})
	router.HandleMethodNotAllowed = true
	p := ginprometheus.NewPrometheus("ginHelpers")
	p.Use(router)
	router.Use(sentrygin.New(sentrygin.Options{}))
	//router.Use(gin.Logger())
	router.Use(timeout.Timeout(timeout.WithTimeout(time.Duration(baseConfig.HandlerTimeout) * time.Second)))
	router.Use(gin.CustomRecovery(ErrorHandler))

	return router
}

// ErrorHandler Обработчик ошибок gin
func ErrorHandler(c *gin.Context, err any) {
	goErr := errors.Wrap(err, 2)
	details := make([]string, 0)

	for _, frame := range goErr.StackFrames() {
		details = append(details, frame.String())
	}

	sentry.CaptureException(goErr)
	c.JSON(http.StatusInternalServerError, gin.H{"message": goErr.Error(), "details": details, "success": false, "service_code": 0})
}

func Error(c *gin.Context, exception *exception.AppException) {
	c.Set("exception", exception)
	c.Status(exception.Code)
}

func Success(c *gin.Context, data any) {
	c.Set("data", data)
}

func SetAppInfo(c *gin.Context, baseConfig *config.BaseConfig) {
	c.Set("app_info", getInstanceAppInfo(c, baseConfig))
}

func getInstanceAppInfo(c *gin.Context, baseConfig *config.BaseConfig) *config.AppInfo {
	appInfo := &config.AppInfo{}
	setBaseDataToAppInfo(c, appInfo)
	appInfo.AppEnv = "UNKNOWN (maybe you not used RequestMiddleware)"
	appInfo.ServiceName = "UNKNOWN (maybe you not used RequestMiddleware)"

	if baseConfig != nil {
		appInfo.AppEnv = baseConfig.AppEnv
		appInfo.ServiceName = baseConfig.Name
	}

	return appInfo
}

func setBaseDataToAppInfo(c *gin.Context, appInfo *config.AppInfo) {
	var err error
	appInfo.RequestId = c.GetHeader(constants.RequestIdHeaderName)
	// если request id не пришел с заголовком, генерим его, чтобы прокидывать дальше при http запросах
	if appInfo.RequestId == "" {
		appInfo.GenerateRequestId()
		c.Request.Header.Add(constants.RequestIdHeaderName, appInfo.RequestId)
	}

	appInfo.LanguageCode = c.GetHeader(constants.LanguageHeaderName)

	if appInfo.LanguageCode == "" {
		appInfo.LanguageCode = constants.LangCodeRu
	}

	appInfo.CityId, _ = strconv.Atoi(c.GetHeader(constants.CityHeaderName))

	if appInfo.CityId == 0 {
		appInfo.CityId = 443
	}

	appInfo.UserId, err = strconv.Atoi(c.GetHeader(constants.UserHeaderName))

	if err != nil {
		appInfo.UserId = 0
	}

	appInfo.RequestUrl = c.Request.URL.Path
	appInfo.RequestMethod = c.Request.Method
	appInfo.RequestScheme = c.Request.URL.Scheme
	appInfo.RequestHost = c.Request.Host
}

func GetAppInfo(c *gin.Context) *config.AppInfo {
	value, exists := c.Get("app_info")

	if exists {
		appInfo := value.(*config.AppInfo)

		return appInfo
	}

	return getInstanceAppInfo(c, nil)
}

// ValidateRequestQuery - Валидация GET параметров HTTP реквеста
func ValidateRequestQuery(c *gin.Context, request validation.IRequest) bool {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		for n, f := range request.CustomValidationRules() {
			v.RegisterValidation(n, f)
		}
	}

	if err := c.BindQuery(request); err != nil {
		var ve validator.ValidationErrors

		if errors.As(err, &ve) {
			out := make(map[string]any, len(ve))

			for _, fe := range ve {
				msg := request.CustomValidationMessage(fe)

				if msg == fe.Tag() {
					msg = request.ValidationMessage(fe)
				}

				out[strcase.ToSnake(fe.Field())] = msg
			}

			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, errors.New("validation error"), out)

			return false
		}

		// Обработка ошибок unmarshal
		var unmarshalTypeError *json.UnmarshalTypeError

		if errors.As(err, &unmarshalTypeError) {
			out := make(map[string]any)
			out[strcase.ToSnake(unmarshalTypeError.Field)] = fmt.Sprintf("Invalid type expected %s but got %s", unmarshalTypeError.Type, unmarshalTypeError.Value)
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, errors.New("validation error"), out)

			return false
		}

		// Обработка ошибок unmarshal
		var parseNumTypeError *strconv.NumError

		if errors.As(err, &parseNumTypeError) {
			out := make(map[string]any)
			out[strcase.ToSnake(parseNumTypeError.Num)] = parseNumTypeError.Error()
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, parseNumTypeError, out)

			return false
		}

		helpers.ErrorResponse(c, http.StatusInternalServerError, err, nil)

		return false
	}

	return true
}

// ValidateRequestBody - Валидация тела HTTP реквеста
func ValidateRequestBody(c *gin.Context, request validation.IRequest) bool {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		for n, f := range request.CustomValidationRules() {
			v.RegisterValidation(n, f)
		}
	}

	if err := c.ShouldBind(&request); err != nil {
		var ve validator.ValidationErrors

		if errors.As(err, &ve) {
			out := make(map[string]any, len(ve))

			for _, fe := range ve {
				msg := request.CustomValidationMessage(fe)

				if msg == fe.Tag() {
					msg = request.ValidationMessage(fe)
				}

				out[strcase.ToSnake(fe.Field())] = msg
			}

			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, errors.New("validation error"), out)

			return false
		}

		// Обработка ошибок unmarshal
		var unmarshalTypeError *json.UnmarshalTypeError

		if errors.As(err, &unmarshalTypeError) {
			out := make(map[string]any)
			out[strcase.ToSnake(unmarshalTypeError.Field)] = fmt.Sprintf("Invalid type expected %s but got %s", unmarshalTypeError.Type, unmarshalTypeError.Value)

			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, errors.New("validation error"), out)

			return false
		}

		helpers.ErrorResponse(c, http.StatusInternalServerError, err, nil)

		return false
	}

	return true
}
