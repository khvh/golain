package golain

import (
	"embed"

	"github.com/imdario/mergo"
)

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
	WithDefaultMiddleware() AppRouter
	WithRoute(method, path string, fn []HandlerFunc) AppRouter
	WithTracing(url ...string) AppRouter
	WithMetrics() AppRouter
	WithFrontend(data embed.FS) AppRouter
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
