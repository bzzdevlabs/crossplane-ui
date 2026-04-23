package oidc_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	gwoidc "github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/oidc"
)

// fakeVerifier is a hand-rolled stub that satisfies gwoidc.Verifier.
// Each test swaps the Verify function in to simulate success/failure.
type fakeVerifier struct {
	verify func(ctx context.Context, raw string) (gwoidc.IDToken, error)
}

func (f *fakeVerifier) Verify(ctx context.Context, raw string) (gwoidc.IDToken, error) {
	return f.verify(ctx, raw)
}

// fakeIDToken serves pre-baked claims as JSON.
type fakeIDToken struct{ payload []byte }

func (t *fakeIDToken) Claims(v any) error { return json.Unmarshal(t.payload, v) }

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestMiddlewareRejectsMissingToken(t *testing.T) {
	t.Parallel()

	mw := gwoidc.Middleware(&fakeVerifier{}, newLogger())
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler reached despite missing token")
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestMiddlewareRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	v := &fakeVerifier{
		verify: func(_ context.Context, _ string) (gwoidc.IDToken, error) {
			return nil, errors.New("expired")
		},
	}
	mw := gwoidc.Middleware(v, newLogger())
	h := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler reached despite invalid token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestMiddlewareInjectsUserOnSuccess(t *testing.T) {
	t.Parallel()

	payload := []byte(
		`{"sub":"abc","email":"alice@example.com",` +
			`"preferred_username":"alice","groups":["admins","devs"]}`,
	)
	v := &fakeVerifier{
		verify: func(_ context.Context, _ string) (gwoidc.IDToken, error) {
			return &fakeIDToken{payload: payload}, nil
		},
	}

	var got gwoidc.User
	mw := gwoidc.Middleware(v, newLogger())
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := gwoidc.UserFromContext(r.Context())
		if !ok {
			t.Fatal("user missing from context")
		}
		got = u
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got.Subject != "abc" || got.PreferredUsername != "alice" ||
		got.Email != "alice@example.com" || len(got.Groups) != 2 {
		t.Errorf("unexpected user: %+v", got)
	}
	name, groups := got.Kubernetes()
	if name != "alice" || len(groups) != 2 {
		t.Errorf("Kubernetes() mismatch: %q %v", name, groups)
	}
}

func TestDevPassthroughInjectsAdminUser(t *testing.T) {
	t.Parallel()

	var got gwoidc.User
	h := gwoidc.DevPassthrough(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := gwoidc.UserFromContext(r.Context())
		if !ok {
			t.Fatal("user missing from context")
		}
		got = u
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got.PreferredUsername != "dev-admin" {
		t.Errorf("dev user = %q, want dev-admin", got.PreferredUsername)
	}
	if len(got.Groups) != 1 || got.Groups[0] != "system:masters" {
		t.Errorf("dev groups = %v, want [system:masters]", got.Groups)
	}
}
