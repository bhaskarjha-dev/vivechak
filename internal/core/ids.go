package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// NormalizeDecisionID normalizes IDs like "d-9" or "D-1" to canonical "D-009", "D-001".
func NormalizeDecisionID(id string) string {
	trimmed := strings.TrimSpace(id)
	if m := regexp.MustCompile(`^(?i)D-(\d+)$`).FindStringSubmatch(trimmed); len(m) > 1 {
		num, _ := strconv.Atoi(m[1])
		return fmt.Sprintf("D-%03d", num)
	}
	return trimmed
}

// NormalizeComparisonID normalizes IDs like "c-1" or "C-01" to canonical "C-001".
func NormalizeComparisonID(id string) string {
	trimmed := strings.TrimSpace(id)
	if m := regexp.MustCompile(`^(?i)C-(\d+)$`).FindStringSubmatch(trimmed); len(m) > 1 {
		num, _ := strconv.Atoi(m[1])
		return fmt.Sprintf("C-%03d", num)
	}
	return trimmed
}
