package engine

import (
	"encoding/json"
	"net/http"
)

// HandlerFunc is the scaffold handler signature.
type HandlerFunc func(*Context)

// Middleware wraps a HandlerFunc and returns a new HandlerFunc.
type Middleware func(HandlerFunc) HandlerFunc

// Context wraps the request and response for a single HTTP transaction.
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	store   map[string]any
	aborted bool
}

// New creates a Context from a standard http request/response pair.
func New(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
	}
}

// JSON writes data as a JSON response with the given status code.
func (c *Context) JSON(status int, data any) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	json.NewEncoder(c.Writer).Encode(data)
}

// Error writes a JSON error response: {"error": msg}
func (c *Context) Error(status int, msg string) {
	c.JSON(status, map[string]string{"error": msg})
}

// AbortWithStatusJSON writes a JSON response and marks the context as aborted.
// Subsequent middleware or handlers should check c.IsAborted() before proceeding.
func (c *Context) AbortWithStatusJSON(status int, data any) {
	c.aborted = true
	c.JSON(status, data)
}

// Abort marks the context as aborted without writing a response.
func (c *Context) Abort() {
	c.aborted = true
}

// IsAborted returns whether the context has been aborted.
func (c *Context) IsAborted() bool {
	return c.aborted
}

// Set stores a key-value pair on the context for the duration of the request.
func (c *Context) Set(key string, value any) {
	if c.store == nil {
		c.store = make(map[string]any)
	}
	c.store[key] = value
}

// Get retrieves a value by key from the context store.
// Returns the value and whether the key exists.
func (c *Context) Get(key string) (any, bool) {
	if c.store == nil {
		return nil, false
	}
	val, ok := c.store[key]
	return val, ok
}

// MustGet retrieves a value by key. Panics if the key doesn't exist.
func (c *Context) MustGet(key string) any {
	val, ok := c.Get(key)
	if !ok {
		panic("engine: key '" + key + "' not found in context")
	}
	return val
}

// BindJSON decodes the request body into dst.
func (c *Context) BindJSON(dst any) error {
	defer c.Request.Body.Close()
	return json.NewDecoder(c.Request.Body).Decode(dst)
}

// Param returns the path parameter value by name.
func (c *Context) Param(name string) string {
	return c.Request.PathValue(name)
}

// Query returns the query parameter value by name.
func (c *Context) Query(name string) string {
	return c.Request.URL.Query().Get(name)
}

// Status writes the status code header.
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// Header sets a response header.
func (c *Context) Header(key, value string) {
	c.Writer.Header().Set(key, value)
}

// String writes a plain text response.
func (c *Context) String(status int, s string) {
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(status)
	c.Writer.Write([]byte(s))
}
