package httpserver

import (
	"fmt"
	"strings"

	"github.com/solum-sp/aps-be-common/common/logger"

	"github.com/gin-gonic/gin"

)

type HttpServerConfig struct {
	Host             string   `env:"HTTP_HOST" envDefault:"localhost"`
	Port             int      `env:"HTTP_PORT" envDefault:"8020"`
	AllowOrigins     []string `env:"HTTP_CORS_ALLOW_ORIGINS" envSeparator:"," envDefault:"*"`
	AllowMethods     []string `env:"HTTP_CORS_ALLOW_METHODS" envSeparator:"," envDefault:"GET,POST,PUT,DELETE,OPTIONS"`
	AllowHeaders     []string `env:"HTTP_CORS_ALLOW_HEADERS" envSeparator:"," envDefault:"Origin,Content-Type,Authorization"`
	AllowCredentials bool     `env:"HTTP_CORS_ALLOW_CREDENTIALS" envDefault:"false"`
}
type RouteRegistrar interface {
	Register(server *HTTPServer, group *gin.RouterGroup)
}
type RouteConfig struct {
	Method      string
	Path        string
	Handler     gin.HandlerFunc
	Description string
	Params      string
	Middlewares []gin.HandlerFunc
}


var DefaultConfig = HttpServerConfig{
	Host: "localhost",
	Port: 8080,
}

type HTTPServer struct {
	config HttpServerConfig
	logger logger.ILogger
	router *gin.Engine
	registrars map[string][]RouteRegistrar
	globalMiddlewares []gin.HandlerFunc
}

type Option func(*HTTPServer)

func NewHTTPServer(opts ...Option) *HTTPServer {
	gin.SetMode(gin.ReleaseMode)
	cfg := HttpServerConfig{}
	hs := &HTTPServer{
		config: cfg,
		router: gin.Default(),
	}

	// Apply functional options
	for _, opt := range opts {
		opt(hs)
	}

	if len(hs.globalMiddlewares) > 0 {
		hs.router.Use(hs.globalMiddlewares...)
	}
	// Final setup
	hs.SetupRouter()

	return hs
}

func (s *HTTPServer) Initialize() *HTTPServer {
	s.SetupRouter()
	return s
}

func (s *HTTPServer) SetupRouter() {
	for version, regs := range s.registrars {
		var group *gin.RouterGroup
		if version == "root" {
			group = s.router.Group("")
		} else {
			group = s.router.Group("/api/" + version) // just /v1, /v2
		}
		for _, reg := range regs {
			reg.Register(s, group)
		}
	}
}


func (s *HTTPServer) AddRouteV2(group *gin.RouterGroup, cfg RouteConfig) {
	handler := cfg.Handler
	if len(cfg.Middlewares) > 0 {
		handler = chainMiddlewares(cfg.Handler, cfg.Middlewares...)
	}
	params := cfg.Params
	if strings.TrimSpace(params) == "" {
		params = "no required params"
	}
	joinPaths := s.joinPaths(group.BasePath(), cfg.Path)
	if group == nil {
		s.router.Handle(cfg.Method, cfg.Path, handler)
		s.logger.Info("Route initialized", "Method", cfg.Method, "Path", cfg.Path, "Description", cfg.Description, "Params", params)
	} else {
		group.Handle(cfg.Method, cfg.Path, handler)
		s.logger.Info("Route initialized", "Method", cfg.Method, "Path", joinPaths, "Description", cfg.Description, "Params", params)
	}
}

func chainMiddlewares(final gin.HandlerFunc, middlewares ...gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, m := range middlewares {
			m(ctx)
			if ctx.IsAborted() {
				return
			}
		}
		final(ctx)
	}
}
func (s *HTTPServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	s.logger.Info(fmt.Sprintf("Starting HTTP server on address: %v...", addr))
	if err := s.router.Run(addr); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to start HTTP server: %s", err))
		return err
	}

	return nil
}

// === Optional configuration like logger, system config,.... ===
func WithLogger(logger logger.ILogger) Option {
	return func(s *HTTPServer) {
		s.logger = logger
	}
}

func WithConfig(c HttpServerConfig) Option {
	return func(s *HTTPServer) {
		if c.Host == "" && c.Port == 0 {
			s.config = DefaultConfig
		} else {
			s.config = c
		}

		// ✅ Ensure AllowOrigins & AllowHeaders are not nil, fallback to defaults
		if len(s.config.AllowOrigins) == 0 {
			s.config.AllowOrigins = DefaultConfig.AllowOrigins
		}
		if len(s.config.AllowHeaders) == 0 {
			s.config.AllowHeaders = DefaultConfig.AllowHeaders
		}
	}
}

func WithVersionedRegistrars(regs map[string][]RouteRegistrar) Option {
	return func(s *HTTPServer) {
		s.registrars = regs
	}
}

func WithGlobalMiddlewares(middlewares ...gin.HandlerFunc) Option {
	return func(s *HTTPServer) {
		s.globalMiddlewares = append(s.globalMiddlewares, middlewares...)
	}
}
func (s *HTTPServer) joinPaths(base, path string) string {
	if base == "" {
		return path
	}
	if path == "" {
		return base
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}