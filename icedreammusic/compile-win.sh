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
export CPATH="$thirdparty_includes:$thirdparty_includes/winfsp:$thirdparty_includes/windows/splat/sdk/Include/10.0.22621:$thirdparty_includes/windows/splat/sdk/Include/10.0.22621/winrt:$thirdparty_includes/libnp/include:/usr/${TARGET_TRIPLET}/include"
# export CGO_CFLAGS="-I$thirdparty_includes -I$thirdparty_includes/winfsp -I/usr/${TARGET_TRIPLET}/include"
# export PATH="$GOBIN:$PATH"
export CGO_ENABLED=1

mkdir -p "$thirdparty_includes"
if [ ! -d "$thirdparty_includes/libnp" ]; then
    wget -q -O libnp.zip https://github.com/delthas/libnp/archive/291aeb5d56d5b90f89ef8a271d0803a698488ca6.zip
    7z x libnp.zip 'libnp-291aeb5d56d5b90f89ef8a271d0803a698488ca6/*' -o"$thirdparty_includes"/
    rm libnp.zip
    mv "$thirdparty_includes/libnp-291aeb5d56d5b90f89ef8a271d0803a698488ca6" "$thirdparty_includes/libnp"
    (
        cd "$thirdparty_includes/libnp/"
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
            -DCMAKE_C_FLAGS="-I$thirdparty_includes/windows/splat/sdk/Include/10.0.22621/winrt -I/usr/${TARGET_TRIPLET}/include" \
            -DCMAKE_CXX_FLAGS="-I$thirdparty_includes/windows/splat/sdk/Include/10.0.22621/winrt -I/usr/${TARGET_TRIPLET}/include" \
            -DCMAKE_FIND_ROOT_PATH_MODE_PROGRAM=NEVER \
            -DCMAKE_FIND_ROOT_PATH_MODE_LIBRARY=ONLY \
            -DCMAKE_FIND_ROOT_PATH_MODE_INCLUDE=ONLY \
            ..
        make
    )
fi
if [ ! -d "$thirdparty_includes/winfsp" ]; then
    wget -q -O winfsp.zip https://github.com/billziss-gh/winfsp/archive/release/1.2.zip
    7z e winfsp.zip 'winfsp-release-1.2/inc/fuse/*' -o"$thirdparty_includes"/winfsp
    rm winfsp.zip
fi

for bin in auto-restart-voicemeeter foobar2000 tunadish tunaposter prime4 np; do
    cd "$SCRIPT_DIR/$bin"
    go build -ldflags "-s -w" -o "$GOBIN/$bin$ext" -v
done
