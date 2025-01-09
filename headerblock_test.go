package headerblock_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	tbua "github.com/wzator/headerblock"
)

const pluginName = "headerBlock"

type noopHandler struct{}

func (n noopHandler) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	rw.WriteHeader(http.StatusTeapot)
}

func TestPlugin(t *testing.T) {
	// Test for requests without any User-Agent headers
	t.Run("NoUserAgents", func(t *testing.T) {
		cfg := tbua.CreateConfig()
		p, err := tbua.New(context.Background(), noopHandler{}, cfg, pluginName)
		if err != nil {
			t.Fatalf("unexpected error during plugin creation: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/foobar", nil)
		rr := httptest.NewRecorder()
		p.ServeHTTP(rr, req)

		if rr.Code != http.StatusTeapot {
			t.Fatalf("unexpected status: got %v, expected %v", rr.Code, http.StatusTeapot)
		}
	})

	// Test for requests with a non-blocked User-Agent header
	t.Run("ValidUserAgent", func(t *testing.T) {
		cfg := tbua.CreateConfig()
		cfg.RequestHeaders = append(cfg.RequestHeaders, tbua.HeaderConfig{
			Name:  "User-Agent",
			Value: "SpamBot",
		})

		p, err := tbua.New(context.Background(), noopHandler{}, cfg, pluginName)
		if err != nil {
			t.Fatalf("unexpected error during plugin creation: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/foobar", nil)
		req.Header.Set("User-Agent", "ValidUserAgent")
		rr := httptest.NewRecorder()
		p.ServeHTTP(rr, req)

		if rr.Code != http.StatusTeapot {
			t.Fatalf("unexpected status: got %v, expected %v for header %v", rr.Code, http.StatusTeapot, req.Header.Get("User-Agent"))
		}
	})

	// Test for requests with a blocked User-Agent header
	t.Run("ForbiddenUserAgent", func(t *testing.T) {
		cfg := tbua.CreateConfig()
		cfg.RequestHeaders = append(cfg.RequestHeaders, tbua.HeaderConfig{
			Name:  "User-Agent",
			Value: "Googlebot",
		})

		p, err := tbua.New(context.Background(), noopHandler{}, cfg, pluginName)
		if err != nil {
			t.Fatalf("unexpected error during plugin creation: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/foobar", nil)
		req.Header.Set("User-Agent", "Googlebot")
		rr := httptest.NewRecorder()
		p.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("unexpected status: got %v, expected %v for header %v", rr.Code, http.StatusForbidden, req.Header.Get("User-Agent"))
		}
	})
}
