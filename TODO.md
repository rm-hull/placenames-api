# Improvements and suggestions for placenames-api (main.go & internal/trie.go)

Below are prioritized, actionable suggestions grouped by area: correctness & robustness, performance & scalability, API & ergonomics, testing & benchmarks, observability & operational, security & robustness, small dev-experience improvements, and higher-level refactors. Use this text as an issue description or checklist for implementation.

## Summary

Reading the code, the implementation is concise and functional. The main areas to consider for improvement are: memory usage (Places stored at every node), robustness around CSV handling and startup, observability/operational readiness, clearer API ergonomics and request validation, and tests/benchmarks to quantify trade-offs.

## Correctness & robustness

- Handle missing file gracefully at startup: consider non-fatal failure modes (empty-data mode or retry/backoff) instead of immediate `log.Fatalf`.
- Configurable strictness for CSV parsing: add strict vs permissive modes. In permissive mode, log and skip malformed rows rather than failing startup.
- ~~CSV line-number accuracy: your `line` counter increments after reading; ensure error messages clearly document which numbering scheme is used (header = line 1).~~
- Input normalization: apply Unicode normalization (NFC) and trimming when storing and when querying to avoid mismatches on composed/decomposed characters.
- ~~Concurrency safety: the `Trie` is not synchronized. If you ever mutate it after startup or build it concurrently, protect it with a RWMutex or avoid post-start writes.~~ **WONT DO: unnecessary**

## Performance & scalability

- ~~Reduce duplicated Place copies:~~
  - ~~Currently each node stores full `Place` values along the path, which duplicates memory per character. Consider:~~
    - ~~Storing pointers (`*Place`) instead of values, or~~
    - ~~Storing indices/IDs into a global slice/map of `Place` objects.~~
- Top-K bounding:
  - If clients request only small result sets, maintain only the top-K items per node (by relevancy) during insertion (use a min-heap/bounded slice). This bounds memory and avoids sorting large slices.
- Sorting costs:
  - Sorting all node slices after bulk insert is fine, but sorting many large slices can be expensive. Consider incremental top-K maintenance to avoid large sorts.
- Compression for trie:
  - Consider a radix/compressed trie to reduce node count and pointer overhead for long common prefixes.
- Unicode and rune handling:
  - Normalize both stored names and queries. Verify rune iteration is correct for your dataset (surrogate handling, combining marks).
- CSV robustness:
  - Use `csvReader.FieldsPerRecord = -1` if records have variable numbers of fields; otherwise validate and fail or skip based on configured strictness.

## API & ergonomics

- Endpoint query parameter vs path param:
  - Consider using query params (e.g., `GET /v1/place-names?prefix=...&max_results=...`) instead of a path param to ease handling of spaces and special characters.
- ~~Cap `max_results` and validate:~~
  - ~~Add an upper bound (e.g., 1000) to avoid expensive large responses.~~
- Response shape:
  - Use a response struct and include metadata: `{results: [], count: N, query: "...", max_results: M}` for clarity and future pagination.
- Consistent error format:
  - Return structured error objects (code + message) so clients can parse errors programmatically.
- Add optional filters/capabilities:
  - Plan for future features like fuzzy matching, region/type filters, or alternate ranking options.

## Testing & benchmarks

- Unit tests:
  - Tests for `Trie.Insert`, `SortAllNodes`, `FindByPrefix` including unicode, ties in relevancy, case insensitivity, duplicates and edge cases (empty prefix).
- ~~Integration tests:~~
  - ~~Small gzipped CSV fixtures to test `loadData` in strict/permissive modes.~~
- Benchmarks:
  - Bench `Insert + SortAllNodes` and `FindByPrefix` at realistic sizes (62k and larger). Measure allocations and peak memory.
- Regression tests:
  - Add tests that verify ordering tie-break rules: relevancy desc, then shorter name wins.

## Observability & operational

- ~~Add structured logging:~~
  - ~~Log counts after load (total rows processed, skipped, errors).~~
- ~~Metrics:~~
  - ~~Expose Prometheus metrics: records loaded, invalid records, request count, latency (histogram), and result count distribution.~~
- ~~Health/readiness:~~
  - ~~Add `/health` and `/ready` endpoints; readiness should check the trie is loaded.~~
- Request instrumentation:
  - Middleware for latency, logging, and request IDs.

## Security & robustness

- Input validation / DoS protections:
  - Rate-limit endpoints or add per-IP quotas.
  - ~~Enforce upper `max_results` cap.~~
- Error handling:
  - Avoid returning internal errors to clients. Log details internally, return friendly messages externally.

## Small code-style & maintainability suggestions

- ~~Move server setup/handlers to `server.go` (or `http/server.go`) and keep `main.go` minimal.~~
- Use dependency injection for handlers (struct with trie, logger, metrics) to make testing easier.
- Extract constants for defaults (port, default max results, max allowed results).
- Add comments documenting complexity and tie-break rules for `SortAllNodes`.
- Add `go:generate` or scripts to produce sample gzipped CSVs for tests.

## Potential higher-level refactors (for larger scale)

- Compact trie (radix or DAWG) to reduce memory/perf overhead.
- Maintain top-K per node during insertion (min-heap) rather than storing all matches.
- Switch to a search-index (Bleve or similar) if you need fuzzy search, tokenization, full-text ranking; tradeoff: increased dependencies/operational complexity.
- Serialize a pre-built index to disk (binary snapshot) to avoid reparsing CSV on every start.

## Prioritized checklist (small wins first)

1. Add unit tests for the Trie methods (Insert, SortAllNodes, FindByPrefix).
2. ~~Add logging in `loadData` with counts of processed/skipped rows and overall duration.~~
3. ~~Cap `max_results` and return clear, structured validation errors.~~
4. ~~Replace per-node `Place` values with pointers or indices if memory profiling shows high usage.~~
5. Add a benchmark for `FindByPrefix` and insertion+sort to measure performance at scale.

## Suggested next steps

- Draft unit tests and a small gzipped CSV fixture and run them.
- Implement top-K maintenance per node (bounded memory) and associated tests/benchmarks.
- ~~Refactor nodes to store `*Place` pointers or indices plus a small conversion utility to keep public API the same.~~
