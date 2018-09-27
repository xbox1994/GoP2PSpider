#!/usr/bin/env bash
while true; do
    if [ `ps aux | grep go | wc -l` != "5" ]; then
    	./restart.sh
    	echo restart success
    fi
    echo sleep 10
    sleep 10
done
