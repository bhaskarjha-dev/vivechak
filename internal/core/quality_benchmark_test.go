package core

import "testing"

var benchmarkSessionContent = []byte(`---
session_id: T1-01
title: Analytical Datastore Selection
date: 2026-10-13
status: complete
---

# T1-01: Analytical Datastore Selection

## Key Findings
- [Grade A] https://clickhouse.com/docs/en/intro ClickHouse columnar storage provides 12x zstd compression.
- [Grade B] https://github.com/ClickHouse/ClickHouse/benchmarks Ingest benchmark demonstrated 75k events/sec at 180MB RAM.
- [Grade A] RFC 9110 compliant HTTP query endpoint for analytics queries.

## Discovered Concerns
Memory footprint spikes if bulk insert flush buffer is configured greater than 50MB per tenant.

## Delta
Contradicted initial assumption that distributed Kafka ingestion cluster was strictly required for buffering.

## Recommendation
Deploy ClickHouse single-binary daemonset on Kubernetes worker nodes.
`)

func BenchmarkObserveSessionQuality(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		obs := ObserveSessionQuality(benchmarkSessionContent)
		_ = obs
	}
}
