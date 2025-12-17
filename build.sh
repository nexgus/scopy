#!/bin/bash
TARGET=scopy
COMMIT=$(git describe --match=NeVeRmAtCh --always --abbrev=8 --dirty)
GOVER=$(go version | cut -d ' ' -f 3)
VER=$(grep 'VersionString' cmd/$TARGET/version.go | sed -E 's/.*"([^"]+)".*/\1/')

GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o bin/$TARGET-$VER-darwin-arm64 \
    -ldflags '-w -extldflags "-static" -X "main.GitCommitHash='"${COMMIT}"'" -X "main.GoVersion='"${GOVER}"'" ' \
    $TARGET/cmd/$TARGET
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/$TARGET-$VER-linux-amd64 \
    -ldflags '-w -extldflags "-static" -X "main.GitCommitHash='"${COMMIT}"'" -X "main.GoVersion='"${GOVER}"'" ' \
    $TARGET/cmd/$TARGET
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/$TARGET-$VER-windows-amd64.exe \
    -ldflags '-w -extldflags "-static" -X "main.GitCommitHash='"${COMMIT}"'"' \
    $TARGET/cmd/$TARGET

myos=$(uname | awk '{print tolower($0)}')
myarch=$(uname -m)
case "$myarch" in
    x86_64)
        myarch="amd64";;
    arm64|aarch64)
        myarch="arm64";;
    *)
        myarch="unknown";;
esac

ln -fs $TARGET-$VER-$myos-$myarch bin/$TARGET
