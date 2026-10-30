package app

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Files remain available after a gateway restart so agents can read sent paths.
func serveFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !sameOrigin(r) {
		http.Error(w, "cross-origin request denied", http.StatusForbidden)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 20<<20)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "File exceeds 20 MB or could not be read", http.StatusRequestEntityTooLarge)
		return
	}
	name := filepath.Base(strings.ReplaceAll(r.URL.Query().Get("name"), "\\", "/"))
	if name == "." || name == "/" || name == "" {
		name = "attachment"
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		http.Error(w, "File storage unavailable", 500)
		return
	}
	dir := filepath.Join(cache, "herdr-tty", "files")
	if err := os.MkdirAll(dir, 0700); err != nil {
		http.Error(w, "File storage unavailable", 500)
		return
	}
	uploadDir, err := os.MkdirTemp(dir, "upload-")
	if err != nil {
		http.Error(w, "File storage unavailable", 500)
		return
	}
	f, err := os.OpenFile(filepath.Join(uploadDir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		http.Error(w, "File storage unavailable", 500)
		return
	}
	_, err = f.Write(data)
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		os.RemoveAll(uploadDir)
		http.Error(w, "File could not be saved", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"path": f.Name()})
}
