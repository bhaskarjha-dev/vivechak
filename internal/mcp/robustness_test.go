package mcputil

import (
	"strings"
	"testing"
)

func TestExtractComparisonOptions_Robustness(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Latin Extended with Kelvin Sign and vs",
			input:    "ȺȺ vs B",
			expected: "ȺȺ, B",
		},
		{
			name:     "Turkish dotted I uppercase and Stripe",
			input:    "İyzico vs Stripe",
			expected: "İyzico, Stripe",
		},
		{
			name:     "Multi-option comparison with versus and or",
			input:    "PostgreSQL versus MySQL or CockroachDB",
			expected: "PostgreSQL, MySQL, CockroachDB",
		},
		{
			name:     "Comparison with contextual suffix",
			input:    "Compare Pinot vs ClickHouse for analytical queries",
			expected: "Pinot, ClickHouse",
		},
		{
			name:     "Kelvin Sign Kafka vs Pulsar",
			input:    "\u212Aafka vs Pulsar",
			expected: "\u212Aafka, Pulsar",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractComparisonOptions(tc.input)
			if got != tc.expected {
				t.Errorf("extractComparisonOptions(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func FuzzExtractComparisonOptions(f *testing.F) {
	seeds := []string{
		"A vs B",
		"ȺȺ vs B",
		"İyzico vs Stripe",
		"Compare PostgreSQL vs MySQL for event logs",
		"foo against bar",
		"Kafka versus Pulsar or RabbitMQ",
		"React vs Vue vs Svelte in production",
		"Which is better: Go vs Rust for CLI tools",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Must not panic on arbitrary Unicode or malformed inputs
		opts := extractComparisonOptions(input)
		if opts == "" {
			return
		}

		parts := strings.Split(opts, ", ")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			// Every extracted option must be a substring of the original input
			if !strings.Contains(input, part) {
				t.Fatalf("extracted option %q is not a substring of input %q (opts=%q)", part, input, opts)
			}
		}
	})
}
