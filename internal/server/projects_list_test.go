package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hacrex/Archon-AI/internal/api"
)

// multiProjectStore returns a deterministic set of projects ordered by name.
type multiProjectStore struct {
	*fakeStore
	projects []api.Project
}

func newMultiProjectStore(names ...string) *multiProjectStore {
	projects := make([]api.Project, 0, len(names))
	for i, name := range names {
		p := api.Project{}
		p.Metadata.Name = name
		if i%2 == 0 {
			p.Status.Phase = "Ready"
		} else {
			p.Status.Phase = "Pending"
		}
		projects = append(projects, p)
	}
	return &multiProjectStore{fakeStore: &fakeStore{}, projects: projects}
}

func (m *multiProjectStore) ListProjects(_ context.Context) ([]api.Project, error) {
	return m.projects, nil
}

func doList(t *testing.T, s *multiProjectStore, target string) map[string]any {
	t.Helper()
	srv := New(s)
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: status %d body %s", target, rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return body
}

func itemNames(body map[string]any) []string {
	raw, _ := body["items"].([]any)
	names := make([]string, 0, len(raw))
	for _, it := range raw {
		m := it.(map[string]any)
		meta := m["metadata"].(map[string]any)
		names = append(names, meta["name"].(string))
	}
	return names
}

func TestListProjectsPagination(t *testing.T) {
	store := newMultiProjectStore("alpha", "bravo", "charlie", "delta")

	body := doList(t, store, "/v1/projects?limit=2")
	names := itemNames(body)
	if len(names) != 2 || names[0] != "alpha" || names[1] != "bravo" {
		t.Fatalf("first page = %v, want [alpha bravo]", names)
	}
	cursor, ok := body["next_cursor"].(string)
	if !ok || cursor != "bravo" {
		t.Fatalf("next_cursor = %v, want bravo", body["next_cursor"])
	}

	body2 := doList(t, store, "/v1/projects?limit=2&cursor="+cursor)
	names2 := itemNames(body2)
	if len(names2) != 2 || names2[0] != "charlie" || names2[1] != "delta" {
		t.Fatalf("second page = %v, want [charlie delta]", names2)
	}
	if _, exists := body2["next_cursor"]; exists {
		t.Fatalf("last page must not carry next_cursor, got %v", body2["next_cursor"])
	}
}

func TestListProjectsFiltering(t *testing.T) {
	store := newMultiProjectStore("alpha", "bravo", "charlie")

	body := doList(t, store, "/v1/projects?phase=ready")
	names := itemNames(body)
	if len(names) != 2 || names[0] != "alpha" || names[1] != "charlie" {
		t.Fatalf("phase filter = %v, want [alpha charlie]", names)
	}

	body = doList(t, store, "/v1/projects?name=bravo")
	names = itemNames(body)
	if len(names) != 1 || names[0] != "bravo" {
		t.Fatalf("name filter = %v, want [bravo]", names)
	}

	body = doList(t, store, "/v1/projects?phase=pending&name=bravo")
	names = itemNames(body)
	if len(names) != 1 || names[0] != "bravo" {
		t.Fatalf("combined filter = %v, want [bravo]", names)
	}

	body = doList(t, store, "/v1/projects?phase=ready&name=bravo")
	if names := itemNames(body); len(names) != 0 {
		t.Fatalf("contradictory filter = %v, want []", names)
	}
}

func TestListProjectsRejectsBadPaginationParams(t *testing.T) {
	store := newMultiProjectStore("alpha")

	srv := New(store)
	for _, target := range []string{"/v1/projects?limit=0", "/v1/projects?limit=abc", "/v1/projects?limit=201", "/v1/projects?cursor=missing"} {
		req := httptest.NewRequest(http.MethodGet, target, nil)
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("GET %s: status %d, want 422 (%s)", target, rec.Code, rec.Body.String())
		}
	}
}

func TestPaginateProjectsEmptyAndExactPage(t *testing.T) {
	projects := newMultiProjectStore("a", "b").projects

	page, next, err := paginateProjects(projects, url.Values{})
	if err != nil || len(page) != 2 || next != "" {
		t.Fatalf("default page: len=%d next=%q err=%v", len(page), next, err)
	}

	page, next, err = paginateProjects(projects, mustParseQuery(t, "limit=2"))
	if err != nil || len(page) != 2 || next != "" {
		t.Fatalf("exact-size page should have no cursor: len=%d next=%q err=%v", len(page), next, err)
	}

	page, next, err = paginateProjects(projects, mustParseQuery(t, "limit=1&cursor=b"))
	if err != nil || len(page) != 0 || next != "" {
		t.Fatalf("cursor at end: len=%d next=%q err=%v", len(page), next, err)
	}
}

func mustParseQuery(t *testing.T, raw string) url.Values {
	t.Helper()
	u, err := url.Parse("/v1/projects?" + raw)
	if err != nil {
		t.Fatalf("parse query %q: %v", raw, err)
	}
	return u.Query()
}
