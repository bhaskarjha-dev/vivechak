package core

import (
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
