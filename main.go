package main

import (
	"sync"

	"github.com/khvh/golain/golain"
	"github.com/khvh/golain/logger"
)

// TestType ...
type TestType struct {
	ID string `json:"id"`
}

func t1(c *golain.Ctx) *golain.Res {
	return c.JSON(&TestType{ID: "1"})
}

func main() {
	logger.Init(true)

	wg := new(sync.WaitGroup)

	wg.Add(2)

	go func() {
		golain.
			New(
				golain.WithFiber(6666, golain.AppRouterOptions{Banner: true, ID: "fib"}),
			).
			EnableMetrics().
			EnableTracing().
			Register(func(g *golain.Golain) {
				g.RegisterRoutes(
					golain.Get[TestType]("/test-path", t1),
				)
			}).
			Run()
		wg.Done()
	}()

	go func() {
		golain.
			New(golain.WithEcho(7777, golain.AppRouterOptions{Banner: true, ID: "ech"})).
			EnableMetrics().
			EnableTracing().
			Register(func(g *golain.Golain) {
				g.RegisterRoutes(
					golain.Get[TestType]("/test-path", t1),
				)
			}).
			Run()
		wg.Done()
	}()

	wg.Wait()
}
