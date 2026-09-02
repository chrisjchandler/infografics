#!/usr/bin/env sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
image_tag=${1:-infografics:local}

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required to build the Infografics image" >&2
  exit 1
fi
if [ ! -f "$repo_root/Dockerfile" ]; then
  echo "Dockerfile not found at $repo_root/Dockerfile" >&2
  exit 1
fi

echo "Building $image_tag from $repo_root"
exec docker build --tag "$image_tag" "$repo_root"
