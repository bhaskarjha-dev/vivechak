package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/bhaskarjha-dev/vivechak/internal/core"
)

// TestValidateComparison_SaveReviewParity_LiveSubprocess checks on the built binary over stdio that vivechak_review
// (mode="validate", artifact_type="comparison") judges a comparison exactly as vivechak_save(kind="comparison") does:
// for two broken comparisons (weights that do not sum to 1; a matrix with one option) both report the same blocking
// codes, both say invalid, and the save stores the file as a draft (HIGH-18).
func TestValidateComparison_SaveReviewParity_LiveSubprocess(t *testing.T) {
	binPath := resolveOrBuildTestBinary(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dir := wgNewProject(t)
	cs, cleanup := startLiveMCPServer(t, ctx, binPath, dir)
	defer cleanup()

	valid := wgComparison("Pino vs Winston", "Pino", "Winston")
	matrix := "| Criterion | Weight | Pino | Winston |\n|---|---|---|---|\n| Throughput | 0.6 | 5 | 3 |\n| Ecosystem | 0.4 | 3 | 5 |\n| **Weighted total** | 1.0 | 4.2 | 3.8 |"
	oneOption := "| Criterion | Weight | Pino |\n|---|---|---|\n| Throughput | 0.6 | 5 |\n| Ecosystem | 0.4 | 3 |\n| **Weighted total** | 1.0 | 4.2 |"
	if !strings.Contains(valid, matrix) {
		t.Fatalf("fixture error: the comparison fixture's matrix changed")
	}

	codes := func(v any) []string {
		raw, _ := json.Marshal(v)
		var issues []struct {
			Code string `json:"code"`
		}
		_ = json.Unmarshal(raw, &issues)
		var out []string
		for _, i := range issues {
			out = append(out, i.Code)
		}
		sort.Strings(out)
		return out
	}

	for _, tc := range []struct {
		name, content, wantCode string
	}{
		{"weights sum to 1.3", strings.Replace(valid, "| Throughput | 0.6 | 5 | 3 |", "| Throughput | 0.9 | 5 | 3 |", 1), "V-CMP-WEIGHTS"},
		{"one option", strings.Replace(valid, matrix, oneOption, 1), "V-CMP-OPTIONS"},
	} {
		review, isErr := wgCall(t, ctx, cs, "vivechak_review", map[string]any{"mode": "validate", "artifact_type": "comparison", "content": tc.content})
		if isErr {
			t.Fatalf("%s: vivechak_review failed: %s", tc.name, review.Message)
		}
		save, isErr := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "comparison", "content": tc.content})
		if isErr {
			t.Fatalf("%s: vivechak_save failed: %s", tc.name, save.Message)
		}
		validation, _ := save.Data["validation"].(map[string]any)
		saveCodes, reviewCodes := codes(validation["issues"]), codes(review.Data["issues"])
		if !reflect.DeepEqual(saveCodes, []string{tc.wantCode}) || !reflect.DeepEqual(reviewCodes, saveCodes) {
			t.Errorf("%s: expected save and review to report exactly [%s], got save %v, review %v", tc.name, tc.wantCode, saveCodes, reviewCodes)
		}
		if save.Data["valid"] != false || review.Data["valid"] != false || review.Data["blocking_count"] != float64(1) {
			t.Errorf("%s: expected both to say invalid (one blocking issue), got save valid=%v, review valid=%v blocking=%v",
				tc.name, save.Data["valid"], review.Data["valid"], review.Data["blocking_count"])
		}
		fp, _ := save.Data["file_path"].(string)
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(fp)))
		if fm, _, _ := core.ParseFrontmatter(data); err != nil || fm.GetString("status") != "draft" {
			t.Errorf("%s: expected the broken comparison stored as a draft at %q (err %v)", tc.name, fp, err)
		}
	}

	// The valid comparison passes both, with no blocking issue.
	review, _ := wgCall(t, ctx, cs, "vivechak_review", map[string]any{"mode": "validate", "artifact_type": "comparison", "content": valid})
	save, _ := wgCall(t, ctx, cs, "vivechak_save", map[string]any{"kind": "comparison", "content": valid})
	if review.Data["valid"] != true || save.Data["valid"] != true || review.Data["blocking_count"] != float64(0) {
		t.Errorf("the valid comparison: expected valid from both, got save valid=%v, review valid=%v blocking=%v (%v)",
			save.Data["valid"], review.Data["valid"], review.Data["blocking_count"], review.Data["issues"])
	}
}
