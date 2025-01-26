#!/bin/bash
trap "rm geecahces;kill 0" EXIT

go build -o geecahces
./geecahces -port=8001 &
./geecahces -port=8002 &
./geecahces -port=8003 -api=1 &

sleep 2
echo ">>> start test"
./geecahces -type=client

wait