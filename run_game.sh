#!/bin/sh

set -e
mv MyBot MyBotOld
go build -o MyBot

./halite --replay-directory replays/ -vvv --width 32 --height 32 "./MyBot" "./MyBotOld"
