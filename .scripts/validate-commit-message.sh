#!/usr/bin/env bash
set -euo pipefail

message_file="${1:?commit message file path required}"
subject="$(head -n 1 "$message_file")"

if printf '%s\n' "$subject" | grep -Eq '^(revert: .+|((build|chore|ci|docs|feat|fix|perf|refactor|style|test)(\([[:alnum:]._-]+\))?(!)?: .+))$'; then
  exit 0
fi

cat >&2 <<'EOF'
Commit messages must follow Conventional Commits.

The first line is the important part and must use a valid type like:
  feat: add hook support
  fix(api): validate commit message
  chore!: drop legacy hook config

You may add a longer body after a blank line if you want to explain the change in more detail.
EOF

exit 1
