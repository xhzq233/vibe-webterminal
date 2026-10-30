package app

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const filePreviewLimit = 256 << 10

type browserFile struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Directory bool      `json:"directory"`
	Size      int64     `json:"size"`
	Modified  time.Time `json:"modified"`
}

func fileBrowserPath(value string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if value == "" || value == "~" {
		return home, nil
	}
	if strings.HasPrefix(value, "~/") {
		value = filepath.Join(home, value[2:])
	}
	if !filepath.IsAbs(value) {
		value = filepath.Join(home, value)
	}
	return filepath.Clean(value), nil
}

func fileBrowserError(w http.ResponseWriter, err error) {
	status, message := http.StatusInternalServerError, "无法读取文件"
	if os.IsNotExist(err) {
		status, message = http.StatusNotFound, "路径不存在"
	}
	if os.IsPermission(err) {
		status, message = http.StatusForbidden, "没有访问权限"
	}
	http.Error(w, message, status)
}

// The browser has the same filesystem permissions as the terminal's OS user.
// Downloads are attachments; remote HTML and scripts never execute in this origin.
func serveFileBrowser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path, err := fileBrowserPath(r.URL.Query().Get("path"))
	if err != nil {
		fileBrowserError(w, err)
		return
	}
	initialInfo, err := os.Stat(path)
	if err != nil {
		fileBrowserError(w, err)
		return
	}
	if !initialInfo.IsDir() && !initialInfo.Mode().IsRegular() {
		http.Error(w, "不支持访问此文件类型", http.StatusBadRequest)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		fileBrowserError(w, err)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		fileBrowserError(w, err)
		return
	}
	if r.URL.Path == "/_herdr/download" {
		if !info.Mode().IsRegular() {
			http.Error(w, "只能下载普通文件", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": info.Name()}))
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeContent(w, r, info.Name(), info.ModTime(), f)
		return
	}
	result := map[string]any{"path": path, "parent": filepath.Dir(path), "name": info.Name(), "directory": info.IsDir(), "size": info.Size()}
	if info.IsDir() {
		entries, err := f.ReadDir(-1)
		if err != nil {
			fileBrowserError(w, err)
			return
		}
		files := make([]browserFile, 0, len(entries))
		for _, entry := range entries {
			entryPath := filepath.Join(path, entry.Name())
			entryInfo, err := os.Stat(entryPath)
			if err != nil {
				continue
			}
			files = append(files, browserFile{entry.Name(), entryPath, entryInfo.IsDir(), entryInfo.Size(), entryInfo.ModTime()})
		}
		sort.Slice(files, func(i, j int) bool {
			if files[i].Directory != files[j].Directory {
				return files[i].Directory
			}
			return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
		})
		result["entries"] = files
	} else if info.Mode().IsRegular() {
		data, err := io.ReadAll(io.LimitReader(f, filePreviewLimit+4))
		if err != nil {
			fileBrowserError(w, err)
			return
		}
		result["contentType"] = http.DetectContentType(data)
		truncated := len(data) > filePreviewLimit
		if truncated {
			data = data[:filePreviewLimit]
			for len(data) > 0 && !utf8.Valid(data) && len(data) > filePreviewLimit-4 {
				data = data[:len(data)-1]
			}
		}
		if utf8.Valid(data) && !bytes.ContainsRune(data, 0) {
			result["text"] = string(data)
		}
		result["truncated"] = truncated
	} else {
		http.Error(w, "不支持预览此文件类型", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodHead {
		_ = json.NewEncoder(w).Encode(result)
	}
}
