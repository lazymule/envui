#!/usr/bin/env bash
# Builds every Windows download into dist/:
#
#   envui-windows-amd64.exe   plain executable
#   envui-setup.exe           installer (Start menu, uninstall, optional Mesa)
#   envui-portable.zip        portable edition (settings next to the exe)
#   envui-portable-vm.zip     portable edition + Mesa software OpenGL for
#                             virtual machines / Remote Desktop
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

readme() { # $1 = with Mesa (yes/no)
  local tr="" en="" lic=""
  if [ "$1" = yes ]; then
    tr="SANAL MAKİNE / UZAK MASAÜSTÜ\n----------------------------\nBu pakette Mesa'nın yazılımsal OpenGL sürücüsü (opengl32.dll,\nlibgallium_wgl.dll) vardır. 'WGL: The driver does not appear to support\nOpenGL' hatası veren bilgisayarlarda çizimi işlemci yapar. Dosyaları\nenvui.exe ile aynı klasörde tutun. Diğer programlar etkilenmez.\n"
    en="\nThis package also contains Mesa's software OpenGL renderer for virtual\nmachines and Remote Desktop (fixes 'WGL: The driver does not appear to\nsupport OpenGL'). Keep the DLLs next to envui.exe."
    lic="Mesa $MESA_VERSION (opengl32.dll, libgallium_wgl.dll): MIT Lisansı - https://www.mesa3d.org/"
  fi
  sed -e "s|@VM_NOTE_TR@|$tr|" -e "s|@VM_NOTE_EN@|$en|" -e "s|@MESA_LICENSE@|$lic|" "$ROOT/tools/portable-readme.txt" \
    | sed 's/\\n/\n/g'
}

portable() { # $1 = name, $2 = with Mesa
  local dir="$WORK/$1"
  mkdir -p "$dir"
  cp "$WORK/envui.exe" "$dir/"
  echo '{}' > "$dir/envui-portable.json"
  readme "$2" > "$dir/OKUBENI.txt"
  if [ "$2" = yes ]; then cp "$WORK/mesa/"*.dll "$dir/"; fi
  rm -f "$DIST/$1.zip"
  (cd "$WORK" && zip -9 -r -q "$DIST/$1.zip" "$1")
  echo "   $DIST/$1.zip"
}

echo "== portable"
portable envui-portable no
portable envui-portable-vm yes

echo "== installer"
mkdir -p "$WORK/inst"
cp "$WORK/envui.exe" "$WORK/mesa/"*.dll "$WORK/inst/"
(cd "$ROOT/installer" && makensis -V2 -DVERSION="$VERSION" -DSRCDIR="$WORK/inst" -DOUTFILE="$DIST/envui-setup.exe" envui.nsi)
echo "   $DIST/envui-setup.exe"
