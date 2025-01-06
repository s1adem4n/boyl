#!/bin/bash

cd frontend || exit
bun install
bun run build
cd ..

build_windows() {
  GOOS=windows
  GOARCH=amd64
  CGO_ENABLED=1
  CC=x86_64-w64-mingw32-gcc
  CXX=x86_64-w64-mingw32-g++

  WEBVIEW2_INCLUDE_PATH="$(pwd)/libs/webview2/build/native/include"
  MIGW_INCLUDE_PATH="$(pwd)/libs/mingw"
  CGO_CFLAGS="-I${WEBVIEW2_INCLUDE_PATH} -I${MIGW_INCLUDE_PATH}"
  CGO_CXXFLAGS="-I${WEBVIEW2_INCLUDE_PATH} -I${MIGW_INCLUDE_PATH}"

  OUTPUT_NAME=boyl_${REF_NAME}_windows_amd64
  BINARY_PATH=build/${OUTPUT_NAME}.exe

  GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=${CGO_ENABLED} CC=${CC} CXX=${CXX} CGO_CFLAGS=${CGO_CFLAGS} CGO_CXXFLAGS=${CGO_CXXFLAGS} go build -o "${BINARY_PATH}" main.go
  zip -r build/"${OUTPUT_NAME}".zip "${BINARY_PATH}"
}

build_linux() {
  GOOS=linux
  GOARCH=amd64
  CGO_ENABLED=1
  OUTPUT_NAME=boyl_${REF_NAME}_linux_amd64
  BINARY_PATH=build/${OUTPUT_NAME}

  GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=${CGO_ENABLED} go build -o "${BINARY_PATH}" main.go
  zip -r build/"${OUTPUT_NAME}".zip "${BINARY_PATH}"
}

build_windows
build_linux
