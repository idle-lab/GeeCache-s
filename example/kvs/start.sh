#!/bin/bash
trap "rm kvs;kill 0" EXIT

go build -o kvs -ldflags -w
./kvs -port=8001 &
./kvs -port=8002 &
./kvs -port=8003 -api=1 &

wait