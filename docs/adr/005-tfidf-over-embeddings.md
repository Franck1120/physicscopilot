# ADR-005: TF-IDF Keyword Search Instead of Vector Embeddings for RAG

**Status:** Accepted
**Date:** 2026-04-12

---

## Context

The RAG (Retrieval-Augmented Generation) layer enriches AI prompts with
relevant equipment records from the knowledge base (`kb/data/problems.json`,
~66 entries at launch). The retrieval step runs on every WebSocket frame that
includes a text query, so it must complete in well under 1 ms to stay off the
critical path.

Retrieval approaches considered:

| Approach | Accuracy | External dependency | Latency (66 docs) | Cost |
|----------|----------|---------------------|-------------------|------|
| **TF-IDF (in-memory)** | Good for keyword queries | None | < 1 ms | Free |
| Dense embeddings (OpenAI) | Excellent | OpenAI API | 50–200 ms | ~$0.01/1k tokens |
| Dense embeddings (local, e.g. sentence-transformers) | Excellent | Python runtime / model file | 10–50 ms | Model size |
| pgvector | Excellent | Postgres extension | 1–5 ms | DB cost |

At 66 documents the semantic gap between TF-IDF and dense embeddings is
small: equipment problems described with specific technical keywords
(e.g. "overheating", "belt slip", "capacitor") map cleanly to TF-IDF terms.
Semantic search would help most when queries are paraphrased or multilingual,
a requirement not yet prioritised for the MVP.

The implementation (`MemoryVectorStore`) computes TF-IDF vectors at startup
and scans the full corpus on each `Search` call. At 66 entries this takes
< 1 ms. The `VectorStore` interface decouples the retrieval algorithm from
the `RAGService` consumer.

---

## Decision

Use **TF-IDF in-memory scoring** (`MemoryVectorStore`) as the default
`VectorStore` implementation. Fields are weighted by importance:
`name + id + category` (3×), `description` (2×), `visual_symptoms` (1×).

The `VectorStore` interface (`server/internal/services/vector_store.go`)
is the migration seam: a future `PgVectorStore` or `SQLiteVectorStore`
implementation can be injected into `RAGService` without changing any caller.

---

## Consequences

### Positive

- **Zero external dependencies at runtime.** No embedding API key, no
  model file, no sidecar process. The server binary is self-contained.
- **Sub-millisecond retrieval.** Full corpus scan at 66 entries completes
  in < 1 ms, adding negligible latency to the WebSocket frame pipeline.
- **No cold-start cost.** TF-IDF vectors are computed once at startup from
  the bundled JSON file; there is no model loading step.
- **Transparent and debuggable.** TF-IDF scores are easy to reason about
  and tune (adjust field weights, add stop words) without retraining.
- **Interface-ready for upgrade.** `VectorStore` is the only contract
  `RAGService` depends on. `SQLiteVectorStore` (sqlite-vec) is the planned
  next implementation when the corpus exceeds ~5 000 entries.

### Negative

- **No semantic understanding.** "The motor is burning" will not match
  "overheating" unless both terms appear in the document. Paraphrased or
  multilingual queries may miss relevant entries.
- **Linear scan does not scale.** O(N) scan is fine at 66 documents but
  becomes a bottleneck at tens of thousands. The `VectorStore` interface
  accommodates an indexed backend before this becomes critical.
- **No learning or personalisation.** TF-IDF weights are fixed at build
  time; user feedback or click-through signals cannot improve retrieval
  without a reindexing step.
- **Field weighting is hand-tuned.** The 3×/2×/1× multipliers were chosen
  by inspection, not by offline evaluation against a labelled query set.
  They may need adjustment as the KB grows.
