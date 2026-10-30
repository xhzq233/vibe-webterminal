package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestTabDirectorySelectsFocusedPaneAndLiveDirectory(t *testing.T) {
	other, shell, foreground := t.TempDir(), t.TempDir(), t.TempDir()
	data, _ := json.Marshal(map[string]any{"result": map[string]any{"snapshot": map[string]any{
		"focused_pane_id": "w2:p7",
		"panes": []map[string]string{
			{"pane_id": "w1:p1", "tab_id": "w1:t1", "cwd": other},
			{"pane_id": "w2:p7", "tab_id": "w2:t3", "cwd": shell, "foreground_cwd": foreground},
		},
	}}})
	got, err := directoryFromSnapshot(data)
	if err != nil || got.Path != foreground || got.TabID != "w2:t3" || got.PaneID != "w2:p7" {
		t.Fatalf("wrong focused directory: %+v %v", got, err)
	}
	if err := os.Remove(foreground); err != nil {
		t.Fatal(err)
	}
	got, err = directoryFromSnapshot(data)
	if err != nil || got.Path != shell {
		t.Fatalf("shell cwd fallback failed: %+v %v", got, err)
	}
}

func TestTabDirectoryDoesNotGuessAnotherPane(t *testing.T) {
	for _, data := range []string{
		`{"error":{"code":"unavailable"}}`,
		`{"result":{"snapshot":{"focused_pane_id":"missing","panes":[{"pane_id":"other","cwd":"/"}]}}}`,
	} {
		if got, err := directoryFromSnapshot([]byte(data)); err == nil {
			t.Fatalf("guessed directory: %+v", got)
		}
	}
}

func TestTabDirectoryEndpointUsesConfiguredSessionAndRequiresLogin(t *testing.T) {
	dir := t.TempDir()
	data, _ := json.Marshal(map[string]any{"result": map[string]any{"snapshot": map[string]any{
		"focused_pane_id": "w1:p2",
		"panes":           []map[string]string{{"pane_id": "w1:p2", "tab_id": "w1:t2", "cwd": dir}},
	}}})
	fixture := filepath.Join(dir, "snapshot.json")
	if err := os.WriteFile(fixture, data, 0600); err != nil {
		t.Fatal(err)
	}
	helper := filepath.Join(dir, "herdr")
	script := "#!/bin/sh\n[ \"$1\" = --session ] && [ \"$2\" = web-test ] && [ \"$3\" = api ] && [ \"$4\" = snapshot ] || exit 1\ncat \"$(dirname \"$0\")/snapshot.json\"\n"
	if err := os.WriteFile(helper, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	auth := testAuthenticator(t)
	handler := newGatewayHandler(auth, http.NotFoundHandler(), Config{Herdr: helper, Session: "web-test"})
	r := httptest.NewRequest(http.MethodGet, "/_herdr/cwd", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("anonymous cwd: %d", w.Code)
	}
	r.AddCookie(&http.Cookie{Name: cookieName, Value: auth.token()})
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	var got tabDirectory
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || got.Path != dir {
		t.Fatalf("cwd request failed: %d %+v", w.Code, got)
	}
}
