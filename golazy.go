package golazy

import (
	"context"
	"io/fs"
	"net/http"
	"runtime"

	"golazy.dev/layerfs"
	"golazy.dev/lazyapp"
	"golazy.dev/lazyassets"
	"golazy.dev/lazydispatch"
	"golazy.dev/lazyhttp"
	"golazy.dev/lazyview"
	"golazy.dev/lazyview/engines/raw"
	"golazy.dev/lazyview/engines/tpl"
)

type GolazyApp struct {
	App        lazyapp.LazyApp
	Server     lazyhttp.HttpService
	Assets     lazyassets.Server
	Views      lazyview.Views
	Dispatcher *lazydispatch.Dispatcher
}

func NewApp(name, version string) *GolazyApp {
	return (&GolazyApp{
		App: lazyapp.New(name, version),
	}).init()
}

func NewWithContext(ctx context.Context, name, version string) *GolazyApp {
	return (&GolazyApp{
		App: lazyapp.NewWithContext(ctx, name, version),
	}).init()
}
func (b *GolazyApp) init() *GolazyApp {

	// Views
	b.Views.Engines = map[string]lazyview.Engine{
		"tpl": &tpl.Engine{},
		"txt": &raw.Engine{},
	}
	b.Views.FS = layerfs.New()
	lazyapp.AppSet(b.App, &b.Views)

	// Dispatcher
	b.Dispatcher = lazydispatch.New()
	lazyapp.AppSet(b.App, b.Dispatcher)

	// Assets
	b.Assets.Storage = &lazyassets.Storage{}
	b.Assets.NextHandler = b.Dispatcher
	lazyapp.AppSet(b.App, &b.Assets)
	lazyapp.AppSet(b.App, &b.Assets.Storage)

	// Server
	b.Server.Addr = ":2000"
	b.Server.Handler = &b.Assets
	b.App.AddService(&b.Server)

	return b
}

func (b *GolazyApp) DrawRoutes(fn func(r *lazydispatch.Scope)) *lazydispatch.Scope {
	return b.Dispatcher.Draw(fn)
}

func (b *GolazyApp) Use(middleware func(http.Handler) http.Handler) {
	b.Dispatcher.Use(middleware)
}

func (b *GolazyApp) AddAssets(fs fs.FS) {
	b.Assets.AddFS(fs)
}
func (b *GolazyApp) AddAsset(path string, content []byte) {
	b.Assets.AddFile(path, content)
}

func (b *GolazyApp) AddViews(fs fs.FS) {
	b.Views.FS.(*layerfs.FS).Add(fs)
}

func (b *GolazyApp) Run() error {
	return b.App.Run()
}

func (b *GolazyApp) Start() <-chan (error) {
	errCh := make(chan (error))
	go func() {
		errCh <- b.App.Run()
	}()
	runtime.Gosched()
	return errCh

}
