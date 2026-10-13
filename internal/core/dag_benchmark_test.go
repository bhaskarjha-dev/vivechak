package core

import (
	"strings"
	"testing"
)

func BenchmarkParsePipeline(b *testing.B) {
	data := []byte(samplePipeline)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		dag, err := ParsePipeline(data)
		if err != nil {
			b.Fatalf("failed to parse pipeline: %v", err)
		}
		if len(dag.Sessions) != 4 {
			b.Fatalf("expected 4 sessions, got %d", len(dag.Sessions))
		}
	}
}

func BenchmarkAssignPromptToSession(b *testing.B) {
	lines := strings.Split(`
# RESEARCH BRIEF: T1-01 Real-Time Stream Processor
Evaluate Kafka vs NATS JetStream given T0-01 ingestion findings.
## Context
Ingestion pipeline runs at 50k events/sec.
`, "\n")

	currentSession := &Session{ID: "T1-01", Title: "Stream Processor"}
	sessionsByID := map[string]*Session{
		"T1-01": currentSession,
		"T1-02": {ID: "T1-02", Title: "Analytics Store"},
	}
	sessionOrder := []string{"T1-01", "T1-02"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		currentSession.Prompt = ""
		s := assignPromptToSession(lines, currentSession, sessionsByID, sessionOrder)
		if s == nil || s.Prompt == "" {
			b.Fatalf("expected assigned prompt")
		}
	}
}
