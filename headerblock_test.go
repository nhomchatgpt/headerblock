package headerblock_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	tbua "github.com/nhomchatgpt/headerblock"
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
		cfg.Log = true
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

	// Test for requests with CF-IPCountry whitelist
	t.Run("WhitelistCfIpCountry", func(t *testing.T) {
		cfg := tbua.CreateConfig()
		cfg.WhitelistRequestHeaders = append(cfg.WhitelistRequestHeaders, tbua.HeaderConfig{
			Name:  "Cf-Ipcountry",
			Value: "VN",
		})

		p, err := tbua.New(context.Background(), noopHandler{}, cfg, pluginName)
		if err != nil {
			t.Fatalf("unexpected error during plugin creation: %v", err)
		}

		// Allowed: CF-IPCountry is VN
		reqVN := httptest.NewRequest(http.MethodGet, "/foobar", nil)
		reqVN.Header.Set("Cf-Ipcountry", "VN")
		rrVN := httptest.NewRecorder()

		log.Printf("Request Headers: %+v", reqVN.Header) // Log request headers
		p.ServeHTTP(rrVN, reqVN)

		if rrVN.Code != http.StatusTeapot {
			t.Fatalf("unexpected status: got %v, expected %v for header %v", rrVN.Code, http.StatusTeapot, reqVN.Header.Get("CF-IPCountry"))
		}

		// Blocked: CF-IPCountry is FR
		reqFR := httptest.NewRequest(http.MethodGet, "/foobar", nil)
		reqFR.Header.Set("Cf-Ipcountry", "FR")
		rrFR := httptest.NewRecorder()
		p.ServeHTTP(rrFR, reqFR)

		if rrFR.Code != http.StatusForbidden {
			t.Fatalf("unexpected status: got %v, expected %v for header %v", rrFR.Code, http.StatusForbidden, reqFR.Header.Get("CF-IPCountry"))
		}

		// Test request with no headers
		req := httptest.NewRequest(http.MethodGet, "/foobar", nil)
		rr := httptest.NewRecorder()
		p.ServeHTTP(rr, req)

		// Expect 403 Forbidden for requests with no headers
		if rr.Code != http.StatusForbidden {
			t.Fatalf("unexpected status: got %v, expected %v for request with no headers", rr.Code, http.StatusTeapot)
		}
	})

	// Test for blocking Transfer-Encoding: chunked header
	t.Run("BlockTransferEncodingChunked", func(t *testing.T) {
		cfg := tbua.CreateConfig()
		cfg.RequestHeaders = append(cfg.RequestHeaders, tbua.HeaderConfig{
			Name:  "Transfer-Encoding",
			Value: "chunked",
		})

		p, err := tbua.New(context.Background(), noopHandler{}, cfg, pluginName)
		if err != nil {
			t.Fatalf("unexpected error during plugin creation: %v", err)
		}

		// Blocked: Transfer-Encoding is chunked
		req := httptest.NewRequest(http.MethodPost, "/foobar", nil)
		req.Header.Set("Transfer-Encoding", "chunked")
		rr := httptest.NewRecorder()
		p.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("unexpected status: got %v, expected %v for header %v", rr.Code, http.StatusForbidden, req.Header.Get("Transfer-Encoding"))
		}

		// Allowed: No Transfer-Encoding header
		reqNoTE := httptest.NewRequest(http.MethodPost, "/foobar", nil)
		rrNoTE := httptest.NewRecorder()
		p.ServeHTTP(rrNoTE, reqNoTE)

		if rrNoTE.Code != http.StatusTeapot {
			t.Fatalf("unexpected status: got %v, expected %v for request without Transfer-Encoding", rrNoTE.Code, http.StatusTeapot)
		}
	})
}
