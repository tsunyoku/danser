#!/bin/bash
export GOOS=darwin
export GOARCH=amd64
export CGO_ENABLED=1
export CC=gcc
export CXX=g++

exec=$1
build=$1
if [ $2 != "" ]
then
  exec+='-s'$2
  build+='-snapshot'$2
fi

go run tools/assets/assets.go ./
go build -ldflags "-s -w -X 'github.com/tsunyoku/danser/build.VERSION=$build' -X 'github.com/tsunyoku/danser/build.Stream=Release'" -o danser -v -x
go run tools/pack/pack.go danser-$exec-osx.zip danser libbass.dylib libbass_fx.dylib libbassenc.dylib libbassmix.dylib assets.dpak
rm -f danser assets.dpak