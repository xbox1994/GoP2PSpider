#!/usr/bin/env bash
kill `ps aux | grep go | awk '{print $2}'`
cd ~/go/src/GoP2PSpider
nohup go run data/server/main.go > data.log 2>&1 &
nohup go run worker/server/main.go -wc 10 > worker.log 2>&1 &