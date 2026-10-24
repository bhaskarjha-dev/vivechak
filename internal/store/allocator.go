package store

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type manifestData struct {
	Schema    int            `json:"schema,omitempty"`
	Scope     string         `json:"scope,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
	Counters  map[string]int `json:"counters,omitempty"`
}

var idNumRe = regexp.MustCompile(`^([A-Za-z]+)-(\d+)`)

func scanHighestID(ws *Workspace, prefix string) int {
	normPrefix := strings.ToUpper(prefix)
	highest := 0

	scanDir := func(dir string) {
		entries, err := ws.ListDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			m := idNumRe.FindStringSubmatch(strings.ToUpper(e.Name()))
			if len(m) == 3 && m[1] == normPrefix {
				if n, err := strconv.Atoi(m[2]); err == nil && n > highest {
					highest = n
				}
			}
		}
	}

	scanDir("research")
	scanDir("research/comparisons")
	scanDir("research/decisions")
	scanDir("research/sessions")

	return highest
}

// AllocateID safely increments the monotonic counter for the given prefix (e.g. "C", "D")
// stored in research/.vivechak.json, and returns an allocated ID (e.g. "C-001", "D-001").
func AllocateID(ws *Workspace, prefix string) (string, error) {
	normPrefix := strings.ToUpper(strings.TrimSpace(prefix))
	if normPrefix == "" {
		normPrefix = "D"
	}

	manifestRel := "research/.vivechak.json"
	manifestLockPath := filepath.Join(ws.AbsPath(), manifestRel)

	unlock, err := LockFile(context.Background(), manifestLockPath, 5*time.Second)
	if err != nil {
		return "", fmt.Errorf("acquiring manifest lock: %w", err)
	}
	defer func() { _ = unlock() }()

	var m manifestData
	m.Schema = 1
	m.Counters = make(map[string]int)

	if ws.FileExists(manifestRel) {
		data, err := ws.ReadFile(manifestRel)
		if err == nil && len(data) > 0 {
			_ = json.Unmarshal(data, &m)
			if m.Counters == nil {
				m.Counters = make(map[string]int)
			}
		}
	} else {
		m.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	current := m.Counters[normPrefix]
	if current == 0 {
		highest := scanHighestID(ws, normPrefix)
		if highest > current {
			current = highest
		}
	}

	next := current + 1
	m.Counters[normPrefix] = next

	updatedBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling updated manifest: %w", err)
	}

	if err := ws.MkdirAll("research", 0o755); err != nil {
		return "", fmt.Errorf("ensuring research dir: %w", err)
	}

	if err := WriteFileAtomic(ws.Root(), manifestRel, updatedBytes, 0o644); err != nil {
		return "", fmt.Errorf("writing updated manifest: %w", err)
	}

	return fmt.Sprintf("%s-%03d", normPrefix, next), nil
}
