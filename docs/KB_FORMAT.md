# Knowledge Base Format

PhysicsCopilot supports a Retrieval-Augmented Generation (RAG) knowledge base stored in `kb/`. The AI uses this to provide device-specific guidance.

---

## Directory Structure

```
kb/
├── devices/
│   ├── prusa-mk4/
│   │   ├── manual.md          # repair procedures
│   │   ├── error-codes.md     # error code → fix mapping
│   │   └── parts.md           # component descriptions
│   └── bambu-lab-x1/
│       └── ...
├── general/
│   ├── soldering.md           # general soldering techniques
│   ├── electronics.md         # component identification
│   └── safety.md              # safety guidelines
└── index.yaml                 # KB metadata and embeddings config
```

---

## Document Format

Each KB document is a Markdown file with YAML frontmatter:

```markdown
---
device: Prusa MK4
category: error-codes
tags: [layer-shift, motors, calibration]
updated: 2026-01-15
source: https://help.prusa3d.com/...
---

# Layer Shift Errors

## Symptom
Layers are offset mid-print, typically after a knock or vibration.

## Causes
1. Loose belt tension (most common)
2. Stepper motor current too low
3. Print speed too high for the current acceleration settings

## Fix
1. Check belt tension: ...
```

---

## Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `device` | Yes | Exact device name (must match `devices` catalog) |
| `category` | Yes | `error-codes`, `maintenance`, `assembly`, `calibration`, `troubleshooting` |
| `tags` | Recommended | Keywords for retrieval |
| `updated` | Yes | ISO 8601 date of last update |
| `source` | Optional | Original documentation URL |

---

## Adding New Knowledge

1. Create a new `.md` file in the appropriate `kb/` subdirectory
2. Add YAML frontmatter
3. Write content in clear, step-by-step format
4. Commit and push — the server rebuilds embeddings on startup
5. Test by asking the AI about the new device/topic

---

## Embedding Pipeline

The RAG service (`server/internal/services/rag_service.go`) embeds KB documents using Gemini's embedding model. Embeddings are stored in-memory at startup (no external vector DB required for small KBs).

For large KBs (> 1000 documents), consider:
- pgvector extension on Supabase
- Pinecone or Weaviate for production scale

---

## Writing Effective KB Content

- **Be specific**: "Turn screw 3 clockwise 90°" > "adjust the screw"
- **Use consistent terminology**: match the device's official manual
- **Include error messages verbatim**: the AI matches against exact strings
- **Keep sections short**: aim for < 200 words per section
- **Add visual context**: describe what a correct/incorrect state looks like
