package app

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Images remain available after a gateway restart so agents can read sent paths.
func serveImageUpload(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Image exceeds 20 MB or could not be read", http.StatusRequestEntityTooLarge)
		return
	}
	extensions := map[string]string{"image/png": ".png", "image/jpeg": ".jpg", "image/gif": ".gif", "image/webp": ".webp"}
	ext, ok := extensions[http.DetectContentType(data)]
	if !ok {
		http.Error(w, "Use PNG, JPEG, GIF or WebP", http.StatusUnsupportedMediaType)
		return
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		http.Error(w, "Image storage unavailable", 500)
		return
	}
	dir := filepath.Join(cache, "herdr-tty", "images")
	if err := os.MkdirAll(dir, 0700); err != nil {
		http.Error(w, "Image storage unavailable", 500)
		return
	}
	f, err := os.CreateTemp(dir, "image-*"+ext)
	if err != nil {
		http.Error(w, "Image storage unavailable", 500)
		return
	}
	_, err = f.Write(data)
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		os.Remove(f.Name())
		http.Error(w, "Image could not be saved", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"path": f.Name()})
}
