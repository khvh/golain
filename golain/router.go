package golain

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/khvh/golain/oas"
	"github.com/swaggest/openapi-go/openapi3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	path     string
	method   string
	spec     *oas.OAS
	handlers []HandlerFunc
}

// Router holds routes
type Router struct {
	prefix string
	group  string
	routes []*Route
}

// NewRouter is a singleton returning method for Router
func NewRouter() *Router {
	return &Router{
		routes: []*Route{},
	}
}

// OASOptions ...
type OASOptions struct {
	Title       string
	Description string
	Version     string
	OASVersion  string
	AuthURL     string
	AuthClient  string
	AuthSecret  string
}

// InitReflector ...
func InitReflector(port int, addresses []string, opts *OASOptions) *openapi3.Reflector {
	ref := &openapi3.Reflector{}

	if opts.OASVersion == "" {
		opts.OASVersion = "3.1.0"
	}

	ref.Spec = &openapi3.Spec{Openapi: opts.OASVersion}

	servers := []openapi3.Server{}

	for _, host := range addresses {
		servers = append(servers, openapi3.Server{
			URL: fmt.Sprintf("http://%s:%d", host, port),
		})
	}

	ref.Spec.WithServers(servers...)

	ref.Spec.Info.
		WithTitle(opts.Title).
		WithVersion(opts.Version).
		WithDescription(opts.Description)

	if opts.AuthURL != "" {

	}

	return ref
}

// WithOIDC ...
func WithOIDC(ref *openapi3.Reflector, url, client, secret string) {
	ref.SpecEns().ComponentsEns().SecuritySchemesEns().WithMapOfSecuritySchemeOrRefValuesItem(
		"bearer",
		openapi3.SecuritySchemeOrRef{
			SecurityScheme: &openapi3.SecurityScheme{
				OAuth2SecurityScheme: (&openapi3.OAuth2SecurityScheme{}).
					WithFlows(openapi3.OAuthFlows{
						Implicit: &openapi3.ImplicitOAuthFlow{
							AuthorizationURL: url,
							Scopes:           map[string]string{},
						},
					}),
			},
		},
	)
}

// Register registers one or more routes
func (r *Router) Register(routes ...*Route) *Router {

	for _, route := range routes {
		r.routes = append(r.routes, route)
	}

	return r
}

// Prefix adds an url prefix
func (r *Router) Prefix(url string) *Router {
	r.prefix = url

	return r
}

// Group groups routes under a common tag
func (r *Router) Group(name string) *Router {
	r.group = name

	return r
}

// Get creates a GET route
func Get[T any](path string, handlers ...HandlerFunc) *Route {
	pc, _, _, _ := runtime.Caller(1)

	var t T

	return &Route{
		path:     path,
		spec:     oas.Of(path, getPackage(pc)).Get(t),
		method:   http.MethodGet,
		handlers: handlers,
	}
}

// Delete creates a DELETE route
func Delete[T interface{}](path string, handlers ...HandlerFunc) *Route {
	pc, _, _, _ := runtime.Caller(1)

	var t T

	return &Route{
		path:     path,
		spec:     oas.Of(path, getPackage(pc)).Delete(t),
		method:   http.MethodDelete,
		handlers: handlers,
	}
}

// Post creates a POST route
func Post[T interface{}, D interface{}](path string, handlers ...HandlerFunc) *Route {
	pc, _, _, _ := runtime.Caller(1)

	var (
		t T
		d D
	)

	return &Route{
		path:     path,
		spec:     oas.Of(path, getPackage(pc)).Post(t, d),
		method:   http.MethodPost,
		handlers: handlers,
	}
}

// Put creates a PUT route
func Put[T interface{}, D interface{}](path string, handlers ...HandlerFunc) *Route {
	pc, _, _, _ := runtime.Caller(1)

	var (
		t T
		d D
	)

	return &Route{
		path:     path,
		spec:     oas.Of(path, getPackage(pc)).Put(t, d),
		method:   http.MethodPut,
		handlers: handlers,
	}
}

// Patch creates a PATCH route
func Patch[T interface{}, D interface{}](path string, handlers ...HandlerFunc) *Route {
	pc, _, _, _ := runtime.Caller(1)

	var (
		t T
		d D
	)

	return &Route{
		path:     path,
		spec:     oas.Of(path, getPackage(pc)).Patch(t, d),
		method:   http.MethodPatch,
		handlers: handlers,
	}
}

func getPackage(pc uintptr) string {
	funcName := runtime.FuncForPC(pc).Name()
	lastSlash := strings.LastIndexByte(funcName, '/')

	if lastSlash < 0 {
		lastSlash = 0
	}

	lastDot := strings.LastIndexByte(funcName[lastSlash:], '.') + lastSlash

	caser := cases.Title(language.English)

	return caser.String(strings.ToLower(funcName[:lastDot]))
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
