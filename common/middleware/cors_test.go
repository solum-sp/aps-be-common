package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	t.Run("Default CORS options", func(t *testing.T) {
		opts := DefaultCORS()
		handler := CORSMiddleware(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
			rr.Header().Get("Access-Control-Allow-Headers"))
	})

	t.Run("Custom CORS options", func(t *testing.T) {
		opts := &CORSOptions{
			AllowedOrigins:     "http://example.com",
			AllowedCredentials: "false",
			AllowedMethods:     "GET, POST",
			AllowedHeaders:     "Content-Type",
		}
		handler := CORSMiddleware(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, "http://example.com", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "false", rr.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "GET, POST", rr.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type", rr.Header().Get("Access-Control-Allow-Headers"))
	})

	t.Run("OPTIONS preflight request", func(t *testing.T) {
		opts := DefaultCORS()
		handler := CORSMiddleware(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("OPTIONS", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Regular request passes through", func(t *testing.T) {
		called := false
		opts := DefaultCORS()
		handler := CORSMiddleware(opts)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.True(t, called)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
