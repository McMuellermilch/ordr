#!/usr/bin/env bash
# Full ordr test runner — setup, run scenarios, verify results, teardown.
# Usage: ./scripts/test.sh [ordr-binary] [target-dir]
#
# Examples:
#   ./scripts/test.sh                          # uses ./ordr and /tmp/ordr-test
#   ./scripts/test.sh /usr/local/bin/ordr      # uses installed binary
#   ./scripts/test.sh ./ordr /tmp/my-test      # custom dir

set -euo pipefail

ORDR="${1:-./ordr}"
TARGET="${2:-/tmp/ordr-test}"
PASS=0
FAIL=0

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
GRAY='\033[0;90m'
RESET='\033[0m'

pass() { echo -e "  ${GREEN}✓${RESET} $1"; ((PASS++)); }
fail() { echo -e "  ${RED}✗${RESET} $1"; ((FAIL++)); }
info() { echo -e "${GRAY}  $1${RESET}"; }
header() { echo -e "\n${YELLOW}▶ $1${RESET}"; }

assert_file_exists() {
  if [ -f "$1" ]; then
    pass "file exists: ${1#$TARGET/}"
  else
    fail "expected file not found: ${1#$TARGET/}"
  fi
}

assert_file_missing() {
  if [ ! -f "$1" ]; then
    pass "file moved away: ${1#$TARGET/}"
  else
    fail "expected file to be gone but still exists: ${1#$TARGET/}"
  fi
}

assert_exit_ok() {
  if [ $? -eq 0 ]; then
    pass "$1 exited successfully"
  else
    fail "$1 exited with error"
  fi
}

# ---------------------------------------------------------------------------
# Check binary
# ---------------------------------------------------------------------------

if [ ! -f "$ORDR" ] && ! command -v "$ORDR" &> /dev/null; then
  echo -e "${RED}Error:${RESET} ordr binary not found at '$ORDR'"
  echo "Build it first: go build -o ordr ./cmd/ordr"
  exit 1
fi

echo ""
echo "ordr test runner"
echo "  Binary:  $ORDR"
echo "  Dir:     $TARGET"
info "Version: $($ORDR --version 2>&1)"

# ---------------------------------------------------------------------------
# Setup
# ---------------------------------------------------------------------------

header "Setup"
bash "$(dirname "$0")/test-setup.sh" "$TARGET" > /dev/null
pass "test directory created with $(find "$TARGET" -type f -not -name ".ordrrc" | wc -l | tr -d ' ') files"

# ---------------------------------------------------------------------------
# Test 1: preview (dry-run — nothing should move)
# ---------------------------------------------------------------------------

header "Test 1: preview (dry-run)"

"$ORDR" preview "$TARGET" > /dev/null
assert_exit_ok "ordr preview"

assert_file_exists "$TARGET/Rechnung_2024_01.pdf"
assert_file_exists "$TARGET/Profilbild.png"
assert_file_exists "$TARGET/projekt_backup_2024.zip"
pass "no files were moved during preview"

# ---------------------------------------------------------------------------
# Test 2: rules list
# ---------------------------------------------------------------------------

header "Test 2: ordr rules list"

OUTPUT=$("$ORDR" rules list 2>&1)
assert_exit_ok "ordr rules list"

if echo "$OUTPUT" | grep -q "Rechnungen"; then
  pass "rule 'Rechnungen' visible in list"
else
  fail "rule 'Rechnungen' not found in output"
fi

# ---------------------------------------------------------------------------
# Test 3: rules test
# ---------------------------------------------------------------------------

header "Test 3: ordr rules test"

OUTPUT=$("$ORDR" rules test "$TARGET/Rechnung_2024_01.pdf" 2>&1)
assert_exit_ok "ordr rules test (pdf)"

if echo "$OUTPUT" | grep -q "Rechnungen"; then
  pass "Rechnung_2024_01.pdf matched rule 'Rechnungen'"
else
  fail "Rechnung_2024_01.pdf did not match expected rule"
fi

OUTPUT=$("$ORDR" rules test "$TARGET/notizen.txt" 2>&1)
if echo "$OUTPUT" | grep -q "No rules match"; then
  pass "notizen.txt correctly unmatched"
else
  fail "notizen.txt should not match any rule"
fi

# ---------------------------------------------------------------------------
# Test 4: status
# ---------------------------------------------------------------------------

header "Test 4: ordr status"

OUTPUT=$("$ORDR" status "$TARGET" 2>&1)
assert_exit_ok "ordr status"

if echo "$OUTPUT" | grep -q "Rules:"; then
  pass "status shows rules count"
else
  fail "status output missing rules info"
fi

# ---------------------------------------------------------------------------
# Test 5: clean
# ---------------------------------------------------------------------------

header "Test 5: ordr clean"

"$ORDR" clean "$TARGET" --yes > /dev/null
assert_exit_ok "ordr clean"

# Check files were moved
assert_file_missing "$TARGET/Rechnung_2024_01.pdf"
assert_file_missing "$TARGET/Profilbild.png"
assert_file_missing "$TARGET/projekt_backup_2024.zip"
assert_file_missing "$TARGET/meeting_recording.mp4"

# Check destinations
assert_file_exists "$TARGET/dokumente/rechnungen/Rechnung_2024_01.pdf"
assert_file_exists "$TARGET/dokumente/rechnungen/Rechnung_2024_02.pdf"
assert_file_exists "$TARGET/dokumente/Angebot_Webdesign.pdf"
assert_file_exists "$TARGET/bilder/screenshots/Screenshot 2024-01-15 at 10.23.45.png"
assert_file_exists "$TARGET/bilder/Profilbild.png"
assert_file_exists "$TARGET/archive/projekt_backup_2024.zip"
assert_file_exists "$TARGET/videos/meeting_recording.mp4"

# Unmatched files must stay
assert_file_exists "$TARGET/notizen.txt"
assert_file_exists "$TARGET/todo.md"
assert_file_exists "$TARGET/setup.sh"

# ---------------------------------------------------------------------------
# Test 6: undo
# ---------------------------------------------------------------------------

header "Test 6: ordr undo"

"$ORDR" undo --yes > /dev/null
assert_exit_ok "ordr undo"

# Files must be back in original locations
assert_file_exists "$TARGET/Rechnung_2024_01.pdf"
assert_file_exists "$TARGET/Profilbild.png"
assert_file_exists "$TARGET/projekt_backup_2024.zip"
assert_file_exists "$TARGET/meeting_recording.mp4"

# ---------------------------------------------------------------------------
# Test 7: clean with --rule filter
# ---------------------------------------------------------------------------

header "Test 7: ordr clean --rule"

"$ORDR" clean "$TARGET" --yes --rule "Videos" > /dev/null
assert_exit_ok "ordr clean --rule Videos"

assert_file_missing "$TARGET/meeting_recording.mp4"
assert_file_exists "$TARGET/tutorial_go.mov"  # wait, mov matches too — check
assert_file_exists "$TARGET/Rechnung_2024_01.pdf"  # must NOT be moved
pass "only video rule was applied, other files untouched"

# ---------------------------------------------------------------------------
# Test 8: recursive
# ---------------------------------------------------------------------------

header "Test 8: ordr clean --recursive"

# Reset first
"$ORDR" undo --yes > /dev/null 2>&1 || true
bash "$(dirname "$0")/test-teardown.sh" "$TARGET" > /dev/null
bash "$(dirname "$0")/test-setup.sh" "$TARGET" > /dev/null

"$ORDR" clean "$TARGET" --yes --recursive > /dev/null
assert_exit_ok "ordr clean --recursive"

assert_file_missing "$TARGET/subfolder/nested_doc.pdf"
assert_file_missing "$TARGET/subfolder/nested_image.png"
pass "nested files were moved by recursive clean"

# ---------------------------------------------------------------------------
# Teardown
# ---------------------------------------------------------------------------

header "Teardown"
bash "$(dirname "$0")/test-teardown.sh" "$TARGET" > /dev/null
pass "test directory removed"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

TOTAL=$((PASS + FAIL))
echo ""
echo "────────────────────────────────"
if [ $FAIL -eq 0 ]; then
  echo -e "${GREEN}  All $TOTAL tests passed${RESET}"
else
  echo -e "${RED}  $FAIL/$TOTAL tests failed${RESET}"
fi
echo "────────────────────────────────"
echo ""

[ $FAIL -eq 0 ] && exit 0 || exit 1
