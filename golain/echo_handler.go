package golain

import (
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/khvh/golain/telemetry"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// EchoRouter ...
type EchoRouter struct {
	app  *echo.Echo
	opts *AppRouterOptions
}

func newEchoRouter(opts *AppRouterOptions) AppRouter {
	r := &EchoRouter{
		app: echo.New(),
	}

	r.app.HideBanner = opts.Banner
	r.app.HidePort = opts.Banner

	r.opts = opts

	return r
}

func mapEchoCtxToGolainCtx(c echo.Context) *Ctx {
	headers := map[string]string{}
	params := map[string]string{}
	query := map[string]string{}
	bts := []byte{}

	for k := range c.Request().Header {
		headers[k] = c.Request().Header.Get(k)
	}

	for _, k := range c.ParamNames() {
		params[k] = c.Param(k)
	}

	for k := range c.QueryParams() {
		query[k] = c.QueryParam(k)
	}

	switch c.Request().Method {
	case http.MethodPost:
	case http.MethodPatch:
	case http.MethodPut:
		_, err := c.Request().Body.Read(bts)
		if err != nil {
			log.Trace().Err(err).Send()
		}
	}

	return NewCtx().
		SetHeaders(headers).
		SetParams(params).
		SetQuery(query).
		SetBody(bts)
}

func mapGolainHandlerToEchoHandler(handler HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		res := handler(mapEchoCtxToGolainCtx(c))

		return c.JSON(res.code, res.data)
	}
}

// Use ...
func (f *EchoRouter) Use(fn func(r *AppRouter)) AppRouter {
	return f
}

// WithDefaultMiddleware ...
func (f *EchoRouter) WithDefaultMiddleware() AppRouter {
	f.app.Use(middleware.RequestID())
	f.app.Use(middleware.CORS())
	f.app.Use(middleware.Recover())

	return f
}

// WithRequestLogger ...
func (f *EchoRouter) WithRequestLogger() AppRouter {
	f.app.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Trace().
				Str("method", c.Request().Method).
				Int("code", v.Status).
				Str("uri", v.URI).
				Str("from", c.Request().RemoteAddr).
				Send()

			return nil
		},
	}))

	return f
}

// WithTracing ...
func (f *EchoRouter) WithTracing(url ...string) AppRouter {
	id := strings.ReplaceAll(f.opts.ID, "-", "_")
	u := "http://localhost:14268/api/traces"

	if len(url) > 0 {
		u = url[0]
	}

	telemetry.New(id, u)

	f.app.Use(otelecho.Middleware(id))

	return f
}

// WithFrontend ...
func (f *EchoRouter) WithFrontend(data embed.FS) AppRouter {
	return f
}

// WithMetrics ...
func (f *EchoRouter) WithMetrics() AppRouter {
	prometheus.NewPrometheus(f.opts.ID, nil).Use(f.app)

	return f
}

// WithRoute ...
func (f *EchoRouter) WithRoute(method, path string, fn []HandlerFunc) AppRouter {
	handlers := []echo.MiddlewareFunc{}

	for i, h := range fn {
		if i > 0 {
			handlers = append(handlers, func(next echo.HandlerFunc) echo.HandlerFunc {
				return mapGolainHandlerToEchoHandler(h)
			})
		}
	}

	switch method {
	case http.MethodGet:
		f.app.GET(path, mapGolainHandlerToEchoHandler(fn[0]), handlers...)
	}

	return f
}

// Run ...
func (f *EchoRouter) Run() {
	log.
		Info().
		Str("id", f.opts.ID).
		Str("URL", fmt.Sprintf("http://0.0.0.0:%d", f.opts.Port)).
		Str("OpenAPI", fmt.Sprintf("http://0.0.0.0:%d/docs", f.opts.Port)).
		Send()

	for _, host := range addresses() {
		log.
			Info().
			Str("id", f.opts.ID).
			Str("URL", fmt.Sprintf("http://%s:%d", host, f.opts.Port)).
			Str("OpenAPI", fmt.Sprintf("http://%s:%d/docs", host, f.opts.Port)).
			Send()
	}

	log.Info().Msgf("%s started with Echo 🚀", f.opts.ID)

	f.app.Start(fmt.Sprintf("%s:%d", f.opts.Host, f.opts.Port))
}

// WithEcho ...
func WithEcho(port int, opts ...AppRouterOptions) Option {
	return WithAppRouter(newEchoRouter(mergeOptions(append(opts, AppRouterOptions{Port: port})...)))
}
