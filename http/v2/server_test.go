package httpserver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpserver "github.com/brunoluiz/x/http/v2"
)

func TestServerImplementsHTTPHandlerAndRegistersRoutes(t *testing.T) {
	server := httpserver.New()
	server.HandleFunc("GET /ready", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ready", nil)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("response status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestRunShutsDownWhenContextIsCanceled(t *testing.T) {
	server := httpserver.New(httpserver.WithAddr("127.0.0.1", 0))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx)
	}()
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not stop after context cancellation")
	}
}
