package router

import (
	"net/http"

	sctx "scaffold/internal/context"
)

// Router is the top-level router built on net/http.ServeMux.
type Router struct {
	mux         *http.ServeMux
	middlewares []sctx.Middleware
}

// New creates a new Router.
func New() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// Use appends global middleware to the router.
func (r *Router) Use(mw ...sctx.Middleware) {
	r.middlewares = append(r.middlewares, mw...)
}

// Group creates a route group with the given prefix.
func (r *Router) Group(prefix string, fn func(g *Group)) {
	g := &Group{
		prefix:      prefix,
		router:      r,
		middlewares: make([]sctx.Middleware, len(r.middlewares)),
	}
	copy(g.middlewares, r.middlewares)
	fn(g)
}

// Handle registers a handler for the given method and pattern.
func (r *Router) Handle(method, pattern string, handler sctx.HandlerFunc, mw ...sctx.Middleware) {
	all := make([]sctx.Middleware, 0, len(r.middlewares)+len(mw))
	all = append(all, r.middlewares...)
	all = append(all, mw...)
	final := chain(handler, all)
	r.mux.HandleFunc(method+" "+pattern, toHTTPHandler(final))
}

func (r *Router) Get(pattern string, h sctx.HandlerFunc, mw ...sctx.Middleware) {
	r.Handle("GET", pattern, h, mw...)
}
func (r *Router) Post(pattern string, h sctx.HandlerFunc, mw ...sctx.Middleware) {
	r.Handle("POST", pattern, h, mw...)
}
func (r *Router) Put(pattern string, h sctx.HandlerFunc, mw ...sctx.Middleware) {
	r.Handle("PUT", pattern, h, mw...)
}
func (r *Router) Patch(pattern string, h sctx.HandlerFunc, mw ...sctx.Middleware) {
	r.Handle("PATCH", pattern, h, mw...)
}
func (r *Router) Delete(pattern string, h sctx.HandlerFunc, mw ...sctx.Middleware) {
	r.Handle("DELETE", pattern, h, mw...)
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
	middlewares []sctx.Middleware
}

// Use appends middleware scoped to this group.
func (g *Group) Use(mw ...sctx.Middleware) {
	g.middlewares = append(g.middlewares, mw...)
}

// Group creates a nested group within this group.
func (g *Group) Group(prefix string, fn func(g *Group)) {
	child := &Group{
		prefix:      g.prefix + prefix,
		router:      g.router,
		middlewares: make([]sctx.Middleware, len(g.middlewares)),
	}
	copy(child.middlewares, g.middlewares)
	fn(child)
}

// Handle registers a handler for the given method and prefixed pattern.
func (g *Group) Handle(method, pattern string, handler sctx.HandlerFunc, mw ...sctx.Middleware) {
	all := make([]sctx.Middleware, 0, len(g.middlewares)+len(mw))
	all = append(all, g.middlewares...)
	all = append(all, mw...)
	final := chain(handler, all)
	g.router.mux.HandleFunc(method+" "+g.prefix+pattern, toHTTPHandler(final))
}

func (g *Group) Get(pattern string, h sctx.HandlerFunc, mw ...sctx.Middleware) {
	g.Handle("GET", pattern, h, mw...)
}
func (g *Group) Post(pattern string, h sctx.HandlerFunc, mw ...sctx.Middleware) {
	g.Handle("POST", pattern, h, mw...)
}
func (g *Group) Put(pattern string, h sctx.HandlerFunc, mw ...sctx.Middleware) {
	g.Handle("PUT", pattern, h, mw...)
}
func (g *Group) Patch(pattern string, h sctx.HandlerFunc, mw ...sctx.Middleware) {
	g.Handle("PATCH", pattern, h, mw...)
}
func (g *Group) Delete(pattern string, h sctx.HandlerFunc, mw ...sctx.Middleware) {
	g.Handle("DELETE", pattern, h, mw...)
}

// chain wraps a handler with middleware in reverse order so that
// the first middleware added is the outermost (runs first).
func chain(handler sctx.HandlerFunc, middlewares []sctx.Middleware) sctx.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// toHTTPHandler converts a scaffold HandlerFunc to a standard http.HandlerFunc.
func toHTTPHandler(h sctx.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := sctx.New(w, r)
		h(c)
	}
}
