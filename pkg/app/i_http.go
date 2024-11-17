package app

type IHttp interface {
	PrepareHttp(app *App) error
}
