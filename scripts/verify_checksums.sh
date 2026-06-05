#!/usr/bin/env bash
#
# verify_checksums.sh — verify integrity of downloaded benchmark datasets.
#
# Checks size and SHA-1 for each known dataset. MD5 is checked when a tool is
# available. With no arguments or with 'all', only files that exist are checked.
# With explicit dataset names, missing files are treated as errors.
#
# Usage:
#   scripts/verify_checksums.sh                 # verify whatever exists
#   scripts/verify_checksums.sh enwik8          # verify only enwik8
#   scripts/verify_checksums.sh enwik8 enwik9   # verify the named datasets
#   scripts/verify_checksums.sh all             # verify whatever exists
#   scripts/verify_checksums.sh /path/to/enwik8 # verify a file at an explicit path
#   scripts/verify_checksums.sh -h              # show this help
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
DOWNLOAD_DIR="${REPO_ROOT}/testdata/downloads"

usage() {
  cat <<USAGE
Usage: $(basename "$0") [enwik8|enwik9|all|/path/to/file ...]

With no arguments, verifies known datasets that currently exist in testdata/downloads/.
With 'all', same behaviour: verifies whatever exists and does not fail on absent files.
With dataset names, verifies only those requested files and fails if absent.
With absolute paths, verifies the file at that path.
Checks file size and SHA-1. Also checks MD5 when md5sum or md5 is available.
USAGE
}

expected_for() {
  case "$1" in
    enwik8) echo "100000000 57b8363b814821dc9d47aa4d41f58733519076b2 a1fa5ffddb56f4953e226637dabbb36a" ;;
    enwik9) echo "1000000000 2996e86fb978f93cca8f566cc56998923e7fe581 e206c3450ac99950df65bf70ef61a12d" ;;
    *) return 1 ;;
  esac
}

ALL_DATASETS=(enwik8 enwik9)

file_size() {
  local f="$1"
  if stat -c%s "$f" >/dev/null 2>&1; then
    stat -c%s "$f"
  else
    stat -f%z "$f"
  fi
}

sha1_of() {
  local f="$1"
  if command -v sha1sum >/dev/null 2>&1; then
    sha1sum "$f" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 1 "$f" | awk '{print $1}'
  else
    echo "ERROR: no sha1sum/shasum tool found" >&2
    return 2
  fi
}

md5_of() {
  local f="$1"
  if command -v md5sum >/dev/null 2>&1; then
    md5sum "$f" | awk '{print $1}'
  elif command -v md5 >/dev/null 2>&1; then
    md5 -q "$f"
  else
    return 1
  fi
}

verify_file() {
  local name="$1" file="$2" absent_policy="${3:-error}"

  if [[ ! -f "$file" ]]; then
    if [[ "$absent_policy" == "skip_if_absent" ]]; then
      echo "SKIP : ${name} (not present)"
      return 0
    fi

    echo "ERROR: requested dataset '${name}' not found at ${file}" >&2
    return 1
  fi

  local exp exp_size exp_sha1 exp_md5
  if ! exp="$(expected_for "$name")"; then
    echo "ERROR: unknown dataset '${name}'" >&2
    return 1
  fi

  exp_size="$(echo "$exp" | awk '{print $1}')"
  exp_sha1="$(echo "$exp" | awk '{print $2}')"
  exp_md5="$(echo "$exp" | awk '{print $3}')"

  echo "CHECK: ${name} (${file})"

  local ok=1
  local actual_size actual_sha1 actual_md5

  actual_size="$(file_size "$file")"
  if [[ "$actual_size" == "$exp_size" ]]; then
    echo "  size OK   (${actual_size} bytes)"
  else
    echo "  size FAIL (expected ${exp_size}, got ${actual_size})"
    ok=0
  fi

  actual_sha1="$(sha1_of "$file")"
  if [[ "$actual_sha1" == "$exp_sha1" ]]; then
    echo "  sha1 OK   (${actual_sha1})"
  else
    echo "  sha1 FAIL (expected ${exp_sha1}, got ${actual_sha1})"
    ok=0
  fi

  if actual_md5="$(md5_of "$file")"; then
    if [[ "$actual_md5" == "$exp_md5" ]]; then
      echo "  md5  OK   (${actual_md5})"
    else
      echo "  md5  FAIL (expected ${exp_md5}, got ${actual_md5})"
      ok=0
    fi
  else
    echo "  md5  SKIP (no md5 tool available)"
  fi

  [[ "$ok" -eq 1 ]]
}

checked=0
failures=0

run_verify() {
  local name="$1" file="$2" absent_policy="$3"
  checked=$((checked + 1))

  if ! verify_file "$name" "$file" "$absent_policy"; then
    failures=$((failures + 1))
  fi
}

verify_existing_known_files() {
  local found=0
  local dataset file

  for dataset in "${ALL_DATASETS[@]}"; do
    file="${DOWNLOAD_DIR}/${dataset}"
    if [[ -f "$file" ]]; then
      found=1
      run_verify "$dataset" "$file" "error"
    fi
  done

  if [[ "$found" -eq 0 ]]; then
    echo "No known uncompressed datasets found in ${DOWNLOAD_DIR}; nothing to verify."
  fi
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" || "${1:-}" == "help" ]]; then
  usage
  exit 0
fi

if [[ "$#" -eq 0 ]]; then
  verify_existing_known_files
else
  for arg in "$@"; do
    case "$arg" in
      all)
        verify_existing_known_files
        ;;
      enwik8|enwik9)
        run_verify "$arg" "${DOWNLOAD_DIR}/${arg}" "error"
        ;;
      /*)
        base="$(basename "$arg")"
        case "$base" in
          enwik8|enwik9)
            run_verify "$base" "$arg" "error"
            ;;
          enwik8.zip|enwik9.zip)
            echo "ERROR: checksums apply to uncompressed ${base%.zip}, not ${base}" >&2
            failures=$((failures + 1))
            ;;
          *)
            echo "ERROR: unsupported file: ${arg}" >&2
            failures=$((failures + 1))
            ;;
        esac
        ;;
      *)
        echo "ERROR: unknown argument '${arg}'" >&2
        usage >&2
        failures=$((failures + 1))
        ;;
    esac
  done
fi

echo "----------------------------------------"
echo "Checked: ${checked} Failures: ${failures}"

if [[ "$failures" -gt 0 ]]; then
  echo "Checksum verification FAILED." >&2
  exit 1
fi

if [[ "$checked" -eq 0 ]]; then
  echo "Nothing to verify (no target files present)."
fi

echo "Checksum verification passed."
