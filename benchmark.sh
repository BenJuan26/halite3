#!/bin/bash

if [ "$1" = "" ]; then
    echo "Must specify an opponent bot path"
    exit 1
fi

if [ ! -f "./$1" ]; then
    echo "No such file $1"
    exit 1
fi

set -e
go build -o MyBot

count=0
victories=0
halite=0
while [ $count -lt 20 ]; do
    json_results=$(./halite --replay-directory replays/ --width 32 --height 32 --results-as-json "./MyBot" "./$1")
    replay=$(echo $json_results | jq -r .replay)
    rm $replay
    rank=$(echo $json_results | jq -r '.stats."0".rank')
    score=$(echo $json_results | jq -r '.stats."0".score')
    halite=$((halite + score))
    if [ $rank -eq 1 ]; then
        echo "Game $count: Win"
        ((victories++))
    else
        echo "Game $count: Loss"
    fi
    ((count++))
done

percent=$(echo "scale=1; ($victories/$count)*100" | bc)
avg_halite=$(echo "$halite/$count" | bc)

echo "-----------------------------"
echo "Victories: $victories"
echo "Total games: $count"
echo "Avg Halite produced: $avg_halite"
echo "Win percent: $percent%"
