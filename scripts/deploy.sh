#!/usr/bin/env bash
set -euo pipefail

# ── Colors ───────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
RESET='\033[0m'

# ── Defaults ─────────────────────────────────────────────────────────────────
DRY_RUN=false
SERVICE_NAME="physicscopilot-server"
POLL_INTERVAL=10
MAX_WAIT=300  # 5 minutes

# ── Usage ────────────────────────────────────────────────────────────────────
usage() {
  cat <<USAGE
Usage: $(basename "$0") [OPTIONS]

Deploy the PhysicsCopilot server to Render.com via Deploy Hook API.

Options:
  --dry-run              Print what would happen without deploying
  --service=<name>       Service name (default: physicscopilot-server)
  -h, --help             Show this help message

Required environment variables:
  RENDER_API_KEY           Render API key for polling deploy status
  RENDER_DEPLOY_HOOK_URL   Render Deploy Hook URL (POST to trigger deploy)

Examples:
  ./scripts/deploy.sh
  ./scripts/deploy.sh --dry-run
  ./scripts/deploy.sh --service=physicscopilot-server
USAGE
}

# ── Parse arguments ──────────────────────────────────────────────────────────
for arg in "$@"; do
  case "$arg" in
    --dry-run)
      DRY_RUN=true
      ;;
    --service=*)
      SERVICE_NAME="${arg#*=}"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo -e "${RED}Unknown option: ${arg}${RESET}" >&2
      usage
      exit 1
      ;;
  esac
done

# ── Prerequisites ────────────────────────────────────────────────────────────
check_prerequisites() {
  if ! command -v curl &>/dev/null; then
    echo -e "${RED}Error: curl is not installed.${RESET}" >&2
    exit 1
  fi

  if [[ -z "${RENDER_API_KEY:-}" ]]; then
    echo -e "${RED}Error: RENDER_API_KEY environment variable is not set.${RESET}" >&2
    exit 1
  fi

  if [[ -z "${RENDER_DEPLOY_HOOK_URL:-}" ]]; then
    echo -e "${RED}Error: RENDER_DEPLOY_HOOK_URL environment variable is not set.${RESET}" >&2
    exit 1
  fi
}

# ── Trigger deploy ───────────────────────────────────────────────────────────
trigger_deploy() {
  echo -e "${YELLOW}Triggering deploy for service: ${SERVICE_NAME}...${RESET}"

  local response
  response=$(curl -sf -X POST "$RENDER_DEPLOY_HOOK_URL" 2>&1) || {
    echo -e "${RED}Error: Failed to trigger deploy via hook URL.${RESET}" >&2
    exit 1
  }

  echo -e "${GREEN}Deploy triggered successfully.${RESET}"
  echo "$response"
}

# ── Poll deploy status ───────────────────────────────────────────────────────
wait_for_deploy() {
  echo -e "${YELLOW}Waiting for deploy to complete (max ${MAX_WAIT}s)...${RESET}"

  local elapsed=0

  while (( elapsed < MAX_WAIT )); do
    # Fetch the latest deploy for the service using Render API
    local deploys
    deploys=$(curl -sf \
      -H "Authorization: Bearer ${RENDER_API_KEY}" \
      -H "Accept: application/json" \
      "https://api.render.com/v1/services/${SERVICE_NAME}/deploys?limit=1" 2>&1) || {
      echo -e "${YELLOW}Warning: Could not fetch deploy status. Retrying...${RESET}"
      sleep "$POLL_INTERVAL"
      elapsed=$((elapsed + POLL_INTERVAL))
      continue
    }

    local status
    status=$(echo "$deploys" | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4)

    case "$status" in
      live)
        echo -e "${GREEN}Deploy completed successfully!${RESET}"
        return 0
        ;;
      build_failed|update_failed|canceled|deactivated)
        echo -e "${RED}Deploy failed with status: ${status}${RESET}" >&2
        exit 1
        ;;
      *)
        echo -e "${YELLOW}  Status: ${status:-unknown} (${elapsed}s elapsed)${RESET}"
        ;;
    esac

    sleep "$POLL_INTERVAL"
    elapsed=$((elapsed + POLL_INTERVAL))
  done

  echo -e "${RED}Error: Deploy did not complete within ${MAX_WAIT}s.${RESET}" >&2
  exit 1
}

# ── Print service URL ────────────────────────────────────────────────────────
print_service_url() {
  echo ""
  echo -e "${GREEN}Service deployed: https://${SERVICE_NAME}.onrender.com${RESET}"
}

# ── Main ─────────────────────────────────────────────────────────────────────
main() {
  check_prerequisites

  if [[ "$DRY_RUN" == true ]]; then
    echo -e "${YELLOW}[DRY RUN] Would deploy service: ${SERVICE_NAME}${RESET}"
    echo -e "${YELLOW}[DRY RUN] Deploy hook URL: ${RENDER_DEPLOY_HOOK_URL}${RESET}"
    echo -e "${YELLOW}[DRY RUN] Would poll for completion (max ${MAX_WAIT}s)${RESET}"
    echo -e "${YELLOW}[DRY RUN] Service URL: https://${SERVICE_NAME}.onrender.com${RESET}"
    exit 0
  fi

  trigger_deploy
  wait_for_deploy
  print_service_url
}

main
