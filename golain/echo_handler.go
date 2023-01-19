package golain

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
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

// EnableTracing ...
func (f *EchoRouter) EnableTracing(url string) AppRouter {
	return f
}

// MountFrontend ...
func (f *EchoRouter) MountFrontend(data embed.FS) AppRouter {
	return f
}

// RegisterRoute ...
func (f *EchoRouter) RegisterRoute(method, path string, fn []HandlerFunc) AppRouter {
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
