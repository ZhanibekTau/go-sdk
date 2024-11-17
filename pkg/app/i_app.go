package app

type IApp interface {
	PrepareComponents(app *App) error
	PrepareConfigs(app *App) error
}
