#!/bin/sh

set -e
go build -o MyBot

./halite --replay-directory replays/ -vvv --width 32 --height 32 "./MyBot" "./MyBot"
