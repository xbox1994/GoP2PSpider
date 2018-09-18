#!/usr/bin/env bash
kill `ps aux | grep go | awk '{print $2}'`