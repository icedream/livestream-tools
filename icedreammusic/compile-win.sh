#!/bin/bash
set -e
set -x
set -u

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
export GOOS=windows
export GOARCH=amd64
export GOBIN="$SCRIPT_DIR"/bin/${GOOS}-${GOARCH}
ext=.exe
TARGET_TRIPLET=x86_64-w64-mingw32
export CC="${TARGET_TRIPLET}"-gcc
export CXX="${TARGET_TRIPLET}"-g++
export PKG_CONFIG_PATH="${TARGET_TRIPLET}"/lib/pkgconfig
thirdparty_includes="$SCRIPT_DIR"/include
export CPATH="$thirdparty_includes:$thirdparty_includes/winfsp:/usr/${TARGET_TRIPLET}/include"
# export CGO_CFLAGS="-I$thirdparty_includes -I$thirdparty_includes/winfsp -I/usr/${TARGET_TRIPLET}/include"
# export PATH="$GOBIN:$PATH"
export CGO_ENABLED=1

mkdir -p "$thirdparty_includes"
if [ ! -d "$thirdparty_includes/winfsp" ]
then
    wget -q -O winfsp.zip https://github.com/billziss-gh/winfsp/archive/release/1.2.zip
    7z e winfsp.zip 'winfsp-release-1.2/inc/fuse/*' -o"$thirdparty_includes"/winfsp
    rm winfsp.zip
fi

for bin in foobar2000 tunadish tunaposter prime4
do
    cd "$SCRIPT_DIR/$bin"
    go build -ldflags "-s -w" -o "$GOBIN/$bin$ext" -v
done
