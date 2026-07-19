#!/usr/bin/env bash
# Creates a realistic test directory structure for ordr.
# Usage: ./scripts/test-setup.sh [target-dir]
#
# Default target: /tmp/ordr-test

set -euo pipefail

TARGET="${1:-/tmp/ordr-test}"

echo "Setting up test environment in: $TARGET"

# Clean slate
rm -rf "$TARGET"
mkdir -p "$TARGET"

# ---------------------------------------------------------------------------
# Documents
# ---------------------------------------------------------------------------
touch "$TARGET/Rechnung_2024_01.pdf"
touch "$TARGET/Rechnung_2024_02.pdf"
touch "$TARGET/Angebot_Webdesign.pdf"
touch "$TARGET/Lebenslauf.pdf"
touch "$TARGET/Vertrag_Miete.pdf"

# ---------------------------------------------------------------------------
# Images
# ---------------------------------------------------------------------------
touch "$TARGET/Urlaub_Mallorca.jpg"
touch "$TARGET/Profilbild.png"
touch "$TARGET/Logo_Entwurf.png"
touch "$TARGET/Screenshot 2024-01-15 at 10.23.45.png"
touch "$TARGET/Screenshot 2024-03-08 at 14.55.01.png"
touch "$TARGET/foto_familie.jpeg"

# ---------------------------------------------------------------------------
# Archives
# ---------------------------------------------------------------------------
touch "$TARGET/projekt_backup_2024.zip"
touch "$TARGET/old_photos.tar.gz"
touch "$TARGET/software_v2.zip"

# ---------------------------------------------------------------------------
# Videos
# ---------------------------------------------------------------------------
touch "$TARGET/meeting_recording.mp4"
touch "$TARGET/tutorial_go.mov"

# ---------------------------------------------------------------------------
# Misc / unmatched
# ---------------------------------------------------------------------------
touch "$TARGET/notizen.txt"
touch "$TARGET/todo.md"
touch "$TARGET/setup.sh"

# ---------------------------------------------------------------------------
# Subdirectory (to test recursive)
# ---------------------------------------------------------------------------
mkdir -p "$TARGET/subfolder"
touch "$TARGET/subfolder/nested_doc.pdf"
touch "$TARGET/subfolder/nested_image.png"
touch "$TARGET/subfolder/random.log"

# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------
cat > "$TARGET/.ordrrc" << 'EOF'
version: 1

rules:
  - name: "Rechnungen"
    target: "dokumente/rechnungen"
    match:
      extensions: [".pdf"]
      pattern: "^Rechnung.*"

  - name: "PDFs"
    target: "dokumente"
    match:
      extensions: [".pdf"]

  - name: "Screenshots"
    target: "bilder/screenshots"
    match:
      extensions: [".png"]
      pattern: "^Screenshot.*"

  - name: "Bilder"
    target: "bilder"
    match:
      extensions: [".jpg", ".jpeg", ".png"]

  - name: "Archive"
    target: "archive"
    match:
      extensions: [".zip", ".tar", ".gz", ".tar.gz"]

  - name: "Videos"
    target: "videos"
    match:
      extensions: [".mp4", ".mov", ".mkv"]
EOF

echo ""
echo "Test environment ready."
echo ""
echo "Files created:"
find "$TARGET" -not -name ".ordrrc" -type f | sort | sed "s|$TARGET/||" | sed 's/^/  /'
echo ""
echo "Config: $TARGET/.ordrrc"
echo ""
echo "Next steps:"
echo "  ordr preview $TARGET"
echo "  ordr clean $TARGET --yes"
echo "  ordr undo --yes"
