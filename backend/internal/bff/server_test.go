package bff_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/bff"
)

// Given a fresh Server with no routes or middleware
// When a request hits Handler()
// Then the embedded mux responds 404 (proves the composition seam works)
func TestGiven_EmptyServer_When_RequestUnknownPath_Then_404(t *testing.T) {
	srv := bff.New()
	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 from empty mux, got %d", rec.Code)
	}
}

// Given a Server with two middlewares registered in order A, B
// When a request flows through Handler()
// Then A runs first (outermost), B runs second, mux runs last
func TestGiven_TwoMiddlewares_When_RequestFlows_Then_FIFOOutermostOrder(t *testing.T) {
	var order []string
	srv := bff.New()
	srv.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "A")
			next.ServeHTTP(w, r)
		})
	})
	srv.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "B")
			next.ServeHTTP(w, r)
		})
	})
	srv.Mux().HandleFunc("GET /x", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "mux")
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	srv.Handler().ServeHTTP(httptest.NewRecorder(), req)

	if len(order) != 3 || order[0] != "A" || order[1] != "B" || order[2] != "mux" {
		t.Fatalf("expected [A B mux], got %v", order)
	}
}
