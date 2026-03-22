#!/bin/bash
# Generate embedded grammar bundles from courses/* (each course with bundle.target + config/generation-status.json).
# Usage:
#   ./scripts/generate-grammar-bundle.sh           # all discovered courses
#   ./scripts/generate-grammar-bundle.sh all       # same
#   ./scripts/generate-grammar-bundle.sh list      # print course dir -> embed id
#   ./scripts/generate-grammar-bundle.sh en|es|…  # single bundle id (matches bundle.target)

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

find_course_dir_for_bundle_id() {
	local want
	want=$(echo "$1" | tr '[:upper:]' '[:lower:]' | tr -d ' \r\n\t')
	for bt in "$PROJECT_ROOT/courses"/*/bundle.target; do
		[ -f "$bt" ] || continue
		id=$(tr '[:upper:]' '[:lower:]' <"$bt" | tr -d ' \r\n\t')
		if [ "$id" = "$want" ]; then
			dirname "$bt"
			return 0
		fi
	done
	return 1
}

list_bundles() {
	echo "Courses -> internal/grammarbundle/<id> (requires bundle.target + config/generation-status.json):"
	for bt in "$PROJECT_ROOT/courses"/*/bundle.target; do
		[ -f "$bt" ] || continue
		cdir=$(dirname "$bt")
		id=$(tr -d ' \r\n\t' <"$bt")
		gs="$cdir/config/generation-status.json"
		if [ -f "$gs" ]; then
			echo "  $id  <=  $cdir"
		else
			echo "  (skip) $cdir — missing config/generation-status.json"
		fi
	done
}

generate_all_discovered() {
	local found=0
	for bt in "$PROJECT_ROOT/courses"/*/bundle.target; do
		[ -f "$bt" ] || continue
		course_dir=$(dirname "$bt")
		out_sub=$(tr -d ' \r\n\t' <"$bt")
		if [ -z "$out_sub" ]; then
			echo "Warning: empty bundle.target in $course_dir"
			continue
		fi
		if [ ! -f "$course_dir/config/generation-status.json" ]; then
			echo "Warning: skip $course_dir (no config/generation-status.json)"
			continue
		fi
		generate_bundle "$course_dir" "$out_sub"
		found=$((found + 1))
	done
	if [ "$found" -eq 0 ]; then
		echo "Error: no courses found. Add courses/<name>/bundle.target (embed id, e.g. en) and config/generation-status.json"
		exit 1
	fi
}

TARGET="${1:-all}"

case "$TARGET" in
all | "")
	generate_all_discovered
	;;
list)
	list_bundles
	;;
*)
	if course_dir=$(find_course_dir_for_bundle_id "$TARGET"); then
		generate_bundle "$course_dir" "$(echo "$TARGET" | tr '[:upper:]' '[:lower:]')"
	else
		echo "Unknown bundle id '$TARGET'. Known ids from bundle.target files:"
		list_bundles
		exit 1
	fi
	;;
esac
