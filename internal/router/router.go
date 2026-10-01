package router

import (
	"net/http"

	"scaffold/internal/engine"
)

// Router is the top-level router built on net/http.ServeMux.
type Router struct {
	mux         *http.ServeMux
	middlewares []engine.Middleware
}

// New creates a new Router.
func New() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// Use appends global middleware to the router.
func (r *Router) Use(mw ...engine.Middleware) {
	r.middlewares = append(r.middlewares, mw...)
}

// Group creates a route group with the given prefix.
func (r *Router) Group(prefix string, fn func(g *Group)) {
	g := &Group{
		prefix:      prefix,
		router:      r,
		middlewares: make([]engine.Middleware, len(r.middlewares)),
	}
	copy(g.middlewares, r.middlewares)
	fn(g)
}

// Handle registers a route for the given method and pattern. steps is zero
// or more engine.Middleware followed by exactly one engine.HandlerFunc,
// e.g. Handle("GET", "/x", middleware.Auth, handler).
func (r *Router) Handle(method, pattern string, steps ...any) {
	mw, handler := splitSteps(steps)
	all := make([]engine.Middleware, 0, len(r.middlewares)+len(mw))
	all = append(all, r.middlewares...)
	all = append(all, mw...)
	final := chain(handler, all)
	r.mux.HandleFunc(method+" "+pattern, toHTTPHandler(final))
}

func (r *Router) Get(pattern string, steps ...any) {
	r.Handle("GET", pattern, steps...)
}
func (r *Router) Post(pattern string, steps ...any) {
	r.Handle("POST", pattern, steps...)
}
func (r *Router) Put(pattern string, steps ...any) {
	r.Handle("PUT", pattern, steps...)
}
func (r *Router) Patch(pattern string, steps ...any) {
	r.Handle("PATCH", pattern, steps...)
}
func (r *Router) Delete(pattern string, steps ...any) {
	r.Handle("DELETE", pattern, steps...)
}

// Handler returns the underlying http.Handler for use with http.Server.
func (r *Router) Handler() http.Handler {
	return r.mux
}

// Serve starts the HTTP server on the given address.
func (r *Router) Serve(addr string) error {
	return http.ListenAndServe(addr, r.mux)
}

// Group represents a route group with a shared prefix and middleware stack.
type Group struct {
	prefix      string
	router      *Router
	middlewares []engine.Middleware
}

// Use appends middleware scoped to this group.
func (g *Group) Use(mw ...engine.Middleware) {
	g.middlewares = append(g.middlewares, mw...)
}

// Group creates a nested group within this group.
func (g *Group) Group(prefix string, fn func(g *Group)) {
	child := &Group{
		prefix:      g.prefix + prefix,
		router:      g.router,
		middlewares: make([]engine.Middleware, len(g.middlewares)),
	}
	copy(child.middlewares, g.middlewares)
	fn(child)
}

// Handle registers a route for the given method and prefixed pattern. steps
// is zero or more engine.Middleware followed by exactly one engine.HandlerFunc,
// e.g. Handle("GET", "/x", middleware.Auth, handler).
func (g *Group) Handle(method, pattern string, steps ...any) {
	mw, handler := splitSteps(steps)
	all := make([]engine.Middleware, 0, len(g.middlewares)+len(mw))
	all = append(all, g.middlewares...)
	all = append(all, mw...)
	final := chain(handler, all)
	g.router.mux.HandleFunc(method+" "+g.prefix+pattern, toHTTPHandler(final))
}

func (g *Group) Get(pattern string, steps ...any) {
	g.Handle("GET", pattern, steps...)
}
func (g *Group) Post(pattern string, steps ...any) {
	g.Handle("POST", pattern, steps...)
}
func (g *Group) Put(pattern string, steps ...any) {
	g.Handle("PUT", pattern, steps...)
}
func (g *Group) Patch(pattern string, steps ...any) {
	g.Handle("PATCH", pattern, steps...)
}
func (g *Group) Delete(pattern string, steps ...any) {
	g.Handle("DELETE", pattern, steps...)
}

// splitSteps separates a route's steps into its middleware and handler.
// steps must contain exactly one engine.HandlerFunc, as the last meaningful
// argument; it panics otherwise.
func splitSteps(steps []any) ([]engine.Middleware, engine.HandlerFunc) {
	mw := make([]engine.Middleware, 0, len(steps))
	var handler engine.HandlerFunc
	var handlerCount int

	for _, s := range steps {
		switch v := s.(type) {
		case engine.HandlerFunc:
			handler = v
			handlerCount++
		case func(*engine.Context):
			handler = v
			handlerCount++
		case engine.Middleware:
			mw = append(mw, v)
		case func(engine.HandlerFunc) engine.HandlerFunc:
			mw = append(mw, v)
		default:
			panic("router: unsupported route argument type")
		}
	}

	if handlerCount == 0 {
		panic("router: route is missing a handler")
	}
	if handlerCount > 1 {
		panic("router: route has multiple handlers")
	}

	return mw, handler
}

// chain wraps a handler with middleware in reverse order so that
// the first middleware added is the outermost (runs first).
func chain(handler engine.HandlerFunc, middlewares []engine.Middleware) engine.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// toHTTPHandler converts a scaffold HandlerFunc to a standard http.HandlerFunc.
func toHTTPHandler(h engine.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := engine.New(w, r)
		h(c)
	}
}
