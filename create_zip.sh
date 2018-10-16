#!/bin/bash

set -e
rm MyBot.zip
zip MyBot MyBot.go
dir=$PWD
pushd $GOPATH > /dev/null
zip -r ${dir}/MyBot src/hlt
popd > /dev/null
