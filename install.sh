#!/bin/sh
# Install the latest ccr release binary.
#   curl -fsSL https://raw.githubusercontent.com/Aky97567/ccr/main/install.sh | sh
set -eu

REPO="Aky97567/ccr"
BINDIR="${CCR_BINDIR:-$HOME/.local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac

tag=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
  | grep -o '"tag_name": *"[^"]*"' | head -1 | cut -d'"' -f4)
[ -n "$tag" ] || { echo "could not resolve the latest release of $REPO" >&2; exit 1; }

url="https://github.com/$REPO/releases/download/$tag/ccr_${tag#v}_${os}_${arch}.tar.gz"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "downloading $url"
curl -fsSL "$url" | tar -xz -C "$tmp"

mkdir -p "$BINDIR" "$HOME/.local/share/ccr"
install -m 0755 "$tmp/ccr" "$BINDIR/ccr"
"$BINDIR/ccr" install-skill || true

echo "installed ccr $tag to $BINDIR/ccr"
case ":$PATH:" in
  *":$BINDIR:"*) ;;
  *) echo "note: $BINDIR is not on your PATH" >&2 ;;
esac
