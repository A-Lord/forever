#!/bin/sh
# Verifies that a skill still describes a tree that exists.
#
#   sh .github/skills/check_paths.sh .github/skills/<skill>    (from anywhere in the repo)
#   make skills-check                                          (every skill)
#
# Checks four things:
#   1. every repo path the skill names in backticks still exists
#   2. every references/*.md is routed to from SKILL.md, and every file SKILL.md
#      routes to exists (a references/ name written with a path in front belongs
#      to another skill and is checked as a path by check 1)
#   3. the frontmatter `name:` still matches the directory name
#   4. every `wowsims-*` skill the markdown names is a directory under .github/skills
#      (wowsims-forever, wowsims-tbc-new and wowsims-sod are repository names and pass)
#
# A path is only checked when it is unambiguous: it must start with a real
# top-level directory or be a known root file, and contain no glob or
# placeholder character (* < … space). One `{a,b}` group is expanded, so
# `dir/{a,b}.ts` checks both files. A trailing `:<line>` is stripped, so a
# `file.ts:123` citation is checked as the file. A path that git ignores is
# skipped because it is absent in a clean checkout by design: that covers
# ui/generated/proto, dist, tmp, binary_dist, node_modules, tools/database/wowsims.db
# and the generated protobuf files, while committed generated files stay checked.
# Write a directory with a trailing slash: git misses a directory-only ignore
# pattern for a path that does not exist unless it ends in a slash.
# Anything the skill deliberately names as *not* existing goes in its
# absent_by_design.txt, one path per line.

set -eu

here=$(CDPATH= cd -- "$1" && pwd)
repo=$(CDPATH= cd -- "$here/../../.." && pwd)
cd "$repo"

ABSENT_BY_DESIGN=$(cat "$here/absent_by_design.txt" 2>/dev/null || true)

# SKILL.md plus references/*.md, when the directory exists.
set -- "$here"/references/*.md
[ -e "$1" ] || set --
docs="$here/SKILL.md $*"

fail=0

expand_braces() {
	case $1 in
	*'{'*'}'*)
		pre=${1%%\{*}
		rest=${1#*\{}
		body=${rest%%\}*}
		post=${rest#*\}}
		echo "$body" | tr ',' '\n' | while read -r alt; do echo "$pre$alt$post"; done
		;;
	*) echo "$1" ;;
	esac
}

# --- 1. paths -------------------------------------------------------------
# shellcheck disable=SC2086
paths=$(cat $docs |
	grep -o '`[^`]*`' |
	tr -d '`' |
	grep -E '^(ui|tools|assets|schemas|proto|sim|cmd|docs)/|^\.github/|^(package\.json|tsconfig\.json|makefile|test-locales\.mjs|vite\.[a-z.-]*mts|vitest\.config\.mts|\.oxlintrc\.json|\.oxfmtrc\.json)$' |
	grep -Ev '[*< ]|…' |
	while read -r p; do expand_braces "$p"; done |
	sed -E 's/:[0-9]+$//' |
	sort -u)

ignored=$(printf '%s\n' "$paths" | git -C "$repo" check-ignore --stdin || true)
count=0
for p in $paths; do
	if printf '%s\n' "$ignored" | grep -qxF -- "$p"; then
		continue
	fi
	if printf '%s\n' "$ABSENT_BY_DESIGN" | grep -qxF -- "$p"; then
		continue
	fi
	count=$((count + 1))
	# Strip a trailing slash so a directory reference checks the directory.
	if [ ! -e "${p%/}" ]; then
		echo "MISSING PATH  $p"
		fail=1
	fi
done

# --- 2. routing -----------------------------------------------------------
for f in "$here"/references/*.md; do
	[ -e "$f" ] || continue
	name="references/$(basename "$f")"
	grep -qE "(^|[^/a-z-])references/$(basename "$f" .md)\.md" "$here/SKILL.md" || {
		echo "UNROUTED      $name is not mentioned in SKILL.md"
		fail=1
	}
done
for name in $(grep -oE '(^|[^/a-z-])references/[a-z-]+\.md' "$here/SKILL.md" | sed -E 's/^[^r]//' | sort -u); do
	[ -f "$here/$name" ] || {
		echo "DANGLING      SKILL.md routes to $name, which does not exist"
		fail=1
	}
done

# --- 3. frontmatter -------------------------------------------------------
declared=$(sed -n '2s/^name: *//p' "$here/SKILL.md")
[ "$declared" = "$(basename "$here")" ] || {
	echo "BAD NAME      frontmatter name '$declared' != directory '$(basename "$here")'"
	fail=1
}

# --- 4. skill names -------------------------------------------------------
# shellcheck disable=SC2086
for token in $(cat $docs | grep -oE 'wowsims-[a-z-]+' | sed -E 's/-+$//' | sort -u); do
	case $token in
	wowsims-forever | wowsims-tbc-new | wowsims-sod) continue ;;
	esac
	[ -d "$repo/.github/skills/$token" ] || {
		echo "UNKNOWN SKILL $token"
		fail=1
	}
done

[ "$fail" = 0 ] && echo "$(basename "$here"): OK ($count paths checked)"
exit "$fail"
