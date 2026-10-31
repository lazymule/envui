#!/usr/bin/env bash
# Builds dist/envui-windows-vm.zip: envui.exe bundled with Mesa's software
# OpenGL renderer (llvmpipe) for Windows machines without an OpenGL driver
# (virtual machines, Remote Desktop sessions, servers).
#
# Needs: go, x86_64-w64-mingw32-gcc, curl, 7z, zip
# Usage: tools/package_vm.sh [mesa-version]   (default: 26.2.5)
set -euo pipefail

MESA_VERSION="${1:-26.2.5}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WORK="$(mktemp -d)"
OUT="$ROOT/dist/envui-windows-vm"
trap 'rm -rf "$WORK"' EXIT

echo "Building envui.exe..."
(cd "$ROOT" && CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags "-H=windowsgui -s -w" -o "$WORK/envui.exe" .)

echo "Downloading Mesa $MESA_VERSION..."
curl -fL -o "$WORK/mesa.7z" \
  "https://github.com/pal1000/mesa-dist-win/releases/download/$MESA_VERSION/mesa3d-$MESA_VERSION-release-msvc.7z"
7z e -y -o"$WORK" "$WORK/mesa.7z" x64/opengl32.dll x64/libgallium_wgl.dll >/dev/null

rm -rf "$OUT" "$OUT.zip"
mkdir -p "$OUT"
cp "$WORK/envui.exe" "$WORK/opengl32.dll" "$WORK/libgallium_wgl.dll" "$OUT/"
sed "s/@MESA_VERSION@/$MESA_VERSION/g" "$ROOT/tools/vm-readme.txt" > "$OUT/OKUBENI.txt"
(cd "$ROOT/dist" && zip -9 -r -q "envui-windows-vm.zip" "envui-windows-vm")
echo "Created $OUT.zip"
