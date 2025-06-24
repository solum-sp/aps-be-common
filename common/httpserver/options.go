package httpserver

func WithHost(host string) Option {
	return func(s *HTTPServer) {
		s.config.Host = host
	}
}

func WithPort(port int) Option {
	return func(s *HTTPServer) {
		s.config.Port = port
	}
}

func WithAllowOrigins(origins []string) Option {
	return func(s *HTTPServer) {
		s.config.AllowOrigins = origins
	}
}

func WithAllowMethods(methods []string) Option {
	return func(s *HTTPServer) {
		s.config.AllowMethods = methods
	}
}

func WithAllowHeaders(headers []string) Option {
	return func(s *HTTPServer) {
		s.config.AllowHeaders = headers
	}
}

func WithAllowCredentials(enabled bool) Option {
	return func(s *HTTPServer) {
		s.config.AllowCredentials = enabled
	}
}
