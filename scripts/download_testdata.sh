#!/usr/bin/env bash
#
# download_testdata.sh — download benchmark datasets into testdata/downloads/.
#
# These datasets (enwik8 / enwik9) are large Wikipedia-derived corpora and are
# NEVER committed to the repository (see .gitignore and docs/data_policy.md).
# They are fetched on demand only.
#
# By default, only enwik8 is downloaded.
#
# Usage:
#   scripts/download_testdata.sh          # downloads enwik8 (default)
#   scripts/download_testdata.sh enwik8   # downloads enwik8
#   scripts/download_testdata.sh enwik9   # downloads enwik9
#   scripts/download_testdata.sh all      # downloads enwik8 and enwik9
#   scripts/download_testdata.sh -h       # show this help
#
# After downloading, this script verifies integrity via
# scripts/verify_checksums.sh.
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
DOWNLOAD_DIR="${REPO_ROOT}/testdata/downloads"
VERIFY_SCRIPT="${SCRIPT_DIR}/verify_checksums.sh"

usage() {
  cat <<USAGE
Usage: $(basename "$0") [enwik8|enwik9|all]

Downloads benchmark datasets into testdata/downloads/ and verifies checksums.
Default: enwik8

Downloaded archives and extracted datasets must not be committed.
USAGE
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" || "${1:-}" == "help" ]]; then
  usage
  exit 0
fi

url_for() {
  case "$1" in
    enwik8) echo "https://mattmahoney.net/dc/enwik8.zip" ;;
    enwik9) echo "https://mattmahoney.net/dc/enwik9.zip" ;;
    *) return 1 ;;
  esac
}

arg="${1:-enwik8}"
case "$arg" in
  enwik8) TARGETS=(enwik8) ;;
  enwik9) TARGETS=(enwik9) ;;
  all) TARGETS=(enwik8 enwik9) ;;
  *)
    echo "ERROR: unknown argument '${arg}'. Use: enwik8 | enwik9 | all" >&2
    usage >&2
    exit 2
    ;;
esac

mkdir -p "${DOWNLOAD_DIR}"

fetch() {
  local url="$1" out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fSL --retry 3 -o "${out}.tmp" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -O "${out}.tmp" "$url"
  else
    echo "ERROR: neither curl nor wget is available" >&2
    return 3
  fi
  mv "${out}.tmp" "${out}"
}

extract() {
  local zip="$1" member="$2" dest="$3"
  if command -v unzip >/dev/null 2>&1; then
    unzip -p "$zip" "$member" > "${dest}.tmp"
    mv "${dest}.tmp" "${dest}"
  elif command -v bsdtar >/dev/null 2>&1; then
    bsdtar -x -f "$zip" -C "$(dirname "$dest")" "$member"
  else
    echo "ERROR: neither unzip nor bsdtar is available" >&2
    return 4
  fi
}

for name in "${TARGETS[@]}"; do
  url="$(url_for "$name")"
  zip_path="${DOWNLOAD_DIR}/${name}.zip"
  out_path="${DOWNLOAD_DIR}/${name}"

  if [[ -f "$out_path" ]]; then
    echo "Already present: ${out_path} (skipping download)"
    continue
  fi

  echo "Downloading ${name} from ${url} ..."
  fetch "$url" "$zip_path"

  echo "Extracting ${name} ..."
  extract "$zip_path" "$name" "$out_path"

  rm -f "$zip_path"

  if [[ ! -f "$out_path" ]]; then
    echo "ERROR: expected extracted file not found: ${out_path}" >&2
    exit 5
  fi

  echo "Downloaded: ${out_path}"
done

echo "Verifying checksums ..."
"${VERIFY_SCRIPT}" "${TARGETS[@]}"

echo "Done. Downloaded data lives in ${DOWNLOAD_DIR} and is NOT committed."
