#!/usr/bin/env bash
# Builds every Windows download into dist/:
#
#   envui-windows-amd64.exe   plain executable (no software-OpenGL fallback)
#   envui-setup.exe           installer (Start menu, uninstall)
#   envui-portable.zip        portable edition (settings next to the exe)
#
# The installer and the portable zip include a "softgl" folder with a copy of
# envui.exe and Mesa's software OpenGL renderer. envui.exe switches to it by
# itself on machines without a usable OpenGL driver (virtual machines,
# Remote Desktop), so one package works everywhere.
#
# Needs: go, x86_64-w64-mingw32-gcc, curl, 7z, zip, makensis
# Usage: tools/package_windows.sh [version] [mesa-version]
set -euo pipefail

VERSION="${1:-$(sed -n 's/^var Version = "\(.*\)"/\1/p' "$(dirname "$0")/../ui/about.go")}"
VERSION="${VERSION#v}"
MESA_VERSION="${2:-26.2.5}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST="$ROOT/dist"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
mkdir -p "$DIST"

echo "== envui $VERSION (Mesa $MESA_VERSION)"
(cd "$ROOT" && CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags "-H=windowsgui -s -w -X github.com/mrf/envui/ui.Version=$VERSION" -o "$WORK/envui.exe" .)
cp "$WORK/envui.exe" "$DIST/envui-windows-amd64.exe"

echo "== Mesa"
curl -fsSL -o "$WORK/mesa.7z" \
  "https://github.com/pal1000/mesa-dist-win/releases/download/$MESA_VERSION/mesa3d-$MESA_VERSION-release-msvc.7z"
7z e -y -o"$WORK/mesa" "$WORK/mesa.7z" x64/opengl32.dll x64/libgallium_wgl.dll >/dev/null

echo "== portable"
P="$WORK/envui-portable"
mkdir -p "$P/softgl"
cp "$WORK/envui.exe" "$P/"
cp "$WORK/envui.exe" "$WORK/mesa/"*.dll "$P/softgl/"
echo '{}' > "$P/envui-portable.json"
sed "s|@MESA_VERSION@|$MESA_VERSION|g" "$ROOT/tools/portable-readme.txt" > "$P/OKUBENI.txt"
rm -f "$DIST/envui-portable.zip" "$DIST/envui-portable-vm.zip"
(cd "$WORK" && zip -9 -r -q "$DIST/envui-portable.zip" envui-portable)
echo "   $DIST/envui-portable.zip"

echo "== installer"
mkdir -p "$WORK/inst/softgl"
cp "$WORK/envui.exe" "$WORK/inst/"
cp "$WORK/envui.exe" "$WORK/mesa/"*.dll "$WORK/inst/softgl/"
(cd "$ROOT/installer" && makensis -V2 -DVERSION="$VERSION" -DSRCDIR="$WORK/inst" -DOUTFILE="$DIST/envui-setup.exe" envui.nsi)
echo "   $DIST/envui-setup.exe"
