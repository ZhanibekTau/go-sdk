package app

type IConsumer interface {
	PrepareConsumer(app *App) error
}
