#!/bin/bash

rm MyBot.zip
zip MyBot MyBot.go
dir=$PWD
pushd $GOPATH > /dev/null
zip -r ${dir}/MyBot src/github.com/BenJuan26/hlt
popd > /dev/null
