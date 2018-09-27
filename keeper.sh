#!/usr/bin/env bash
while true; do
    if [ `ps aux | grep go | wc -l` != "5" ]; then
    	./restart.sh
    fi
    sleep 10
done
