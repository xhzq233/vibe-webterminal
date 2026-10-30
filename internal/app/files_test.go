package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestFileBrowserAuthenticatedBrowseAndDownload(t *testing.T) {
	dir := t.TempDir()
	name := "中文 file #1.html"
	content := []byte("<script>alert('must download, never execute')</script>")
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "folder"), 0700); err != nil {
		t.Fatal(err)
	}
	auth := testAuthenticator(t)
	server := httptest.NewServer(newGatewayHandler(auth, http.NotFoundHandler()))
	defer server.Close()
	request := func(route, path string, authenticated bool) *http.Response {
		t.Helper()
		r, _ := http.NewRequest(http.MethodGet, server.URL+route+"?"+url.Values{"path": {path}}.Encode(), nil)
		if authenticated {
			r.AddCookie(&http.Cookie{Name: cookieName, Value: auth.token()})
		}
		client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		response, err := client.Do(r)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { response.Body.Close() })
		return response
	}
	for _, route := range []string{"/_herdr/files", "/_herdr/download"} {
		if response := request(route, path, false); response.StatusCode != http.StatusSeeOther {
			t.Fatalf("anonymous %s: %d", route, response.StatusCode)
		}
	}
	response := request("/_herdr/files", dir, true)
	var listing struct {
		Path    string
		Entries []browserFile
	}
	if err := json.NewDecoder(response.Body).Decode(&listing); err != nil {
		t.Fatal(err)
	}
	if len(listing.Entries) != 2 || !listing.Entries[0].Directory || listing.Entries[1].Path != path {
		t.Fatalf("unexpected directory listing: %+v", listing)
	}
	response = request("/_herdr/download", path, true)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 || !bytes.Equal(body, content) {
		t.Fatalf("download failed: status %d", response.StatusCode)
	}
	if response.Header.Get("Content-Type") != "application/octet-stream" || !strings.HasPrefix(response.Header.Get("Content-Disposition"), "attachment;") {
		t.Fatal("download could execute active content")
	}
	if response.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("file response allows caching")
	}
	response = request("/_herdr/files", filepath.Join(dir, "does-not-exist"), true)
	if response.StatusCode != 404 {
		t.Fatalf("missing file: %d", response.StatusCode)
	}
}

func TestFileBrowserBoundedPreviewAndBinary(t *testing.T) {
	dir := t.TempDir()
	text := strings.Repeat("中", filePreviewLimit)
	for _, item := range []struct {
		name    string
		content []byte
		binary  bool
	}{
		{"large.txt", []byte(text), false},
		{"binary.bin", []byte{0, 1, 255, 20}, true},
	} {
		path := filepath.Join(dir, item.name)
		if err := os.WriteFile(path, item.content, 0600); err != nil {
			t.Fatal(err)
		}
		r := httptest.NewRequest(http.MethodGet, "/_herdr/files?"+url.Values{"path": {path}}.Encode(), nil)
		w := httptest.NewRecorder()
		serveFileBrowser(w, r)
		var result map[string]json.RawMessage
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		preview, exists := result["text"]
		if item.binary {
			if exists {
				t.Fatal("binary file offered as text")
			}
			continue
		}
		var decoded string
		if err := json.Unmarshal(preview, &decoded); err != nil {
			t.Fatal(err)
		}
		if !utf8.ValidString(decoded) || len(decoded) > filePreviewLimit || !strings.HasPrefix(text, decoded) {
			t.Fatal("preview corrupted multibyte text")
		}
		var truncated bool
		_ = json.Unmarshal(result["truncated"], &truncated)
		if !truncated {
			t.Fatal("large preview did not report truncation")
		}
	}
}

func TestFileBrowserHomePathsAndReadOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, value := range []string{"", "~", "~/folder/.."} {
		got, err := fileBrowserPath(value)
		if err != nil || got != home {
			t.Fatalf("home expansion %q: %q %v", value, got, err)
		}
	}
	w := httptest.NewRecorder()
	serveFileBrowser(w, httptest.NewRequest(http.MethodPost, "/_herdr/files", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatal("browser accepted write operation")
	}
}
