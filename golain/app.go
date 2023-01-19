package golain

import (
	"context"
	"embed"
	"net"
	"net/http"
	"os"

	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
)

// Ctx ...
type Ctx struct {
	Params  map[string]string
	Query   map[string]string
	Headers map[string]string
	Body    []byte
	Context context.Context
}

// NewCtx ...
func NewCtx() *Ctx {
	return &Ctx{}
}

// SetParams is a setter for Ctx.Params
func (c *Ctx) SetParams(p map[string]string) *Ctx {
	c.Params = p

	return c
}

// SetQuery is a setter for Ctx.Query
func (c *Ctx) SetQuery(q map[string]string) *Ctx {
	c.Query = q

	return c
}

// SetHeaders is a setter for Ctx.Headers
func (c *Ctx) SetHeaders(h map[string]string) *Ctx {
	c.Headers = h

	return c
}

// SetBody is a setter for Ctx.Body
func (c *Ctx) SetBody(b []byte) *Ctx {
	c.Body = b

	return c
}

// SetContext is a setter for Ctx.Context
func (c *Ctx) SetContext(ctx context.Context) *Ctx {
	c.Context = ctx

	return c
}

// Res ...
type Res struct {
	data any
	code int
	c    *Ctx
}

// JSON ...
func (c *Ctx) JSON(data any, statusCode ...int) *Res {
	code := 200

	if len(statusCode) > 0 {
		code = statusCode[0]
	}

	return &Res{
		data,
		code,
		c,
	}
}

// HandlerFunc ...
type HandlerFunc func(c *Ctx) *Res

// Route ...
type Route struct {
	path   string
	method string
	// spec    *spec.OAS
	handlers []HandlerFunc
}

// Get creates a GET route
func Get[T any](path string, handlers ...HandlerFunc) *Route {
	// var t T

	return &Route{
		path: path,
		// spec:    spec.Of(path, getPackage(pc)).Get(t),
		method:   http.MethodGet,
		handlers: handlers,
	}
}

// AppRouterOptions ...
type AppRouterOptions struct {
	ID            string
	Host          string
	Port          int
	Banner        bool
	RequestLogger bool
}

// AppRouter ...
type AppRouter interface {
	Use(fn func(r *AppRouter)) AppRouter
	RegisterRoute(method, path string, fn []HandlerFunc) AppRouter
	EnableTracing(url string) AppRouter
	MountFrontend(data embed.FS) AppRouter
	Run()
}

func mergeOptions(opts ...AppRouterOptions) *AppRouterOptions {
	o := &AppRouterOptions{}

	for _, opt := range opts {
		_ = mergo.Merge(o, &opt, mergo.WithSliceDeepCopy)
	}

	return o
}

// WithAppRouter ...
func WithAppRouter(r AppRouter) Option {
	return func(g *Golain) error {
		g.r = r

		return nil
	}
}

// Golain ...
type Golain struct {
	r AppRouter
}

// Option represents option function
type Option func(g *Golain) error

// New ...
func New(opts ...Option) *Golain {
	instance := &Golain{}

	for _, opt := range opts {
		if err := opt(instance); err != nil {
			log.Err(err).Send()
		}
	}

	return instance
}

// Register ...
func (g *Golain) Register(fn func(g *Golain)) *Golain {
	fn(g)

	return g
}

// RegisterRoutes ...
func (g *Golain) RegisterRoutes(routes ...*Route) *Golain {
	for _, r := range routes {
		g.r.RegisterRoute(r.method, r.path, r.handlers)
	}

	return g
}

// Run ...
func (g *Golain) Run() {
	g.r.Run()
}

// Params return path parameters from Ctx
func Params[P any](c *Ctx) P {
	var p P

	return p
}

// Query returns query parameters from Ctx
func Query[Q any](c *Ctx) Q {
	var q Q

	return q
}

// Body returns body from Ctx
func Body[B any](c *Ctx) B {
	var b B

	return b
}

// Headers returns query headers from Ctx
func Headers[H any](c *Ctx) H {
	var h H

	return h
}

// Addresses returns addresses the server can bind to
func addresses() []string {
	host, _ := os.Hostname()
	addresses, _ := net.LookupIP(host)

	var hosts []string

	for _, addr := range addresses {
		if ipv4 := addr.To4(); ipv4 != nil {
			hosts = append(hosts, ipv4.String())
		}
	}

	return hosts
}
