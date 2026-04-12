# Knowledge Base

The knowledge base (`kb/data/problems.json`) is **curated manually**.
Each entry describes a known failure mode, its symptoms, and the recommended
repair steps. The RAGService performs keyword-based retrieval over this file
to inject relevant context into Gemini prompts.

## File format

```json
[
  {
    "id": "unique-slug",
    "title": "Short problem title",
    "keywords": ["keyword1", "keyword2"],
    "problem": "Description of the failure mode.",
    "solution": "Step-by-step repair guidance.",
    "vertical": "3d-printer"
  }
]
```

## Adding entries

Edit `kb/data/problems.json` directly. Keep entries focused: one problem per
entry, concrete keywords, actionable solution steps.

## Scrapers (archived)

The `scraper/` directory previously contained Python scripts for automated
collection of 3D printer repair documentation (Prusa KB, Bambu Lab docs,
Creality wiki, Reddit r/3Dprinting). These were replaced by manual curation
because automated scraping produced noisy, low-quality entries that degraded
RAG retrieval quality.

If you want to revive scraping, start from the original sources listed above
and apply strict filtering before adding entries to `problems.json`.
