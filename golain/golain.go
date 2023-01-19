package golain

import (
	"net"
	"os"

	"github.com/khvh/golain/queue"
	"github.com/rs/zerolog/log"
)

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
		g.r.WithRoute(r.method, r.path, r.handlers)
	}

	return g
}

// EnableMetrics ...
func (g *Golain) EnableMetrics() *Golain {
	g.r.WithMetrics()

	return g
}

// EnableTracing ...
func (g *Golain) EnableTracing(url ...string) *Golain {
	g.r.WithTracing(url...)
	return g
}

// EnableQueue ...
func (g *Golain) EnableQueue(url, pw string, opts queue.Queues, fn func(q *queue.Queue)) *Golain {
	g.r.WithQueue(url, pw, opts, fn)

	return g
}

// Run ...
func (g *Golain) Run() {
	g.r.Run()
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
