package app

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"go-sdk/pkg/config"
	ginHelper "go-sdk/pkg/gin"
	"go-sdk/pkg/tracer"
	"time"
)

func NewApp(appExt IApp) *App {
	return &App{
		AppExt: appExt,
	}
}

// App Приложение
type App struct {
	TraceClient *tracer.Tracer
	BaseConfig  *config.BaseConfig
	Router      *gin.Engine
	isInit      bool
	AppExt      interface{}
	Location    *time.Location
}

func (app *App) InitBaseConfig() (*config.BaseConfig, error) {
	baseConfig := &config.BaseConfig{HandlerTimeout: 30}
	err := config.InitConfig(baseConfig)

	if err != nil {
		return nil, err
	}

	spew.Dump(baseConfig)

	return baseConfig, nil
}

func (app *App) initConfig() error {
	envErr := config.ReadEnv()

	if envErr != nil {
		fmt.Println(envErr.Error())
	}

	baseConfig, err := app.InitBaseConfig()

	if err != nil {
		return err
	}

	app.BaseConfig = baseConfig

	if iApp, ok := app.AppExt.(IApp); ok {
		if cErr := iApp.PrepareConfigs(app); cErr != nil {
			return cErr
		}
	} else {
		fmt.Println("App does not implement IApp, skipping PrepareConfigs.")
	}

	return nil
}

func (app *App) initApp() error {
	if app.isInit {
		return nil
	}

	err := app.initConfig()

	if err != nil {
		return err
	}

	if app.BaseConfig.TimeZone != "" {
		location, lErr := time.LoadLocation(app.BaseConfig.TimeZone)

		if lErr != nil {
			return lErr
		}

		app.Location = location
	}

	tErr := app.initTraceClient()

	if tErr != nil {
		return tErr
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: app.BaseConfig.SentryDsn,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	if iApp, ok := app.AppExt.(IApp); ok {
		if cErr := iApp.PrepareComponents(app); cErr != nil {
			return cErr
		}
	} else {
		fmt.Println("App does not implement IApp, skipping PrepareComponents.")
	}

	app.isInit = true

	return nil
}

// RunHttp Запуск веб сервера
func (app *App) RunHttp() error {
	if !app.isInit {
		iErr := app.initApp()

		if iErr != nil {
			return iErr
		}
	}

	//инициализация ginHelpers
	app.Router = ginHelper.InitRouter(app.BaseConfig)

	if iHttp, ok := app.AppExt.(IHttp); ok {
		if cErr := iHttp.PrepareHttp(app); cErr != nil {
			return cErr
		}
	} else {
		fmt.Println("App does not implement IHttp, skipping PrepareHttp.")
	}

	//запускаем сервер
	gErr := app.Router.Run(app.BaseConfig.ServerAddress)

	if gErr != nil {
		return gErr
	}

	return nil
}

func (app *App) RunConsumer() error {
	if !app.isInit {
		iErr := app.initApp()

		if iErr != nil {
			return iErr
		}
	}

	if iConsumer, ok := app.AppExt.(IConsumer); ok {
		if pcErr := iConsumer.PrepareConsumer(app); pcErr != nil {
			return pcErr
		}
	} else {
		fmt.Println("App does not implement IConsumer, skipping PrepareConsumer.")
	}

	return nil
}

// InitTraceClient - инициализация трейсера
func (app *App) initTraceClient() error {
	traceClient, err := tracer.InitTraceClient()

	if err != nil {
		fmt.Println("Соединение с трассировкой - ошибка : ", err.Error())
	}

	app.TraceClient = traceClient

	return nil
}
