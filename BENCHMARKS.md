# Performance Benchmarks

This file records performance work that affects the template parser and
renderer. It includes experiments that were kept and approaches that were
removed or rejected after evaluation.

## Reproduction

The measurements below were collected on 2026-08-12 with Go 1.26.5 on an
Apple M1 (`darwin/arm64`). Each result is the median of eight one-second runs
with one logical processor:

```bash
GOMAXPROCS=1 go test -run '^$' \
  -bench '^(BenchmarkEngine_Parse|BenchmarkTemplate_Render|BenchmarkTemplate_RenderStructProperty|BenchmarkTemplate_RenderIncludes)$' \
  -benchmem -benchtime=1s -count=8 .
```

Absolute timings vary across machines and thermal conditions. Allocation
counts and large relative changes are more stable.

## Overall result

| Benchmark | Time before | Time after | Change | Bytes before | Bytes after | Change | Allocations before | Allocations after | Change |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Parse | 48.47 ms | 41.16 ms | -15% | 62.61 MB | 13.74 MB | -78% | 136,095 | 130,077 | -4% |
| Loop render | 540.87 µs | 441.59 µs | -18% | 315,880 B | 147,856 B | -53% | 12,248 | 11,247 | -8% |
| Struct property render | 462.43 µs | 340.41 µs | -26% | 207,880 B | 95,856 B | -54% | 8,748 | 7,747 | -11% |
| 100 repeated includes | 467.51 µs | 97.05 µs | -79% | 1,331,160 B | 84,272 B | -94% | 4,610 | 1,433 | -69% |

## Incremental trials

The table records adjacent benchmark sweeps, so each row isolates one change
more closely than the overall comparison.

| Trial | Benchmark | Time | Bytes/op | Allocs/op | Decision |
| --- | --- | ---: | ---: | ---: | --- |
| Pool generated yacc parsers | Parse | 48.47 → 36.21 ms | 62.61 → 19.61 MB | 136,095 → 128,100 | Kept: large time and memory reduction |
| Preallocate scanned tokens | Parse | 36.21 → 36.81 ms | 19.61 → 13.71 MB | 128,100 → 128,077 | Kept: neutral time, 30% fewer bytes, three-line change |
| Pass render contexts by pointer | Loop render | 540.87 → 477.73 µs | 315,880 → 163,888 B | 12,248 → 12,249 | Kept: 12% faster and 48% fewer bytes |
| Precompute literal values | Loop render | 477.73 → 430.79 µs | 163,888 → 147,856 B | 12,249 → 11,247 | Kept: 10% faster and 1,002 fewer allocations |
| Compile include arguments once | Repeated includes | 461.66 → 377.95 µs | 272,883 → 232,078 B | 4,511 → 3,811 | Kept: 18% faster |
| Stream partial output | Repeated includes | 377.95 → 343.03 µs | 232,078 → 219,277 B | 3,811 → 3,411 | Kept: 9% faster and avoids a temporary result string |
| Render-scoped compiled partial cache | Repeated includes | 343.03 → 107.12 µs | 219,277 → 100,273 B | 3,411 → 1,633 | Kept: 69% faster while preserving mutable stores |
| Cache struct metadata | Struct property render | 441.93 → 336.65 µs | 207,904 → 95,904 B | 8,748 → 7,748 | Kept: 24% faster and 54% fewer bytes |

Literal values now allocate their runtime wrappers during parsing instead of
during every render. This slightly offsets the parser pool's allocation-count
reduction, but benefits every subsequent render of a compiled template.

## Tried and not kept

- **Process-wide compiled partial cache:** Rejected after the design trial. A
  cache shared across renders would require template-store invalidation,
  concurrency control, and configuration revision tracking. The render-scoped
  cache obtains the large repeated-partial win, rereads the store, compares the
  source bytes, and cannot become stale across top-level renders.
- **Eager render-scoped cache allocation:** Implemented initially, then removed.
  It charged templates that never render partials and allocated a discarded map
  for every child context. The final implementation initializes the map only on
  the first partial compilation.
- **Caching missing struct properties:** Implemented initially, then removed.
  It could retain an unbounded set of user-controlled property names in a
  process-wide map. The final cache stores only successful lookups, whose count
  is bounded by the fields and methods of encountered types.
- **Handwritten default-delimiter scanner:** Rejected before implementation.
  After parser pooling, regular-expression matching was no longer the dominant
  allocator. Reimplementing delimiter, trimming, and malformed-token behavior
  would add substantial compatibility risk for a smaller remaining CPU target.
