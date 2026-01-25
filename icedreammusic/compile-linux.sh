#!/bin/bash
set -e
set -x
set -u

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
export GOOS=linux
export GOARCH=${GOARCH:-amd64}
export GOBIN="$SCRIPT_DIR"/bin/${GOOS}-${GOARCH}
ext=
TARGET_TRIPLET=x86_64-linux-gnu
export CC="${TARGET_TRIPLET}"-gcc
export CXX="${TARGET_TRIPLET}"-g++
export AR="${TARGET_TRIPLET}"-ar
export LD="${TARGET_TRIPLET}"-ld
export OBJCOPY="${TARGET_TRIPLET}"-objcopy
export RANLIB="${TARGET_TRIPLET}"-ranlib
export SIZE="${TARGET_TRIPLET}"-size
export STRIP="${TARGET_TRIPLET}"-strip
export PKG_CONFIG_PATH="${TARGET_TRIPLET}"/lib/pkgconfig
# thirdparty_includes="$SCRIPT_DIR"/include
# export CPATH="$thirdparty_includes:$thirdparty_includes/winfsp:/usr/${TARGET_TRIPLET}/include"
# # export CGO_CFLAGS="-I$thirdparty_includes -I$thirdparty_includes/winfsp -I/usr/${TARGET_TRIPLET}/include"
# # export PATH="$GOBIN:$PATH"
# export CGO_ENABLED=1

# mkdir -p "$thirdparty_includes"
# if [ ! -d "$thirdparty_includes/winfsp" ]; then
#     wget -q -O winfsp.zip https://github.com/winfsp/winfsp/archive/refs/tags/v2.1.zip
#     7z e winfsp.zip 'winfsp-2.1/inc/fuse/*' -o"$thirdparty_includes"/winfsp
#     rm winfsp.zip
# fi

for bin in foobar2000 tunadish tunaposter prime4 np; do
    cd "$SCRIPT_DIR/$bin"
    go build -ldflags "-s -w" -o "$GOBIN/$bin$ext" -v
done
