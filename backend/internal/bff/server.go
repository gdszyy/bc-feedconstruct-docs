// Package bff is the dispatch layer: WebSocket hub + REST endpoints exposed
// to the Next.js frontend. Frontend never connects to FeedConstruct directly.
//
// This file defines the wiring seam each parallel BFF track plugs into:
//
//   srv := bff.New()
//   bff.RegisterBetsRoutes(srv.Mux(), betsMgr)        // M13/M14 — implemented
//   bff.RegisterHealthzRoutes(srv.Mux(), readiness)    // B1/B2  — Wave 10-C
//   bff.RegisterMatchRoutes(srv.Mux(), catalogRepo)    // B4     — Wave 10-D
//   bff.RegisterWebSocketRoutes(srv.Mux(), feedBus)    // B5     — Wave 10-E
//   srv.Use(bff.OriginGuard(allowed))                  // S2     — Wave 10-F
//   srv.Use(bff.RateLimit(60))                          // S3     — Wave 10-F
//   http.ListenAndServe(addr, srv.Handler())
//
// Each Register* lives in its own file owned by exactly one Wave-10 track
// (see /root/.claude/plans/tdd-wise-flurry.md). Do not add module-specific
// logic to server.go — only the composition seam.
package bff

import "net/http"

// Server composes the stdlib mux with an ordered middleware chain.
// Modules register their routes onto Mux() directly; cross-cutting
// concerns (origin, rate-limit, telemetry) wrap the whole tree via Use.
type Server struct {
	mux         *http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

// New returns a Server with an empty mux. Routes and middleware are
// attached by Wave-10 tracks via Mux() and Use().
func New() *Server {
	return &Server{mux: http.NewServeMux()}
}

// Mux exposes the underlying ServeMux so module Register* helpers can
// attach handlers. Go 1.22+ path patterns ({id}, METHOD prefixes) are
// supported.
func (s *Server) Mux() *http.ServeMux { return s.mux }

// Use appends a middleware that will wrap the final handler. Middleware
// is applied in reverse registration order so the first Use call runs
// outermost (closest to the network).
func (s *Server) Use(mw func(http.Handler) http.Handler) {
	if mw == nil {
		return
	}
	s.middlewares = append(s.middlewares, mw)
}

// Handler returns the mux wrapped by every registered middleware. Call
// once at boot time after all routes and middleware are attached.
func (s *Server) Handler() http.Handler {
	var h http.Handler = s.mux
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		h = s.middlewares[i](h)
	}
	return h
}
