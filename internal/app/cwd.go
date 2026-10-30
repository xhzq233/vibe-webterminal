package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type tabDirectory struct {
	Path   string `json:"path"`
	TabID  string `json:"tab_id"`
	PaneID string `json:"pane_id"`
}

func directoryFromSnapshot(data []byte) (tabDirectory, error) {
	var response struct {
		Result struct {
			Snapshot struct {
				FocusedPaneID string `json:"focused_pane_id"`
				Panes         []struct {
					PaneID        string `json:"pane_id"`
					TabID         string `json:"tab_id"`
					CWD           string `json:"cwd"`
					ForegroundCWD string `json:"foreground_cwd"`
				} `json:"panes"`
			} `json:"snapshot"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return tabDirectory{}, err
	}
	snapshot := response.Result.Snapshot
	for _, pane := range snapshot.Panes {
		if pane.PaneID == "" || pane.PaneID != snapshot.FocusedPaneID {
			continue
		}
		for _, path := range []string{pane.ForegroundCWD, pane.CWD} {
			if !filepath.IsAbs(path) {
				continue
			}
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				return tabDirectory{path, pane.TabID, pane.PaneID}, nil
			}
		}
	}
	return tabDirectory{}, errors.New("focused pane has no accessible directory")
}

func serveTabDirectory(w http.ResponseWriter, r *http.Request, config Config) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	args := []string{}
	if config.Session != "" {
		args = append(args, "--session", config.Session)
	}
	args = append(args, "api", "snapshot")
	command := exec.CommandContext(ctx, config.Herdr, args...)
	command.Env = childEnvironment(os.Environ())
	data, err := command.Output()
	if err == nil {
		var directory tabDirectory
		directory, err = directoryFromSnapshot(data)
		if err == nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(directory)
			return
		}
	}
	http.Error(w, "无法识别当前 Tab 的目录，请手动输入路径或点主目录", http.StatusServiceUnavailable)
}
