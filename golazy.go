// Package golazy is the go framework or building web applications
package lazyapp

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"runtime"

	"golazy.dev/layerfs"
	"golazy.dev/lazyassets"
	"golazy.dev/lazycontext"
	"golazy.dev/lazydispatch"
	"golazy.dev/lazyhttp"
	"golazy.dev/lazyservice"
	"golazy.dev/lazyview"
	"golazy.dev/lazyview/engines/raw"
	"golazy.dev/lazyview/engines/tpl"
)

// GoLazyApp it's the glue of all golazy modules
type GoLazyApp struct {
	LazyService  lazyservice.Manager
	LazyHTTP     lazyhttp.HTTPService
	LazyAssets   lazyassets.Server
	LazyView     lazyview.Views
	LazyDispatch *lazydispatch.Dispatcher
}

var nameKey = "app.name"
var versionKey = "app.version"

// New creates a new GolazyApp instance
// See [golazy.dev/lazyapp.New]
func New(name, version string) *GoLazyApp {
	s := lazyservice.New()
	s.AddValue(nameKey, name).AddValue(versionKey, version)
	return (&GoLazyApp{
		LazyService: s,
	}).init()
}

// NewWithContext creates a new GolazyApp instance with the provided context.
// See [golazy.dev/lazyapp.NewWithContext]
func NewWithContext(ctx context.Context, name, version string) *GoLazyApp {
	srv := lazyservice.NewWithContext(ctx)
	srv.AddValue(nameKey, name).AddValue(versionKey, version)
	return (&GoLazyApp{
		LazyService: srv,
	}).init()
}

func (b *GoLazyApp) init() *GoLazyApp {
	lazycontext.Set(b.LazyService, &b.LazyService)

	// Views
	b.LazyView.Engines = map[string]lazyview.Engine{
		"tpl": &tpl.Engine{},
		"txt": &raw.Engine{},
	}
	b.LazyView.FS = layerfs.New()
	lazycontext.Set(b.LazyService, &b.LazyView)

	// Dispatcher
	b.LazyDispatch = lazydispatch.New()
	lazycontext.Set(b.LazyService, b.LazyDispatch)

	// Assets
	b.LazyAssets.Storage = &lazyassets.Storage{}
	b.LazyAssets.NextHandler = b.LazyDispatch
	lazycontext.Set(b.LazyService, &b.LazyAssets)
	lazycontext.Set(b.LazyService, b.LazyAssets.Storage)

	// Server
	b.LazyHTTP.Addr = ":2000"
	b.LazyHTTP.Handler = &b.LazyAssets
	b.LazyService.AddService(&b.LazyHTTP)

	return b
}

// AddService adds a service to the app
// See [golazy.dev/lazyapp.AddService]
func (b *GoLazyApp) AddService(srv lazyservice.Service) {
	b.LazyService.AddService(srv)
}

// Draw draws the http routes of the server
// See [golazy.dev/lazydispatch.Draw]
func (b *GoLazyApp) Draw(fn func(r *lazydispatch.Scope)) *lazydispatch.Scope {
	return b.LazyDispatch.Draw(fn)
}

// Use adds a middleware to the server
// See [golazy.dev/lazydispatch.Use]
func (b *GoLazyApp) Use(middleware func(http.Handler) http.Handler) {
	b.LazyDispatch.Use(middleware)
}

// Public adds all the public files and assets
// It expects all the public files to be in a directory called "public"
// See [golazy.dev/lazyassets.AddFS]
func (b *GoLazyApp) Public(fs fs.FS) {
	b.LazyAssets.AddFS(dir(fs, "public"))
}

// Views adds all the views
// It expects all the views to be in a directory called "views"
// See [golazy.dev/lazyview.Views] and [golazy.dev/layerfs.FS.Add]
func (b *GoLazyApp) Views(fs fs.FS) {
	b.LazyView.FS.(*layerfs.FS).Add(dir(fs, "views"))
}

// Run will start the application
// See [golazy.dev/lazyapp.Run]
func (b *GoLazyApp) Run() error {
	return b.LazyService.Run()
}

// Start will start the application and return a channel with the error
func (b *GoLazyApp) Start() <-chan (error) {
	errCh := make(chan (error))
	go func() {
		errCh <- b.LazyService.Run()
	}()
	runtime.Gosched()
	return errCh

}

func dir(files fs.FS, dir string) fs.FS {
	files, err := fs.Sub(files, dir)
	if err != nil {
		fmt.Printf("Error: Subdirectory %s not found: %s\n", dir, err.Error())
		os.Exit(-1)
	}
	return files
}
