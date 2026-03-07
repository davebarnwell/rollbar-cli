#!/usr/bin/env bash
set -euo pipefail

prefix="${1:-v}"
date_part="$(date -u +%y.%m.%d)"
base="${prefix}${date_part}"

next_seq=1

while IFS= read -r tag; do
  if [[ "$tag" =~ ^${base//./\\.}\.([0-9]+)$ ]]; then
    candidate=$((BASH_REMATCH[1] + 1))
    if (( candidate > next_seq )); then
      next_seq=$candidate
    fi
  fi
done < <(git tag -l "${base}.*")

printf '%s.%d\n' "$base" "$next_seq"
