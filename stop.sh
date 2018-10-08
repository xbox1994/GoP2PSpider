#!/usr/bin/env bash
kill `ps aux | grep go | awk '{print $2}'`
kill `ps aux | grep ./keeper.sh | awk '{print $2}'`