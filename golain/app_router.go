package golain

import (
	"embed"

	"github.com/imdario/mergo"
	"github.com/khvh/golain/queue"
)

// AppRouterOptions ...
type AppRouterOptions struct {
	ID            string
	Version       string
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
	WithQueue(url, pw string, opts queue.Queues, fn func(q *queue.Queue)) AppRouter
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
