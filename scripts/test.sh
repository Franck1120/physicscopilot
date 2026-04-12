#!/usr/bin/env bash
# test.sh — Run the full test suite for PhysicsCopilot.
#
# Usage:
#   bash scripts/test.sh              # all tests
#   bash scripts/test.sh --go         # Go server only
#   bash scripts/test.sh --flutter    # Flutter app only
#   bash scripts/test.sh --coverage   # with HTML coverage reports
#
# Exit code 0 = all tests passed.

set -euo pipefail

RESET='\033[0m'
GREEN='\033[0;32m'
RED='\033[0;31m'
BOLD='\033[1m'

info()  { echo -e "${GREEN}[test]${RESET} $*"; }
fail()  { echo -e "${RED}[FAIL]${RESET} $*" >&2; }
step()  { echo -e "\n${BOLD}▸ $*${RESET}"; }

RUN_GO=true
RUN_FLUTTER=true
COVERAGE=false

for arg in "$@"; do
  case $arg in
    --go)       RUN_FLUTTER=false ;;
    --flutter)  RUN_GO=false ;;
    --coverage) COVERAGE=true ;;
  esac
done

PASS=0
FAIL=0

run_go() {
  step "Go server tests"
  if (cd server && go test -v -race -count=1 -coverprofile=coverage.out ./...); then
    info "Go tests PASSED ✓"
    PASS=$((PASS+1))
    if $COVERAGE; then
      (cd server && go tool cover -html=coverage.out -o coverage.html)
      info "Coverage report: server/coverage.html"
    fi
  else
    fail "Go tests FAILED"
    FAIL=$((FAIL+1))
  fi
}

run_flutter() {
  step "Flutter app tests"
  local flutter_args="--coverage"
  if (cd app && flutter test $flutter_args); then
    info "Flutter tests PASSED ✓"
    PASS=$((PASS+1))
    if $COVERAGE; then
      if command -v lcov &>/dev/null; then
        (cd app && genhtml coverage/lcov.info -o coverage/html --quiet)
        info "Coverage report: app/coverage/html/index.html"
      else
        warn "lcov not installed — skipping HTML coverage for Flutter"
      fi
    fi
  else
    fail "Flutter tests FAILED"
    FAIL=$((FAIL+1))
  fi
}

$RUN_GO      && run_go
$RUN_FLUTTER && run_flutter

echo ""
echo -e "${BOLD}Results: ${GREEN}${PASS} passed${RESET}, ${RED}${FAIL} failed${RESET}"

[[ $FAIL -eq 0 ]] || exit 1
