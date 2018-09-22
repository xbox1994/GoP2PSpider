#!/usr/bin/env bash
kill `ps aux | grep go | awk '{print $2}'`
cd ~/go/src/GoP2PSpider
nohup go run data/server/main.go > data.log 2>&1 &
sleep 5
nohup go run worker/server/main.go > worker.log 2>&1 &

# curl -X POST \
#  http://localhost:9200/p2p/t/_delete_by_query \
#  -H 'content-type: application/json' \
#  -d '{"query":{"match_all":{}}}'