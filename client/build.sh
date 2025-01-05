#!/bin/bash

cd frontend || exit
bun install
bun run build
cd ..

build_windows() {
  export GOOS=windows
  export GOARCH=amd64
  export CGO_ENABLED=1
  export CC=x86_64-w64-mingw32-gcc
  export CXX=x86_64-w64-mingw32-g++

  WEBVIEW2_INCLUDE_PATH="$(pwd)/libs/webview2/build/native/include"
  MIGW_INCLUDE_PATH="$(pwd)/libs/mingw"
  export CGO_CFLAGS="-I${WEBVIEW2_INCLUDE_PATH} -I${MIGW_INCLUDE_PATH}"
  export CGO_CXXFLAGS="-I${WEBVIEW2_INCLUDE_PATH} -I${MIGW_INCLUDE_PATH}"

  go build -o build/boyl_windows_amd64.exe main.go
  zip -r build/boyl_"${REF_NAME}"_windows_amd64.zip build/boyl_windows_amd64.exe
}

build_linux() {
  export GOOS=linux
  export GOARCH=amd64
  export CGO_ENABLED=1
  export CC=
  export CXX=
  export CGO_CFLAGS=
  export CGO_CXXFLAGS=

  go build -o build/boyl_linux_amd64 main.go
  zip -r build/boyl_"${REF_NAME}"_linux_amd64.zip build/boyl_linux_amd64
}

build_windows
build_linux
