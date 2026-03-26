package api

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

// newTestClient returns an API client pointed at the given httptest server.
func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	c := NewClient("test-api-key", server.URL)
	return c
}

// ---------------------------------------------------------------------------
// HandleError
// ---------------------------------------------------------------------------

func TestHandleError_401(t *testing.T) {
	err := HandleError(401, nil, "do something")
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("unexpected 401 error message: %v", err)
	}
}

func TestHandleError_403_WithBody(t *testing.T) {
	body := []byte(`{"error": "premium required"}`)
	err := HandleError(403, body, "create alias")
	if err == nil {
		t.Fatal("expected error for 403")
	}
	if !strings.Contains(err.Error(), "premium required") {
		t.Errorf("expected error body message, got: %v", err)
	}
}

func TestHandleError_403_NoBody(t *testing.T) {
	err := HandleError(403, []byte("not json"), "create alias")
	if err == nil {
		t.Fatal("expected error for 403")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("expected 'forbidden' fallback, got: %v", err)
	}
}

func TestHandleError_404(t *testing.T) {
	err := HandleError(404, nil, "get alias")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found', got: %v", err)
	}
	if !strings.Contains(err.Error(), "get alias") {
		t.Errorf("expected action in error, got: %v", err)
	}
}

func TestHandleError_500_WithErrorField(t *testing.T) {
	body := []byte(`{"error": "internal server error"}`)
	err := HandleError(500, body, "list aliases")
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !strings.Contains(err.Error(), "internal server error") {
		t.Errorf("expected body error message, got: %v", err)
	}
}

func TestHandleError_500_NoErrorField(t *testing.T) {
	body := []byte(`{"message": "something went wrong"}`)
	err := HandleError(500, body, "list aliases")
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Errorf("expected HTTP status code in error, got: %v", err)
	}
}

func TestHandleError_500_InvalidJSON(t *testing.T) {
	err := HandleError(500, []byte("not json"), "list aliases")
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Errorf("expected HTTP status fallback, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ResolveAliasID
// ---------------------------------------------------------------------------

func TestResolveAliasID_IntegerInput(t *testing.T) {
	// Should parse directly without making any HTTP request
	c := NewClient("unused", "")
	id, err := c.ResolveAliasID("12345")
	if err != nil {
		t.Fatalf("ResolveAliasID('12345'): %v", err)
	}
	if id != 12345 {
		t.Errorf("expected 12345, got %d", id)
	}
}

func TestResolveAliasID_EmailInput_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/aliases" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// Verify query param is passed
		q := r.URL.Query().Get("query")
		if q != "test@sl.local" {
			t.Errorf("expected query=test@sl.local, got %q", q)
		}
		// Verify authentication header
		if r.Header.Get("Authentication") != "test-api-key" {
			t.Errorf("missing or wrong Authentication header: %q", r.Header.Get("Authentication"))
		}

		resp := AliasListResponse{
			Aliases: []Alias{
				{ID: 42, Email: "test@sl.local"},
				{ID: 99, Email: "other@sl.local"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	id, err := c.ResolveAliasID("test@sl.local")
	if err != nil {
		t.Fatalf("ResolveAliasID: %v", err)
	}
	if id != 42 {
		t.Errorf("expected ID 42, got %d", id)
	}
}

func TestResolveAliasID_EmailInput_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := AliasListResponse{Aliases: []Alias{}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.ResolveAliasID("nonexistent@sl.local")
	if err == nil {
		t.Fatal("expected error for not-found email")
	}
	if !strings.Contains(err.Error(), "alias not found") {
		t.Errorf("expected 'alias not found', got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Client.do - HTTP integration with httptest
// ---------------------------------------------------------------------------

func TestClient_GetUserInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/user_info" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		resp := UserInfo{
			Name:      "Test User",
			Email:     "test@example.com",
			IsPremium: true,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	info, _, err := c.GetUserInfo()
	if err != nil {
		t.Fatalf("GetUserInfo: %v", err)
	}
	if info.Name != "Test User" {
		t.Errorf("Name = %q, want %q", info.Name, "Test User")
	}
	if info.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", info.Email, "test@example.com")
	}
	if !info.IsPremium {
		t.Error("expected IsPremium to be true")
	}
}

func TestClient_GetUserInfo_AuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, _, err := c.GetUserInfo()
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("expected auth error, got: %v", err)
	}
}

func TestClient_ListAliases(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// Verify query parameters
		q := r.URL.Query()
		if q.Get("page_id") != "0" {
			t.Errorf("expected page_id=0, got %s", q.Get("page_id"))
		}
		// Check that filter params are present in raw query
		raw := r.URL.RawQuery
		if !strings.Contains(raw, "pinned") {
			t.Error("expected pinned flag in query")
		}
		if !strings.Contains(raw, "enabled") {
			t.Error("expected enabled flag in query")
		}

		resp := AliasListResponse{
			Aliases: []Alias{
				{ID: 1, Email: "a@sl.local", Enabled: true},
				{ID: 2, Email: "b@sl.local", Enabled: false},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	aliases, _, err := c.ListAliases(0, true, false, true, "")
	if err != nil {
		t.Fatalf("ListAliases: %v", err)
	}
	if len(aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(aliases))
	}
}

func TestClient_ToggleAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/aliases/42/toggle" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"enabled": false}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	enabled, _, err := c.ToggleAlias(42)
	if err != nil {
		t.Fatalf("ToggleAlias: %v", err)
	}
	if enabled {
		t.Error("expected enabled=false after toggle")
	}
}

func TestClient_DeleteAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/aliases/7" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"deleted": true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.DeleteAlias(7)
	if err != nil {
		t.Fatalf("DeleteAlias: %v", err)
	}
}

func TestClient_DeleteAlias_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.DeleteAlias(999)
	if err == nil {
		t.Fatal("expected error for 404 delete")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestClient_CreateRandomAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/alias/random/new" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(201)
		alias := Alias{ID: 100, Email: "random123@sl.local"}
		_ = json.NewEncoder(w).Encode(alias)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	alias, _, err := c.CreateRandomAlias("test note")
	if err != nil {
		t.Fatalf("CreateRandomAlias: %v", err)
	}
	if alias.ID != 100 {
		t.Errorf("expected ID 100, got %d", alias.ID)
	}
	if alias.Email != "random123@sl.local" {
		t.Errorf("expected email random123@sl.local, got %q", alias.Email)
	}
}

func TestClient_GetStats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stats" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		stats := Stats{NbAlias: 10, NbBlock: 5, NbForward: 100, NbReply: 3}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	stats, _, err := c.GetStats()
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.NbAlias != 10 {
		t.Errorf("NbAlias = %d, want 10", stats.NbAlias)
	}
	if stats.NbForward != 100 {
		t.Errorf("NbForward = %d, want 100", stats.NbForward)
	}
}

func TestClient_ListMailboxes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/mailboxes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		resp := MailboxListResponse{
			Mailboxes: []Mailbox{
				{ID: 1, Email: "primary@example.com", Default: true, Verified: true},
				{ID: 2, Email: "secondary@example.com", Default: false, Verified: true},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	mailboxes, _, err := c.ListMailboxes()
	if err != nil {
		t.Fatalf("ListMailboxes: %v", err)
	}
	if len(mailboxes) != 2 {
		t.Fatalf("expected 2 mailboxes, got %d", len(mailboxes))
	}
	if !mailboxes[0].Default {
		t.Error("expected first mailbox to be default")
	}
}

// ---------------------------------------------------------------------------
// isConnectionRefused helper
// ---------------------------------------------------------------------------

func TestIsConnectionRefused(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"exact match", fmt.Errorf("connection refused"), true},
		{"substring match", fmt.Errorf("dial tcp: connection refused"), true},
		{"no match", fmt.Errorf("timeout"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isConnectionRefused(tt.err)
			if got != tt.want {
				t.Errorf("isConnectionRefused(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Client.do - request body serialization
// ---------------------------------------------------------------------------

func TestClient_UpdateAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/aliases/5" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// Verify JSON body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["note"] != "updated note" {
			t.Errorf("expected note='updated note', got %v", body["note"])
		}

		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	note := "updated note"
	err := c.UpdateAlias(5, &UpdateAliasRequest{Note: &note})
	if err != nil {
		t.Fatalf("UpdateAlias: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ListAllAliases - pagination
// ---------------------------------------------------------------------------

func TestClient_ListAllAliases_Pagination(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp AliasListResponse
		switch callCount {
		case 1:
			resp = AliasListResponse{
				Aliases: []Alias{{ID: 1, Email: "a@sl.local"}, {ID: 2, Email: "b@sl.local"}},
			}
		case 2:
			resp = AliasListResponse{
				Aliases: []Alias{{ID: 3, Email: "c@sl.local"}},
			}
		default:
			// Empty page signals end of pagination
			resp = AliasListResponse{Aliases: []Alias{}}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	aliases, err := c.ListAllAliases(false, false, false, "")
	if err != nil {
		t.Fatalf("ListAllAliases: %v", err)
	}
	if len(aliases) != 3 {
		t.Errorf("expected 3 aliases across pages, got %d", len(aliases))
	}
	if callCount != 3 {
		t.Errorf("expected 3 API calls (2 with data + 1 empty), got %d", callCount)
	}
}

// ---------------------------------------------------------------------------
// NewClient defaults
// ---------------------------------------------------------------------------

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("my-key", "")
	if c.apiKey != "my-key" {
		t.Errorf("apiKey = %q, want %q", c.apiKey, "my-key")
	}
	if c.baseURL != BaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, BaseURL)
	}
	if c.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestNewClient_CustomBaseURL(t *testing.T) {
	c := NewClient("my-key", "https://sl.example.com")
	if c.baseURL != "https://sl.example.com" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "https://sl.example.com")
	}
}

// ---------------------------------------------------------------------------
// Authentication header
// ---------------------------------------------------------------------------

func TestClient_AuthenticationHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authentication")
		if auth != "secret-key" {
			t.Errorf("Authentication header = %q, want %q", auth, "secret-key")
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}
		w.WriteHeader(200)
		fmt.Fprint(w, `{"nb_alias":0,"nb_block":0,"nb_forward":0,"nb_reply":0}`)
	}))
	defer srv.Close()

	c := NewClient("secret-key", srv.URL)
	_, _, err := c.GetStats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// wrapNetworkError
// ---------------------------------------------------------------------------

// timeoutErr is a helper that satisfies net.Error with Timeout() == true.
type timeoutErr struct{}

func (timeoutErr) Error() string   { return "i/o timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return false }

func TestWrapNetworkError_DNS(t *testing.T) {
	orig := &url.Error{
		Op:  "Get",
		URL: "https://app.simplelogin.io",
		Err: &net.DNSError{Name: "app.simplelogin.io", Err: "no such host"},
	}
	wrapped := wrapNetworkError(orig)
	if !strings.Contains(wrapped.Error(), "could not resolve app.simplelogin.io") {
		t.Errorf("expected DNS error message containing hostname, got: %v", wrapped)
	}
	// Original error must be preserved in the chain
	var dnsErr *net.DNSError
	if !errors.As(wrapped, &dnsErr) {
		t.Error("original *net.DNSError should be preserved in the error chain")
	}
}

func TestWrapNetworkError_ConnectionRefused(t *testing.T) {
	orig := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &os.SyscallError{Syscall: "connect", Err: fmt.Errorf("connection refused")},
	}
	wrapped := wrapNetworkError(orig)
	if !strings.Contains(wrapped.Error(), "could not connect to SimpleLogin API") {
		t.Errorf("expected connection refused message, got: %v", wrapped)
	}
	var opErr *net.OpError
	if !errors.As(wrapped, &opErr) {
		t.Error("original *net.OpError should be preserved in the error chain")
	}
}

func TestWrapNetworkError_Timeout_OpError(t *testing.T) {
	orig := &net.OpError{
		Op:  "read",
		Net: "tcp",
		Err: timeoutErr{},
	}
	wrapped := wrapNetworkError(orig)
	if !strings.Contains(wrapped.Error(), "request timed out") {
		t.Errorf("expected timeout message, got: %v", wrapped)
	}
	var opErr *net.OpError
	if !errors.As(wrapped, &opErr) {
		t.Error("original *net.OpError should be preserved in the error chain")
	}
}

func TestWrapNetworkError_TLS_RemoteError(t *testing.T) {
	orig := &net.OpError{
		Op:  "remote error",
		Net: "tcp",
		Err: fmt.Errorf("tls: handshake failure"),
	}
	wrapped := wrapNetworkError(orig)
	if !strings.Contains(wrapped.Error(), "TLS handshake failed") {
		t.Errorf("expected TLS error message, got: %v", wrapped)
	}
	var opErr *net.OpError
	if !errors.As(wrapped, &opErr) {
		t.Error("original *net.OpError should be preserved in the error chain")
	}
}

func TestWrapNetworkError_TLS_RecordHeaderError(t *testing.T) {
	orig := &net.OpError{
		Op:  "read",
		Net: "tcp",
		Err: &tls.RecordHeaderError{
			Msg: "first record does not look like a TLS handshake",
		},
	}
	wrapped := wrapNetworkError(orig)
	if !strings.Contains(wrapped.Error(), "TLS handshake failed") {
		t.Errorf("expected TLS error message, got: %v", wrapped)
	}
}

func TestWrapNetworkError_Timeout_URLError(t *testing.T) {
	orig := &url.Error{
		Op:  "Get",
		URL: "https://app.simplelogin.io",
		Err: timeoutErr{},
	}
	wrapped := wrapNetworkError(orig)
	if !strings.Contains(wrapped.Error(), "request timed out") {
		t.Errorf("expected timeout message, got: %v", wrapped)
	}
	var urlErr *url.Error
	if !errors.As(wrapped, &urlErr) {
		t.Error("original *url.Error should be preserved in the error chain")
	}
}

func TestWrapNetworkError_Unknown(t *testing.T) {
	orig := fmt.Errorf("some unknown error")
	wrapped := wrapNetworkError(orig)
	if wrapped != orig {
		t.Errorf("expected unknown error to be returned as-is, got: %v", wrapped)
	}
}

// ---------------------------------------------------------------------------
// CreateCustomAlias
// ---------------------------------------------------------------------------

func TestClient_CreateCustomAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v3/alias/custom/new" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body CreateCustomAliasRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.AliasPrefix != "my-prefix" {
			t.Errorf("expected alias_prefix='my-prefix', got %q", body.AliasPrefix)
		}
		if body.SignedSuffix != ".abcdef@sl.local" {
			t.Errorf("expected signed_suffix='.abcdef@sl.local', got %q", body.SignedSuffix)
		}
		if len(body.MailboxIDs) != 2 || body.MailboxIDs[0] != 1 || body.MailboxIDs[1] != 3 {
			t.Errorf("expected mailbox_ids=[1,3], got %v", body.MailboxIDs)
		}
		if body.Note != "test note" {
			t.Errorf("expected note='test note', got %q", body.Note)
		}

		w.WriteHeader(201)
		alias := Alias{ID: 55, Email: "my-prefix.abcdef@sl.local"}
		_ = json.NewEncoder(w).Encode(alias)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	alias, _, err := c.CreateCustomAlias(&CreateCustomAliasRequest{
		AliasPrefix:  "my-prefix",
		SignedSuffix: ".abcdef@sl.local",
		MailboxIDs:   []int{1, 3},
		Note:         "test note",
	})
	if err != nil {
		t.Fatalf("CreateCustomAlias: %v", err)
	}
	if alias.ID != 55 {
		t.Errorf("expected ID 55, got %d", alias.ID)
	}
	if alias.Email != "my-prefix.abcdef@sl.local" {
		t.Errorf("expected email 'my-prefix.abcdef@sl.local', got %q", alias.Email)
	}
}

func TestClient_CreateCustomAlias_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"error":"premium required"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, _, err := c.CreateCustomAlias(&CreateCustomAliasRequest{
		AliasPrefix:  "test",
		SignedSuffix: ".xyz@sl.local",
		MailboxIDs:   []int{1},
	})
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	if !strings.Contains(err.Error(), "premium required") {
		t.Errorf("expected 'premium required' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// DeleteMailbox
// ---------------------------------------------------------------------------

func TestClient_DeleteMailbox_NoTransfer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/mailboxes/5" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// When transferTo is nil, the body should be empty (no JSON)
		bodyBytes, _ := io.ReadAll(r.Body)
		if len(bodyBytes) > 0 {
			t.Errorf("expected no request body when transferTo is nil, got: %s", string(bodyBytes))
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"deleted": true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.DeleteMailbox(5, nil)
	if err != nil {
		t.Fatalf("DeleteMailbox: %v", err)
	}
}

func TestClient_DeleteMailbox_WithTransfer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/mailboxes/5" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body DeleteMailboxRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.TransferAliasesTo == nil || *body.TransferAliasesTo != 10 {
			t.Errorf("expected transfer_aliases_to=10, got %v", body.TransferAliasesTo)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"deleted": true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	transferTo := 10
	err := c.DeleteMailbox(5, &transferTo)
	if err != nil {
		t.Fatalf("DeleteMailbox with transfer: %v", err)
	}
}

func TestClient_DeleteMailbox_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.DeleteMailbox(999, nil)
	if err == nil {
		t.Fatal("expected error for 404 delete")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetAllAliasActivities - pagination
// ---------------------------------------------------------------------------

func TestClient_GetAllAliasActivities_Pagination(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/aliases/42/activities") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		callCount++
		var resp ActivityResponse
		switch callCount {
		case 1:
			resp = ActivityResponse{
				Activities: []Activity{
					{Action: "forward", From: "a@example.com", To: "me@sl.local"},
					{Action: "block", From: "spam@example.com", To: "me@sl.local"},
				},
			}
		case 2:
			resp = ActivityResponse{
				Activities: []Activity{
					{Action: "reply", From: "me@sl.local", To: "b@example.com"},
				},
			}
		default:
			resp = ActivityResponse{Activities: []Activity{}}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	activities, err := c.GetAllAliasActivities(42)
	if err != nil {
		t.Fatalf("GetAllAliasActivities: %v", err)
	}
	if len(activities) != 3 {
		t.Errorf("expected 3 activities across pages, got %d", len(activities))
	}
	if callCount != 3 {
		t.Errorf("expected 3 API calls (2 with data + 1 empty), got %d", callCount)
	}
	if activities[0].Action != "forward" {
		t.Errorf("expected first activity action='forward', got %q", activities[0].Action)
	}
	if activities[2].Action != "reply" {
		t.Errorf("expected third activity action='reply', got %q", activities[2].Action)
	}
}

func TestClient_GetAllAliasActivities_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ActivityResponse{Activities: []Activity{}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	activities, err := c.GetAllAliasActivities(42)
	if err != nil {
		t.Fatalf("GetAllAliasActivities: %v", err)
	}
	if len(activities) != 0 {
		t.Errorf("expected 0 activities, got %d", len(activities))
	}
}

// ---------------------------------------------------------------------------
// ExportData / ExportAliases
// ---------------------------------------------------------------------------

func TestClient_ExportData(t *testing.T) {
	rawData := `{"aliases": [{"email": "a@sl.local"}], "mailboxes": [{"email": "me@example.com"}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/export/data" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(rawData))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	data, err := c.ExportData()
	if err != nil {
		t.Fatalf("ExportData: %v", err)
	}
	if string(data) != rawData {
		t.Errorf("expected raw data %q, got %q", rawData, string(data))
	}
}

func TestClient_ExportData_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.ExportData()
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestClient_ExportAliases(t *testing.T) {
	csvData := "alias,enabled,note\na@sl.local,true,test\nb@sl.local,false,"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/export/aliases" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(csvData))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	data, err := c.ExportAliases()
	if err != nil {
		t.Fatalf("ExportAliases: %v", err)
	}
	if string(data) != csvData {
		t.Errorf("expected CSV data %q, got %q", csvData, string(data))
	}
}

func TestClient_ExportAliases_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.ExportAliases()
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
