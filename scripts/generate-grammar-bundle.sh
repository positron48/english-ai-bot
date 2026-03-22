#!/bin/bash
# Generate embedded grammar bundles from courses/english-grammar and courses/spanish-grammar
# Usage: ./scripts/generate-grammar-bundle.sh [en|es|all]
# Default: en (backward compatible)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

generate_bundle() {
	local COURSE_DIR="$1"
	local OUT_SUB="$2"
	local BUNDLE_DIR="$PROJECT_ROOT/internal/grammarbundle/${OUT_SUB}"

	if [ ! -d "$COURSE_DIR" ]; then
		echo "Error: Course directory not found at $COURSE_DIR"
		exit 1
	fi

	echo "Generating grammar bundle (${OUT_SUB}) from $COURSE_DIR ..."

	rm -rf "$BUNDLE_DIR/chapters"
	mkdir -p "$BUNDLE_DIR/chapters"

	if [ -f "$COURSE_DIR/config/generation-status.json" ]; then
		cp "$COURSE_DIR/config/generation-status.json" "$BUNDLE_DIR/sections.json"
		echo "✓ Copied sections.json for ${OUT_SUB}"
	else
		echo "Error: generation-status.json not found in $COURSE_DIR/config/"
		exit 1
	fi

	local CHAPTER_COUNT=0
	local INDEX_DATA="{"

	for CHAPTER_DIR in "$COURSE_DIR/chapters"/*/; do
		if [ ! -d "$CHAPTER_DIR" ]; then
			continue
		fi

		local FINAL_FILE=""
		if [ -f "$CHAPTER_DIR/04-final.json" ]; then
			FINAL_FILE="$CHAPTER_DIR/04-final.json"
		elif [ -f "$CHAPTER_DIR/05-final.json" ]; then
			FINAL_FILE="$CHAPTER_DIR/05-final.json"
		fi

		if [ -z "$FINAL_FILE" ]; then
			echo "Warning: No final.json found in $(basename "$CHAPTER_DIR"), skipping"
			continue
		fi

		local CHAPTER_NAME
		CHAPTER_NAME=$(basename "$CHAPTER_DIR")

		local CHAPTER_ID
		CHAPTER_ID=$(grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' "$FINAL_FILE" | head -1 | sed 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' || echo "")

		if [ -z "$CHAPTER_ID" ]; then
			CHAPTER_ID=$(echo "$CHAPTER_NAME" | sed 's/^[0-9]*\.//')
		fi

		cp "$FINAL_FILE" "$BUNDLE_DIR/chapters/${CHAPTER_ID}.json"

		if [ $CHAPTER_COUNT -gt 0 ]; then
			INDEX_DATA+=","
		fi
		INDEX_DATA+="\"${CHAPTER_ID}\":\"${CHAPTER_ID}.json\""

		CHAPTER_COUNT=$((CHAPTER_COUNT + 1))
	done

	INDEX_DATA+="}"

	cat >"$BUNDLE_DIR/index.json" <<EOF
{
  "version": "1.0.0",
  "generated_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "chapters": $INDEX_DATA
}
EOF

	echo "✓ Generated index.json for ${OUT_SUB} with $CHAPTER_COUNT chapters"
	echo "✓ Bundle complete: $BUNDLE_DIR"
}

TARGET="${1:-en}"

case "$TARGET" in
en)
	generate_bundle "$PROJECT_ROOT/courses/english-grammar" "en"
	;;
es)
	generate_bundle "$PROJECT_ROOT/courses/spanish-grammar" "es"
	;;
all)
	generate_bundle "$PROJECT_ROOT/courses/english-grammar" "en"
	generate_bundle "$PROJECT_ROOT/courses/spanish-grammar" "es"
	;;
*)
	echo "Usage: $0 [en|es|all]"
	exit 1
	;;
esac
