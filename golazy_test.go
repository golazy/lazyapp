package golazy

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestAppBuilder(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app := NewWithContext(ctx, "test", "1.0.0")
	app.AddAsset("index.html", []byte("Hello, World!"))

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
