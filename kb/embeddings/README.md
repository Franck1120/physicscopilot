# Embedding Generation

Scripts to generate vector embeddings from processed KB data and upload to Supabase pgvector.

## Pipeline
1. `scraper/` → raw HTML/PDF → `data/raw/`
2. Chunking + cleaning → `data/processed/`
3. Embedding (Gemini text-embedding-004) → `embeddings/*.jsonl`
4. Upload to Supabase `kb_chunks` table
