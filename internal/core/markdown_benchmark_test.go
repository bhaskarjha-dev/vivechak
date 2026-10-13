package core

import "testing"

var benchmarkMarkdownDocument = `
# Session Document: Primary Datastore Evaluation

## Context & Problem Statement
We need an analytical columnar engine ingesting 50k events/sec with sub-second queries.
The datastore must maintain persistent compression across 90 days of retention.

### Sub-criteria
1. Tenant data isolation
2. Memory footprint under 256MB RAM per node
3. SQL compatibility

## Evaluated Options
- Option 1: ClickHouse (embedded single binary)
- Option 2: VictoriaLogs
- Option 3: DuckDB

## Recommendation
Adopt ClickHouse embedded single binary mode. Provides 10x-14x zstd compression on structured JSON logs.
Secondary indices allow fast tenant isolation queries without memory overhead.

## Discovered Concerns
Recursive CTE traversal degrades beyond 10 hops without depth bounding.
Memory footprint spikes under concurrent ingest if block flush intervals are misconfigured.

## Sources
- [Grade A] https://clickhouse.com/docs/en/about-us/architecture
- [Grade B] https://github.com/ClickHouse/ClickHouse/tree/master/benchmark
`

func BenchmarkExtractSection(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sec := ExtractSection(benchmarkMarkdownDocument, "Recommendation")
		if len(sec) == 0 {
			b.Fatalf("expected non-empty recommendation")
		}
	}
}

func BenchmarkExtractSection_DiscoveredConcerns(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sec := ExtractSection(benchmarkMarkdownDocument, "Discovered Concerns")
		if len(sec) == 0 {
			b.Fatalf("expected non-empty concerns")
		}
	}
}
