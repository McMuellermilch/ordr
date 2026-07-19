#!/usr/bin/env bash
# Removes the test directory and clears the ordr undo log.
# Usage: ./scripts/test-teardown.sh [target-dir]

set -euo pipefail

TARGET="${1:-/tmp/ordr-test}"

echo "Cleaning up test environment: $TARGET"

if [ -d "$TARGET" ]; then
  rm -rf "$TARGET"
  echo "  Removed $TARGET"
else
  echo "  $TARGET does not exist, nothing to remove."
fi

UNDO_LOG="$HOME/.ordr/undo.json"
if [ -f "$UNDO_LOG" ]; then
  rm -f "$UNDO_LOG"
  echo "  Cleared undo log ($UNDO_LOG)"
fi

echo "Done."
