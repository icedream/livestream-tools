#!/bin/bash
set -e
set -x
set -u

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
export GOOS=windows
export GOARCH=amd64
export GOBIN="$SCRIPT_DIR"/bin/${GOOS}-${GOARCH}
ext=.exe
TARGET_TRIPLET=x86_64-w64-mingw32
export CC="${TARGET_TRIPLET}"-gcc
export CXX="${TARGET_TRIPLET}"-g++
export AR="${TARGET_TRIPLET}"-ar
export LD="${TARGET_TRIPLET}"-ld
export OBJCOPY="${TARGET_TRIPLET}"-objcopy
export RANLIB="${TARGET_TRIPLET}"-ranlib
export SIZE="${TARGET_TRIPLET}"-size
export STRIP="${TARGET_TRIPLET}"-strip
export PKG_CONFIG_PATH="${TARGET_TRIPLET}"/lib/pkgconfig
thirdparty_includes="$SCRIPT_DIR"/include
winsdk_version=10.0.26100
export CPATH="$thirdparty_includes:$thirdparty_includes/winfsp:$thirdparty_includes/windows/splat/sdk/Include/${winsdk_version}:$thirdparty_includes/windows/splat/sdk/Include/${winsdk_version}/winrt:$thirdparty_includes/libnp/include:/usr/${TARGET_TRIPLET}/include"
# export CGO_CFLAGS="-I$thirdparty_includes -I$thirdparty_includes/winfsp -I/usr/${TARGET_TRIPLET}/include"
# export PATH="$GOBIN:$PATH"
export CGO_ENABLED=1

if [ ! -d "$thirdparty_includes/windows/splat/sdk/Include/${winsdk_version}" ]; then
    if command -v xwin 2>/dev/null >/dev/null; then
        xwin --sdk-version="$winsdk_version" --cache-dir="$thirdparty_includes/windows/.xwin-cache" splat --preserve-ms-arch-notation --output="$thirdparty_includes/windows/splat"
    else
        echo "ERROR: xwin not found and $thirdparty_includes/windows/splat does not exist. Please install the xwin tool from https://github.com/Jake-Shadle/xwin." >&2
        exit 1
    fi

    # patch pragma error
    srcdir="$(pwd)"
    (cd "$thirdparty_includes/windows/splat/sdk/Include/${winsdk_version}/winrt" && patch -u -p1 -N -i "$srcdir/winrt.1.patch")
fi

mkdir -p "$thirdparty_includes"
if [ ! -d "$thirdparty_includes/libnp" ]; then
    wget -q -O libnp.zip https://github.com/delthas/libnp/archive/291aeb5d56d5b90f89ef8a271d0803a698488ca6.zip
    wget -q -O libnp.1.patch https://github.com/delthas/libnp/pull/1.patch
    srcdir="$(pwd)"
    7z x libnp.zip 'libnp-291aeb5d56d5b90f89ef8a271d0803a698488ca6/*' -o"$thirdparty_includes"/
    rm libnp.zip
    mv "$thirdparty_includes/libnp-291aeb5d56d5b90f89ef8a271d0803a698488ca6" "$thirdparty_includes/libnp"
    (
        cd "$thirdparty_includes/libnp/"
        patch -p1 -N -i "$srcdir/libnp.1.patch"
        mkdir -p build
        cd build
        cmake \
            -DCMAKE_SYSTEM_NAME=Windows \
            -DCMAKE_SYSTEM_PROCESSOR= \
            -DCMAKE_AR="$AR" \
            -DCMAKE_ASM_COMPILER="$CC" \
            -DCMAKE_C_COMPILER="$CC" \
            -DCMAKE_CXX_COMPILER="$CXX" \
            -DCMAKE_LINKER="$LD" \
            -DCMAKE_OBJCOPY="$OBJCOPY" \
            -DCMAKE_RANLIB="$RANLIB" \
            -DCMAKE_SIZE="$SIZE" \
            -DCMAKE_STRIP="$STRIP" \
            -DCMAKE_C_FLAGS="-I$thirdparty_includes/windows/splat/sdk/Include/${winsdk_version}/winrt -I/usr/${TARGET_TRIPLET}/include" \
            -DCMAKE_CXX_FLAGS="-I$thirdparty_includes/windows/splat/sdk/Include/${winsdk_version}/winrt -I/usr/${TARGET_TRIPLET}/include" \
            -DCMAKE_FIND_ROOT_PATH_MODE_PROGRAM=NEVER \
            -DCMAKE_FIND_ROOT_PATH_MODE_LIBRARY=ONLY \
            -DCMAKE_FIND_ROOT_PATH_MODE_INCLUDE=ONLY \
            ..
        make
    )
    mkdir -p "$GOBIN"
    cp "$thirdparty_includes/libnp/build"/*.dll "$GOBIN"
fi
if [ ! -d "$thirdparty_includes/winfsp" ]; then
    wget -q -O winfsp.zip https://github.com/winfsp/winfsp/archive/refs/tags/v2.1.zip
    7z e winfsp.zip 'winfsp-2.1/inc/fuse/*' -o"$thirdparty_includes"/winfsp
    rm winfsp.zip
fi

for bin in auto-restart-voicemeeter foobar2000 tunadish tunaposter prime4 np; do
    cd "$SCRIPT_DIR/$bin"
    go build -ldflags "-s -w" -o "$GOBIN/$bin$ext" -v
done
