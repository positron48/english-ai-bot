#!/bin/bash
# Collect per-course training packs into internal/grammartrainingpack/<id>.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

copy_pack() {
	local COURSE_DIR="$1"
	local OUT_SUB="$2"
	local SRC_DIR="$COURSE_DIR/training_pack"
	local DST_DIR="$PROJECT_ROOT/internal/grammartrainingpack/${OUT_SUB}"

	if [ ! -d "$SRC_DIR" ]; then
		echo "Warning: no training_pack in $COURSE_DIR; creating empty index"
		mkdir -p "$DST_DIR"
		cat >"$DST_DIR/index.json" <<EOF
{
  "version": "1.0.0",
  "language": "${OUT_SUB}",
  "course_id": "",
  "generated_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "chapters": {}
}
EOF
		return
	fi

	rm -rf "$DST_DIR"
	mkdir -p "$DST_DIR"
	cp -R "$SRC_DIR"/. "$DST_DIR"/
	echo "✓ copied training pack ${OUT_SUB} from ${COURSE_DIR}"
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

TARGET="${1:-all}"
if [ "$TARGET" = "all" ]; then
	for bt in "$PROJECT_ROOT/courses"/*/bundle.target; do
		[ -f "$bt" ] || continue
		course_dir=$(dirname "$bt")
		out_sub=$(tr -d ' \r\n\t' <"$bt")
		copy_pack "$course_dir" "$out_sub"
	done
else
	if course_dir=$(find_course_dir_for_bundle_id "$TARGET"); then
		copy_pack "$course_dir" "$(echo "$TARGET" | tr '[:upper:]' '[:lower:]')"
	else
		echo "Unknown bundle id: $TARGET"
		exit 1
	fi
fi

