#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
go_root=/tmp/crayons-go
node_root=/tmp/crayons-node
go_archive=/tmp/go1.24.6.linux-amd64.tar.gz
node_archive=/tmp/node-v22.17.1-linux-x64.tar.xz

download() {
  local url=$1
  local out=$2
  python3 - "$url" "$out" <<'PY'
import os
import sys
import urllib.request

url, out = sys.argv[1], sys.argv[2]
os.makedirs(os.path.dirname(out), exist_ok=True)
urllib.request.urlretrieve(url, out)
print(out)
PY
}

if [[ ! -x "$go_root/go/bin/go" ]]; then
  [[ -f "$go_archive" ]] || download "https://go.dev/dl/go1.24.6.linux-amd64.tar.gz" "$go_archive"
  rm -rf "$go_root"
  mkdir -p "$go_root"
  tar -xzf "$go_archive" -C "$go_root"
fi

if [[ ! -x "$node_root/node-v22.17.1-linux-x64/bin/node" ]]; then
  [[ -f "$node_archive" ]] || download "https://nodejs.org/dist/v22.17.1/node-v22.17.1-linux-x64.tar.xz" "$node_archive"
  rm -rf "$node_root"
  mkdir -p "$node_root"
  python3 - "$node_archive" "$node_root" <<'PY'
import lzma
import sys
import tarfile
from pathlib import Path

archive = Path(sys.argv[1])
out_dir = Path(sys.argv[2])
with lzma.open(archive, "rb") as compressed:
    with tarfile.open(fileobj=compressed, mode="r:") as tar:
        tar.extractall(out_dir)
PY
fi

printf 'Toolchain ready. Use PATH="%s/.tools/bin:$PATH"\n' "$repo_root"
