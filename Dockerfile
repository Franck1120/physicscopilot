# ── PhysicsCopilot — Root Dockerfile (Render / full build) ──────────────────
#
# Build context: repo root (so kb/ is accessible).
# Use this Dockerfile for Render and any deployment that needs the KB bundled.
#
# For local server-only development use server/Dockerfile instead.

# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install certificates for HTTPS calls to Gemini API
RUN apk add --no-cache ca-certificates tzdata

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /server ./cmd/server

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget && \
    adduser -D -u 10001 appuser

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /server /server

# Knowledge base — optional but recommended for KB-enriched Gemini prompts.
# Loaded at runtime via KB_PATH (set below).
COPY kb/data/problems.json /kb/data/problems.json

USER appuser

EXPOSE 8080

ENV KB_PATH=/kb/data/problems.json

HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/server"]
