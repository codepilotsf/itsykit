#!/bin/sh
# check-corruption.sh — catch duplicated-token corruption before it is committed.
#
# Context: an agent `edit` tool can intermittently duplicate the token it
# changes when the replacement is numeric (e.g. 900px -> 900px900px,
# 0.82 -> 0.828282...). The same class of bug has produced duplicate lines and
# an orphaned CSS brace. Whole-file writes were unaffected.
#
# Usage:
#   scripts/check-corruption.sh --staged    # scan staged files (default)
#   scripts/check-corruption.sh FILE...     # scan the given files
#
# Exit status: 1 if anything suspicious is found, else 0.
set -u
bad=0

if [ "$#" -eq 0 ] || [ "${1:-}" = "--staged" ]; then
  files=$(git diff --cached --name-only --diff-filter=ACM)
else
  files="$*"
fi
[ -z "$files" ] && exit 0

flag() { echo "check-corruption: $1"; bad=1; }

for f in $files; do
  case "$f" in
    *.css|*.html|*.htm|*.json|*.jsonc|*.md|*.txt|*.js|*.mjs|*.cjs|*.ts|*.tsx|*.svg|*.yml|*.yaml|*.toml|*.gitignore|.gitignore) ;;
    *) continue ;;
  esac
  [ -f "$f" ] || continue

  # 1. doubled numeric value + unit (900px900px, 0.5rem0.5rem, 50%50%)
  m1=$(grep -nE '([0-9][0-9.]*(px|rem|em|ex|ch|vw|vh|vmin|vmax|svh|svw|%|fr|deg|ms|s))\1' "$f" 2>/dev/null)
  [ -n "$m1" ] && flag "duplicated value in $f:" && echo "$m1" | head -3

  # 2. repeated digit group (828282, 900900, 525252)
  m2=$(grep -nE '([0-9]{3,6})\1' "$f" 2>/dev/null)
  [ -n "$m2" ] && flag "repeated digits in $f:" && echo "$m2" | head -3

  # 3. adjacent duplicate non-trivial lines (line-duplication corruption)
  m3=$(awk 'NR>1 && $0==prev && length($0)>3 {print NR": "$0} {prev=$0}' "$f" 2>/dev/null)
  [ -n "$m3" ] && flag "duplicate line in $f:" && echo "$m3" | head -3

  # 4. unbalanced braces in CSS (orphaned media query / dropped brace)
  case "$f" in
    *.css)
      o=$(tr -cd '{' < "$f" | wc -c | tr -d ' ')
      c=$(tr -cd '}' < "$f" | wc -c | tr -d ' ')
      [ "$o" != "$c" ] && flag "unbalanced braces in $f ($o '{' vs $c '}')"
      ;;
  esac
done

if [ "$bad" != "0" ]; then
  echo "check-corruption: blocked. Fix the flagged files, or re-check by hand and commit with --no-verify."
  exit 1
fi
exit 0
