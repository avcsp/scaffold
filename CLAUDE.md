# scaffold

A minimal Go HTTP framework named "Scaffold" built on `net/http.ServeMux`. This file explains the structure so you don't accidentally edit framework internals or break upgrade compatibility.

## Structure

```
app.go          entrypoint: config, logger, db, router, server wiring — framework-owned
routes.go       route definitions — app-owned, this is where you work
internal/       framework packages — framework-owned, never edit directly
handlers/       app handlers, one package per {version}/{package} — app-owned
middleware/     app middleware, one file per concern — app-owned
scaffold        CLI to pull framework updates from github.com/avcsp/scaffold (./scaffold upgrade)
```

Note the app's root `middleware/` is a different package from the framework's `internal/middleware/` (which holds `Recovery`/`CORS`) — same name, different owner, don't confuse the two.

### `internal/` packages

| Package       | Responsibility                                                                       |
|---------------|--------------------------------------------------------------------------------------|
| `engine`      | `Context` (request/response wrapper), `HandlerFunc`, `Middleware` types              |
| `router`      | `Router`/`Group` on top of `http.ServeMux`; route registration, middleware chaining  |
| `middleware`  | `Recovery`, `CORS` — cross-cutting concerns wrapped around the whole mux             |
| `server`      | `http.Server` lifecycle: graceful shutdown, `/probe/live`, `/probe/ready`            |
| `config`      | `.env` loading, `Getenv` with fallback                                               |
| `database`    | optional GORM connection (Postgres/MySQL/MariaDB via `DATABASE_DSN`)                 |
| `logger`      | `slog` setup, JSON to stdout, level from `ENV`                                       |

## Off-limits for app developers

**Never edit files under `internal/`.** They are overwritten wholesale by `./scaffold upgrade`, which pulls `internal/`, `app.go`, `.air.toml`, `Dockerfile`, `Makefile`, `.dockerignore`, `.gitignore`, `.env.example`, and `CLAUDE.md` from upstream and replaces them verbatim. Any local change to those files is silently lost on the next upgrade.

**Treat `app.go` as framework-owned too** — it's in the sync path. If you need app-specific startup logic, add an `OnShutdown` hook or a new file in the `main` package rather than editing `app.go` directly; expect manual re-application after an upgrade if you do touch it.

**Safe to edit freely** (never touched by `./scaffold upgrade`):
    - `routes.go` — your route wiring
    - `handlers/` — your handlers
    - `middleware/` — your middleware (auth, logging, rate limiting, etc.)
    - `.env` — your local secrets/config (never committed)
    - any new `.go` file you add to the `main` package, or new app packages outside `internal/` (e.g. `models/`)

If a framework package is missing something you need, don't patch it in place — that diff will be clobbered on the next upgrade. Either open it upstream in `github.com/avcsp/scaffold`, or build the extra behavior in your own app-owned code that calls into the framework's public API.

## Handler conventions

Handlers are `func(c *engine.Context)` — no return value, no error. Signal failure by writing a response, not by returning one.

```go
r.Get("/users/{id}", func(c *engine.Context) {
    id := c.Param("id")
    c.JSON(http.StatusOK, map[string]string{"message": "user: " + id})
})
```

- **Path params**: `c.Param("id")` (backed by Go 1.22+ `ServeMux` `{id}` patterns)
- **Query params**: `c.Query("key")`
- **JSON body in**: `c.BindJSON(&dst)`
- **JSON body out**: `c.JSON(status, data)`
- **Plain text out**: `c.String(status, body)`
- **Status only**: `c.Status(code)`
- **Per-request state**: `c.Set(key, val)` / `c.Get(key)` / `c.MustGet(key)` — not a global map, scoped to the request
- **Built-in services**: `c.Import("db")` returns `*gorm.DB` (nil if `DATABASE_DSN` unset) — don't import `internal/database` directly from app code, go through `Import`

### Middleware

A middleware is `func(engine.HandlerFunc) engine.HandlerFunc`. Register it:

- Globally: `r.Use(mw)` on the `*router.Router` before defining routes
- Per-group: `g.Use(mw)` inside a `r.Group(prefix, func(g *router.Group) {...})` callback
- Per-route, inline, before the handler: `r.Get("/x", someMiddleware, handler)`

Route registration (`Handle`/`Get`/`Post`/...) takes a variadic list of middleware followed by exactly one handler — order matters, middleware runs in the order listed, outermost first.

`CORS` and `Recovery` are **not** route middleware — they wrap the whole `http.Handler` in `server.New()`, so they run before routing/mux matching happens (this is why `OPTIONS` preflight requests are handled for every route automatically, even ones that never register an `OPTIONS` method).

### Handler file layout

Handlers live under `handlers/{version}/{package}/`, one file per action, named after the action. A package's list handler (`GET /{version}/{package}`) goes in `list.go`:

```
handlers/v1.0/users/list.go   ->  func List(c *engine.Context) { ... }  ->  GET /v1.0/users
```

`routes.go` stays the single place wiring paths to handlers — it imports the handler packages and registers them, it doesn't contain handler bodies:

```go
import "scaffold-app/handlers/v1.0/users"

r.Group("/v1.0", func(r *router.Group) {
    r.Get("/users", users.List)
})
```

### Aborting a chain

Call `c.Abort()` (or `c.AbortWithStatusJSON(status, data)`) in middleware to stop the chain. Downstream middleware/handlers should check `c.IsAborted()` before doing work if they run after something that might abort.

## Routing

Routes live in `routes.go`, registered via `router.New()` → `Routes(r)` (called from `app.go`). Use `r.Group(prefix, func(g *router.Group) {...})` to share a path prefix and middleware stack across related routes.

```go
r.Group("/v1.0", func(r *router.Group) {
    r.Get("/users", listUsers)
    r.Get("/users/{id}", getUser)
})
```

Patterns follow `net/http.ServeMux` syntax (`{id}`, `{path...}` wildcards, exact method matching per route).

## Upgrading the framework

```
./scaffold upgrade
```

Requires a clean working tree. Fetches `github.com/avcsp/scaffold` (branch via `SCAFFOLD_BRANCH`, default `dev`) and overwrites the framework-owned paths listed above. `go.mod`/`go.sum` are **not** touched — run `go mod tidy` after upgrading to pick up any new framework dependencies without disturbing your own.
