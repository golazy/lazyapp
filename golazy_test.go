package lazyapp

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"golazy.dev/lazyapp"
	"golazy.dev/lazyassets"
	"golazy.dev/lazyview"
)

func TestAppBuilder(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app := NewWithContext(ctx, "test", "1.0.0")
	app.LazyAssets.AddFile("index.html", []byte("Hello, World!"))

	errCh := app.Start()

	resp, err := http.Get("http://localhost:2000/index.html")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "Hello, World!" {
		t.Errorf("Expected: Hello, World! Got: %s", string(body))
	}
	cancel()

	if err := <-errCh; err != nil {
		t.Errorf("Error: %v", err)
	}
}

func TestAppHasContexts(t *testing.T) {
	app := New("test", "1.0.0")

	if value := lazyapp.AppGet[*lazyassets.Storage](app.LazyApp); value == nil {
		t.Fatal("storage is nil")
	}

	if value := lazyapp.AppGet[*lazyassets.Server](app.LazyApp); value == nil {
		t.Fatal("lazyassets.Server is nil")
	}

	if value := lazyapp.AppGet[*lazyapp.LazyApp](app.LazyApp); value == nil {
		t.Fatal("*lazyapp.LazyApp is nil")
	}

	if value := lazyapp.AppGet[*lazyview.Views](app.LazyApp); value == nil {
		t.Fatal("lazyview.Views is nil")
	}

}
