package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimiter(t *testing.T) {
	makeRequest := func(handler http.Handler, remoteAddr string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = remoteAddr
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		return rr
	}

	t.Run("single request allowed", func(t *testing.T) {
		rl := NewRateLimiter(rate.Limit(1), 1)
		handler := rl.RateLimitMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rr := makeRequest(handler, "192.0.2.1:1234")
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("exceeding rate limit", func(t *testing.T) {
		rl := NewRateLimiter(rate.Limit(1), 1)
		handler := rl.RateLimitMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// First request should succeed
		rr1 := makeRequest(handler, "192.0.2.2:1234")
		assert.Equal(t, http.StatusOK, rr1.Code)

		// Second immediate request should fail
		rr2 := makeRequest(handler, "192.0.2.2:1234")
		assert.Equal(t, http.StatusTooManyRequests, rr2.Code)
	})

	t.Run("different IPs independent limits", func(t *testing.T) {
		rl := NewRateLimiter(rate.Limit(1), 1)
		handler := rl.RateLimitMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Both requests from different IPs should succeed
		rr1 := makeRequest(handler, "192.0.2.3:1234")
		rr2 := makeRequest(handler, "192.0.2.4:1234")

		assert.Equal(t, http.StatusOK, rr1.Code)
		assert.Equal(t, http.StatusOK, rr2.Code)
	})

	t.Run("rate recovery", func(t *testing.T) {
		rl := NewRateLimiter(rate.Limit(2), 1)
		handler := rl.RateLimitMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// First request succeeds
		rr1 := makeRequest(handler, "192.0.2.5:1234")
		assert.Equal(t, http.StatusOK, rr1.Code)

		// Second immediate request fails
		rr2 := makeRequest(handler, "192.0.2.5:1234")
		assert.Equal(t, http.StatusTooManyRequests, rr2.Code)

		// Wait for rate limit to reset
		time.Sleep(501 * time.Millisecond)

		// Third request should succeed
		rr3 := makeRequest(handler, "192.0.2.5:1234")
		assert.Equal(t, http.StatusOK, rr3.Code)
	})

	t.Run("burst handling", func(t *testing.T) {
		rl := NewRateLimiter(rate.Limit(1), 3)
		handler := rl.RateLimitMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Should allow burst of 3 requests
		for i := 0; i < 3; i++ {
			rr := makeRequest(handler, "192.0.2.6:1234")
			assert.Equal(t, http.StatusOK, rr.Code)
		}

		// Fourth request should fail
		rr := makeRequest(handler, "192.0.2.6:1234")
		assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	})
}
