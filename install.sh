#!/bin/sh
# Install the sectors CLI from GitHub releases.
#
#   curl -fsSL https://raw.githubusercontent.com/emrsyah/sectors-cli/main/install.sh | sh
#
# Env overrides:
#   SECTORS_VERSION       version tag to install (default: latest release)
#   SECTORS_INSTALL_DIR   install directory (default: /usr/local/bin, else ~/.local/bin)
set -eu

REPO="emrsyah/sectors-cli"
BIN="sectors"

die() { echo "install: $*" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

# --- detect platform -------------------------------------------------------
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux | darwin) ;;
  *) die "unsupported OS '$os' (download a binary from https://github.com/$REPO/releases)";;
esac

arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch="amd64" ;;
  aarch64 | arm64) arch="arm64" ;;
  *) die "unsupported architecture '$arch'";;
esac

# --- resolve version -------------------------------------------------------
have curl || have wget || die "need curl or wget"
fetch() { if have curl; then curl -fsSL "$1"; else wget -qO- "$1"; fi; }

tag="${SECTORS_VERSION:-}"
if [ -z "$tag" ]; then
  tag=$(fetch "https://api.github.com/repos/$REPO/releases/latest" \
        | grep '"tag_name":' | head -1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
  [ -n "$tag" ] || die "could not determine latest release; set SECTORS_VERSION"
fi
ver="${tag#v}" # archives are named without the leading 'v'

archive="sectors-cli_${ver}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/${tag}/${archive}"

# --- choose install dir ----------------------------------------------------
dir="${SECTORS_INSTALL_DIR:-/usr/local/bin}"
if [ ! -d "$dir" ] || [ ! -w "$dir" ]; then
  if [ "${SECTORS_INSTALL_DIR:-}" = "" ]; then
    dir="$HOME/.local/bin"
    mkdir -p "$dir"
  fi
fi

# --- download & install ----------------------------------------------------
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
echo "Downloading $archive ..."
if have curl; then curl -fSL "$url" -o "$tmp/$archive"; else wget -O "$tmp/$archive" "$url"; fi
tar -xzf "$tmp/$archive" -C "$tmp"

if [ -w "$dir" ]; then
  mv "$tmp/$BIN" "$dir/$BIN"
elif have sudo; then
  echo "Installing to $dir (needs sudo) ..."
  sudo mv "$tmp/$BIN" "$dir/$BIN"
else
  die "cannot write to $dir; set SECTORS_INSTALL_DIR to a writable directory"
fi
chmod +x "$dir/$BIN" 2>/dev/null || true

echo "Installed $BIN $tag to $dir/$BIN"
case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "note: $dir is not on your PATH — add it, e.g.  export PATH=\"$dir:\$PATH\"";;
esac
"$dir/$BIN" --version 2>/dev/null || true
