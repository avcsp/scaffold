package router

import "net/http"

// Middleware wraps an http.Handler and returns a new http.Handler.
type Middleware func(http.Handler) http.Handler

// Router is the top-level router built on net/http.ServeMux.
type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
}

// New creates a new Router.
func New() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// Use appends global middleware to the router.
func (r *Router) Use(mw ...Middleware) {
	r.middlewares = append(r.middlewares, mw...)
}

// Group creates a route group with the given prefix.
// The callback receives a Group that inherits the router's middleware.
func (r *Router) Group(prefix string, fn func(g *Group)) {
	g := &Group{
		prefix:      prefix,
		router:      r,
		middlewares: make([]Middleware, len(r.middlewares)),
	}
	copy(g.middlewares, r.middlewares)
	fn(g)
}

// Handle registers a handler for the given method and pattern.
// Optional per-route middleware is applied after global middleware.
func (r *Router) Handle(method, pattern string, handler http.HandlerFunc, mw ...Middleware) {
	all := make([]Middleware, 0, len(r.middlewares)+len(mw))
	all = append(all, r.middlewares...)
	all = append(all, mw...)
	r.mux.Handle(method+" "+pattern, chain(handler, all))
}

func (r *Router) Get(pattern string, h http.HandlerFunc, mw ...Middleware) {
	r.Handle("GET", pattern, h, mw...)
}
func (r *Router) Post(pattern string, h http.HandlerFunc, mw ...Middleware) {
	r.Handle("POST", pattern, h, mw...)
}
func (r *Router) Put(pattern string, h http.HandlerFunc, mw ...Middleware) {
	r.Handle("PUT", pattern, h, mw...)
}
func (r *Router) Patch(pattern string, h http.HandlerFunc, mw ...Middleware) {
	r.Handle("PATCH", pattern, h, mw...)
}
func (r *Router) Delete(pattern string, h http.HandlerFunc, mw ...Middleware) {
	r.Handle("DELETE", pattern, h, mw...)
}

// Serve starts the HTTP server on the given address.
func (r *Router) Serve(addr string) error {
	return http.ListenAndServe(addr, r.mux)
}

// Group represents a route group with a shared prefix and middleware stack.
type Group struct {
	prefix      string
	router      *Router
	middlewares []Middleware
}

// Use appends middleware scoped to this group.
func (g *Group) Use(mw ...Middleware) {
	g.middlewares = append(g.middlewares, mw...)
}

// Group creates a nested group within this group.
func (g *Group) Group(prefix string, fn func(g *Group)) {
	child := &Group{
		prefix:      g.prefix + prefix,
		router:      g.router,
		middlewares: make([]Middleware, len(g.middlewares)),
	}
	copy(child.middlewares, g.middlewares)
	fn(child)
}

// Handle registers a handler for the given method and prefixed pattern.
// Optional per-route middleware is applied after group middleware.
func (g *Group) Handle(method, pattern string, handler http.HandlerFunc, mw ...Middleware) {
	all := make([]Middleware, 0, len(g.middlewares)+len(mw))
	all = append(all, g.middlewares...)
	all = append(all, mw...)
	g.router.mux.Handle(method+" "+g.prefix+pattern, chain(handler, all))
}

func (g *Group) Get(pattern string, h http.HandlerFunc, mw ...Middleware) {
	g.Handle("GET", pattern, h, mw...)
}
func (g *Group) Post(pattern string, h http.HandlerFunc, mw ...Middleware) {
	g.Handle("POST", pattern, h, mw...)
}
func (g *Group) Put(pattern string, h http.HandlerFunc, mw ...Middleware) {
	g.Handle("PUT", pattern, h, mw...)
}
func (g *Group) Patch(pattern string, h http.HandlerFunc, mw ...Middleware) {
	g.Handle("PATCH", pattern, h, mw...)
}
func (g *Group) Delete(pattern string, h http.HandlerFunc, mw ...Middleware) {
	g.Handle("DELETE", pattern, h, mw...)
}

// chain wraps a handler with middleware in reverse order so that
// the first middleware added is the outermost (runs first).
func chain(handler http.Handler, middlewares []Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
