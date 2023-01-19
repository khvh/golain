package golain

import (
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/khvh/golain/telemetry"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
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

// Use ...
func (f *FiberRouter) Use(fn func(r *AppRouter)) AppRouter {
	return f
}

// WithDefaultMiddleware ...
func (f *FiberRouter) WithDefaultMiddleware() AppRouter {
	f.app.Use(requestid.New())
	f.app.Use(recover.New())
	f.app.Use(cors.New())
	f.app.Get("/monitor", monitor.New(monitor.Config{Title: f.opts.ID}))

	return f
}

// WithRequestLogger ...
func (f *FiberRouter) WithRequestLogger() AppRouter {
	return f
}

// WithTracing ...
func (f *FiberRouter) WithTracing(url ...string) AppRouter {
	id := strings.ReplaceAll(f.opts.ID, "-", "_")
	u := "http://localhost:14268/api/traces"

	otel.Tracer(id)

	if len(url) > 0 {
		u = url[0]
	}

	telemetry.New(id, u)

	f.app.Use(otelfiber.Middleware(otelfiber.WithServerName(f.opts.ID)))

	return f
}

// WithFrontend ...
func (f *FiberRouter) WithFrontend(data embed.FS) AppRouter {
	return f
}

// WithMetrics ...
func (f *FiberRouter) WithMetrics() AppRouter {
	id := strings.ReplaceAll(f.opts.ID, "-", "_")
	prometheus := fiberprometheus.New(id)

	prometheus.RegisterAt(f.app, "/metrics")

	f.app.Use(prometheus.Middleware)

	return f
}

// WithRoute ...
func (f *FiberRouter) WithRoute(method, path string, fn []HandlerFunc) AppRouter {
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
