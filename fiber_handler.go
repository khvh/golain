package golain

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// FiberRouter ...
type FiberRouter struct {
	app  *fiber.App
	opts *AppRouterOptions
}

func newFiberRouter(opts *AppRouterOptions) AppRouter {
	r := &FiberRouter{
		app: fiber.New(fiber.Config{DisableStartupMessage: opts.Banner}),
	}

	r.opts = opts

	return r
}

func mapFiberCtxToGolainCtx(c *fiber.Ctx) *Ctx {
	q := map[string]string{}

	c.QueryParser(&q)

	return NewCtx().
		SetHeaders(c.GetReqHeaders()).
		SetParams(c.AllParams()).
		SetQuery(q).
		SetBody(c.Body())
}

func mapGolainHandlerToFiber(handler HandlerFunc) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(handler(mapFiberCtxToGolainCtx(c)).data)
	}
}

// UseRoute ...
func (f *FiberRouter) UseRoute(method, path string, fn []HandlerFunc) AppRouter {
	handlers := []fiber.Handler{}

	for _, h := range fn {
		handlers = append(handlers, mapGolainHandlerToFiber(h))
	}

	switch method {
	case http.MethodGet:
		f.app.Get(path, handlers...)
	}

	return f
}

// Run ...
func (f *FiberRouter) Run() {
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

	log.Info().Msgf("%s started with Fiber 🚀", f.opts.ID)

	f.app.Listen(fmt.Sprintf("%s:%d", f.opts.Host, f.opts.Port))
}

// WithFiber ...
func WithFiber(port int, opts ...AppRouterOptions) Option {
	return WithAppRouter(newFiberRouter(mergeOptions(append(opts, AppRouterOptions{Port: port})...)))
}
