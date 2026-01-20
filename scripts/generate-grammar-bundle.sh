#!/bin/bash
# Generate grammar course bundle from submodule
# This script reads the grammar course submodule and creates a minimal bundle
# that will be embedded into the Go binary

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BUNDLE_DIR="$PROJECT_ROOT/internal/grammarbundle"
COURSE_DIR="$PROJECT_ROOT/courses/english-grammar"

if [ ! -d "$COURSE_DIR" ]; then
    echo "Error: Course submodule not found at $COURSE_DIR"
    echo "Please run: git submodule update --init --recursive"
    exit 1
fi

echo "Generating grammar bundle..."

# Clean bundle directory (keep .gitkeep if exists)
rm -rf "$BUNDLE_DIR/chapters"
mkdir -p "$BUNDLE_DIR/chapters"

# Copy sections config (normalized)
if [ -f "$COURSE_DIR/config/generation-status.json" ]; then
    cp "$COURSE_DIR/config/generation-status.json" "$BUNDLE_DIR/sections.json"
    echo "✓ Copied sections.json"
else
    echo "Error: generation-status.json not found"
    exit 1
fi

# Scan chapters and copy final JSON files
CHAPTER_COUNT=0
INDEX_DATA="{"

# Find all chapter directories
for CHAPTER_DIR in "$COURSE_DIR/chapters"/*/; do
    if [ ! -d "$CHAPTER_DIR" ]; then
        continue
    fi
    
    CHAPTER_NAME=$(basename "$CHAPTER_DIR")
    
    # Try 04-final.json first, then 05-final.json
    FINAL_FILE=""
    if [ -f "$CHAPTER_DIR/04-final.json" ]; then
        FINAL_FILE="$CHAPTER_DIR/04-final.json"
    elif [ -f "$CHAPTER_DIR/05-final.json" ]; then
        FINAL_FILE="$CHAPTER_DIR/05-final.json"
    fi
    
    if [ -z "$FINAL_FILE" ]; then
        echo "Warning: No final.json found in $CHAPTER_NAME, skipping"
        continue
    fi
    
    # Extract chapter_id from JSON (first try to read it)
    CHAPTER_ID=$(grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' "$FINAL_FILE" | head -1 | sed 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' || echo "")
    
    if [ -z "$CHAPTER_ID" ]; then
        # Fallback: use folder name without prefix numbers
        CHAPTER_ID=$(echo "$CHAPTER_NAME" | sed 's/^[0-9]*\.//')
    fi
    
    # Copy to bundle with chapter_id as filename
    cp "$FINAL_FILE" "$BUNDLE_DIR/chapters/${CHAPTER_ID}.json"
    
    if [ $CHAPTER_COUNT -gt 0 ]; then
        INDEX_DATA+=","
    fi
    INDEX_DATA+="\"${CHAPTER_ID}\":\"${CHAPTER_ID}.json\""
    
    CHAPTER_COUNT=$((CHAPTER_COUNT + 1))
done

INDEX_DATA+="}"

# Generate index.json
cat > "$BUNDLE_DIR/index.json" <<EOF
{
  "version": "1.0.0",
  "generated_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "chapters": $INDEX_DATA
}
EOF

echo "✓ Generated index.json with $CHAPTER_COUNT chapters"
echo "✓ Bundle generation complete: $BUNDLE_DIR"
