package middleware

import "net/http"

// CORSOptions defines allowed origins, headers, and methods
type CORSOptions struct {
	AllowedOrigins     string
	AllowedCredentials string
	AllowedMethods     string
	AllowedHeaders     string
}

// DefaultCORS returns a default CORS configuration
func DefaultCORS() *CORSOptions {
	return &CORSOptions{
		AllowedOrigins:     "*",
		AllowedCredentials: "true",
		AllowedMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowedHeaders:     "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	}
}

func CORSMiddleware(opts *CORSOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set allowed origins
			w.Header().Set("Access-Control-Allow-Origin", opts.AllowedOrigins)
			w.Header().Set("Access-Control-Allow-Credentials", opts.AllowedCredentials)
			w.Header().Set("Access-Control-Allow-Methods", opts.AllowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", opts.AllowedHeaders)

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
