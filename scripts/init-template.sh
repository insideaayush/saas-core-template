#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <new-project-name>"
  echo "Example: $0 acme-core"
  exit 1
fi

NEW_NAME="$1"
SOURCE_NAME="saas-core-template"
SOURCE_DB_NAME="saas_core_template"
TARGET_DB_NAME="${NEW_NAME//-/_}"

if [[ -z "$NEW_NAME" ]]; then
  echo "Project name cannot be empty."
  exit 1
fi

if [[ "$NEW_NAME" == "$SOURCE_NAME" ]]; then
  echo "Project name is already $SOURCE_NAME. Nothing to do."
  exit 0
fi

if ! command -v git >/dev/null 2>&1; then
  echo "git is required."
  exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required."
  exit 1
fi

echo "Initializing template..."
echo "Replacing '$SOURCE_NAME' -> '$NEW_NAME'"
echo "Replacing '$SOURCE_DB_NAME' -> '$TARGET_DB_NAME'"

python3 - "$SOURCE_NAME" "$NEW_NAME" "$SOURCE_DB_NAME" "$TARGET_DB_NAME" <<'PY'
import pathlib
import subprocess
import sys

source_name, target_name, source_db_name, target_db_name = sys.argv[1:5]
workspace = pathlib.Path(".").resolve()

binary_extensions = {
    ".png", ".jpg", ".jpeg", ".gif", ".webp", ".ico", ".pdf", ".zip",
    ".gz", ".tar", ".woff", ".woff2", ".eot", ".ttf", ".otf"
}

paths = subprocess.check_output(["git", "ls-files"], text=True).splitlines()
changed = 0

for relative_path in paths:
    path = workspace / relative_path
    if not path.is_file():
        continue
    if path.suffix.lower() in binary_extensions:
        continue

    try:
        content = path.read_text(encoding="utf-8")
    except UnicodeDecodeError:
        continue

    updated = content.replace(source_name, target_name).replace(source_db_name, target_db_name)
    if updated != content:
        path.write_text(updated, encoding="utf-8")
        changed += 1

print(f"Updated {changed} tracked files.")
PY

echo "Initialization complete."
echo
echo "Next steps:"
echo "  1) Review changes: git diff --stat && git diff"
echo "  2) Update any provider/project-specific secrets in .env files"
echo "  3) Reinstall frontend deps if package name changed: cd frontend && npm install"
