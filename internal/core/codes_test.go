package core

import (
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestAllCodes_ExactSourceParity(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}

	reCode := regexp.MustCompile(`\b(?:E|V|W|Q)-[A-Z0-9]+(?:-[A-Z0-9]+)*\b`)
	reQualityCat := regexp.MustCompile(`Category:\s*"([a-z_]+)"`)

	emitted := make(map[string]bool)
	for _, sub := range []string{"cmd", "internal"} {
		walkRoot := filepath.Join(repoRoot, sub)
		err := filepath.WalkDir(walkRoot, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, _ := filepath.Rel(repoRoot, path)
			relSlash := filepath.ToSlash(rel)
			if relSlash == "internal/core/codes.go" {
				return nil
			}
			src, rErr := os.ReadFile(path)
			if rErr != nil {
				return rErr
			}
			if relSlash == "internal/core/quality.go" {
				for _, m := range reQualityCat.FindAllStringSubmatch(string(src), -1) {
					emitted["Q-"+strings.ToUpper(m[1])] = true
				}
			}

			fset := token.NewFileSet()
			file := fset.AddFile(path, fset.Base(), len(src))
			var s scanner.Scanner
			s.Init(file, src, nil, 0)
			for {
				_, tok, lit := s.Scan()
				if tok == token.EOF {
					break
				}
				if tok != token.STRING {
					continue
				}
				unq, uErr := strconv.Unquote(lit)
				if uErr != nil {
					continue
				}
				for _, m := range reCode.FindAllString(unq, -1) {
					switch m {
					case "E-A", "E-B", "E-C", "E-D", "E-E":
						continue
					}
					emitted[m] = true
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", sub, err)
		}
	}

	entries := AllCodes()
	registered := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if registered[entry.Code] {
			t.Errorf("duplicate code in AllCodes(): %s", entry.Code)
		}
		if entry.Level == "" || entry.Message == "" || entry.Fix == "" {
			t.Errorf("incomplete CodeEntry for %s: %+v", entry.Code, entry)
		}
		registered[entry.Code] = true
	}

	var missingFromRegistry, deadInRegistry []string
	for c := range emitted {
		if !registered[c] {
			missingFromRegistry = append(missingFromRegistry, c)
		}
	}
	for c := range registered {
		if !emitted[c] {
			deadInRegistry = append(deadInRegistry, c)
		}
	}
	sort.Strings(missingFromRegistry)
	sort.Strings(deadInRegistry)

	if len(missingFromRegistry) > 0 {
		t.Errorf("codes emitted in source but missing from core.AllCodes(): %v", missingFromRegistry)
	}
	if len(deadInRegistry) > 0 {
		t.Errorf("dead codes in core.AllCodes() emitted nowhere in source: %v", deadInRegistry)
	}

	codesDocBytes, err := os.ReadFile(filepath.Join(repoRoot, "docs", "reference", "codes.md"))
	if err != nil {
		t.Fatalf("reading docs/reference/codes.md: %v", err)
	}
	reDocCodeRow := regexp.MustCompile("(?m)^\\|\\s*`([EVWQ]-[A-Z0-9-]+)`\\s*\\|")
	docCodes := make(map[string]bool)
	for _, m := range reDocCodeRow.FindAllStringSubmatch(string(codesDocBytes), -1) {
		docCodes[m[1]] = true
	}
	for c := range registered {
		if !docCodes[c] {
			t.Errorf("code %s missing from docs/reference/codes.md (run go run ./tools/docgen)", c)
		}
	}
	for c := range docCodes {
		if !registered[c] {
			t.Errorf("unregistered code %s present in docs/reference/codes.md", c)
		}
	}
}
