#!/usr/bin/env bash
# setup.sh — Bootstrap a full PhysicsCopilot development environment.
#
# Usage:
#   bash scripts/setup.sh          # full setup
#   bash scripts/setup.sh --check  # verify prerequisites only
#
# Requirements: go 1.25+, flutter 3.x, docker, make

set -euo pipefail

RESET='\033[0m'
BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'

info()    { echo -e "${GREEN}[setup]${RESET} $*"; }
warn()    { echo -e "${YELLOW}[warn]${RESET}  $*"; }
error()   { echo -e "${RED}[error]${RESET} $*" >&2; exit 1; }
step()    { echo -e "\n${BOLD}▸ $*${RESET}"; }

# ── Prerequisites ─────────────────────────────────────────────────────────────

check_cmd() {
  command -v "$1" &>/dev/null || error "$1 not found. Install it first: $2"
}

check_prerequisites() {
  step "Checking prerequisites"
  check_cmd go      "https://go.dev/dl"
  check_cmd flutter "https://docs.flutter.dev/get-started/install"
  check_cmd docker  "https://docs.docker.com/get-docker"
  check_cmd make    "apt install make / brew install make"

  GO_VER=$(go version | awk '{print $3}' | sed 's/go//')
  info "Go $GO_VER ✓"

  FLUTTER_VER=$(flutter --version 2>/dev/null | head -1 | awk '{print $2}')
  info "Flutter $FLUTTER_VER ✓"

  info "All prerequisites satisfied."
}

if [[ "${1:-}" == "--check" ]]; then
  check_prerequisites
  exit 0
fi

check_prerequisites

# ── Server setup ──────────────────────────────────────────────────────────────

step "Go server: downloading dependencies"
(cd server && go mod download)
info "go mod download ✓"

step "Go server: installing dev tools"
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
info "golangci-lint, govulncheck, gosec ✓"

# ── Flutter setup ─────────────────────────────────────────────────────────────

step "Flutter app: installing dependencies"
(cd app && flutter pub get)
info "flutter pub get ✓"

# ── Environment file ──────────────────────────────────────────────────────────

step "Environment: checking .env"
if [[ ! -f .env ]]; then
  if [[ -f .env.example ]]; then
    cp .env.example .env
    warn ".env created from .env.example — fill in your secrets before running."
  else
    warn "No .env.example found. Create a .env file manually (see docs/DEVELOPMENT.md)."
  fi
else
  info ".env already exists ✓"
fi

# ── Git hooks ─────────────────────────────────────────────────────────────────

step "Git hooks: installing pre-commit"
if [[ -d .git ]]; then
  cat > .git/hooks/pre-commit <<'HOOK'
#!/usr/bin/env bash
set -e
cd server && go vet ./... && go build ./...
cd ../app && flutter analyze --fatal-warnings
HOOK
  chmod +x .git/hooks/pre-commit
  info "pre-commit hook installed ✓"
fi

# ── Docker ────────────────────────────────────────────────────────────────────

step "Docker: pulling base images"
docker pull golang:1.25-alpine 2>/dev/null && info "golang:1.25-alpine ✓" || warn "Docker pull failed (offline?)"

echo ""
info "Setup complete. Next steps:"
echo "  1. Fill in .env with SUPABASE_URL, SUPABASE_JWT_SECRET, GEMINI_API_KEY"
echo "  2. Run: make run"
echo "  3. In another terminal: cd app && flutter run"
echo ""
echo "See docs/DEVELOPMENT.md for the full guide."
